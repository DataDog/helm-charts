# Changelog

## 2.12.0

- Official image `2.12.0`

## 2.11.1

- Add support for custom annotations on PersistentVolumeClaims (PVCs) via `persistence.annotations` in `values.yaml`

## 2.11.0

- Official image `2.11.0`

## 2.10.0

- Official image `2.10.0`

## 2.9.1

- Official image `2.9.1`

## 2.9.0

- Official image `2.9.0`

## 2.8.1

- Official image `2.8.1`

## 2.8.0

- Official image `2.8.0`

## 2.7.0

- Official image `2.7.0`

## 2.6.0

- Official image `2.6.0`

## 2.5.2

- Official image `2.5.2`

## 2.5.1

- Official image `2.5.1`

## 2.5.0

- Official image `2.5.0`

## 2.4.2

- Official image `2.4.2`

## 2.4.1

- Official image `2.4.1`

## 2.4.0

- Official image `2.4.0`

## 2.3.0

- Official image `2.3.0`

## 2.2.3

- Official image `2.2.3`

## 2.2.2

- Official image `2.2.2`

## 2.2.1

- Official image `2.2.1`

## 2.2.0

- Official image `2.2.0`

## 2.1.2

- Official image `2.1.2`

## 2.1.1

- Official image `2.1.1`

## 2.1.0

- Official image `2.1.0`

## 2.0.2

- Official image `2.0.2`

## 2.0.1

- Official image `2.0.1`

## 2.0.0

- GA release of Observability Pipelines Worker v2
- Removed `datadog.remoteConfigurationEnabled` and `pipelineConfig` values

## 1.8.1

- Migrate from `kubeval` to `kubeconform` for ci chart validation.

## 1.8.0

- Official image `1.8.0`

## 1.7.1

- Official image `1.7.1`

## 1.7.0

- Official image `1.7.0`

## 1.6.0

- Official image `1.6.0`

## 1.5.2

- Dropped ArtifactHub license designation to avoid confusion

## 1.5.1

- Official image `1.5.1`

## 1.5.0

- Official image `1.5.0`

## 1.4.0

- Official image `1.4.0`

## 1.4.0-rc.0

- Nightly image representative of `1.4.0`
- Add `datadog.workerAPI.enabled`, `datadog.workerAPI.playground`, `datadog.workerAPI.address` for Worker API configuration
- Expose Worker API port in pod and through service if enabled
- Remove deprecated `datadog.configKey`

## 1.3.1

- Official image `1.3.1`

## 1.3.0

- Official image `1.3.0`
- Add AP1 Site Comment in `values.yaml`.

## 1.2.1

- Official image `1.2.1`

## 1.2.0

- Official image `1.2.0`

## 1.2.0-rc.1

- Nightly image `2023-05-04`

## 1.2.0-rc.0

- Rename `config` to `pipelineConfig` in values
- Add `datadog.pipelineId` value to replace `datadog.configKey`. `configKey` is still supported for backwards compatability.
- Add new `datadog.remoteConfigurationEnabled` and `datadog.dataDir` values

## 1.1.1

- Update `args` to use the `run` subcommand
- Update default for `DATA_DIR`
- `1.1.1` release

## 1.0.0

- GA release

## 0.1.0

- Initial version
