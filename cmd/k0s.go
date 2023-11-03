package cmd

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

func installK0s(config string) error {
	log.Debug("installing k0s with config: ", config)
	if err := cmdRun("k0sctl", "apply", "--config", config, "--no-wait"); err != nil {
		return fmt.Errorf("failed to install k0s: %w", err)
	}

	// create kubeconfig
	if err := createKubeConfig(config); err != nil {
		return fmt.Errorf("failed to create kubeconfig: %w", err)
	}
	log.Debugf("kubeconfig file for k0s cluster: %s", KubeConfigFile)

	return nil
}
