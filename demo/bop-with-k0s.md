# Install Boundless Operator with K0s

### Pre-requisites:

Ensure that AWS machine already exists. Follow [Create machines on AWS](..%2FREADME.md#create-machines-on-aws) to create the machines.

### Install Boundless Operator with K0s

#### Boundless Operator Blueprint

Create a file named `blueprint.yaml` with the following content. Ensure that `infra` section has the correct IP addresses of the machines.

```yaml
apiVersion: boctl.mirantis.com/v1alpha1
kind: Cluster
metadata:
  name: bop-cluster
spec:
  infra:
    hosts:
    - ssh:
        address: 3.80.73.246
        keyPath: ./example/aws-tf/aws_private.pem
        port: 22
        user: ubuntu
      role: controller
    - ssh:
        address: 3.226.255.130
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
      - name: example-server
        kind: MKEAddon
        enabled: true
        namespace: default
        chart:
          name: nginx
          repo: https://charts.bitnami.com/bitnami
          version: 15.1.1
          values: |
            service:
              type: ClusterIP
```

#### Apply Blueprint

```bash
boctl apply --config blueprint.yaml
```

#### Connect to cluster

```bash
export KUBECONFIG=./kubeconfig
```

#### Verify k0s Installation

```bash
boctl check --config blueprint.yaml
```
