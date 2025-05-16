# Changelog

## 0.1.3

* Adds a config parameter to keep-alive connections. By default, it is enabled.
* Isolates cloudprem the grpc endpoint as a different server running on a different port, to reduce risks of misconfiguration. 
For backward compatibility, the cloudprem endpoint is still available on the regular gRPC port too.
* Added average aggregation
* Support missing options on the attribute remapper:
    * tags can be used as a source and target via `source_type`/`target_type`
    * `target_format` tries to cast attributes into `string`, `integer` or `double`
    * `override_on_conflict`: override if the attribute/tag already exists
* Remap all core attributes in the preprocessing step (remapping did not cover all aliases before)

## 0.1.2

* Add pipelinesConfig property to values.yaml https://github.com/DataDog/pomsky-helm-charts/pull/4
* Fix sort order for same-second documents
* Indexing pomsky's traces in pomsky by default

## 0.1.1

* Load index config from file instead of inline definition
* Switch to gRPC health check for public ALB
* Upgrade image to v0.1.1

## 0.1.0

* Initial version
