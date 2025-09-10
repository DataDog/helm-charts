
# How to preserve CRDs

# 1st iteration with manually replacing annotations

```sh
# install poc1
❯ helm install datadog ./charts/datadog -f charts/datadog/migration-poc1.yaml

# update CRD annotations to use `datadog-operator` as release
❯ kg crd -l "app.kubernetes.io/name=datadogCRDs" -o name | xargs -n1 -I{} kubectl annotate --overwrite {} meta.helm.sh/release-name=datadog-operator meta.helm.sh/release-namespace=datadog-agent

# install Operator chart with `datadog-operator` release name
❯ helm install datadog-operator datadog/datadog-operator --set datadogAgentProfile.enabled=true --set datadogCRDs.crds.datadogAgentProfiles=true

# Uninstall datadog 
❯ helm uninstall datadog
```

## 2nd iteration


```sh
❯ helm install datadog ./charts/datadog -f charts/datadog/migration-poc1.yaml

# CRD metadata
❯ kgyaml crd datadogagents.datadoghq.com | yq .metadata
annotations:
  controller-gen.kubebuilder.io/version: v0.16.3
  meta.helm.sh/release-name: datadog
  meta.helm.sh/release-namespace: datadog-agent
creationTimestamp: "2025-09-10T18:53:26Z"
generation: 1
labels:
  app.kubernetes.io/instance: datadog
  app.kubernetes.io/managed-by: Helm
  app.kubernetes.io/name: datadogCRDs
  helm.sh/chart: datadogCRDs-2.11.0
name: datadogagents.datadoghq.com
resourceVersion: "162068"
uid: 3a6b9360-2c72-4055-a337-33a70c9cf5b9

# add resource policy
❯ helm upgrade datadog ./charts/datadog -f charts/datadog/migration-poc1.yaml --set datadog-operator.datadogCRDs.keepInstalledCrds=true

❯ kgyaml crd datadogagents.datadoghq.com | yq .metadata
annotations:
  controller-gen.kubebuilder.io/version: v0.16.3
  helm.sh/resource-policy: keep                     # <- added
  meta.helm.sh/release-name: datadog
  meta.helm.sh/release-namespace: datadog-agent
creationTimestamp: "2025-09-10T18:53:26Z"
generation: 1
labels:
  app.kubernetes.io/instance: datadog
  app.kubernetes.io/managed-by: Helm
  app.kubernetes.io/name: datadogCRDs
  helm.sh/chart: datadogCRDs-2.11.0
name: datadogagents.datadoghq.com
resourceVersion: "162492"
uid: 3a6b9360-2c72-4055-a337-33a70c9cf5b9

# update annotation
❯ helm upgrade datadog ./charts/datadog -f charts/datadog/migration-poc1.yaml --set datadog-operator.datadogCRDs.keepInstalledCrds=true --set datadog-operator.datadogCRDs.orphanInstalledCrds=true

❯ kgyaml crd datadogagents.datadoghq.com | yq .metadata
annotations:
  controller-gen.kubebuilder.io/version: v0.16.3
  helm.sh/resource-policy: keep
  meta.helm.sh/release-name: datadog-operator           # <- switched to datadog-operator, release name we use with operator
  meta.helm.sh/release-namespace: datadog-agent
creationTimestamp: "2025-09-10T18:53:26Z"
generation: 1
labels:
  app.kubernetes.io/instance: datadog
  app.kubernetes.io/managed-by: Helm
  app.kubernetes.io/name: datadogCRDs
  helm.sh/chart: datadogCRDs-2.11.0
name: datadogagents.datadoghq.com
resourceVersion: "162977"
uid: 3a6b9360-2c72-4055-a337-33a70c9cf5b9

❯ helm install datadog-operator datadog/datadog-operator

# instance is now datadog-operator not datadog
❯ kgyaml crd datadogagents.datadoghq.com | yq .metadata
annotations:
  controller-gen.kubebuilder.io/version: v0.16.3
  helm.sh/resource-policy: keep
  meta.helm.sh/release-name: datadog-operator
  meta.helm.sh/release-namespace: datadog-agent
creationTimestamp: "2025-09-10T18:53:26Z"
generation: 1
labels:
  app.kubernetes.io/instance: datadog-operator       # <- changed to operator release
  app.kubernetes.io/managed-by: Helm
  app.kubernetes.io/name: datadogCRDs
  helm.sh/chart: datadogCRDs-2.11.0
name: datadogagents.datadoghq.com
resourceVersion: "163388"
uid: 3a6b9360-2c72-4055-a337-33a70c9cf5b9

# two operator pods
❯ kgp
NAME                                        READY   STATUS    RESTARTS   AGE
datadog-7ts9q                               3/3     Running   0          3m12s
datadog-84tv2                               2/3     Running   0          23s
datadog-9jp2m                               3/3     Running   0          3m12s
datadog-cluster-agent-57869db9f7-hjfps      1/1     Running   0          85s
datadog-cluster-agent-57869db9f7-tslf6      1/1     Running   0          55s
datadog-clusterchecks-66ccc74f64-bx4cr      1/1     Running   0          55s
datadog-clusterchecks-66ccc74f64-h9h6p      1/1     Running   0          85s
datadog-datadog-operator-749847cdfc-p52tr   1/1     Running   0          3m12s
datadog-kpwgr                               3/3     Running   0          54s
datadog-operator-58bd7c68f5-6226j           1/1     Running   0          2s
datadog-qkp66                               3/3     Running   0          2m1s
datadog-rq5sn

# leases points to old pod
❯ kgyaml leases.coordination.k8s.io datadog-operator-lock
apiVersion: coordination.k8s.io/v1
kind: Lease
metadata:
  creationTimestamp: "2025-09-09T18:51:59Z"
  name: datadog-operator-lock
  namespace: datadog-agent
  resourceVersion: "163524"
  uid: 296995da-eac4-4b15-b989-0a9a87591b1e
spec:
  acquireTime: "2025-09-10T18:54:49.377568Z"
  holderIdentity: datadog-datadog-operator-749847cdfc-p52tr_5dd9cd6b-db08-491d-846d-707dca1fc53b
  leaseDurationSeconds: 60
  leaseTransitions: 10
  renewTime: "2025-09-10T18:57:04.437106Z"


❯ helm uninstall datadog
These resources were kept due to the resource policy:
[CustomResourceDefinition] datadogmonitors.datadoghq.com
[CustomResourceDefinition] datadogpodautoscalers.datadoghq.com
[CustomResourceDefinition] datadogagents.datadoghq.com
[CustomResourceDefinition] datadogmetrics.datadoghq.com

release "datadog" uninstalled

# crd still there
❯ kg crd
NAME                                  CREATED AT
datadogagents.datadoghq.com           2025-09-10T18:53:26Z
datadogmetrics.datadoghq.com          2025-09-10T18:53:26Z
datadogmonitors.datadoghq.com         2025-09-10T18:53:26Z
datadogpodautoscalers.datadoghq.com   2025-09-10T18:53:26Z

```