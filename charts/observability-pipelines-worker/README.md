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
| affinity | object | `{}` | Configure [affinity](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#affinity-and-anti-affinity) rules for Vector Pods. |
| args | list | `["--config-dir","/etc/vector/"]` | Override Vector's default arguments. |
| autoscaling.behavior | object | `{}` | Configure separate scale-up and scale-down behaviors. |
| autoscaling.customMetric | object | `{}` | Target a custom metric for autoscaling. |
| autoscaling.enabled | bool | `false` | Create a [HorizontalPodAutoscaler](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/) for Vector. |
| autoscaling.maxReplicas | int | `10` | Maximum replicas for Vector's HPA. |
| autoscaling.minReplicas | int | `1` | Minimum replicas for Vector's HPA. |
| autoscaling.targetCPUUtilizationPercentage | int | `80` | Target CPU utilization for Vector's HPA. |
| autoscaling.targetMemoryUtilizationPercentage | int | `nil` | Target memory utilization for Vector's HPA. |
| command | list | `[]` | Override Vector's default command. |
| commonLabels | object | `{}` | Add additional labels to all created resources. |
| containerPorts | list | `[]` | Manually define Vector's containerPorts, overriding automated generation of containerPorts. |
| customConfig | object | `{}` | Override Vector's default configs, if used **all** options need to be specified. This section supports using helm templates to populate dynamic values. See Vector's [configuration documentation](https://vector.dev/docs/reference/configuration/) for all options. |
| dataDir | string | `""` | Specify the path for Vector's data, only used when existingConfigMaps are used. |
| dnsConfig | object | `{}` | Specify the [dnsConfig](https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#pod-dns-config) options for Vector Pods. |
| dnsPolicy | string | `"ClusterFirst"` | Specify the [dnsPolicy](https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#pod-s-dns-policy) for Vector Pods. |
| env | list | `[]` | Set environment variables for Vector containers. |
| envFrom | list | `[]` | Define environment variables from Secrets or ConfigMaps. |
| existingConfigMaps | list | `[]` | List of existing ConfigMaps for Vector's configuration instead of creating a new one. Requires dataDir to be set. Additionally, containerPorts, service.ports, and serviceHeadless.ports should be specified based on your supplied configuration. If set, this parameter takes precedence over customConfig and the chart's default configs. |
| extraContainers | list | `[]` | Extra Containers to be added to the Vector Pods. |
| extraVolumeMounts | list | `[]` | Additional Volume to mount into Vector Containers. |
| extraVolumes | list | `[]` | Additional Volumes to use with Vector Pods. |
| fullnameOverride | string | `""` | Override the full name of resources. |
| image.pullPolicy | string | `"IfNotPresent"` | The [pullPolicy](https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy) for Vector's image. |
| image.pullSecrets | list | `[]` | The [imagePullSecrets](https://kubernetes.io/docs/concepts/containers/images/#specifying-imagepullsecrets-on-a-pod) to reference for the Vector Pods. |
| image.repository | string | `"timberio/vector"` | Override default registry and name for Vector's image. |
| image.sha | string | `""` | The SHA to use for Vector's image. |
| image.tag | string | Derived from the Chart's appVersion. | The tag to use for Vector's image. |
| ingress.annotations | object | `{}` | Set annotations on the Ingress. |
| ingress.className | string | `""` | Specify the [ingressClassName](https://kubernetes.io/blog/2020/04/02/improvements-to-the-ingress-api-in-kubernetes-1.18/#specifying-the-class-of-an-ingress), requires Kubernetes >= 1.18 |
| ingress.enabled | bool | `false` | If true, create and use an Ingress resource. |
| ingress.hosts | list | `[]` | Configure the hosts and paths for the Ingress. |
| ingress.tls | list | `[]` | Configure TLS for the Ingress. |
| initContainers | list | `[]` | Init Containers to be added to the Vector Pods. |
| lifecycle | object | `{}` | Set lifecycle hooks for Vector containers. |
| livenessProbe | object | `{}` | Override default liveness probe settings. If customConfig is used, requires customConfig.api.enabled to be set to true. |
| nameOverride | string | `""` | Override the name of resources. |
| nodeSelector | object | `{}` | Configure a [nodeSelector](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#nodeselector) for Vector Pods. |
| persistence.accessModes | list | `["ReadWriteOnce"]` | Specifies the accessModes for PersistentVolumeClaims. |
| persistence.enabled | bool | `false` | If true, create and use PersistentVolumeClaims. |
| persistence.existingClaim | string | `""` | Name of an existing PersistentVolumeClaim to use. |
| persistence.finalizers | list | `["kubernetes.io/pvc-protection"]` | Specifies the finalizers of PersistentVolumeClaims. |
| persistence.selectors | object | `{}` | Specifies the selectors for PersistentVolumeClaims. |
| persistence.size | string | `"10Gi"` | Specifies the size of PersistentVolumeClaims. |
| podAnnotations | object | `{}` | Set annotations on Vector Pods. |
| podDisruptionBudget.enabled | bool | `false` | Enable a [PodDisruptionBudget](https://kubernetes.io/docs/tasks/run-application/configure-pdb/) for Vector. |
| podDisruptionBudget.maxUnavailable | int | `nil` | The number of Pods that can be unavailable after an eviction. |
| podDisruptionBudget.minAvailable | int | `1` | The number of Pods that must still be available after an eviction. |
| podHostNetwork | bool | `false` | Configure hostNetwork on Vector Pods. |
| podLabels | object | `{"vector.dev/exclude":"true"}` | Set labels on Vector Pods. |
| podManagementPolicy | string | `"OrderedReady"` | Specify the [podManagementPolicy](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#pod-management-policies) for the StatefulSet. |
| podPriorityClassName | string | `""` | Set the [priorityClassName](https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/#priorityclass) on Vector Pods. |
| podSecurityContext | object | `{}` | Allows you to overwrite the default [PodSecurityContext](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/) for Vector Pods. |
| readinessProbe | object | `{}` | Override default readiness probe settings. If customConfig is used, requires customConfig.api.enabled to be set to true. |
| replicas | int | `1` | Specify the number of Pods to create. |
| resources | object | `{}` | Set Vector resource requests and limits. |
| rollWorkload | bool | `true` | Add a checksum of the generated ConfigMap to workload annotations. |
| secrets.generic | object | `{}` | Each Key/Value will be added to the Secret's data key, each value should be raw and NOT base64 encoded. Any secrets can be provided here. It's commonly used for credentials and other access related values. **NOTE: Don't commit unencrypted secrets to git!** |
| securityContext | object | `{}` | Specify securityContext on Vector containers. |
| service.annotations | object | `{}` | Set annotations on Vector's Service. |
| service.enabled | bool | `true` | If true, create and provide a Service resource for Vector. |
| service.externalTrafficPolicy | string | `""` | Specify the [externalTrafficPolicy](https://kubernetes.io/docs/tasks/access-application-cluster/create-external-load-balancer/#preserving-the-client-source-ip). |
| service.ipFamilies | list | `[]` | Configure [IPv4/IPv6 dual-stack](https://kubernetes.io/docs/concepts/services-networking/dual-stack/). |
| service.ipFamilyPolicy | string | `""` | Configure [IPv4/IPv6 dual-stack](https://kubernetes.io/docs/concepts/services-networking/dual-stack/). |
| service.loadBalancerIP | string | `""` | Specify the [loadBalancerIP](https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer). |
| service.ports | list | `[]` | Manually set the Service ports, overriding automated generation of Service ports. |
| service.topologyKeys | list | `[]` | Specify the [topologyKeys](https://kubernetes.io/docs/concepts/services-networking/service-topology/#using-service-topology) field on Vector's Service. |
| service.type | string | `"ClusterIP"` | Set the type for Vector's Service. |
| serviceAccount.annotations | object | `{}` | Annotations to add to Vector's ServiceAccount. |
| serviceAccount.automountToken | bool | `true` | Automount API credentials for Vector's ServiceAccount. |
| serviceAccount.create | bool | `true` | If true, create a ServiceAccount for Vector. |
| serviceAccount.name | string | `nil` | The name of the ServiceAccount to use. If not set and serviceAccount.create is true, a name is generated using the fullname template. |
| serviceHeadless.enabled | bool | `true` | If true, create and provide a Headless Service resource for Vector. |
| terminationGracePeriodSeconds | int | `60` | Override Vector's terminationGracePeriodSeconds. |
| tolerations | list | `[]` | Configure Vector Pods to be scheduled on [tainted](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/) nodes. |
| topologySpreadConstraints | list | `[]` | Configure [topology spread constraints](https://kubernetes.io/docs/concepts/scheduling-eviction/topology-spread-constraints/) for Vector Pods. |
| updateStrategy | object | `{}` | Customize the [updateStrategy](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/stateful-set-v1/#StatefulSetSpec) used to replace Vector Pods. |
