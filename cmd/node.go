package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func WaitForNodes() error {
	ctx := context.Background()
	clientset, err := createKubeClient()
	if err != nil {
		return err
	}
	return waitForNodes(ctx, clientset)
}

func WaitForPods() error {
	ctx := context.Background()
	clientset, err := createKubeClient()
	if err != nil {
		return err
	}
	return waitForPods(ctx, clientset)
}

func waitForPods(ctx context.Context, clientset kubernetes.Interface) error {
	// wait for all pods
	timeoutCtx, cancelFunc := context.WithTimeout(ctx, 5*time.Minute)
	defer cancelFunc()

	return wait.PollUntilContextCancel(timeoutCtx, 5*time.Second, true, func(ctx context.Context) (bool, error) {
		pods, err := clientset.CoreV1().Pods(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
		if err != nil {
			return false, fmt.Errorf("failed to list pods: %v", err)
		}

		if len(pods.Items) == 0 {
			return false, nil
		}

		allRunning := true
		for _, pod := range pods.Items {
			if pod.Status.Phase != v1.PodRunning {
				allRunning = false
				break
			}
		}
		return allRunning, nil
	})
}

func waitForNodes(ctx context.Context, clientset kubernetes.Interface) error {
	timeoutCtx, cancelFunc := context.WithTimeout(ctx, 5*time.Minute)
	defer cancelFunc()

	return wait.PollUntilContextCancel(timeoutCtx, 5*time.Second, true, func(ctx context.Context) (bool, error) {
		// wait for node to be ready
		nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		if err != nil {
			return false, fmt.Errorf("failed to list nodes: %v", err)
		}

		if len(nodes.Items) == 0 {
			return false, nil
		}

		allReady := true
		for _, node := range nodes.Items {
			ready := false
			for _, condition := range node.Status.Conditions {
				if condition.Type == v1.NodeReady && condition.Status == v1.ConditionTrue {
					ready = true
					break
				}
			}

			allReady = allReady && ready
		}
		return allReady, nil
	})
}

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
