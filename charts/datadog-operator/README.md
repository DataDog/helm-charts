# Datadog Operator

![Version: 0.1.1](https://img.shields.io/badge/Version-0.1.1-informational?style=flat-square) ![AppVersion: 0.3.1](https://img.shields.io/badge/AppVersion-0.3.1-informational?style=flat-square)

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Allows to specify affinity for Datadog Operator PODs |
| fullnameOverride | string | `""` |  |
| image.pullPolicy | string | `"IfNotPresent"` | Define the pullPolicy for Datadog Operator image |
| image.repository | string | `"datadog/operator"` | Repository to use for Datadog Operator image |
| image.tag | string | `"0.3.1"` | Define the Datadog Operator version to use |
| imagePullSecrets | list | `[]` | Datadog Operator repository pullSecret (ex: specify docker registry credentials) |
| logLevel | string | `"info"` | Set Datadog Operator log level (debug, info, error, panic, fatal) |
| metricsPort | int | `8383` | Port used for OpenMetrics endpoint |
| nameOverride | string | `""` | Override name of app |
| nodeSelector | object | `{}` | Allows to schedule Datadog Operator on specific nodes |
| probesPort | int | `9090` | Port used by readiness/liveness probes |
| rbac.create | bool | `true` | Specifies whether the RBAC resources should be created |
| replicaCount | int | `1` | Number of instances of Datadog Operator |
| resources | object | `{}` | Set resources requests/limits for Datadog Operator PODs |
| secretBackend.command | string | `""` | Specifies the path to the command that implements the secret backend api |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `nil` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template |
| supportExtendedDaemonset | string | `"false"` | If true, supports using ExtendedDeamonSet CRD |
| tolerations | list | `[]` | Allows to schedule Datadog Operator on tainted nodes |