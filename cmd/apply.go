package cmd

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var cmdApply = &cli.Command{
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
}

func applyWrapper(c *cli.Context) error {
	// read the blueprint config
	config, err := initBlueprint(c)
	if err != nil {
		return err
	}

	if config.Spec.Kubernetes != nil {
		log.Infof("Installing Kubernetes distribution: %s", config.Spec.Kubernetes.Provider)
		switch config.Spec.Kubernetes.Provider {
		case "k0s":
			k0sctlConfigPath, err := getK0sctlConfigPath(c)
			if err = installK0s(k0sctlConfigPath); err != nil {
				return err
			}
		case "kind":
			log.Infof("Installing Kubernetes distribution: %s", "kind")
			if err = installKindCluster(config.Metadata.Name, KubeConfigFile); err != nil {
				return err
			}
		default:
			return fmt.Errorf("invalid Kubernetes distribution provider: %s", config.Spec.Kubernetes.Provider)
		}
	}

	log.Infof("Waiting for nodes to be ready")
	if err := WaitForNodes(); err != nil {
		return fmt.Errorf("failed to wait for nodes: %w", err)
	}

	log.Info("Installing Boundless Operator")
	log.Debugf("Installing Boundless Operator using manifest file: %s", BoundlessManifestUrl)
	err = kubectlApply(BoundlessManifestUrl)
	if err != nil {
		return fmt.Errorf("failed to install Boundless Operator: %w", err)
	}

	log.Infof("Waiting for all pods to be ready")
	if err := WaitForPods(); err != nil {
		return fmt.Errorf("failed to wait for pods: %w", err)
	}

	// install components
	log.Info("Applying Boundless Operator resource")
	err = installComponents(config)
	if err != nil {
		return fmt.Errorf("failed to install components: %w", err)
	}

	log.Info("Finished installing Boundless Operator")

	return nil
}
