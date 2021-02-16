# Datadog Operator

![Version: 0.4.1](https://img.shields.io/badge/Version-0.4.1-informational?style=flat-square) ![AppVersion: 0.4.0](https://img.shields.io/badge/AppVersion-0.4.0-informational?style=flat-square)

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Allows to specify affinity for Datadog Operator PODs |
| datadog-crds.crds.datadogAgents | bool | `true` | Set to true to deploy the DatadogAgents CRD |
| datadog-crds.crds.datadogMetrics | bool | `true` | Set to true to deploy the DatadogMetrics CRD |
| fullnameOverride | string | `""` |  |
| image.pullPolicy | string | `"IfNotPresent"` | Define the pullPolicy for Datadog Operator image |
| image.repository | string | `"gcr.io/datadoghq/operator"` | Repository to use for Datadog Operator image |
| image.tag | string | `"0.4.0"` | Define the Datadog Operator version to use |
| imagePullSecrets | list | `[]` | Datadog Operator repository pullSecret (ex: specify docker registry credentials) |
| installCRDs | bool | `true` | Set to true to deploy the Datadog's CRDs |
| logLevel | string | `"info"` | Set Datadog Operator log level (debug, info, error, panic, fatal) |
| metricsPort | int | `8383` | Port used for OpenMetrics endpoint |
| nameOverride | string | `""` | Override name of app |
| nodeSelector | object | `{}` | Allows to schedule Datadog Operator on specific nodes |
| rbac.create | bool | `true` | Specifies whether the RBAC resources should be created |
| replicaCount | int | `1` | Number of instances of Datadog Operator |
| resources | object | `{}` | Set resources requests/limits for Datadog Operator PODs |
| secretBackend.command | string | `""` | Specifies the path to the command that implements the secret backend api |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `nil` | The name of the service account to use. If not set name is generated using the fullname template |
| supportExtendedDaemonset | string | `"false"` | If true, supports using ExtendedDeamonSet CRD |
| tolerations | list | `[]` | Allows to schedule Datadog Operator on tainted nodes |
| podAnnotations | object | `{}` | Allows setting additional annotations for Datadog Operator PODs |
| podLabels | object | `{}` | Allows setting additional labels for for Datadog Operator PODs |
