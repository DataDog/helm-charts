# Changelog

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
