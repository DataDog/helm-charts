# Changelog

## 1.7.2

* Remove XValidation as requires K8S >= 1.25.

## 1.7.1

* Add DPA CRD.

## 1.7.0
* Update CRDs from Datadog Operator v1.7.0 tag.

## 1.6.0
* Update CRDs from Datadog Operator v1.6.0 tag.

## 1.5.0
* Update CRDs from Datadog Operator v1.5.0 tag.

## 1.4.0
* Update CRDs from Datadog Operator v1.4.0 tag.

## 1.3.1
* Migrate from `kubeval` to `kubeconform` for ci chart validation.

## 1.3.0
* Update CRDs from Datadog Operator v1.3.0 tag.

## 1.2.0
* Update CRDs from Datadog Operator v1.2.0 tag.

## 1.1.0
* Update CRDs from Datadog Operator v1.1.0-rc.1 tag.

## 1.0.1

* Update CRDs from Datadog Operator v1.0.3.

## 1.0.0

* Default DatadogAgent stored version is `v2alpha1` to align with the GA of the Datadog Operator.

## 0.6.1

* Add missing `nodeLabelsAsTags` and `namespaceLabelsAsTags` to the v2alpha1 spec. 

## 0.6.0

* Support Certificate Manager.
* Document conversion webhook configuration.

## 0.5.9

* Updating DatadogMonitors CRD and DatadogAgents CRDs.

## 0.5.8

* Updating CRD of the Datadog Operator for Kubernetes cluster < 1.21.0.

## 0.5.7

* Update CRD of DatadogAgent to have new fields for the cws feature.

## 0.5.6

* Introduce option to store DatadogAgent v2alpha1 or v1alpha1.

## 0.5.5

* Fix CI, by renaming `kubeval.yaml` to `kubeval-values.yaml`

## 0.5.4

* Fix semver comparison for minor version corner case.
* Update charts.

## 0.5.3

* Fix the semver comparison so v1beta1 is used on 1.21.

## 0.5.2

* Rely on the Kubernetes version to deploy the CRD v1 or v1beta1.

## 0.5.1

* Remove `preserveUnknownFields` to maintain compatibility with Kubernetes versions <1.15.

## 0.5.0

* Update CRDs from Datadog Operator v0.8.0.

## 0.4.7

* Fix Capabilities.APIVersions check

## 0.4.6

* Nothing

## 0.4.5

* Reduce DatadogAgent CRD size by removing description.

## 0.4.4

* Update CRDs from Datadog Operator v0.7.2.

## 0.4.3

* Cleanup `update-crds.sh` script.

## 0.4.2

* Fixed instructions to run the `update-crds.sh` script.

## 0.4.1

* Cleanup `update-crds.sh` script.

## 0.4.0

* Update CRDs from Datadog Operator v0.7.0.
* Remove Extended Daemon Set CRDs from this chart. They will be direclty located in the ExtendedDaemonset chart.

## 0.3.5

* Add CRDs from Extended Daemon Set v0.7.0.

## 0.3.4

* Include only `v1beta1` CRDs from the EDS v0.6.0 tag.

## 0.3.3

* Add CRDs from Extended Daemon Set v0.6.0 tag.

## 0.3.2

* Set `apiVersion` to `v1` for compatibility with helm 2.

## 0.3.1

* Fix typo in DatadogMetrics CRD

## 0.3.0

* Update all the CRDs from operator v0.6.0 tag.

## 0.2.0

* Update all the CRDs from operator v0.5.0 tag.

## 0.1.1

* Move back `chart.yaml` `apiVersion` to `v1` for compatibily with helm2.

## 0.1.0

* Initial version
* Add `DatadogMetrics` and `DatadogAgents` CRDs
