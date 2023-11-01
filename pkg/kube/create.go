package kube

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/mirantis/boundless-operator/api/v1alpha1"
)

func CreateOrUpdate(obj client.Object) error {
	logf.SetLogger(zap.New())

	scheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(scheme)

	config, err := getKubeConfig()
	if err != nil {
		return err
	}

	//log.Debug("Creating kubernetes client")
	kubeClient, err := client.New(config, client.Options{Scheme: scheme})
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %v", err)
	}

	existing := &v1alpha1.Blueprint{}
	err = kubeClient.Get(context.Background(), client.ObjectKeyFromObject(obj), existing)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("failed to get cluster object: %v", err)
		}
	}
	if existing.Name != "" {
		obj.SetResourceVersion(existing.GetResourceVersion())
		err = kubeClient.Update(context.Background(), obj)
		if err != nil {
			return fmt.Errorf("failed to update cluster object: %v", err)
		}
	} else {
		if err := kubeClient.Create(context.Background(), obj); err != nil {
			return fmt.Errorf("failed to create cluster object: %v", err)
		}
	}

	return nil
}
