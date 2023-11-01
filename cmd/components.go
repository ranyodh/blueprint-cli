package cmd

import (
	"bytes"
	"fmt"

	"github.com/k0sproject/dig"
	log "github.com/sirupsen/logrus"
	yamlDecoder "gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/mirantis/boundless-operator/api/v1alpha1"

	"boundless-cli/pkg/kube"

	"boundless-cli/pkg/config"
)

var DefaultComponents = config.Components{
	Core: config.Core{
		Ingress: &config.CoreComponent{
			Enabled:  true,
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
			Kind:      "HelmAddon",
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

func installComponents(cluster config.Blueprint) error {
	components := cluster.Spec.Components
	ingressConfig, err := yamlValues(components.Core.Ingress.Config)
	if err != nil {
		return fmt.Errorf("failed to convert ingress config to yaml: %w", err)
	}

	var addons []v1alpha1.AddonSpec
	for _, addon := range components.Addons {
		addons = append(addons, v1alpha1.AddonSpec{
			Name:      addon.Name,
			Kind:      addon.Kind,
			Enabled:   addon.Enabled,
			Namespace: addon.Namespace,
			Chart: v1alpha1.Chart{
				Name:    addon.Chart.Name,
				Repo:    addon.Chart.Repo,
				Version: addon.Chart.Version,
				Set:     addon.Chart.Set,
				Values:  addon.Chart.Values,
			},
		})
	}

	c := v1alpha1.Blueprint{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Metadata.Name,
			Namespace: v1.NamespaceDefault,
		},
		Spec: v1alpha1.BlueprintSpec{
			Components: v1alpha1.Component{
				Core: v1alpha1.Core{
					Ingress: v1alpha1.IngressSpec{
						Enabled:  components.Core.Ingress.Enabled,
						Provider: components.Core.Ingress.Provider,
						Config:   ingressConfig,
					},
				},
				Addons: addons,
			},
		},
	}

	log.Info("Applying Blueprint")
	if err := kube.CreateOrUpdate(&c); err != nil {
		return fmt.Errorf("failed to create/update Blueprint object: %v", err)
	}

	return nil
}

func yamlValues(values dig.Mapping) (string, error) {
	valuesYaml := new(bytes.Buffer)

	encoder := yamlDecoder.NewEncoder(valuesYaml)
	err := encoder.Encode(&values)
	if err != nil {
		return "", err
	}
	return valuesYaml.String(), nil
}

func jsonValues(values dig.Mapping) (string, error) {
	valuesYaml := new(bytes.Buffer)

	encoder := yamlDecoder.NewEncoder(valuesYaml)
	err := encoder.Encode(&values)
	if err != nil {
		return "", err
	}

	json, err := yaml.ToJSON(valuesYaml.Bytes())
	if err != nil {
		return "", err
	}
	return string(json), nil
}
