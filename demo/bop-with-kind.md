# Install Boundless Operator on Kind

### Create a kind cluster
`kind create cluster -n mke4`

### Install MKE Operator
~`kubectl apply -f ./manifests/mke-operator.yaml`~
TBD

#### Check if operator is running
```
kubectl get pods -n mke-operator-system

NAME                                               READY   STATUS    RESTARTS   AGE
mke-operator-controller-manager-75b6f5d4db-gc2rp   2/2     Running   0          4m24s
```

### Install core component and ad-on using MKE Operator

To do this, we will create a `Cluster` object specifying which components to install and their configuration:
```
cat <<EOF | kubectl create -f -
apiVersion: mke.mirantis.com/v1alpha1
kind: Cluster
metadata:
  labels:
    app.kubernetes.io/name: mkecluster
    app.kubernetes.io/instance: mkecluster-sample
    app.kubernetes.io/part-of: mke-operator
    app.kubernetes.io/created-by: mke-operator
  name: mkecluster-sample
spec:
  components:
    core:
      ingress:
        enabled: true
        provider: ingress-nginx
        config: |
          controller:
            service:
              nodePorts:
                http: 30000
                https: 30001
              type: NodePort
    addons:
      - name: my-grafana
        kind: MKEAddon
        enabled: true
        namespace: monitoring
        chart:
          name: grafana
          repo: https://grafana.github.io/helm-charts
          version: 6.58.7
          values: |
            ingress:
              enabled: true
EOF
```

After some time, the operator will install the specified components:
```
kubectl get deploy
NAME                       READY   UP-TO-DATE   AVAILABLE   AGE
ingress-nginx-controller   1/1     1            1           47s
nginx                      1/1     1            1           47s
```

### Add another add-on
To add more ad-ons to MKE cluster, we can create `MKEAddon` object:
```
cat <<EOF | kubectl create -f -
apiVersion: mke.mirantis.com/v1alpha1
kind: MkeAddon
metadata:
  name: mkeaddon-example-server
spec:
  name: example-server
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
EOF
```

After a while, the the add-on will be installed
```
kubectl get deploy
NAME                       READY   UP-TO-DATE   AVAILABLE   AGE
grafana                    1/1     1            1           7m44s
ingress-nginx-controller   1/1     1            1           7m39s
nginx                      1/1     1            1           7s
```

#### Note: This can also be done by using `MkeCluster` object with following command
```
cat <<EOF | kubectl create -f -
apiVersion: mke.mirantis.com/v1alpha1
kind: MkeCluster
metadata:
  labels:
    app.kubernetes.io/name: mkecluster
    app.kubernetes.io/instance: mkecluster-sample
    app.kubernetes.io/part-of: mke-operator
    app.kubernetes.io/created-by: mke-operator
  name: mkecluster-sample
spec:
  components:
    core:
      ingress:
        enabled: true
        provider: ingress-nginx
        config: |
          controller:
            service:
              nodePorts:
                http: 30000
                https: 30001
              type: NodePort
    addons:
      - name: my-grafana
        kind: MKEAddon
        enabled: true
        namespace: monitoring
        chart:
          name: grafana
          repo: https://grafana.github.io/helm-charts
          version: 6.58.7
          values: |
            ingress:
              enabled: true
            - name: example-server
      - name: mkeaddon-example-server
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
EOF
```

### Modify existing add-on

Currently, there are two add-on installed:
```
kubectl get mkeaddons
NAME                      AGE
mke-grafana               9m34s
mkeaddon-example-server   113s
```

Let’s modify one of the add-on’s configuration. We will modify `mkeaddon-example-server` and update the configuration:

Currently, the the “nginx” server is using `ClusterIP` as the service type.
```
kubectl get svc nginx
NAME    TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)   AGE
nginx   ClusterIP   10.96.154.133   <none>        80/TCP    14m
```

Lets change that to `NodePort`. This time, using `MkeCluster` object:
```
cat <<EOF | kubectl apply -f -
apiVersion: mke.mirantis.com/v1alpha1
kind: MkeCluster
metadata:
  labels:
    app.kubernetes.io/name: mkecluster
    app.kubernetes.io/instance: mkecluster-sample
    app.kubernetes.io/part-of: mke-operator
    app.kubernetes.io/created-by: mke-operator
  name: mkecluster-sample
spec:
  components:
    core:
      ingress:
        enabled: true
        provider: ingress-nginx
        config: |
          controller:
            service:
              nodePorts:
                http: 30000
                https: 30001
              type: NodePort
    addons:
      - name: my-grafana
        kind: MKEAddon
        enabled: true
        namespace: monitoring
        chart:
          name: grafana
          repo: https://grafana.github.io/helm-charts
          version: 6.58.7
          values: |
            ingress:
              enabled: true
            - name: example-server
      - name: mkeaddon-example-server
        kind: MKEAddon
        enabled: true
        namespace: default
        chart:
          name: nginx
          repo: https://charts.bitnami.com/bitnami
          version: 15.1.1
          values: |
            service:
              type: NodePort
EOF
```

After a while, the `Service` object should should show “NodePort”
```
kubectl get svc nginx
NAME    TYPE       CLUSTER-IP      EXTERNAL-IP   PORT(S)        AGE
nginx   NodePort   10.96.154.133   <none>        80:30429/TCP   54m
```

### Destroy the cluster
`kind delete cluster -n mke4`
