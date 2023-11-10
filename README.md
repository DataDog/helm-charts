# Datadog Helm Charts

[![Artifact HUB](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/datadog)](https://artifacthub.io/packages/search?repo=datadog)

Official Helm charts for Datadog products. Currently supported:
- [Datadog Agents](charts/datadog/README.md) (datadog/datadog)
- [Observability Pipelines Worker](charts/observability-pipelines-worker/README.md) (datadog/observability-pipelines-worker)

## How to use Datadog Helm repository

You need to add this repository to your Helm repositories:

```shell
helm repo add datadog https://helm.datadoghq.com
helm repo update
```
