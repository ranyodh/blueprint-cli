package config

import (
	"github.com/k0sproject/dig"
	v1 "github.com/k3s-io/helm-controller/pkg/apis/helm.cattle.io/v1"
	"sigs.k8s.io/yaml"
)

func ParseK0sCluster(data []byte) (K0sCluster, error) {
	var cluster K0sCluster
	err := yaml.Unmarshal(data, &cluster)
	if err != nil {
		return K0sCluster{}, err
	}
	return cluster, nil
}

func ParseMkeCluster(data []byte) (MkeCluster, error) {
	var mkeCluster MkeCluster
	err := yaml.Unmarshal(data, &mkeCluster)
	if err != nil {
		return MkeCluster{}, err
	}

	return mkeCluster, nil
}

func ParseCoreComponentManifests(data []byte) (v1.HelmChart, error) {
	var helmChart v1.HelmChart
	err := yaml.Unmarshal(data, &helmChart)
	if err != nil {
		return v1.HelmChart{}, err
	}

	return helmChart, nil
}

func ConvertToK0s(mke MkeCluster) K0sCluster {
	return K0sCluster{
		APIVersion: "k0sctl.k0sproject.io/v1beta1",
		Kind:       "Cluster",
		Metadata: Metadata{
			Name: mke.Metadata.Name,
		},
		Spec: K0sClusterSpec{
			Hosts: mke.Spec.Infra.Hosts,
			K0S: K0s{
				Version:       mke.Spec.Kubernetes.Version,
				DynamicConfig: digBool(mke.Spec.Kubernetes.Config, "dynamicConfig"),
				Config:        mke.Spec.Kubernetes.Config,
			},
		},
	}
}

func ConvertToMke(k0s K0sCluster, components Components) MkeCluster {
	return MkeCluster{
		APIVersion: k0s.APIVersion,
		Kind:       "Cluster",
		Metadata: Metadata{
			Name: k0s.Metadata.Name,
		},
		Spec: MkeClusterSpec{
			Infra: Infra{
				Hosts: k0s.Spec.Hosts,
			},
			Kubernetes: Kubernetes{
				Provider: "k0s",
				Version:  k0s.Spec.K0S.Version,
				Config:   k0s.Spec.K0S.Config,
			},
			Mke: Mke{
				Components: components,
			},
		},
	}
}

func DigToString(m dig.Mapping, keys ...string) string {
	val := m.Dig(keys...)
	if val == nil {
		return ""
	}
	return val.(string)
}

func digBool(m dig.Mapping, keys ...string) bool {
	val := m.Dig(keys...)
	if val == nil {
		return false
	}
	return val.(bool)
}
