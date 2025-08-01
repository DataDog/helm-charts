# Datadog Synthetics Private Location

![Version: 0.17.11](https://img.shields.io/badge/Version-0.17.11-informational?style=flat-square) ![AppVersion: 1.58.0](https://img.shields.io/badge/AppVersion-1.58.0-informational?style=flat-square)

[Datadog](https://www.datadoghq.com/) is a hosted infrastructure monitoring platform. This chart adds a Datadog Synthetics Private Location Deployment. For more information about synthetics monitoring with Datadog, please refer to the [Datadog documentation website](https://docs.datadoghq.com/synthetics/private_locations/?tab=helmchart).

## How to use Datadog Helm repository

You need to add this repository to your Helm repositories:

```
helm repo add datadog https://helm.datadoghq.com
helm repo update
```

## Quick start

To install the chart with the release name `<RELEASE_NAME>`, retrieve your Private Location configuration file from your [Synthetics Private Location settings page](https://app.datadoghq.com/synthetics/settings/private-locations/) and save it under `config.json` then run:

```bash
helm install <RELEASE_NAME> datadog/synthetics-private-location --set-file configFile=config.json
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Allows to specify affinity for Datadog Synthetics Private Location PODs |
| commonLabels | object | `{}` | Labels to apply to all resources |
| configConfigMap | string | `""` | Config Map that stores the configuration of the private location worker for the deployment |
| configFile | string | `"{}"` | JSON string containing the configuration of the private location worker |
| configSecret | string | `""` | Name of the secret that stores the configuration of the private location worker for the deployment. Use it only if you want to manage the secret outside of the Helm chart as using `configFile` will create a secret. The `data` inside the secret needs to have the key `synthetics-check-runner.json`. |
| dnsConfig | object | `{}` | DNS Config to set to the Datadog Synthetics Private Location PODs |
| dnsPolicy | string | `"ClusterFirst"` | DNS Policy to set to the Datadog Synthetics Private Location PODs |
| enableStatusProbes | bool | `false` | Enable both liveness and readiness probes (minimal private location image version required: 1.12.0) |
| env | list | `[]` | Set environment variables |
| envFrom | list | `[]` | Set environment variables from configMaps and/or secrets |
| extraVolumeMounts | list | `[]` | Optionally specify extra list of additional volumeMounts for container |
| extraVolumes | list | `[]` | Optionally specify extra list of additional volumes to mount into the pod |
| fullnameOverride | string | `""` | Override the full qualified app name |
| hostAliases | list | `[]` | Add entries to Datadog Synthetics Private Location PODs' /etc/hosts |
| image.pullPolicy | string | `"IfNotPresent"` | Define the pullPolicy for Datadog Synthetics Private Location image |
| image.repository | string | `"gcr.io/datadoghq/synthetics-private-location-worker"` | Repository to use for Datadog Synthetics Private Location image |
| image.tag | string | `"1.58.0"` | Define the Datadog Synthetics Private Location version to use |
| imagePullSecrets | list | `[]` | Datadog Synthetics Private Location repository pullSecret (ex: specify docker registry credentials) |
| nameOverride | string | `""` | Override name of app |
| nodeSelector | object | `{}` | Allows to schedule Datadog Synthetics Private Location on specific nodes |
| podAnnotations | object | `{}` | Annotations to set to Datadog Synthetics Private Location PODs |
| podDisruptionBudget | object | `{"enabled":false,"minAvailable":1}` | Allows to create and configure PodDisruptionBudget for Datadog Synthetics Private Location deployment |
| podLabels | object | `{}` | Labels to be placed on pods managed by the deployment |
| podSecurityContext | object | `{}` | Security context to set to Datadog Synthetics Private Location PODs |
| priorityClassName | string | `""` | Allows to specify PriorityClass for Datadog Synthetics Private Location PODs |
| replicaCount | int | `1` | Number of instances of Datadog Synthetics Private Location |
| resources | object | `{}` | Set resources requests/limits for Datadog Synthetics Private Location PODs |
| securityContext | object | `{}` | Security context to set to the Datadog Synthetics Private Location container |
| serviceAccount.annotations | object | `{}` | Annotations for the service account |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `""` | The name of the service account to use. If not set name is generated using the fullname template |
| tolerations | list | `[]` | Allows to schedule Datadog Synthetics Private Location on tainted nodes |
