# Datadog Operator

![Version: 1.0.2](https://img.shields.io/badge/Version-1.0.2-informational?style=flat-square) ![AppVersion: 1.0.0](https://img.shields.io/badge/AppVersion-1.0.0-informational?style=flat-square)

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Allows to specify affinity for Datadog Operator PODs |
| apiKey | string | `nil` | Your Datadog API key |
| apiKeyExistingSecret | string | `nil` | Use existing Secret which stores API key instead of creating a new one |
| appKey | string | `nil` | Your Datadog APP key |
| appKeyExistingSecret | string | `nil` | Use existing Secret which stores APP key instead of creating a new one |
| collectOperatorMetrics | bool | `true` | Configures an openmetrics check to collect operator metrics |
| containerSecurityContext | object | `{}` | A security context defines privileges and access control settings for a container. |
| datadogAgent.enabled | bool | `true` | Enables Datadog Agetn controller |
| datadogCRDs.crds.datadogAgents | bool | `true` |  |
| datadogCRDs.crds.datadogMetrics | bool | `true` |  |
| datadogCRDs.crds.datadogMonitors | bool | `true` |  |
| datadogCRDs.migration.datadogAgents.conversionWebhook.enabled | bool | `false` |  |
| datadogCRDs.migration.datadogAgents.conversionWebhook.name | string | `"datadog-operator-webhook-service"` |  |
| datadogCRDs.migration.datadogAgents.conversionWebhook.namespace | string | `"default"` |  |
| datadogCRDs.migration.datadogAgents.useCertManager | bool | `false` |  |
| datadogCRDs.migration.datadogAgents.version | string | `"v2alpha1"` |  |
| datadogMonitor.enabled | bool | `false` | Enables the Datadog Monitor controller |
| dd_url | string | `nil` | The host of the Datadog intake server to send Agent data to, only set this option if you need the Agent to send data to a custom URL |
| env | list | `[]` | Define any environment variables to be passed to the operator. |
| fullnameOverride | string | `""` |  |
| image.pullPolicy | string | `"IfNotPresent"` | Define the pullPolicy for Datadog Operator image |
| image.repository | string | `"gcr.io/datadoghq/operator"` | Repository to use for Datadog Operator image |
| image.tag | string | `"1.0.0"` | Define the Datadog Operator version to use |
| imagePullSecrets | list | `[]` | Datadog Operator repository pullSecret (ex: specify docker registry credentials) |
| installCRDs | bool | `true` | Set to true to deploy the Datadog's CRDs |
| logLevel | string | `"info"` | Set Datadog Operator log level (debug, info, error, panic, fatal) |
| maximumGoroutines | string | `nil` | Override default gouroutines threshold for the health check failure. |
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
| serviceAccount.annotations | object | `{}` | Allows setting additional annotations for service account |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `nil` | The name of the service account to use. If not set name is generated using the fullname template |
| site | string | `nil` | The site of the Datadog intake to send data to (documentation: https://docs.datadoghq.com/getting_started/site/) |
| supportExtendedDaemonset | string | `"false"` | If true, supports using ExtendedDaemonSet CRD |
| tolerations | list | `[]` | Allows to schedule Datadog Operator on tainted nodes |
| watchNamespaces | list | `[]` | Restricts the Operator to watch its managed resources on specific namespaces |

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

## Migrating to the version 1.0 of the Datadog Operator

### Disclaimer

As part of the General Availability release of the Datadog Operator, we are offering a migration path for our early adopters to migrate to the GA version of the custom resource, `v2alpha1/DatadogAgent`.

The Datadog Operator v1.X reconciles the version `v2alpha1` of the DatadogAgent custom resource, while the v0.X reconciles `v1alpha1`.

### Requirements

If you are using the v1alpha1 with a v0.X version of the Datadog Operator and would like to upgrade, you will need to use the Conversion Webhook feature.

Start by ensuring that you have the minimum required version of the chart and it's dependencies:

```
NAME                	CHART VERSION	APP VERSION	DESCRIPTION
datadog/datadog-crds	0.6.1        	1          	Datadog Kubernetes CRDs chart
```

and for the Datadog Operator chart:

```
NAME                    	CHART VERSION	APP VERSION	DESCRIPTION
datadog/datadog-operator	1.0.0        	1.0.0      	Datadog Operator
```

Then you will need to install the cert manager if you don't have it already, add the chart:
```
helm repo add jetstack https://charts.jetstack.io
```
and then install it:
```
 helm install \
  cert-manager jetstack/cert-manager \
  --version v1.11.0 \
  --set installCRDs=true
```

### Migration

You can update with the following:

```
helm upgrade \
    datadog-operator datadog/datadog-operator \
    --set image.tag=1.0.0 \
    --set datadogCRDs.migration.datadogAgents.version=v2alpha1 \
    --set datadogCRDs.migration.datadogAgents.useCertManager=true \
    --set datadogCRDs.migration.datadogAgents.conversionWebhook.enabled=true
```

### Notes

Starting at the version 1.0.0 of the datadog-operator chart, the fields `image.tag` has a default values of `1.0.0` and `datadogCRDs.migration.datadogAgents.version` is `v2alpha1`.

We set them in the command here to illustrate the migration of going from a Datadog Operator version < 1.0.0 with a stored version of `v1alpha1` to the GA version of `1.0.0` with a stored version of `v2alpha1`.

### Implementation details

This will create a self-signed `Certificate` (using an `Issuer`) that will be used by the Certificate Manager to mutate the DatadogAgent CRD to document the `caBundle` that the API Server will use to contact the Conversion Webhook.

The Datadog Operator will be running the new reconciler for `v2alpha1` object and will also start a Conversion Webhook Server, exposed on port 9443. This server is the one the API Server will be using to convert v1alpha1 DatadogAgent into v2alpha1.

### Lifecycle

The conversionWebhook is not supposed to be an ever running process, we recommend using it to migrate your objects as a transition.

Once converted, you can store the new version of your DatadogAgent, deactivate the conversion and simply deploy v2alpha1 objects.

### Roadmap

Upon releasing the v2 version of the DatadogAgent object, we will remove v1alpha1 from the CRD as part of a major update of the charts (datadog-crds and datadog-operator).

### Troubleshooting

* I don't see v2alpha1 version of the DatadogAgent resource

The v1alpha1 and the v2alpha1 are `served` so you might need to specify which version you want to see:

```
kubectl get datadogagents.v2alpha1.datadoghq.com datadog-agent
```

* The Conversion is not working

The logs of the Datadog Operator pod should show that the conversion webhook is enabled, the server is running, the certificates are watched.

```
kubectl logs datadog-operator-XXX-YYY
[...]
{"level":"INFO","ts":"2023-02-16T16:47:07Z","logger":"controller-runtime.webhook","msg":"Registering webhook","path":"/convert"}
{"level":"INFO","ts":"2023-02-16T16:47:07Z","logger":"controller-runtime.builder","msg":"Conversion webhook enabled","GVK":"datadoghq.com/v2alpha1, Kind=DatadogAgent"}
{"level":"INFO","ts":"2023-02-16T16:47:07Z","logger":"setup","msg":"starting manager"}
{"level":"INFO","ts":"2023-02-16T16:47:07Z","logger":"controller-runtime.webhook.webhooks","msg":"Starting webhook server"}
{"level":"INFO","ts":"2023-02-16T16:47:07Z","logger":"controller-runtime.certwatcher","msg":"Updated current TLS certificate"}
{"level":"INFO","ts":"2023-02-16T16:47:07Z","logger":"controller-runtime.webhook","msg":"Serving webhook server","host":"","port":9443}
{"level":"INFO","ts":"2023-02-16T16:47:07Z","msg":"Starting server","path":"/metrics","kind":"metrics","addr":"0.0.0.0:8383"}
{"level":"INFO","ts":"2023-02-16T16:47:07Z","msg":"Starting server","kind":"health probe","addr":"0.0.0.0:8081"}
{"level":"INFO","ts":"2023-02-16T16:47:07Z","logger":"controller-runtime.certwatcher","msg":"Starting certificate watcher"}
[...]
```

* Check the service registered for the conversion for a registered Endpoint

```
kubectl describe service datadog-operator-webhook-service
[...]
Name:              datadog-operator-webhook-service
Namespace:         default
[...]
Selector:          app.kubernetes.io/instance=datadog-operator,app.kubernetes.io/name=datadog-operator
[...]
Port:              <unset>  443/TCP
TargetPort:        9443/TCP
Endpoints:         10.88.3.28:9443
```

* Verify the registered service for the conversion webhook

```
kubectl describe crd datadogagents.datadoghq.com
[...]
  Conversion:
    Strategy:  Webhook
    Webhook:
      Client Config:
        Ca Bundle:  LS0t[...]UtLS0tLQo=
        Service:
          Name:       datadog-operator-webhook-service
          Namespace:  default
          Path:       /convert
          Port:       443
      Conversion Review Versions:
        v1
```

* The CRD does not have the `caBundle`

Make sure that the CRD has the correct annotation: `cert-manager.io/inject-ca-from: default/datadog-operator-serving-cert` and check the logs of the `cert-manager-cainjector` pod.

If you do not see anything standing out, setting the log level to 5 (debug) might help:

```
kubectl edit deploy cert-manager-cainjector -n cert-manager
[...]
    spec:
      containers:
      - args:
        - --v=5
[...]
```

You should see logs such as:

```
[...]
I0217 08:11:15.582479       1 controller.go:178] cert-manager/certificate/customresourcedefinition/generic-inject-reconciler "msg"="updated object" "resource_kind"="CustomResourceDefinition" "resource_name"="datadogagents.datadoghq.com" "resource_namespace"="" "resource_version"="v1"
I0217 08:25:24.989209       1 sources.go:98] cert-manager/certificate/customresourcedefinition/generic-inject-reconciler "msg"="Extracting CA from Certificate resource" "certificate"="default/datadog-operator-serving-cert" "resource_kind"="CustomResourceDefinition" "resource_name"="datadogagents.datadoghq.com" "resource_namespace"="" "resource_version"="v1"
[...]
```
### Rollback

If you migrated to the new version of the Datadog Operator using v2alpha1 but want to rollback to the former version, we recommend:
- Scaling the Datadog Operator deployment to 0 replicas.
  ```
  kubectl scale deploy datadog-operator --replicas=0
  ```
- Upgrading the chart to have v1alpha1 stored and for the Datadog Operator to use the 0.8.X image.
  ```
  helm upgrade \
    datadog-operator datadog/datadog-operator \
    --set image.tag=0.8.4 \
    --set datadogCRDs.migration.datadogAgents.version=v1alpha1 \
    --set datadogCRDs.migration.datadogAgents.useCertManager=false \
    --set datadogCRDs.migration.datadogAgents.conversionWebhook.enabled=false
  ```
- Redeploy the previous DatadogAgent v1alpha1 object.

Note: The Daemonset of the Datadog Agents will be rolled out in the process.
