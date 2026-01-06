# Changelog

## 0.1.14

* Add support for PodDisruptionBudget for metastore

## 0.1.13

* Update Docker image to `v0.1.16`
* Update resource requests and limits to match new sizing recommendations
* Add support for customizing cluster ID
* Add support for `topologySpreadConstraints`

## 0.1.12

* Use Docker image version `v0.1.15`
* Fix indentation under control plane resources section

## 0.1.11

* Fix typo in `valueFrom` defining API key environment variable
* Use latest Docker image including new ingest latency metric and minor bugfixes

## 0.1.10

* Enable reverse connection by default
* Parse syslog-formatted events natively

## 0.1.9

* Add support for reverse connection
* Add tokenizer that behaves like the one used in the SaaS products
* Improve CPU utilization for configurations with fewer than 4 vCPUs
* Export metrics from CloudPrem pods to Datadog Agent or DogStatsD server
* Add sensible defaults for indexer resources
* Add ability to set retention period from the helm chart values
* Improve observability
* Fix bug occurring with TableView widget

## 0.1.8

* Add support for Azure

## 0.1.7

* Add support for autoscaling via Horizontal Pod Autoscaler (HPA) for the indexer and search StatefulSets.

## 0.1.6

* Add support for NGINX Ingress Controller

## 0.1.5

* Introduce `aws.partition` parameter to support service account role ARNs in China regions.

## 0.1.4

* Enable preprocessing by default
* Fix some field remapping issues (most notably, remap `msg` field)
* Fix document tiebreaker
* Fix index/source creation bootstrap job

## 0.1.3

* Add a config parameter to keep-alive connections. By default, it is enabled.
* Isolate CloudPrem gRPC endpoint as a different server running on a different port, to reduce risks of misconfiguration.
For backward compatibility, the CloudPrem endpoint is still available on the regular gRPC port too.
* Add average aggregation
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
