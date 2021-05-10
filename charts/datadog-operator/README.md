# Datadog Operator

![Version: 0.6.0](https://img.shields.io/badge/Version-0.6.0-informational?style=flat-square) ![AppVersion: 0.6.0](https://img.shields.io/badge/AppVersion-0.6.0-informational?style=flat-square)

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Allows to specify affinity for Datadog Operator PODs |
| apiKey | string | `nil` | Your Datadog API key |
| apiKeyExistingSecret | string | `nil` | Use existing Secret which stores API key instead of creating a new one |
| appKey | string | `nil` | Your Datadog APP key |
| appKeyExistingSecret | string | `nil` | Use existing Secret which stores APP key instead of creating a new one |
| datadog-crds.crds.datadogAgents | bool | `true` | Set to true to deploy the DatadogAgents CRD |
| datadog-crds.crds.datadogMetrics | bool | `true` | Set to true to deploy the DatadogMetrics CRD |
| datadog-crds.crds.datadogMonitors | bool | `true` | Set to true to deploy the DatadogMonitors CRD |
| datadogMonitor.enabled | bool | `false` | Enables the Datadog Monitor controller |
| fullnameOverride | string | `""` |  |
| image.pullPolicy | string | `"IfNotPresent"` | Define the pullPolicy for Datadog Operator image |
| image.repository | string | `"gcr.io/datadoghq/operator"` | Repository to use for Datadog Operator image |
| image.tag | string | `"0.6.0"` | Define the Datadog Operator version to use |
| imagePullSecrets | list | `[]` | Datadog Operator repository pullSecret (ex: specify docker registry credentials) |
| installCRDs | bool | `true` | Set to true to deploy the Datadog's CRDs |
| logLevel | string | `"info"` | Set Datadog Operator log level (debug, info, error, panic, fatal) |
| metricsPort | int | `8383` | Port used for OpenMetrics endpoint |
| nameOverride | string | `""` | Override name of app |
| nodeSelector | object | `{}` | Allows to schedule Datadog Operator on specific nodes |
| podAnnotations | object | `{}` | Allows setting additional annotations for Datadog Operator PODs |
| podLabels | object | `{}` | Allows setting additional labels for for Datadog Operator PODs |
| rbac.create | bool | `true` | Specifies whether the RBAC resources should be created |
| replicaCount | int | `1` | Number of instances of Datadog Operator |
| resources | object | `{}` | Set resources requests/limits for Datadog Operator PODs |
| secretBackend.arguments | string | `""` | Specifies the space-separated arguments passed to the command that implements the secret backend api |
| secretBackend.command | string | `""` | Specifies the path to the command that implements the secret backend api |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `nil` | The name of the service account to use. If not set name is generated using the fullname template |
| supportExtendedDaemonset | string | `"false"` | If true, supports using ExtendedDeamonSet CRD |
| tolerations | list | `[]` | Allows to schedule Datadog Operator on tainted nodes |