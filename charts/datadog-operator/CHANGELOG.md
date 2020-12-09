# Changelog

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
