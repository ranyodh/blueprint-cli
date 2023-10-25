package kube

import (
	"fmt"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func createKubeClient() (kubernetes.Interface, error) {
	config, err := getKubeConfig()
	if err != nil {
		return nil, err
	}

	forConfig, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return forConfig, nil
}

func getKubeConfig() (*rest.Config, error) {
	kubecconfig, err := filepath.Abs("kubeconfig")
	if err != nil {
		return nil, fmt.Errorf("failed to get path for kubeconfig: %v", err)
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubecconfig)
	if err != nil {
		return nil, err
	}
	return config, nil
}
