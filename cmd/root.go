package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/k0sproject/dig"
	_ "github.com/k0sproject/k0sctl/pkg/apis/k0sctl.k0sproject.io/v1beta1/cluster"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"

	"boundless-cli/pkg/config"
	"boundless-cli/version"
)

const boundlessManifestUrl = "https://raw.githubusercontent.com/mirantis/boundless/main/deploy/static/boundless-operator.yaml"

var DefaultComponents = config.Components{
	Core: config.Core{
		Ingress: &config.CoreComponent{
			Enabled:  false,
			Provider: "ingress-nginx",
			Config: dig.Mapping{
				"controller": dig.Mapping{
					"service": dig.Mapping{
						"type": "NodePort",
						"nodePorts": dig.Mapping{
							"http":  30000,
							"https": 30001,
						},
					},
				},
			},
		},
	},
	Addons: []config.Addons{
		{
			Name:      "example-server",
			Kind:      "MKEAddon",
			Enabled:   true,
			Namespace: "default",
			Chart: config.Chart{
				Name:    "nginx",
				Repo:    "https://charts.bitnami.com/bitnami",
				Version: "15.1.1",
				Values: `"service":
  "type": "ClusterIP"
`,
			},
		},
	},
}

// App is the main urfave/cli.App for boctl
var App = &cli.App{
	Name:  "bocli",
	Usage: "boundless operator management tool",
	Flags: []cli.Flag{
		debugFlag,
		traceFlag,
		redactFlag,
	},
	Before: actions(initLogging),
	Commands: []*cli.Command{
		{
			Name:  "version",
			Usage: "Output bocli version",
			Action: func(ctx *cli.Context) error {
				fmt.Printf("version: %s\n", version.Version)
				return nil
			},
		},
		{
			Name:  "init",
			Usage: "create a cluster",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "version",
					Usage: "Include a skeleton k0s config section",
				},
				&cli.BoolFlag{
					Name:  "k0s",
					Usage: "Include a skeleton k0s config section",
				},
				&cli.StringFlag{
					Name:    "cluster-name",
					Usage:   "Cluster name",
					Aliases: []string{"n"},
					Value:   "k0s-cluster",
				},
				&cli.IntFlag{
					Name:    "controller-count",
					Usage:   "The number of controllers to create when addresses are given",
					Aliases: []string{"C"},
					Value:   1,
				},
				&cli.StringFlag{
					Name:    "user",
					Usage:   "Host user when addresses given",
					Aliases: []string{"u"},
				},
				&cli.StringFlag{
					Name:    "key-path",
					Usage:   "Host key path when addresses given",
					Aliases: []string{"i"},
				},
			},
			Before: actions(initLogging),
			Action: initWrapper,
		},
		{
			Name:  "apply",
			Usage: "apply a cluster",
			Flags: []cli.Flag{
				configFlag,
				debugFlag,
				traceFlag,
				redactFlag,
			},
			Before: actions(initLogging),
			Action: applyWrapper,
		},
		{
			Name:  "update",
			Usage: "update a cluster",
			Flags: []cli.Flag{
				configFlag,
				debugFlag,
			},
			Before: actions(initLogging),
			Action: func(c *cli.Context) error {
				// read the cluster config
				cfg, err := initMKEConfig(c)
				if err != nil {
					return err
				}

				log.Info("Updating Components")
				err = installComponents(cfg.Spec.Mke.Components)
				if err != nil {
					return fmt.Errorf("failed to update components: %w", err)
				}
				return nil
			},
		},
		{
			Name:  "reset",
			Usage: "reset a cluster",
			Flags: []cli.Flag{
				configFlag,
				debugFlag,
				forceFlag,
			},
			Before: actions(initLogging),
			Action: func(c *cli.Context) error {
				return cmdWrapper(c)
			},
		},
	},
	Action: func(c *cli.Context) error {
		return cmdWrapper(c)
	},
}

func cmdWrapper(c *cli.Context) error {
	k0sctlConfigFile, err := getK0sctlConfigPath(c)
	if err != nil {
		return err
	}
	args := append(argInsert(c.Command.Name, c.Args().Slice()), "--config", k0sctlConfigFile)
	cmd := exec.Command("k0sctl", args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func initWrapper(c *cli.Context) error {
	cmd := exec.Command("k0sctl", argInsert("init", c.Args().Slice())...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	buf := new(bytes.Buffer)
	cmd.Stdout = buf
	err := cmd.Run()
	if err != nil {
		return err
	}

	cfg := buf.Bytes()
	k0sConfig, err := config.ParseK0sCluster(cfg)
	if err != nil {
		return err
	}

	mkeConfig := config.ConvertToMke(k0sConfig, DefaultComponents)
	encoder := yaml.NewEncoder(os.Stdout)
	return encoder.Encode(&mkeConfig)
}

func applyWrapper(c *cli.Context) error {
	k0sctlConfigFile, err := getK0sctlConfigPath(c)

	// install k0s
	log.Infof("Installing Kubernetes distribution: %s", "k0s")
	if err = installK0s(k0sctlConfigFile); err != nil {
		return err
	}

	log.Infof("Waiting for nodes to be ready")
	if err := WaitForNodes(); err != nil {
		return fmt.Errorf("failed to wait for nodes: %w", err)
	}

	// install toolkit
	log.Info("Installing MKE 4")

	//log.Debugf("Installing Helm Controller")
	//err = kubectlApply("manifests/helm-controller.yaml")
	//if err != nil {
	//	return fmt.Errorf("failed to install Helm Controller: %w", err)
	//}

	log.Infof("Installing MKE Operator")
	err = kubectlApply(boundlessManifestUrl)
	if err != nil {
		return fmt.Errorf("failed to install MKE Operator: %w", err)
	}

	log.Infof("Waiting for all pods to be ready")
	if err := WaitForPods(); err != nil {
		return fmt.Errorf("failed to wait for pods: %w", err)
	}

	// read the cluster config
	config, err := initMKEConfig(c)
	if err != nil {
		return err
	}

	// install components
	log.Info("Applying MKE Operator resource")
	err = installComponents(config.Spec.Mke.Components)
	if err != nil {
		return fmt.Errorf("failed to install components: %w", err)
	}

	log.Info("Finished installing MKE")

	return nil
}

func argInsert(arg string, args []string) []string {
	return append([]string{arg}, args...)
}
