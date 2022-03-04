# Changelog

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
