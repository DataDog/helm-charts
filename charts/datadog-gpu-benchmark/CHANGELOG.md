# Datadog GPU Benchmark changelog

## 0.1.0

* Initial version: the `datadog-gpu-benchmark` namespace for Datadog GPU Monitoring benchmark runs, with RBAC for the Cluster Agent's Private Action Runner, the runner ServiceAccount, pod overrides and a ValidatingAdmissionPolicy (Kubernetes 1.30+). Optional, off by default: a NetworkPolicy, a ResourceQuota, a LimitRange and a PriorityClass.
