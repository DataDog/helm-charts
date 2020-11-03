# Datadog CRDs

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 1](https://img.shields.io/badge/AppVersion-1-informational?style=flat-square)

This chart was designed to allow others "datadog" charts to share `CustomResourcesDefinitions` such as the `DatadogMetric`.

## How to use Datadog Helm repository

You need to add this repository to your Helm repositories:

```
helm repo add datadog https://helm.datadoghq.com
helm repo update
```

## Prerequisites

This chart can be use with Kubernetes `1.11+` or OpenShift `3.11+` in order to support `CustomResourcesDefinitions`.
But the recommanded Kubernetes version are `1.16+`.

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| crds.datadogAgents | bool | `false` | Set to true to deploy the DatadogAgents CRD |
| crds.datadogMetrics | bool | `false` | Set to true to deploy the DatadogMetrics CRD |
| fullnameOverride | string | `""` | Override the full qualified app name |
| nameOverride | string | `""` | Override name of app |
