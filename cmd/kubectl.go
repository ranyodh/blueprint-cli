package cmd

import (
	"os"
	"os/exec"
)

// call kubectl apply

func kubectlApply(path string) error {
	cmd := exec.Command("kubectl", "apply", "-f", path, "--kubeconfig", KubeConfigFile)
	//cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func kubectlCreate(path string) error {
	cmd := exec.Command("kubectl", "create", "-f", path, "--kubeconfig", KubeConfigFile)
	//cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
