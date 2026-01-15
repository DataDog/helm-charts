# Changelog

## 2.17.0

* Update Datadog Operator chart for 1.22.0.
* Enable DatadogAgentInternal controller and CRD.
* Add ClusterRole RBAC permissions for DatadogAgentInternal. 
* Fix endpoint-config ConfigMap lookup to use exact name instead of suffix matching, preventing value concatenation when multiple Datadog releases exist in the same namespace.

## 2.17.0-dev.3

* Update Datadog Operator chart for 1.22.0-rc.3.
* Enable DatadogAgentInternal controller and CRD.
* Add ClusterRole RBAC permissions for DatadogAgentInternal. 

## 2.17.0-dev.2

* Fix endpoint-config ConfigMap lookup to use exact name instead of suffix matching, preventing value concatenation when multiple Datadog releases exist in the same namespace.

## 2.17.0-dev.1

* Update Datadog Operator chart for 1.22.0-rc.1.

## 2.16.0

* Update Datadog Operator image tag to 1.21.0.

## 2.16.0-dev.7

* Add dnsConfig option

## 2.16.0-dev.6

* Update Datadog Operator chart for 1.21.0-rc.3.

## 2.16.0-dev.5

* Update version of Datadog CRDs to 2.14.0-dev.5.

## 2.16.0-dev.4

* Use values from Datadog chart's endpoint-config configMap,  if present.

## 2.16.0-dev.3

* Update Datadog Operator image tag to 1.21.0-rc.2.

## 2.16.0-dev.2

* Same as 2.16.0-dev.1 and update version of Datadog CRDs to 2.14.0-dev.2 to pick up changes to DatadogPodAutoscaler.

## 2.16.0-dev.1

* Update Datadog Operator image tag to 1.21.0-rc.1.

## 2.15.2

* Revert of Datadog Operator image to 1.20.0 on the stable branch (2.15.1 was missing this fix) and include updated version of Datadog CRDs to 2.13.1 to pick up changes to DatadogPodAutoscaler.

## 2.15.1 (deprecated – do not use)

* This version was missing some required fixes and should not be used.
* Use **2.15.2** instead (or **2.15.0** if you cannot upgrade to 2.15.2).
* (Original change) Update version of Datadog CRDs to 2.13.1 to pick up changes to DatadogPodAutoscaler.

## 2.15.0

* Update Datadog Operator image tag to 1.20.0.

## 2.15.0-dev.3

* Update Datadog Operator image tag to 1.20.0-rc.4.

## 2.15.0-dev.2

* Update Datadog Operator image tag to 1.20.0-rc.2.

## 2.15.0-dev.1

* Update Datadog Operator image tag to 1.20.0-rc.1.

## 2.14.3

* Update Datadog Operator image tag to 1.19.1.

## 2.14.2

* Update Datadog Operator image tag to 1.19.0.

## 2.14.1

* [CASCL-610] Add require RBAC for ArgoRollout support in datadog-operator chart ([#2076](https://github.com/DataDog/helm-charts/pull/2076)).

## 2.14.0-dev.3

* Update Datadog Operator image tag to 1.19.0-rc.3.

## 2.14.0-dev.2

* Update Datadog Operator image tag to 1.19.0-rc.2.

## 2.14.0-dev.1

* Update Datadog Operator image tag to 1.19.0-rc.1.

## 2.13.1

* Add default `initialDelaySeconds: 15` to the Liveness Probe

## 2.13.0

* Update Datadog Operator chart for 1.18.0.

## 2.13.0-dev.5

* Update Datadog Operator image tag to 1.18.0-rc.4.

## 2.13.0-dev.4

* Update Datadog Operator image tag to 1.18.0-rc.3.

## 2.13.0-dev.3

* Update Datadog Operator image tag to 1.18.0-rc.2.

## 2.13.0-dev.2

* Update Datadog Operator image tag to 1.18.0-rc.1.

## 2.13.0-dev.1

* Update Datadog Operator chart for 1.18.0-rc.1.

## 2.12.1

* Update `datadog-crds` dependency to stable version (no-op change).

## 2.12.0

* Update Datadog Operator chart for 1.17.0.

## 2.12.0-dev.4

* Add option to disable service account automountServiceAccountToken. 

## 2.12.0-dev.3

* Update Datadog Operator chart for 1.17.0-rc.3.

## 2.12.0-dev.2

* Update Datadog Operator chart for 1.17.0-rc.2.

## 2.12.0-dev.1

* Update Datadog Operator chart for 1.17.0-rc.1.

## 2.11.1

* Handle Operator image tag with a digest gracefully.

## 2.11.0

* Update Datadog Operator chart for 1.16.0.

## 2.11.0-dev.3

* Document `datadogCRDs.crds.datadogAgentProfiles` option to install the DatadogAgentProfile CRD.

## 2.11.0-dev.2

* Update default image tag for Datadog Operator to `1.16.0-rc.1`.

## 2.11.0-dev.1

* Update Datadog Operator chart for 1.16.0-rc.1.

## 2.10.0

* Update Datadog Operator chart for 1.15.1

## 2.10.0-dev.2

* Update Datadog Operator chart for 1.15.0-rc.2.

## 2.10.0-dev.1

* Fix semverCompare to work with pre-release versions.

## 2.10.0-dev

* Update Datadog Operator chart for 1.15.0-rc.1.

## 2.9.2

* no-op chart bump to sync changlog with chart version.

## 2.9.0

* Update Datadog Operator version to 1.14.0.

## 2.9.0-dev

* Update Datadog Operator version to 1.14.0-rc.3.

## 2.8.0

* Update Datadog Operator version to 1.13.0.

## 2.7.0

* Update Datadog Operator version to 1.12.1.

## 2.6.0

* Update Datadog Operator version to 1.12.0.
* Add DatadogGenericResource configuration.

## 2.5.1

* Expose CRD-specific namespace watch configuration added in Operator 1.8.0 release.

## 2.5.0

* Update Datadog Operator version to 1.11.1.

## 2.4.0

* Add configuration to grant the necessary RBAC to the operator for the CWS Instrumentation Admission Controller feature in the Cluster-Agent.

## 2.3.0

* Update Datadog Operator version to 1.10.0.

## 2.2.0

* Add clusterRole.allowReadAllResources to allow viewing all resources. This is required for collecting custom resources in the Kubernetes Explorer

## 2.1.0

* Update Datadog Operator version to 1.9.0.
* Add DatadogDashboard configuration.

## 2.0.1

* Make Operator `livenessProbe` configurable.

## 2.0.0

* Update Datadog Operator version to 1.8.0.
* Drop support for DatadogAgent `v1alpha1` and conversion webhook.

## 1.8.5

* Update `datadog-crds` dependency to `1.7.2`.

## 1.8.4

* Add option to specify `deployment.annotations`.

## 1.8.3

* Add `image.doNotCheckTag` option to permit skipping operator image tag compatibility.

## 1.8.2

* Deprecate `webhookEnabled` flag for 1.7.0.

## 1.8.1

* Configure tool version.

## 1.8.0

* Update Datadog Operator version to 1.7.0.

## 1.7.1

* Add `DD_TOOL_VERSION` to operator deployment.

## 1.7.0

* Update Datadog Operator version to 1.6.0.

## 1.6.1

* Fix clusterRole when DatadogAgentProfiles are enabled.

## 1.6.0

* Update Datadog Operator version to 1.5.0.

## 1.5.2

* Add deprecation warning for `DatadogAgent` `v1alpha1` CRD version.

## 1.5.1

* Add configuration for Operator flag `introspectionEnabled`: this parameter is used to enable the Introspection. It is disabled by default.

## 1.5.0

* Update Datadog Operator version to 1.4.0.

## 1.4.2

* Migrate from `kubeval` to `kubeconform` for ci chart validation.

## 1.4.1

* Add configuration for Operator flag `datadogSLOEnabled` : this parameter is used to enable the Datadog SLO Controller. It is disabled by default.

## 1.4.0

* Update Datadog Operator version to 1.3.0.

## 1.3.0

* Add configuration to mount volumes (`volumes` and `volumeMounts`) in the container. Empty by default.

## 1.2.2

* Fix that an error occurs when specifying replicaCount using `--set`

## 1.2.1

* Minor spelling corrections in the `datadog-operator` chart.

## 1.2.0

* Update Datadog Operator version to 1.2.0.

## 1.1.2

* Add configuration for Operator flag `operatorMetricsEnabled` : this parameter can be used to disable the Operator metrics forwarder. It is enabled by default.

## 1.1.1

* Add permissions to curl `/metrics/slis` to operator cluster role.

## 1.1.0

* Update Datadog Operator version to 1.1.0.

## 1.0.8

* Minor spelling corrections in the `datadog-operator` chart.

## 1.0.7

* Fix clusterrole to include `extensions` group for `customresourcedefinitions` resource.

## 1.0.6

* Fix conversionWebhook.enabled parameter to correctly set user-configured value when enabling the conversion webhook.

## 1.0.5

* Add AP1 Site Comment in `values.yaml`.

## 1.0.4

* Update Datadog Operator version to 1.0.3.

## 1.0.3

* Add `list` and `watch` permissions of `customresourcedefinitions` for the KSM core check to collect CRD resources.

## 1.0.2

* Use `.Release.Name` for reference to conversion webhook certificate in datadog-operator deployment.yaml


## 1.0.1

* Use `.Release.Name` for conversion webhook certificate / issuer name to align with the certificate name generated in datadog-crds sub-chart

## 1.0.0

* Default image is now `1.0.0`
* Updated documentation.
* Stored Version is v2alpha1 by default:
    * If you are using a chart 0.X, refer to the [Migration Steps](https://github.com/DataDog/helm-charts/blob/main/charts/datadog-operator/README.md#migrating-to-the-version-10-of-the-datadog-operator).
* Added Failure exceptions to avoid breaking changes:
    * Added exception when using unsupported version of the DatadogAgent object for the configured version of the Datadog Operator.

## 0.10.1

* Add configuration for new Operator parameters `maximumGoroutines` and `datadogAgentEnabled`.

## 0.10.0

* Add ability to use the conversion webhook
* Add dependency on the cert manager to manage the certificates of the conversion webhook
* Note that the option to enable the various CRDs has changed from `datadog-crds` to `datadogCRDs`.

## 0.9.2

* Updating CRD dependency to DatadogMonitors and DatadogAgent.
* Update minimum version of the Datadog Operator to 0.8.4.

## 0.9.1

* Updating dependency to CRD to allow all fields.

## 0.9.0

* Add option to deactivate the conversion webhook for usecases where v2alpha1 is solely used.
* Conversion webhook option is not used if the operator version does not support it.
* V2alpha1 is now always served.

## 0.8.8

* Update chart to Datadog Operator tag `0.8.2`.

## 0.8.7

* Add namespaces to all namespace-scoped objects using the HELM standard `Release.namespace`.

## 0.8.6

* Updating dependency to CRD chart.

## 0.8.5

* Updating dependency to CRD chart.

## 0.8.4

* Update dependency on CRD charts to `0.5.2` to allow deployment on Google marketplace.

## 0.8.3

* Update chart to Datadog Operator tag `0.8.1`.

## 0.8.2

* Fix comments in `values.yaml` to allow a seamless `helm-docs` update.

## 0.8.1

* Add arbitrary environment variable definition.

## 0.8.0

* Update chart to Datadog Operator `0.8.0`.

## 0.7.11

* Allow additional service account annotations.

## 0.7.10

* Sync operator RBACs from `datadog-operator` repo to add missing `verticalpodautoscalers` RBACs.

## 0.7.9

* Add missing `datadogmetrics` RBACs.

## 0.7.8

* Fix `PodDisruptionBudget` api version definition when using `helm template`.

## 0.7.7

* Update `PodDisruptionBudget` api version to get rid of `policy/v1beta1 PodDisruptionBudget is deprecated in v1.21+, unavailable in v1.25+; use policy/v1 PodDisruptionBudget` warning.

## 0.7.6

* Nothing

## 0.7.5

* Add a configuration field `containerSecurityContext` to configure a security context for a Container
* Add `site` option to change the Datadog intake site.

## 0.7.4

* Update chart to Datadog CRDs `0.4.5`

## 0.7.3

* Update chart to Datadog Operator `0.7.2` and CRDs `0.4.4`

## 0.7.2

* Add `watchNamespaces` option to configure the namespaces watched by the operator.

## 0.7.1

* Add missing RBAC to the operator to enable the admission controller in the cluster-agent.

## 0.7.0

* Update chart to support the operation version `v0.7.0`

## 0.6.3

* Add missing `poddisruptionbudgets` RBAC when the compliance feature is enabled.

## 0.6.2

* Add a configuration field `collectOperatorMetrics` to disable/enable collecting operator metrics

## 0.6.1

* Update chart for operator release `v0.6.1`
* Support for Datadog API endpoint can change to different region, `dd_url`

## 0.6.0

* Update chart for Operator release `v0.6.0`
* Support Datadog Monitors controller

## 0.5.4

* Add apiKey, apiKeyExistingSecret, appKey, and appKeyExistingSecret values to values.yaml and set their respective env vars using a Kubernetes secret

## 0.5.3

* Only deploy a `PodDisruptionBudget` when `replicaCount` is greater than `1`

## 0.5.2

* Support configuring the secret backend command arguments (requires Datadog Operator v0.5.0+)

## 0.5.1

* Support configuring the secret backend command arguments (requires Datadog Operator v0.5.0+)

## 0.5.0

* Update chart for Operator release `v0.5.0`

## 0.4.1

* Added support for `podAnnotations` and `podLabels` values

## 0.4.0

* BREAKING CHANGES
* Update to work with Operator 0.4: https://github.com/DataDog/datadog-operator/releases/tag/v0.4.0
* Datadog Operator was updated to be based on Operator SDK 1.0. CLI flags are not compatible between 0.x and 0.4

## 0.2.1

* Add "datadog-crds" chart as dependency. It is used to install the datadog's CRDs.

## 0.2.0

* Use `gcr.io` instead of Dockerhub

## 0.1.2

* Fix name of serviceAccount used in Deployment if serviceAccount.name is set

## 0.1.1

* Add automatic README.md generation from `Values.yaml`

## 0.1.0

* Initial version
