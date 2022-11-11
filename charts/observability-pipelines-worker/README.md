# Observability Pipelines Worker

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.24.2-distroless-libc](https://img.shields.io/badge/AppVersion-0.24.2--distroless--libc-informational?style=flat-square)

## How to use Datadog Helm repository

You need to add this repository to your Helm repositories:

```
helm repo add datadog https://helm.datadoghq.com
helm repo update
```

## Requirements

Kubernetes: `>=1.15.0-0`

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
| autoscaling.enabled | bool | `false` | Create a [HorizontalPodAutoscaler](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/) for Vector. Valid for the "Aggregator" and "Stateless-Aggregator" roles. |
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
| haproxy.affinity | object | `{}` | Configure [affinity](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#affinity-and-anti-affinity) rules for HAProxy Pods. |
| haproxy.autoscaling.customMetric | object | `{}` | Target a custom metric for autoscaling. |
| haproxy.autoscaling.enabled | bool | `false` | Create a HorizontalPodAutoscaler for HAProxy. |
| haproxy.autoscaling.maxReplicas | int | `10` | Maximum replicas for HAProxy's HPA. |
| haproxy.autoscaling.minReplicas | int | `1` | Minimum replicas for HAProxy's HPA. |
| haproxy.autoscaling.targetCPUUtilizationPercentage | int | `80` | Target CPU utilization for HAProxy's HPA. |
| haproxy.autoscaling.targetMemoryUtilizationPercentage | int | `nil` | Target memory utilization for HAProxy's HPA. |
| haproxy.containerPorts | list | `[]` | Manually define HAProxy's containerPorts, overrides automated generation of containerPorts. |
| haproxy.customConfig | string | `""` | Override HAProxy's default configs, if used **all** options need to be specified. This parameter supports using Helm templates to insert values dynamically. By default, this chart will parse Vector's configuration from customConfig to generate HAProxy's config, which can be overwritten with haproxy.customConfig. |
| haproxy.enabled | bool | `false` | If true, create a HAProxy load balancer. |
| haproxy.existingConfigMap | string | `""` | Use this existing ConfigMap for HAProxy's configuration instead of creating a new one. Additionally, haproxy.containerPorts and haproxy.service.ports should be specified based on your supplied configuration. If set, this parameter takes precedence over customConfig and the chart's default configs. |
| haproxy.extraContainers | list | `[]` | Extra Containers to be added to the HAProxy Pods. |
| haproxy.extraVolumeMounts | list | `[]` | Additional Volume to mount into HAProxy Containers. |
| haproxy.extraVolumes | list | `[]` | Additional Volumes to use with HAProxy Pods. |
| haproxy.image.pullPolicy | string | `"IfNotPresent"` | HAProxy image pullPolicy. |
| haproxy.image.pullSecrets | list | `[]` | The [imagePullSecrets](https://kubernetes.io/docs/concepts/containers/images/#specifying-imagepullsecrets-on-a-pod) to reference for the HAProxy Pods. |
| haproxy.image.repository | string | `"haproxytech/haproxy-alpine"` | Override default registry and name for HAProxy. |
| haproxy.image.tag | string | `"2.4.17"` | The tag to use for HAProxy's image. |
| haproxy.initContainers | list | `[]` | Init Containers to be added to the HAProxy Pods. |
| haproxy.livenessProbe | object | `{"tcpSocket":{"port":1024}}` | Override default HAProxy liveness probe settings. |
| haproxy.nodeSelector | object | `{}` | Configure a [nodeSelector](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#nodeselector) for HAProxy Pods |
| haproxy.podAnnotations | object | `{}` | Set annotations on HAProxy Pods. |
| haproxy.podLabels | object | `{}` | Set labels on HAProxy Pods. |
| haproxy.podPriorityClassName | string | `""` | Set the priorityClassName on HAProxy Pods. |
| haproxy.podSecurityContext | object | `{}` | Allows you to overwrite the default PodSecurityContext for HAProxy. |
| haproxy.readinessProbe | object | `{"tcpSocket":{"port":1024}}` | Override default HAProxy readiness probe settings. |
| haproxy.replicas | int | `1` | Set the number of HAProxy Pods to create. |
| haproxy.resources | object | `{}` | Set HAProxy resource requests and limits. |
| haproxy.rollWorkload | bool | `true` | Add a checksum of the generated ConfigMap to the HAProxy Deployment. |
| haproxy.securityContext | object | `{}` | Specify securityContext on HAProxy containers. |
| haproxy.service.annotations | object | `{}` | Set annotations on HAProxy's Service. |
| haproxy.service.externalTrafficPolicy | string | `""` | Specify the [externalTrafficPolicy](https://kubernetes.io/docs/tasks/access-application-cluster/create-external-load-balancer/#preserving-the-client-source-ip). |
| haproxy.service.ipFamilies | list | `[]` | Configure [IPv4/IPv6 dual-stack](https://kubernetes.io/docs/concepts/services-networking/dual-stack/). |
| haproxy.service.ipFamilyPolicy | string | `""` | Configure [IPv4/IPv6 dual-stack](https://kubernetes.io/docs/concepts/services-networking/dual-stack/). |
| haproxy.service.loadBalancerIP | string | `""` | Specify the [loadBalancerIP](https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer). |
| haproxy.service.ports | list | `[]` | Manually set HAPRoxy's Service ports, overrides automated generation of Service ports. |
| haproxy.service.topologyKeys | list | `[]` | Specify the [topologyKeys](https://kubernetes.io/docs/concepts/services-networking/service-topology/#using-service-topology) field on HAProxy's Service spec. |
| haproxy.service.type | string | `"ClusterIP"` | Set type of HAProxy's Service. |
| haproxy.serviceAccount.annotations | object | `{}` | Annotations to add to the HAProxy ServiceAccount. |
| haproxy.serviceAccount.automountToken | bool | `true` | Automount API credentials for the HAProxy ServiceAccount. |
| haproxy.serviceAccount.create | bool | `true` | If true, create a HAProxy ServiceAccount. |
| haproxy.serviceAccount.name | string | `nil` | The name of the HAProxy ServiceAccount to use. If not set and create is true, a name is generated using the fullname template. |
| haproxy.strategy | object | `{}` | Customize the [strategy](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/deployment-v1/) used to replace HAProxy Pods. |
| haproxy.terminationGracePeriodSeconds | int | `60` | Override HAProxy's terminationGracePeriodSeconds. |
| haproxy.tolerations | list | `[]` | Configure HAProxy Pods to be scheduled on [tainted](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/) nodes. |
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
| persistence.accessModes | list | `["ReadWriteOnce"]` | Specifies the accessModes for PersistentVolumeClaims. Valid for the "Aggregator" role. |
| persistence.enabled | bool | `false` | If true, create and use PersistentVolumeClaims. |
| persistence.existingClaim | string | `""` | Name of an existing PersistentVolumeClaim to use. Valid for the "Aggregator" role. |
| persistence.finalizers | list | `["kubernetes.io/pvc-protection"]` | Specifies the finalizers of PersistentVolumeClaims. Valid for the "Aggregator" role. |
| persistence.hostPath.path | string | `"/var/lib/vector"` | Override path used for hostPath persistence. Valid for the "Agent" role, persistence is always used for the "Agent" role. |
| persistence.selectors | object | `{}` | Specifies the selectors for PersistentVolumeClaims. Valid for the "Aggregator" role. |
| persistence.size | string | `"10Gi"` | Specifies the size of PersistentVolumeClaims. Valid for the "Aggregator" role. |
| podAnnotations | object | `{}` | Set annotations on Vector Pods. |
| podDisruptionBudget.enabled | bool | `false` | Enable a [PodDisruptionBudget](https://kubernetes.io/docs/tasks/run-application/configure-pdb/) for Vector. |
| podDisruptionBudget.maxUnavailable | int | `nil` | The number of Pods that can be unavailable after an eviction. |
| podDisruptionBudget.minAvailable | int | `1` | The number of Pods that must still be available after an eviction. |
| podHostNetwork | bool | `false` | Configure hostNetwork on Vector Pods. |
| podLabels | object | `{"vector.dev/exclude":"true"}` | Set labels on Vector Pods. |
| podManagementPolicy | string | `"OrderedReady"` | Specify the [podManagementPolicy](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#pod-management-policies) for the StatefulSet. Valid for the "Aggregator" role. |
| podMonitor.additionalLabels | object | `{}` | Adds additional labels to the PodMonitor. |
| podMonitor.enabled | bool | `false` | If true, create a PodMonitor for Vector. |
| podMonitor.honorLabels | bool | `false` | If true, honor_labels is set to true in the [scrape config](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#scrape_config). |
| podMonitor.honorTimestamps | bool | `true` | If true, honor_timestamps is set to true in the [scrape config](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#scrape_config). |
| podMonitor.jobLabel | string | `"app.kubernetes.io/name"` | Override the label to retrieve the job name from. |
| podMonitor.metricRelabelings | list | `[]` | [MetricRelabelConfigs](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#metric_relabel_configs) to apply to samples before ingestion. |
| podMonitor.path | string | `"/metrics"` | Override the path to scrape. |
| podMonitor.port | string | `"prom-exporter"` | Override the port to scrape. |
| podMonitor.relabelings | list | `[]` | [RelabelConfigs](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config) to apply to samples before scraping. |
| podPriorityClassName | string | `""` | Set the [priorityClassName](https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/#priorityclass) on Vector Pods. |
| podSecurityContext | object | `{}` | Allows you to overwrite the default [PodSecurityContext](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/) for Vector Pods. |
| psp.create | bool | `false` | If true, create a [PodSecurityPolicy](https://kubernetes.io/docs/concepts/security/pod-security-policy/) resource. PodSecurityPolicy is deprecated as of Kubernetes v1.21, and will be removed in v1.25. Intended for use with the "Agent" role. |
| rbac.create | bool | `true` | If true, create and use RBAC resources. Only valid for the "Agent" role. |
| readinessProbe | object | `{}` | Override default readiness probe settings. If customConfig is used, requires customConfig.api.enabled to be set to true. |
| replicas | int | `1` | Specify the number of Pods to create. Valid for the "Aggregator" and "Stateless-Aggregator" roles. |
| resources | object | `{}` | Set Vector resource requests and limits. |
| role | string | `"Aggregator"` | [Role](https://vector.dev/docs/setup/deployment/roles/) for this Vector instance, valid options are: "Agent", "Aggregator", and "Stateless-Aggregator". |
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
| topologySpreadConstraints | list | `[]` | Configure [topology spread constraints](https://kubernetes.io/docs/concepts/scheduling-eviction/topology-spread-constraints/) for Vector Pods. Valid for the "Aggregator" and "Stateless-Aggregator" roles. |
| updateStrategy | object | `{}` | Customize the updateStrategy used to replace Vector Pods, this is also used for the DeploymentStrategy for the "Stateless-Aggregators". Valid options depend on the chosen role. |
