# Datadog changelog

## 0.17.11

* Update private location image version to `1.58.0`.

## 0.17.10

* Update private location image version to `1.57.0`.

## 0.17.9

* Update private location image version to `1.56.1`.

## 0.17.8

* Update private location image version to `1.56.0`.

## 0.17.7

* Update private location image version to `1.55.0`.

## 0.17.6

* Add optional annotations for service account.

## 0.17.5

* Update private location image version to `1.54.0`.

## 0.17.4

* Update private location image version to `1.53.0`.

## 0.17.3

* Update private location image version to `1.52.0`.

## 0.17.2

* Update private location image version to `1.51.0`.

## 0.17.1

* Update private location image version to `1.50.0`.

## 0.17.0

* Add `podDisruptionBudget` to allow creating and configuring PodDisruptionBudget for deployment.

## 0.16.4

* Update private location image version to `1.49.0`.

## 0.16.3

* Add dnsConfig to DD private location Pod

## 0.16.2

* Update private location image version to `1.48.0`.

## 0.16.1

* Update private location image version to `1.47.0`.

## 0.16.0

* Add `podLabels` value to allow setting labels that only appear on the pods managed by the deployment.

## 0.15.31

* Fix `env` indentation in Deployment template.

## 0.15.30

* Fix `envFrom` indentation in Deployment template.

## 0.15.29

* Update Kubernetes deployment template to set `DATADOG_WORKER_ENABLE_STATUS_PROBES` environment variable when `enableStatusProbes` value is defined.

## 0.15.28

* Update private location image version to `1.46.0`.

## 0.15.27

* Update private location image version to `1.45.0`.

## 0.15.26

* Migrate from `kubeval` to `kubeconform` for ci chart validation.

## 0.15.25

* Update private location image version to `1.44.0`.

## 0.15.24

* Clarify the usage of `configSecret`

## 0.15.23

* Add `priorityClassName` value to specify PriorityClass for pods.

## 0.15.22

* Update private location image version to `1.43.0`.

## 0.15.21

* Update private location image version to `1.42.0`.

## 0.15.20

* Support `dnsPolicy` configuration.

## 0.15.19

* Update private location image version to `1.41.0`.

## 0.15.18

* Update private location image version to `1.40.0`.

## 0.15.17

* Update private location image version to `1.39.0`.

## 0.15.16

* Update private location image version to `1.38.0`.

## 0.15.15

* Update private location image version to `1.37.0`.

## 0.15.14

* Update private location image version to `1.36.0`.

## 0.15.13

* Update private location image version to `1.35.0`.

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
