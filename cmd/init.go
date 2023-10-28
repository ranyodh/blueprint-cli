package cmd

import (
	"bytes"
	"os"
	"os/exec"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"

	"boundless-cli/pkg/config"
)

var cmdInit = &cli.Command{
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
		&cli.BoolFlag{
			Name:  "kind",
			Usage: "Create a kind cluster",
		},
	},
	Before: actions(initLogging),
	Action: initWrapper,
}

func initWrapper(c *cli.Context) error {
	isKind := c.Bool("kind")
	if isKind {
		return encode(config.ConvertToClusterWithKind("kind-cluster", DefaultComponents))
	}

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

	return encode(config.ConvertToClusterWithK0s(k0sConfig, DefaultComponents))
}

func encode(mkeConfig config.Cluster) error {
	encoder := yaml.NewEncoder(os.Stdout)
	return encoder.Encode(&mkeConfig)
}
