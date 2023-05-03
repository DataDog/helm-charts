# Observability Pipelines Worker

![Version: 1.2.0-rc.0](https://img.shields.io/badge/Version-1.2.0--rc.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 1](https://img.shields.io/badge/AppVersion-1-informational?style=flat-square)

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
  --set datadog.apiKey=<DD_API_KEY> \
  --set datadog.pipelineId=<DD_OP_PIPELINE_ID> \
  datadog/observability-pipelines-worker
```

By default, this chart creates secrets for your Observability Pipelines API key. However, you can use
manually created Secrets by setting the `datadog.apiKeyExistingSecret` values
(see [Creating a Secret](#create-and-provide-a-secret-that-contains-your-datadog-api-key), below).

**Note:** When creating the Secret(s), be sure to name the key fields `api-key`.

After a few minutes, you should see your new pipeline active in Datadog.

**Note:** You can set your [Datadog site](https://docs.datadoghq.com/getting_started/site) using the `datadog.site` option.

```bash
helm install --name <RELEASE_NAME> \
    --set datadog.apiKey=<DD_API_KEY> \
    --set datadog.pipelineId=<DD_OP_PIPELINE_ID> \
    --set datadog.site=<DATADOG_SITE> \
    datadog/observability-pipelines-worker
```

#### Create and provide a Secret that contains your Datadog API Key

To create a Secret that contains your Datadog API key, replace the `<DATADOG_API_KEY>` below with the API key for your
organization. This Secret is used in the manifest to deploy the Observability Pipelines Worker.

```bash
export DATADOG_SECRET_NAME=datadog-secrets
kubectl create secret generic $DATADOG_SECRET_NAME \
    --from-literal api-key="<DD_API_KEY>" \
```

**Note**: This creates a Secret in the **default** Namespace. If you are using a custom Namespace, update the Namespace
flag of the command before running it.

Now, the installation command contains a reference to the Secret.

```bash
helm install --name <RELEASE_NAME> \
  --set datadog.apiKeyExistingSecret=$DATADOG_SECRET_NAME \
  datadog/observability-pipelines-worker
```

### Uninstalling the chart

To uninstall the `<RELEASE_NAME>` release:

```bash
helm delete <RELEASE_NAME>
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Configure [affinity and anti-affinity](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#affinity-and-anti-affinity). |
| args | list | `["run"]` | Override default image arguments. |
| autoscaling.behavior | object | `{}` | Configure separate scale-up and scale-down behaviors. |
| autoscaling.enabled | bool | `false` | If **true**, create a [HorizontalPodAutoscaler](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/). |
| autoscaling.maxReplicas | int | `10` | Specify the maximum number of replicas. |
| autoscaling.minReplicas | int | `1` | Specify the minimum number of replicas. |
| autoscaling.targetCPUUtilizationPercentage | int | `80` | Specify the target CPU utilization. |
| autoscaling.targetMemoryUtilizationPercentage | int | `nil` | Specify the target memory utilization. |
| command | list | `[]` | Override default image command. |
| commonLabels | object | `{}` | Labels to apply to all resources. |
| containerPorts | list | `[]` | Manually define ContainerPort array, overriding automated generation of ContainerPorts. |
| datadog.apiKey | string | `nil` | Specify your Datadog API key. |
| datadog.apiKeyExistingSecret | string | `""` | Specify a preexisting Secret that has your API key instead of creating a new one. The value must be stored under the `api-key`. |
| datadog.configKey | string | `nil` | Specify your Datadog Configuration key. DEPRECATED. Use `datadog.pipelineId` instead. |
| datadog.configKeyExistingSecret | string | `""` | Specify a preexisting Secret that has your configuration key instead of creating a new one. The value must be stored under the `config-key`. DEPRECATED. Use `datadog.pipelineId` instead. |
| datadog.dataDir | string | `"/var/lib/observability-pipelines-worker"` | The data directory for OPW to store runtime data in. |
| datadog.pipelineId | string | `nil` | Specify your Datadog Observability Pipelines pipeline ID |
| datadog.remoteConfigurationEnabled | bool | `false` | Whether to allow remote configuration of the worker from Datadog. |
| datadog.site | string | `"datadoghq.com"` | The [site](https://docs.datadoghq.com/getting_started/site/) of the Datadog intake to send data to. |
| dnsConfig | object | `{}` | Specify the [dnsConfig](https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#pod-dns-config). |
| dnsPolicy | string | `"ClusterFirst"` | Specify the [dnsPolicy](https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#pod-s-dns-policy). |
| env | list | `[]` | Define environment variables. |
| envFrom | list | `[]` | Define environment variables from ConfigMap or Secret data. |
| extraContainers | list | `[]` | Specify extra Containers to be added. |
| extraVolumeMounts | list | `[]` | Specify Additional VolumeMounts to use. |
| extraVolumes | list | `[]` | Specify additional Volumes to use. |
| fullnameOverride | string | `""` | Override the fully qualified app name. |
| image.digest | string | `nil` | Specify the image digest to use; takes precedence over `image.tag`. |
| image.name | string | `"observability-pipelines-worker"` | Specify the image name to use (relative to `image.repository`). |
| image.pullPolicy | string | `"IfNotPresent"` | Specify the [pullPolicy](https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy). |
| image.pullSecrets | list | `[]` | Specify the [imagePullSecrets](https://kubernetes.io/docs/concepts/containers/images/#specifying-imagepullsecrets-on-a-pod). |
| image.repository | string | `"gcr.io/datadoghq"` | Specify the image repository to use. |
| image.tag | string | `"nightly-2023-05-03"` | Specify the image tag to use. |
| ingress.annotations | object | `{}` | Specify annotations for the Ingress. |
| ingress.className | string | `""` | Specify the [ingressClassName](https://kubernetes.io/blog/2020/04/02/improvements-to-the-ingress-api-in-kubernetes-1.18/#specifying-the-class-of-an-ingress), requires Kubernetes >= 1.18. |
| ingress.enabled | bool | `false` | If **true**, create an Ingress resource. |
| ingress.hosts | list | `[]` | Configure the hosts and paths for the Ingress. |
| ingress.tls | list | `[]` | Configure TLS for the Ingress. |
| initContainers | list | `[]` | Specify initContainers to be added. |
| lifecycle | object | `{}` | Specify lifecycle hooks for Containers. |
| livenessProbe | object | `{}` | Specify the livenessProbe [configuration](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#configure-probes). |
| nameOverride | string | `""` | Override the name of the app. |
| nodeSelector | object | `{}` | Configure [nodeSelector](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#nodeselector). |
| persistence.accessModes | list | `["ReadWriteOnce"]` | Specify the accessModes for PersistentVolumeClaims. |
| persistence.enabled | bool | `false` | If **true**, create and use PersistentVolumeClaims. |
| persistence.existingClaim | string | `""` | Name of an existing PersistentVolumeClaim to use. |
| persistence.finalizers | list | `["kubernetes.io/pvc-protection"]` | Specify the finalizers of PersistentVolumeClaims. |
| persistence.selector | object | `{}` | Specify the selectors for PersistentVolumeClaims. |
| persistence.size | string | `"10Gi"` | Specify the size of PersistentVolumeClaims. |
| persistence.storageClassName | string | `nil` | Specify the storageClassName for PersistentVolumeClaims. |
| pipelineConfig | object | `{}` | This section supports using Helm templates to populate dynamic values. See Observability Pipelines' [configuration documentation](https://docs.datadoghq.com/observability_pipelines/reference/) for all options. |
| podAnnotations | object | `{}` | Set annotations on Pods. |
| podDisruptionBudget.enabled | bool | `false` | If **true**, create a [PodDisruptionBudget](https://kubernetes.io/docs/tasks/run-application/configure-pdb/). |
| podDisruptionBudget.maxUnavailable | int | `nil` | Specify the number of Pods that can be unavailable after an eviction. |
| podDisruptionBudget.minAvailable | int | `1` | Specify the number of Pods that must still be available after an eviction. |
| podHostNetwork | bool | `false` | Enable the hostNetwork option on Pods. |
| podLabels | object | `{}` | Set labels on Pods. |
| podManagementPolicy | string | `"OrderedReady"` | Specify the [podManagementPolicy](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#pod-management-policies). |
| podPriorityClassName | string | `""` | Set the [priorityClassName](https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/#priorityclass). |
| podSecurityContext | object | `{}` | Allows you to overwrite the default [PodSecurityContext](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/). |
| readinessProbe | object | `{}` | Specify the readinessProbe [configuration](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#configure-probes). |
| replicas | int | `1` | Specify the number of replicas to create. |
| resources | object | `{}` | Specify resource requests and limits. |
| securityContext | object | `{}` | Specify securityContext for Containers. |
| service.annotations | object | `{}` | Specify annotations for the Service. |
| service.enabled | bool | `true` | If **true**, create a Service resource. |
| service.externalTrafficPolicy | string | `""` | Specify the [externalTrafficPolicy](https://kubernetes.io/docs/tasks/access-application-cluster/create-external-load-balancer/#preserving-the-client-source-ip). |
| service.ipFamilies | list | `[]` | Configure [IPv4/IPv6 dual-stack](https://kubernetes.io/docs/concepts/services-networking/dual-stack/). |
| service.ipFamilyPolicy | string | `""` | Configure [IPv4/IPv6 dual-stack](https://kubernetes.io/docs/concepts/services-networking/dual-stack/). |
| service.loadBalancerIP | string | `""` | Specify the [loadBalancerIP](https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer). |
| service.ports | array | `nil` | Manually set the ServicePort array, overriding automated generation of ServicePorts. |
| service.topologyKeys | array | `nil` | Specify the [topologyKeys](https://kubernetes.io/docs/concepts/services-networking/service-topology/#using-service-topology). |
| service.type | string | `"ClusterIP"` | Specify the type for the Service. |
| serviceAccount.annotations | object | `{}` | Annotations to add to the ServiceAccount, if `serviceAccount.create` is **true**. |
| serviceAccount.create | bool | `true` | If **true**, create a ServiceAccount. |
| serviceAccount.name | string | `"default"` | Specify a preexisting ServiceAccount to use if `serviceAccount.create` is **false**. |
| serviceHeadless.enabled | bool | `true` | If **true**, create a "headless" Service resource. |
| terminationGracePeriodSeconds | int | `60` | Override terminationGracePeriodSeconds. |
| tolerations | list | `[]` | Configure [taints and tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/). |
| topologySpreadConstraints | list | `[]` | Configure [topology spread constraints](https://kubernetes.io/docs/concepts/scheduling-eviction/topology-spread-constraints/). |
| updateStrategy | object | `{}` | Customize the [updateStrategy](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/stateful-set-v1/#StatefulSetSpec). |
