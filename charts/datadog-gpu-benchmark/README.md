# Datadog GPU Benchmark

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![AppVersion: 0.1.0](https://img.shields.io/badge/AppVersion-0.1.0-informational?style=flat-square)

## Overview

[Datadog GPU Monitoring](https://docs.datadoghq.com/gpu_monitoring/) can benchmark your inference endpoints from inside your cluster: it measures time to first token, time per output token and throughput under load, next to your GPU and inference server metrics.

Each benchmark run is one Kubernetes Job, which Datadog creates in the `datadog-gpu-benchmark` namespace through the [Private Action Runner](https://docs.datadoghq.com/actions/private_actions/) of your Cluster Agent, and deletes when the run ends. The Job runs Datadog's benchmark image (based on [NVIDIA AIPerf](https://github.com/ai-dynamo/aiperf)) and reports through its logs: it never calls the Kubernetes API, and Datadog never reads anything else in your cluster.

This chart prepares that namespace:

| Resource | Purpose |
| --- | --- |
| Namespace `datadog-gpu-benchmark` | Pod Security `restricted` |
| Role and RoleBinding | Lets the Cluster Agent's ServiceAccount create, read and delete Jobs, read their pods and logs, and read the pod overrides, in this namespace only |
| ServiceAccount `datadog-gpu-benchmark-run` | Runs benchmark pods, with no role and no token |
| ConfigMap `datadog-gpu-benchmark-pod` | Placement, labels and image mirror for benchmark pods (`podOverrides`) |
| ValidatingAdmissionPolicy (Kubernetes 1.30+) | Admits only benchmark pods of the expected shape: Datadog's image pinned by digest, one container, the expected environment and volumes, no host access, no ServiceAccount token |

Optional guardrails, off by default: a NetworkPolicy (`networkPolicy.enabled`), a ResourceQuota (`resourceQuota.enabled`), a LimitRange (`limitRange.enabled`) and a low PriorityClass that never preempts other pods (`priorityClass.create`).

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
- Kubernetes 1.30+ for the admission policy. On older clusters, enforce the same rules with Kyverno or Gatekeeper.

## Installation

```bash
helm repo add datadog https://helm.datadoghq.com
helm repo update
helm install datadog-gpu-benchmark datadog/datadog-gpu-benchmark \
  --set clusterAgent.serviceAccount.name=<cluster agent ServiceAccount> \
  --set clusterAgent.serviceAccount.namespace=<cluster agent namespace>
```

The Cluster Agent's ServiceAccount is `<release>-cluster-agent` when you install the Agent with the `datadog` chart, and `<DatadogAgent name>-cluster-agent` with the Datadog Operator: `kubectl get serviceaccounts -A | grep cluster-agent` lists it.

## Network policy

To restrict benchmark pods' network, turn on the NetworkPolicy and list the endpoints you benchmark (it needs a CNI that enforces NetworkPolicy), for example a vLLM service in the `inference` namespace:

```yaml
networkPolicy:
  enabled: true
  egress:
    - to:
        - namespaceSelector:
            matchLabels:
              kubernetes.io/metadata.name: inference
      ports:
        - protocol: TCP
          port: 8000
```

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
priorityClass:
  create: true
```

## Image mirror

If your cluster can't pull from Datadog's public registries, mirror the image and set `podOverrides.imageRepository` (and `podOverrides.imagePullSecrets` if needed). Datadog keeps the image digest, so the mirror must hold that exact image. The admission policy allows the mirror automatically.

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| admissionPolicy.allowInlineScript | bool | `true` | Allow the benchmark container's `python -c <script>` command. Datadog runs the benchmark this way until the image has its own entrypoint; set to false then. |
| admissionPolicy.enabled | bool | `true` | Create a ValidatingAdmissionPolicy that only admits benchmark Jobs and pods of the expected shape: Datadog's image, one container, the expected environment variables and volumes, no host access, no ServiceAccount token. Rendered only on Kubernetes 1.30+ (`admissionregistration.k8s.io/v1`); on older clusters, use an equivalent Kyverno or Gatekeeper policy. |
| admissionPolicy.extraEnv | list | `[]` | Extra environment variable names benchmark containers may set. |
| admissionPolicy.imageRepositories | list | `["registry.datadoghq.com/gpu-benchmark","docker.io/datadog/gpu-benchmark","public.ecr.aws/datadog/gpu-benchmark","gcr.io/datadoghq/gpu-benchmark","eu.gcr.io/datadoghq/gpu-benchmark","asia.gcr.io/datadoghq/gpu-benchmark","datadoghq.azurecr.io/gpu-benchmark"]` | Image repositories benchmark pods may use. `podOverrides.imageRepository` is added automatically. |
| clusterAgent.serviceAccount.name | string | `""` | Name of the Cluster Agent's ServiceAccount, whose Private Action Runner creates the benchmark Jobs. Required. With the `datadog` Helm chart it is `<release>-cluster-agent`; with the Datadog Operator, `<DatadogAgent name>-cluster-agent`. |
| clusterAgent.serviceAccount.namespace | string | `""` | Namespace of the Cluster Agent's ServiceAccount. Required. |
| limitRange.enabled | bool | `false` | Cap the size of one benchmark container. |
| limitRange.max | object | `{"cpu":"16","memory":"32Gi"}` | The most one container may request or use. |
| limitRange.max.cpu | string | `"16"` |  |
| limitRange.max.memory | string | `"32Gi"` |  |
| namespace.annotations | object | `{}` | Extra annotations for the namespace. |
| namespace.create | bool | `true` | Create the benchmark namespace, with Pod Security `restricted` labels. Set to false if you create it yourself; it must then carry the same labels. |
| namespace.labels | object | `{}` | Extra labels for the namespace. |
| namespace.name | string | `"datadog-gpu-benchmark"` | The benchmark namespace. Datadog creates the Jobs in this namespace, so it cannot be changed. |
| networkPolicy.dns | object | `{"enabled":true,"namespaceSelector":{"kubernetes.io/metadata.name":"kube-system"},"podSelector":{"k8s-app":"kube-dns"}}` | Allow DNS to kube-dns or CoreDNS. |
| networkPolicy.dns.enabled | bool | `true` |  |
| networkPolicy.dns.namespaceSelector | object | `{"kubernetes.io/metadata.name":"kube-system"}` | Namespace of the cluster DNS. |
| networkPolicy.dns.namespaceSelector."kubernetes.io/metadata.name" | string | `"kube-system"` |  |
| networkPolicy.dns.podSelector | object | `{"k8s-app":"kube-dns"}` | Pods of the cluster DNS. |
| networkPolicy.dns.podSelector.k8s-app | string | `"kube-dns"` |  |
| networkPolicy.egress | list | `[]` | Egress rules (NetworkPolicyEgressRule) to the inference endpoints you benchmark, for example a namespace selector and port. Until you set them, benchmark pods only reach DNS, and runs fail to connect. |
| networkPolicy.enabled | bool | `false` | Restrict benchmark pods' network: no ingress, egress only to DNS and to `networkPolicy.egress`. Needs a CNI that enforces NetworkPolicy. When on, list your inference endpoints in `networkPolicy.egress`, or runs can't reach them. |
| podOverrides.affinity | object | `{}` | Affinity for benchmark pods. |
| podOverrides.annotations | object | `{}` | Annotations added to benchmark pods. `ad.datadoghq.com/logs_exclude` is set by Datadog. |
| podOverrides.imagePullSecrets | list | `[]` | Secrets of the benchmark namespace the kubelet pulls the image with, e.g. `[{name: my-registry}]`. |
| podOverrides.imageRepository | string | `""` | Repository of a mirror of Datadog's benchmark image, without a tag or digest, e.g. `123456789012.dkr.ecr.us-east-1.amazonaws.com/datadog/gpu-benchmark`. Datadog keeps the image digest, so the mirror must hold that exact image. Also allowed by the admission policy. |
| podOverrides.labels | object | `{}` | Labels added to benchmark pods. `app.kubernetes.io/managed-by`, `gpu-benchmark/org` and `gpu-benchmark/run` are set by Datadog. |
| podOverrides.nodeSelector | object | `{}` | Node selector for benchmark pods. Benchmark pods only need CPU: keep them off GPU nodes. |
| podOverrides.priorityClassName | string | `""` | PriorityClass for benchmark pods. Defaults to the chart's PriorityClass when `priorityClass.create` is true. |
| podOverrides.tolerations | list | `[]` | Tolerations for benchmark pods, for example to run on a tainted CPU node pool. |
| podOverrides.topologySpreadConstraints | list | `[]` | Topology spread constraints for benchmark pods. |
| priorityClass.create | bool | `false` | Create a low PriorityClass for benchmark pods, which never preempts other pods. |
| priorityClass.name | string | `"datadog-gpu-benchmark-low"` | Name of the PriorityClass. |
| priorityClass.value | int | `-100` | Priority of benchmark pods. Below 0, they are scheduled after your default-priority workloads. |
| rbac.create | bool | `true` | Create the Role and RoleBinding that let the Cluster Agent's ServiceAccount manage benchmark Jobs in the benchmark namespace, and nothing outside it. |
| resourceQuota.enabled | bool | `false` | Cap what benchmark runs may use in the namespace. The default `hard` allows two concurrent runs of the largest size. |
| resourceQuota.hard | object | `{"pods":"2","count/jobs.batch":"10","requests.cpu":"32","limits.cpu":"32","requests.memory":"64Gi","limits.memory":"64Gi"}` | The quota's limits. |
| resourceQuota.hard."count/jobs.batch" | string | `"10"` |  |
| resourceQuota.hard."limits.cpu" | string | `"32"` |  |
| resourceQuota.hard."limits.memory" | string | `"64Gi"` |  |
| resourceQuota.hard."requests.cpu" | string | `"32"` |  |
| resourceQuota.hard."requests.memory" | string | `"64Gi"` |  |
| resourceQuota.hard.pods | string | `"2"` |  |
