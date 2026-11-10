# Datadog GPU Benchmark

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![AppVersion: 0.1.0](https://img.shields.io/badge/AppVersion-0.1.0-informational?style=flat-square)

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

- The Datadog Cluster Agent 7.84+ with its Private Action Runner enabled, and the benchmark actions allowed. With the `datadog` chart:

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

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| clusterAgent.namespace | string | `""` | Namespace of the Cluster Agent. Required. |
| clusterAgent.serviceAccount | string | `""` | Name of the Cluster Agent's ServiceAccount. Required. With the `datadog` Helm chart it is `<release>-cluster-agent`; with the Datadog Operator, `<DatadogAgent name>-cluster-agent`. |
| podOverrides.affinity | object | `{}` | Affinity for benchmark pods. |
| podOverrides.annotations | object | `{}` | Annotations added to benchmark pods. |
| podOverrides.imagePullSecrets | list | `[]` | Secrets of the benchmark namespace the kubelet pulls the image with, e.g. `[{name: my-registry}]`. |
| podOverrides.imageRepository | string | `""` | Repository of a mirror of Datadog's benchmark image, without a tag or digest, e.g. `123456789012.dkr.ecr.us-east-1.amazonaws.com/datadog/gpu-benchmark`. Datadog keeps the image digest, so the mirror must hold that exact image. |
| podOverrides.labels | object | `{}` | Labels added to benchmark pods. `app.kubernetes.io/managed-by`, `gpu-benchmark/org` and `gpu-benchmark/run` are set by Datadog. |
| podOverrides.nodeSelector | object | `{}` | Node selector for benchmark pods. Benchmark pods only need CPU: keep them off GPU nodes. |
| podOverrides.priorityClassName | string | `""` | PriorityClass for benchmark pods. |
| podOverrides.tolerations | list | `[]` | Tolerations for benchmark pods, for example to run on a tainted CPU node pool. |
| podOverrides.topologySpreadConstraints | list | `[]` | Topology spread constraints for benchmark pods. |
