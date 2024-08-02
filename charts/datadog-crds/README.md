# Datadog CRDs

![Version: 1.7.2](https://img.shields.io/badge/Version-1.7.2-informational?style=flat-square) ![AppVersion: 1](https://img.shields.io/badge/AppVersion-1-informational?style=flat-square)

This chart was designed to allow other "datadog" charts to share `CustomResourceDefinitions` such as the `DatadogMetric`.

## How to use Datadog Helm repository

You need to add this repository to your Helm repositories:

```
helm repo add datadog https://helm.datadoghq.com
helm repo update
```

## Prerequisites

This chart can be used with Kubernetes `1.11+` or OpenShift `3.11+` since  `CustomResourceDefinitions` are supported starting with these versions.
But the recommended Kubernetes versions are `1.16+`.

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| crds.datadogAgentProfiles | bool | `false` | Set to true to deploy the DatadogAgentProfiles CRD |
| crds.datadogAgents | bool | `false` | Set to true to deploy the DatadogAgents CRD |
| crds.datadogMetrics | bool | `false` | Set to true to deploy the DatadogMetrics CRD |
| crds.datadogMonitors | bool | `false` | Set to true to deploy the DatadogMonitors CRD |
| crds.datadogPodAutoscalers | bool | `false` | Set to true to deploy the DatadogPodAutoscalers CRD |
| crds.datadogSLOs | bool | `false` | Set to true to deploy the DatadogSLO CRD |
| fullnameOverride | string | `""` | Override the fully qualified app name |
| migration.datadogAgents.conversionWebhook.enabled | bool | `false` |  |
| migration.datadogAgents.conversionWebhook.name | string | `"datadog-operator-webhook-service"` |  |
| migration.datadogAgents.conversionWebhook.namespace | string | `"default"` |  |
| migration.datadogAgents.useCertManager | bool | `false` |  |
| migration.datadogAgents.version | string | `"v2alpha1"` |  |
| nameOverride | string | `""` | Override name of app |

## Developers

### How to update CRDs

```shell
./update-crds.sh <datadog-operator-tag>
```
