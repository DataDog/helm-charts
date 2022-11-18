# Observability Pipelines Worker

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.1.0](https://img.shields.io/badge/AppVersion-0.1.0-informational?style=flat-square)

## How to use Datadog Helm repository

You need to add this repository to your Helm repositories:

```
helm repo add datadog https://helm.datadoghq.com
helm repo update
```

## Quick start

### Installing the Observability Pipelines Worker chart

To install the chart with the release name `<RELEASE_NAME>` run:

```bash
helm install --name <RELEASE_NAME> \
  datadog/observability-pipelines-worker
```

### Uninstalling the chart

To uninstall/delete the `<RELEASE_NAME>` deployment:

```bash
helm delete <RELEASE_NAME>
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Values

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Configure [affinity and anti-affinity](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#affinity-and-anti-affinity) |
| args | list | `[]` | Override default image arguments |
| autoscaling.behavior | object | `{}` | Configure separate scale-up and scale-down behaviors |
| autoscaling.customMetric | object | `{}` | Target a custom metric for autoscaling |
| autoscaling.enabled | bool | `false` | Create a [HorizontalPodAutoscaler](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/) |
| autoscaling.maxReplicas | int | `10` | Set maximum replicas |
| autoscaling.minReplicas | int | `1` | Set minimum replicas |
| autoscaling.targetCPUUtilizationPercentage | int | `80` | Set target CPU utilization |
| autoscaling.targetMemoryUtilizationPercentage | int | `nil` | Set target memory utilization |
| command | list | `[]` | Override default image command |
| commonLabels | object | `{}` | Labels to apply to all resources |
| containerPorts | list | `[]` | Manually define containerPorts, overriding automated generation of containerPorts |
| customConfig | object | `{}` | Override Vector's default configs, if used **all** options need to be specified. This section supports using helm templates to populate dynamic values. See Vector's [configuration documentation](https://vector.dev/docs/reference/configuration/) for all options. |
| dataDir | string | `""` | Specify the path for Vector's data, only used when existingConfigMaps are used |
| datadog.apiKey | string | `"<DATADOG_API_KEY>"` | Your Datadog API key |
| datadog.apiKeyExistingSecret | string | `""` | Use existing Secret which stores API key instead of creating a new one. The value should be set with the `api-key` key inside the secret. |
| dnsConfig | object | `{}` | Specify the [dnsConfig](https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#pod-dns-config) |
| dnsPolicy | string | `"ClusterFirst"` | Specify the [dnsPolicy](https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#pod-s-dns-policy) |
| env | list | `[]` | Define environment variables |
| envFrom | list | `[]` | Define environment variables from ConfigMap or Secret data |
| existingConfigMaps | list | `[]` | List of existing ConfigMaps for Vector's configuration instead of creating a new one. Requires dataDir to be set. Additionally, containerPorts, service.ports, and serviceHeadless.ports should be specified based on your supplied configuration. If set, this parameter takes precedence over customConfig and the chart's default configs. |
| extraContainers | list | `[]` | Extra Containers to be added |
| extraVolumeMounts | list | `[]` | Additional Volume to mount |
| extraVolumes | list | `[]` | Additional Volumes to use |
| fullnameOverride | string | `""` | Override the full qualified app name |
| image.digest | string | `""` | Image digest to use, takes precedence over `image.tag` |
| image.name | string | `"observability-pipelines-worker"` | Image name to use (relative to `image.repository`) |
| image.pullPolicy | string | `"IfNotPresent"` | Set the [pullPolicy](https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy) |
| image.pullSecrets | list | `[]` | Set the [imagePullSecrets](https://kubernetes.io/docs/concepts/containers/images/#specifying-imagepullsecrets-on-a-pod) |
| image.repository | string | `"gcr.io/datadoghq"` | Image repository to use |
| image.tag | string | `"0.1.0"` | Image tag to use |
| ingress.annotations | object | `{}` | Set annotations on the Ingress |
| ingress.className | string | `""` | Specify the [ingressClassName](https://kubernetes.io/blog/2020/04/02/improvements-to-the-ingress-api-in-kubernetes-1.18/#specifying-the-class-of-an-ingress), requires Kubernetes >= 1.18 |
| ingress.enabled | bool | `false` | If **true**, create and use an Ingress resource |
| ingress.hosts | list | `[]` | Configure the hosts and paths for the Ingress |
| ingress.tls | list | `[]` | Configure TLS for the Ingress |
| initContainers | list | `[]` | Init Containers to be added |
| lifecycle | object | `{}` | Set lifecycle hooks for containers |
| livenessProbe | object | `{}` | Override default liveness probe settings. If customConfig is used, requires customConfig.api.enabled to be set to true. |
| nameOverride | string | `""` | Override the name of app |
| nodeSelector | object | `{}` | Configure [nodeSelector](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#nodeselector) |
| persistence.accessModes | list | `["ReadWriteOnce"]` | Specifies the accessModes for PersistentVolumeClaims |
| persistence.enabled | bool | `false` | If **true**, create and use PersistentVolumeClaims |
| persistence.existingClaim | string | `""` | Name of an existing PersistentVolumeClaim to use |
| persistence.finalizers | list | `["kubernetes.io/pvc-protection"]` | Specifies the finalizers of PersistentVolumeClaims |
| persistence.selectors | object | `{}` | Specifies the selectors for PersistentVolumeClaims |
| persistence.size | string | `"10Gi"` | Specifies the size of PersistentVolumeClaims |
| podAnnotations | object | `{}` | Set annotations on Pods. |
| podDisruptionBudget.enabled | bool | `false` | Enable a [PodDisruptionBudget](https://kubernetes.io/docs/tasks/run-application/configure-pdb/) |
| podDisruptionBudget.maxUnavailable | int | `nil` | The number of Pods that can be unavailable after an eviction |
| podDisruptionBudget.minAvailable | int | `1` | The number of Pods that must still be available after an eviction |
| podHostNetwork | bool | `false` | Enable hostNetwork on Pods. |
| podLabels | object | `{"vector.dev/exclude":"true"}` | Set labels on Pods. |
| podManagementPolicy | string | `"OrderedReady"` | Specify the [podManagementPolicy](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#pod-management-policies) for the StatefulSet |
| podPriorityClassName | string | `""` | Set the [priorityClassName](https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/#priorityclass) |
| podSecurityContext | object | `{}` | Allows you to overwrite the default [PodSecurityContext](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/) |
| readinessProbe | object | `{}` | Override default readiness probe settings. If customConfig is used, requires customConfig.api.enabled to be set to true. |
| replicas | int | `1` | Specify the number of Pods to create |
| resources | object | `{}` | Set resource requests and limits |
| rollWorkload | bool | `true` | Add a checksum of the generated ConfigMap to Pod annotations |
| secrets.generic | object | `{}` | Each Key/Value will be added to the Secret's data key, each value should be raw and NOT base64 encoded. Any secrets can be provided here. It's commonly used for credentials and other access related values. **NOTE: Don't commit unencrypted secrets to git!** |
| securityContext | object | `{}` | Specify securityContext for containers |
| service.annotations | object | `{}` | Set annotations on the Service |
| service.enabled | bool | `true` | If **true**, create and provide a Service resource |
| service.externalTrafficPolicy | string | `""` | Specify the [externalTrafficPolicy](https://kubernetes.io/docs/tasks/access-application-cluster/create-external-load-balancer/#preserving-the-client-source-ip) |
| service.ipFamilies | list | `[]` | Configure [IPv4/IPv6 dual-stack](https://kubernetes.io/docs/concepts/services-networking/dual-stack/) |
| service.ipFamilyPolicy | string | `""` | Configure [IPv4/IPv6 dual-stack](https://kubernetes.io/docs/concepts/services-networking/dual-stack/) |
| service.loadBalancerIP | string | `""` | Specify the [loadBalancerIP](https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer) |
| service.ports | list | `[]` | Manually set the Service ports, overriding automated generation of Service ports |
| service.topologyKeys | list | `[]` | Specify the [topologyKeys](https://kubernetes.io/docs/concepts/services-networking/service-topology/#using-service-topology) |
| service.type | string | `"ClusterIP"` | Set the type for the Service |
| serviceAccount.annotations | object | `{}` | Annotations to add to the ServiceAccount |
| serviceAccount.automountToken | bool | `true` | Automount API credentials for the ServiceAccount |
| serviceAccount.create | bool | `true` | If true, create a ServiceAccount |
| serviceAccount.name | string | `nil` | The name of the ServiceAccount to use. If not set and `serviceAccount.create` is **true**, a name is generated using the fullname template. |
| serviceHeadless.enabled | bool | `true` | If **true**, create and provide a Headless Service resource |
| terminationGracePeriodSeconds | int | `60` | Override terminationGracePeriodSeconds |
| tolerations | list | `[]` | Configure [taints and tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/) |
| topologySpreadConstraints | list | `[]` | Configure [topology spread constraints](https://kubernetes.io/docs/concepts/scheduling-eviction/topology-spread-constraints/) |
| updateStrategy | object | `{}` | Customize the [updateStrategy](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/stateful-set-v1/#StatefulSetSpec) |
