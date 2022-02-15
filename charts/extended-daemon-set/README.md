# Extended DaemonSet

![Version: v0.2.2](https://img.shields.io/badge/Version-v0.2.2-informational?style=flat-square) ![AppVersion: v0.7.0](https://img.shields.io/badge/AppVersion-v0.7.0-informational?style=flat-square)

This chart installs the Extended DaemonSet (EDS). It aims to provide a new implementation of the Kubernetes DaemonSet resource with key features:
- Canary Deployment: Deploy a new DaemonSet version with only a few nodes.
- Custom Rolling Update: Improve the default rolling update logic available in Kubernetes batch/v1 Daemonset.

For more information, please refer to the [EDS repo](https://github.com/DataDog/extendeddaemonset/).

## How to use the Datadog Helm repository

You need to add this repository to your Helm repositories:

```
helm repo add datadog https://helm.datadoghq.com
helm repo update
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Allows to specify affinity for the Extended DaemonSet PODs |
| clusterScope | bool | `false` | Allows ExtendedDaemonset controller to watch all namespaces |
| fullnameOverride | string | `""` | Overrides the full qualified app name |
| image.pullPolicy | string | `"IfNotPresent"` | Defines the pullPolicy for the Extended DaemonSet image |
| image.repository | string | `"datadog/extendeddaemonset"` | Repository to use for the Extended DaemonSet image |
| image.tag | string | `"v0.7.0"` | Defines the Extended DaemonSet version to use |
| imagePullSecrets | list | `[]` | Extended DaemonSet image repository pullSecret (ex: specify docker registry credentials) |
| installCRDs | bool | `true` | Set to true to deploy all the ExtendedDaemonSet CRDs (ExtendedDaemonSet, ExtendedDaemonSetReplicaSet, ExtendedDaemonSettings) |
| logLevel | string | `"info"` | Sets the log level (debug, info, error, panic, fatal) |
| nameOverride | string | `""` | Overrides name of app |
| nodeSelector | object | `{}` | Allows to schedule on specific nodes |
| podSecurityContext | object | `{}` | Sets the pod security context |
| pprof.enabled | bool | `false` | Set to true to enable pprof |
| rbac.create | bool | `true` | Specifies whether the RBAC resources should be created |
| replicaCount | int | `1` | Number of instances of the Extended DaemonSet |
| resources | object | `{}` | Sets resources requests/limits for Datadog Operator PODs |
| securityContext | object | `{}` | Sets the security context |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `nil` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template |
| tolerations | list | `[]` | Allows to schedule on tainted nodes |

## Developers

### How to update CRDs

```shell
./update-crds.sh <extendeddaemonset-tag>
```
