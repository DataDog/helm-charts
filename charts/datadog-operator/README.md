# Datadog Operator

![Version: 2.17.0](https://img.shields.io/badge/Version-2.17.0-informational?style=flat-square) ![AppVersion: 1.22.0](https://img.shields.io/badge/AppVersion-1.22.0-informational?style=flat-square)

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Allows to specify affinity for Datadog Operator PODs |
| apiKey | string | `nil` | Your Datadog API key |
| apiKeyExistingSecret | string | `nil` | Use existing Secret which stores API key instead of creating a new one |
| appKey | string | `nil` | Your Datadog APP key |
| appKeyExistingSecret | string | `nil` | Use existing Secret which stores APP key instead of creating a new one |
| clusterName | string | `nil` | Set a unique cluster name reporting from the Datadog Operator. |
| clusterRole | object | `{"allowCreatePodsExec":false,"allowReadAllResources":false}` | Set specific configuration for the cluster role |
| collectOperatorMetrics | bool | `true` | Configures an openmetrics check to collect operator metrics |
| containerSecurityContext | object | `{}` | A security context defines privileges and access control settings for a container. |
| datadogAgent.enabled | bool | `true` | Enables Datadog Agent controller |
| datadogAgentInternal.enabled | bool | `true` | Enables the Datadog Agent Internal controller |
| datadogAgentProfile.enabled | bool | `false` | If true, enables DatadogAgentProfile controller (beta). Requires v1.5.0+ |
| datadogCRDs.crds.datadogAgentInternals | bool | `true` | Set to true to deploy the DatadogAgentInternals CRD |
| datadogCRDs.crds.datadogAgentProfiles | bool | `false` | Set to true to deploy the DatadogAgentProfile CRD |
| datadogCRDs.crds.datadogAgents | bool | `true` | Set to true to deploy the DatadogAgents CRD |
| datadogCRDs.crds.datadogDashboards | bool | `false` | Set to true to deploy the DatadogDashboard CRD |
| datadogCRDs.crds.datadogGenericResources | bool | `false` | Set to true to deploy the DatadogGenericResource CRD |
| datadogCRDs.crds.datadogMetrics | bool | `true` | Set to true to deploy the DatadogMetrics CRD |
| datadogCRDs.crds.datadogMonitors | bool | `true` | Set to true to deploy the DatadogMonitors CRD |
| datadogCRDs.crds.datadogPodAutoscalers | bool | `true` | Set to true to deploy the DatadogPodAutoscalers CRD |
| datadogCRDs.crds.datadogSLOs | bool | `false` | Set to true to deploy the DatadogSLO CRD |
| datadogDashboard.enabled | bool | `false` | Enables the Datadog Dashboard controller |
| datadogGenericResource.enabled | bool | `false` | Enables the Datadog Generic Resource controller |
| datadogMonitor.enabled | bool | `false` | Enables the Datadog Monitor controller |
| datadogSLO.enabled | bool | `false` | Enables the Datadog SLO controller |
| dd_url | string | `nil` | The host of the Datadog intake server to send Agent data to, only set this option if you need the Agent to send data to a custom URL |
| deployment.annotations | object | `{}` | Allows setting additional annotations for the deployment resource |
| dnsConfig | object | `{}` | Specify DNS configuration options for Datadog Operator PODs |
| env | list | `[]` | Define any environment variables to be passed to the operator. |
| fullnameOverride | string | `""` |  |
| image.doNotCheckTag | bool | `false` | Permit skipping operator image tag compatibility with the chart. |
| image.pullPolicy | string | `"IfNotPresent"` | Define the pullPolicy for Datadog Operator image |
| image.repository | string | `"gcr.io/datadoghq/operator"` | Repository to use for Datadog Operator image |
| image.tag | string | `"1.22.0"` | Define the Datadog Operator version to use |
| imagePullSecrets | list | `[]` | Datadog Operator repository pullSecret (ex: specify docker registry credentials) |
| installCRDs | bool | `true` | Set to true to deploy the Datadog's CRDs |
| introspection.enabled | bool | `false` | If true, enables introspection feature (beta). Requires v1.4.0+ |
| livenessProbe | object | `{"initialDelaySeconds":15,"periodSeconds":10}` | Add default livenessProbe settings. HTTP GET is not configurable as it is hardcoded in the Operator. |
| logLevel | string | `"info"` | Set Datadog Operator log level (debug, info, error, panic, fatal) |
| maximumGoroutines | string | `nil` | Override default goroutines threshold for the health check failure. |
| metricsPort | int | `8383` | Port used for OpenMetrics endpoint |
| nameOverride | string | `""` | Override name of app |
| nodeSelector | object | `{}` | Allows to schedule Datadog Operator on specific nodes |
| operatorMetricsEnabled | string | `"true"` | Enable forwarding of Datadog Operator metrics and events to Datadog. |
| podAnnotations | object | `{}` | Allows setting additional annotations for Datadog Operator PODs |
| podLabels | object | `{}` | Allows setting additional labels for for Datadog Operator PODs |
| rbac.create | bool | `true` | Specifies whether the RBAC resources should be created |
| remoteConfiguration.enabled | bool | `false` | If true, enables Remote Configuration in the Datadog Operator (beta). Requires clusterName, API and App keys to be set. |
| replicaCount | int | `1` | Number of instances of Datadog Operator |
| resources | object | `{}` | Set resources requests/limits for Datadog Operator PODs |
| secretBackend.arguments | string | `""` | Specifies the space-separated arguments passed to the command that implements the secret backend api |
| secretBackend.command | string | `""` | Specifies the path to the command that implements the secret backend api |
| serviceAccount.annotations | object | `{}` | Allows setting additional annotations for service account |
| serviceAccount.automountServiceAccountToken | bool | `true` | Specifies whether the service account token should be automatically mounted |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `nil` | The name of the service account to use. If not set name is generated using the fullname template |
| site | string | `nil` | The site of the Datadog intake to send data to (documentation: https://docs.datadoghq.com/getting_started/site/) |
| supportExtendedDaemonset | string | `"false"` | If true, supports using ExtendedDaemonSet CRD |
| tolerations | list | `[]` | Allows to schedule Datadog Operator on tainted nodes |
| volumeMounts | list | `[]` | Specify additional volumes to mount in the container |
| volumes | list | `[]` | Specify additional volumes to mount in the container |
| watchNamespaces | list | `[]` | Restricts the Operator to watch its managed resources on specific namespaces unless CRD-specific watchNamespaces properties are set |
| watchNamespacesAgent | list | `[]` | Restricts the Operator to watch DatadogAgent resources on specific namespaces. Requires v1.8.0+ |
| watchNamespacesAgentProfile | list | `[]` | Restricts the Operator to watch DatadogAgentProfile resources on specific namespaces. Requires v1.8.0+ |
| watchNamespacesMonitor | list | `[]` | Restricts the Operator to watch DatadogMonitor resources on specific namespaces. Requires v1.8.0+ |
| watchNamespacesSLO | list | `[]` | Restricts the Operator to watch DatadogSLO resources on specific namespaces. Requires v1.8.0+ |

## How to configure which namespaces are watched by the Operator.

By default, the Operator only watches resources (`DatadogAgent`, `DatadogMonitor`) that are present in the same namespace.

It is possible to configure the Operator to watch resources that are present in one or several specific namespaces.

```yaml
watchNamespaces:
- "default"
- "datadog"
```

To watch all namespaces, the following configuration needs to be used:

```yaml
watchNamespaces:
- ""
```