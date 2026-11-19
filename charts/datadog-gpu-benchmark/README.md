# Datadog GPU Benchmark

![Version: 0.2.0](https://img.shields.io/badge/Version-0.2.0-informational?style=flat-square) ![AppVersion: 0.1.0](https://img.shields.io/badge/AppVersion-0.1.0-informational?style=flat-square)

## Overview

[Datadog GPU Monitoring](https://docs.datadoghq.com/gpu_monitoring/) can benchmark your inference endpoints from inside your cluster: it measures time to first token, time per output token and throughput under load, next to your GPU and inference server metrics.

Each benchmark run is one Kubernetes Job, which Datadog creates in the `datadog-gpu-benchmark` namespace through the [Private Action Runner](https://docs.datadoghq.com/actions/private_actions/) of your Cluster Agent, and deletes when the run ends. The Job runs Datadog's benchmark image (based on [NVIDIA AIPerf](https://github.com/ai-dynamo/aiperf)) and reports through its logs: it never calls the Kubernetes API, and Datadog never reads anything else in your cluster.

This chart gives the Cluster Agent what it needs in that namespace:

| Resource | Purpose |
| --- | --- |
| ServiceAccount `datadog-gpu-benchmark-run` | Runs benchmark pods, with no role and no token |
| Role and RoleBinding `cluster-agent-gpu-benchmark` | Lets the Cluster Agent's ServiceAccount create, read and delete Jobs, read their pods and logs, and read the pod overrides, in this namespace only |
| ConfigMap `datadog-gpu-benchmark-pod` | Placement, labels and image mirror for benchmark pods (`podOverrides`) |

## Prerequisites

- The Datadog Cluster Agent 7.85+ (the first release with the `getPodLogs` action) with its Private Action Runner enabled, and the benchmark actions allowed. With the `datadog` chart:

  ```yaml
  clusterAgent:
    privateActionRunner:
      enabled: true
      actionsAllowlist:
        - com.datadoghq.kubernetes.batch.createJob
        - com.datadoghq.kubernetes.batch.getJob
        - com.datadoghq.kubernetes.batch.deleteJob
        - com.datadoghq.kubernetes.core.listPod
        - com.datadoghq.kubernetes.core.getPodLogs
        - com.datadoghq.kubernetes.core.getConfigMap
  ```

- CPU capacity for the runs: benchmark pods need no GPU.

## Installation

```bash
helm repo add datadog https://helm.datadoghq.com
helm repo update
helm install datadog-gpu-benchmark datadog/datadog-gpu-benchmark \
  --namespace datadog-gpu-benchmark --create-namespace \
  --set clusterAgent.namespace=<cluster agent namespace> \
  --set clusterAgent.serviceAccount=<cluster agent ServiceAccount>
```

The chart must be installed in the `datadog-gpu-benchmark` namespace: Datadog creates the benchmark Jobs there. Benchmark pods run non-root, with no capabilities and a read-only root filesystem, so the namespace can enforce Pod Security `restricted`:

```bash
kubectl label namespace datadog-gpu-benchmark pod-security.kubernetes.io/enforce=restricted
```

The Cluster Agent's ServiceAccount is `<release>-cluster-agent` when you install the Agent with the `datadog` chart, and `<DatadogAgent name>-cluster-agent` with the Datadog Operator: `kubectl get serviceaccounts -A | grep cluster-agent` lists it.

## Placement

Keep benchmark pods off your GPU nodes and away from inference workloads with `podOverrides`, for example on a tainted CPU node pool:

```yaml
podOverrides:
  nodeSelector:
    node-pool: cpu
  tolerations:
    - key: dedicated
      operator: Equal
      value: benchmarks
      effect: NoSchedule
```

## Image mirror

If your cluster can't pull from Datadog's public registries, mirror the image and set `podOverrides.imageRepository` (and `podOverrides.imagePullSecrets` if needed). Datadog keeps the image digest, so the mirror must hold that exact image.

## Datasets

Runs can replay your own traffic (Mooncake or Weka traces) instead of synthetic prompts. In your cluster, a dataset stays on a PersistentVolumeClaim of the `datadog-gpu-benchmark` namespace: benchmark pods mount it read-only, and its content never leaves your cluster. Set `datasetVolume.enabled` to have the chart create the `datadog-gpu-benchmark-datasets` claim, or use a claim of your own in the namespace. The claim is kept when the release is uninstalled.

```yaml
datasetVolume:
  enabled: true
  storageClassName: efs   # your cluster's default when empty
  size: 100Gi
```

Copy a dataset onto the claim with a short-lived pod that mounts it:

```bash
kubectl -n datadog-gpu-benchmark run dataset-loader --image=busybox:1.36 --restart=Never \
  --overrides='{"spec":{"containers":[{"name":"dataset-loader","image":"busybox:1.36","command":["sleep","3600"],"volumeMounts":[{"name":"data","mountPath":"/data"}]}],"volumes":[{"name":"data","persistentVolumeClaim":{"claimName":"datadog-gpu-benchmark-datasets"}}]}}'
kubectl -n datadog-gpu-benchmark wait --for=condition=Ready pod/dataset-loader
kubectl -n datadog-gpu-benchmark exec dataset-loader -- mkdir -p /data/2026-11
kubectl -n datadog-gpu-benchmark cp traces.jsonl dataset-loader:/data/2026-11/traces.jsonl
kubectl -n datadog-gpu-benchmark delete pod dataset-loader
```

Then register the dataset in Datadog with the claim name and its path on the volume, here `2026-11/traces.jsonl`. A Mooncake dataset is one JSONL file; a Weka dataset is a directory of per-trace JSON files (`kubectl cp ./weka dataset-loader:/data/weka`). Benchmark pods run as non-root, so files must be readable by everyone (`0644`, directories `0755`).

With a `ReadWriteOnce` volume (EBS, Persistent Disk), delete the loader pod before starting a run: the volume attaches to one node at a time, and a run whose pod can't mount it fails after 10 minutes pending. `ReadWriteMany` volumes (EFS, Filestore, Azure Files) don't have this constraint.

To keep datasets in a bucket instead, serve it through a CSI driver such as [Mountpoint for Amazon S3](https://github.com/awslabs/mountpoint-s3-csi-driver) or [Cloud Storage FUSE](https://cloud.google.com/kubernetes-engine/docs/how-to/cloud-storage-fuse-csi-driver), and bind the claim to its PersistentVolume with `datasetVolume.volumeName`. You then upload datasets with your bucket's own tools.

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| clusterAgent.namespace | string | `""` | Namespace of the Cluster Agent. Required. |
| clusterAgent.serviceAccount | string | `""` | Name of the Cluster Agent's ServiceAccount. Required. With the `datadog` Helm chart it is `<release>-cluster-agent`; with the Datadog Operator, `<DatadogAgent name>-cluster-agent`. |
| datasetVolume.accessModes | list | `["ReadWriteMany"]` | Access modes of the claim. `ReadWriteMany` (EFS, Filestore, Azure Files) lets you copy datasets while runs read them. With `ReadWriteOnce`, delete the pod you copied them with before starting a run. |
| datasetVolume.enabled | bool | `false` | Create the `datadog-gpu-benchmark-datasets` PersistentVolumeClaim. It is kept when the release is uninstalled. |
| datasetVolume.size | string | `"50Gi"` | Size of the claim. Leave room for every dataset you keep on it. |
| datasetVolume.storageClassName | string | `""` | StorageClass of the claim. Empty uses the cluster's default StorageClass. |
| datasetVolume.volumeName | string | `""` | Bind the claim to an existing PersistentVolume, for example one a bucket CSI driver (Mountpoint for Amazon S3, Cloud Storage FUSE) serves. Empty provisions a volume from the StorageClass. |
| podOverrides.affinity | object | `{}` | Affinity for benchmark pods. |
| podOverrides.annotations | object | `{}` | Annotations added to benchmark pods. |
| podOverrides.imagePullSecrets | list | `[]` | Secrets of the benchmark namespace the kubelet pulls the image with, e.g. `[{name: my-registry}]`. |
| podOverrides.imageRepository | string | `""` | Repository of a mirror of Datadog's benchmark image, without a tag or digest, e.g. `123456789012.dkr.ecr.us-east-1.amazonaws.com/datadog/gpu-benchmark`. Datadog keeps the image digest, so the mirror must hold that exact image. |
| podOverrides.labels | object | `{}` | Labels added to benchmark pods. `app.kubernetes.io/managed-by`, `gpu-benchmark/org` and `gpu-benchmark/run` are set by Datadog. |
| podOverrides.nodeSelector | object | `{}` | Node selector for benchmark pods. Benchmark pods only need CPU: keep them off GPU nodes. |
| podOverrides.priorityClassName | string | `""` | PriorityClass for benchmark pods. |
| podOverrides.tolerations | list | `[]` | Tolerations for benchmark pods, for example to run on a tainted CPU node pool. |
| podOverrides.topologySpreadConstraints | list | `[]` | Topology spread constraints for benchmark pods. |
