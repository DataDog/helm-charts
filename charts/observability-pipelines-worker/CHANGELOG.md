# Changelog

## 1.4.0

* Official image `1.4.0`

## 1.4.0-rc.0

* Nightly image representative of `1.4.0`
* Add `datadog.workerAPI.enabled`, `datadog.workerAPI.playground`, `datadog.workerAPI.address` for Worker API configuration
* Expose Worker API port in pod and through service if enabled
* Remove deprecated `datadog.configKey`

## 1.3.1

* Official image `1.3.1`

## 1.3.0

* Official image `1.3.0`
* Add AP1 Site Comment in `values.yaml`.

## 1.2.1

* Official image `1.2.1`

## 1.2.0

* Official image `1.2.0`

## 1.2.0-rc.1

* Nightly image `2023-05-04`

## 1.2.0-rc.0

* Rename `config` to `pipelineConfig` in values
* Add `datadog.pipelineId` value to replace `datadog.configKey`. `configKey` is still supported for backwards compatability.
* Add new `datadog.remoteConfigurationEnabled` and `datadog.dataDir` values

## 1.1.1

* Update `args` to use the `run` subcommand
* Update default for `DATA_DIR`
* `1.1.1` release

## 1.0.0

* GA release

## 0.1.0

* Initial version
