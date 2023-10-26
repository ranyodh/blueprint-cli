# Boundless Operator - Tech Preview

## Introduction

TBD

## Quick Start

### Install on Kind

1. Install Kind
   ````
2. Install Boundless CLI Binary: 
   ```
   
   curl -s -L https://github.com/ranyodh/boundless-tech-preview/releases/download/latest/bocli_darwin_x86_64.tar.gz | tar xvz - -C /usr/local/bin
   ``



### Prerequisites
Ensure that following are installed on the system:
* `k0sctl` (required for installing k0s distribution)
* `terraform` (for creating VMs on AWS)

### Create machines on AWS

There are `terraform` scripts in the `example/` directory that can be used to create machines on AWS.

1. `cd example/aws-tf`
2. Create a `terraform.tfvars` file with the content similar to:
   ```
   cluster_name = "rs-mke4-test"
   controller_count = 1
   worker_count = 1
   cluster_flavor = "m5.large"
   ```
3. `terraform init`
4. `terraform apply`
5. `terraform output --raw bop_cluster > ./blueprint.yaml`

### Install Boundless Operator

#### Compile the `boctl` binary
`make build`

#### Generate the `blueprint.yaml` config file
`make init`

Now, edit the `blueprint.yaml` file to set the `spec.infra.hosts` from the output of `terraform output --raw bop_cluster`.

The host section should look similar to:
```yaml
spec:
  infra:
    hosts:
    - ssh:
        address: 52.91.89.114
        keyPath: ./example/aws-tf/aws_private.pem
        port: 22
        user: ubuntu
      role: controller
    - ssh:
        address: 10.0.0.2
        keyPath: ./example/aws-tf/aws_private.pem
        port: 22
        user: ubuntu
      role: worker
```

##### Boundless Operator Blueprint

The complete `blueprint.yaml` file should look similar to the following. This config file will deploy MKE 4 with `k0s` as the Kubernetes distribution and `ingress-nginx` as the ingress controller and one add-on `nginx`:
```yaml
apiVersion: boctl.mirantis.com/v1beta1
kind: Cluster
metadata:
  name: bop-cluster
spec:
  infra:
    hosts:
    - ssh:
        address: 52.91.89.114
        keyPath: ./example/aws-tf/aws_private.pem
        port: 22
        user: ubuntu
      role: controller
    - ssh:
        address: 10.0.0.2
        keyPath: ./example/aws-tf/aws_private.pem
        port: 22
        user: ubuntu
      role: worker
  kubernetes:
    provider: k0s
    version: 1.27.4+k0s.0
  mke:
    components:
      core:
        ingress:
          enabled: true
          provider: ingress-nginx
          config:
            controller:
              service:
                nodePorts:
                  http: 30000
                  https: 30001
                type: NodePort
      addons:
      - name: example-server
        kind: MKEAddon
        enabled: true
        namespace: default
        chart:
          name: nginx
          repo: https://charts.bitnami.com/bitnami
          version: 15.1.1
          values: |2
            "service":
              "type": "ClusterIP"
```

#### Deploy Blueprint
`make apply`

#### Update Blueprint
`make update`

### Connect to cluster
The `kubeconfig` file will be generated in the current directory. Use this file to connect to the cluster.
```bash
export KUBECONFIG=./kubeconfig
kubectl get nodes
kubectl get pods
```

### Core Components

Currently, you can replace the ingress controller from `ingress-nginx` to `kong` by updating the `blueprint.yaml` file:
```yaml
spec:
  mke:
    components:
      core:
        ingress:
          enabled: true
          provider: kong # ingress-nginx, kong, etc.
```

> If the cluster is already deployed, run `make reset` to destroy the cluster and then run `make apply` to recreate it.

### Add-ons
Update the `blueprint.yaml` file to add add-ons to the cluster. The add-ons are defined in the `spec.mke.components.addons` section.

Any public Helm chart can be used as an add-on.

Use the following configuration to add the `grafana` as an add-on:
```yaml
spec:
  mke:
    components:
      addons:
      - name: my-grafana
        enabled: true
        kind: MKEAddon
        namespace: monitoring
        chart:
          name: grafana
          repo: https://grafana.github.io/helm-charts
          version: 6.58.7
          values: |
            ingress:
              enabled: true
```
and then run `make update` to update the cluster.

### Cleanup

#### Remove Boundless Operator from the cluster
`make reset`

#### Destroy the cluster
```bash
cd example/aws-tf
terraform destroy --auto-approve
```

## Install Boundless Operator on Kind cluster

[Install Boundless Operator on Kind](demo%2Fbop-with-kind.md)








