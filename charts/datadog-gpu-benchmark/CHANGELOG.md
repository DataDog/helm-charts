# Datadog GPU Benchmark changelog

## 0.2.0

* Add `datasetVolume`: an optional `datadog-gpu-benchmark-datasets` PersistentVolumeClaim for your own datasets (Mooncake or Weka traces), which benchmark pods mount read-only. It is kept when the release is uninstalled. `datasetVolume.storageClassName` unset uses the cluster's default StorageClass, and `""` requests none, to bind a classless PersistentVolume with `datasetVolume.volumeName`.

## 0.1.0

* Initial version: what the Cluster Agent's Private Action Runner needs to run Datadog GPU Monitoring benchmark Jobs in the `datadog-gpu-benchmark` namespace: the benchmark pods' ServiceAccount, a Role and RoleBinding for the Cluster Agent, and the pod overrides ConfigMap.
