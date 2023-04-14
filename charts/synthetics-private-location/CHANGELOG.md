# Datadog changelog

## 0.15.12

* Update private location image version to `1.34.1`.

## 0.15.11

* Update private location image version to `1.34.0`.

## 0.15.10

* Update private location image version to `1.33.0`.

## 0.15.9

* Fix commonLabels duplicated in Deployment.

## 0.15.8

* Update private location image version to `1.32.0`.

## 0.15.7

* Update private location image version to `1.31.1`.

## 0.15.6

* Update private location image version to `1.31.0`.

## 0.15.5

* Update private location image version to `1.29.0`.

## 0.15.4

* Support `commonLabels` for resources from Kubernetes deployment

## 0.15.3

* Support `commonlabels` configuration to be able to add common labels on all resources created by the chart.

### 0.15.2

* Update private location image version to `1.28.0`.

### 0.15.1

* Update private location image version to `1.27.0`.

### 0.15.0

* Do not default to `configFile` value for configuration to allow using `extraVolumes` to mount configuration files

### 0.14.4

* Update private location image version to `1.26.0`.

### 0.14.3

* Update private location image version to `1.25.0`.

### 0.14.2

* Add ability to template the ConfigMap/Secret name.

### 0.14.1

* Update private location image version to `1.24.0`.

### 0.14.0

* Replace deprecated liveness probe mechanism with the HTTP-based one.
* Add readiness probe using the HTTP-based mechanism.
* Add `enableStatusProbes` value to enable/disable both liveness and readiness probes. Minimal private location image version required: `1.12.0`.

### 0.13.4

* Update private location image version to `1.23.0`.

### 0.13.3

* Update private location image version to `1.22.0`.

### 0.13.2

* Update private location image version to `1.21.0`.

### 0.13.1

* Update private location image version to `1.20.0`.

### 0.13.0

* Add extra mount (`extraVolumes` and `extraVolumeMounts` ) for supporting private root CA certificates as described in <https://docs.datadoghq.com/synthetics/private_locations/configuration/#private-root-certificates>.

### 0.12.1

* Update private location image version to `1.19.0`.

### 0.12.0

* Add support for adding HostAliases to private location pods.

### 0.11.1

* Update private location image version to `1.18.1`.

### 0.11.0

* Update private location image version to `1.18.0`.

### 0.10.0

* Update private location image version to `1.17.0`.

### 0.9.1

* Nothing

### 0.9.0

* Update private location image version to `1.16.0`.

### 0.8.0

* Update private location image version to `1.14.0`.

### 0.7.0

* Update private location image version to `1.13.0`.

### 0.6.0

* Use secret instead of Config Map for `configFile`.
* Added `configSecret` to support passing the json config using a Secret.

### 0.5.0

* Update private location image version to `1.11.0`.

### 0.4.0

* Add 'envFrom' and 'env' to support configuration via environment variables

### 0.3.0

* Added `configConfigMap` to support passing the json config using a Config Map.
* Update the Synthetics Private Location version to `1.10.0`

### 0.2.0

* Use `gcr.io` instead of `Dockerhub`

### 0.1.0

* Initial version
