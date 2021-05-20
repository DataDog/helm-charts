# Changelog

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
