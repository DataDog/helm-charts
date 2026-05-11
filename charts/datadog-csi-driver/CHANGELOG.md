# Changelog

## 0.15.0

* Enable SSI (Single Step Instrumentation) on GKE Autopilot by always rendering the `storage-dir` volume/mount and the `DD_APM_ENABLED` env var, and by adding a new `datadog-datadog-csi-driver-daemonset-exemption-v1.1.0` allowlist entry.

## 0.14.0

* Set the `csi.datadoghq.com/apm-enabled` annotation on the `k8s.csi.datadoghq.com` `CSIDriver` resource based on `apm.enabled`. The cluster-agent admission controller reads this annotation (via the `workloadmeta-kubeapiserver` CSIDriver collector) to decide whether SSI library injection can use CSI mode for this driver.

## 0.13.0

* Add `driver.resources` value to configure resource requests and limits for the CSI driver container.

## 0.12.0

* Add `labels` value to configure labels on CSI driver daemonset pods.

## 0.11.0

* Registry allow list is now configured via `global.apmRegistryAllowList` in the parent `datadog` chart. When set, the CSI driver enforces the list via `DD_REGISTRY_ALLOW_LIST` and the admission controller enforces it via `DD_ADMISSION_CONTROLLER_AUTO_INSTRUMENTATION_CONTAINER_REGISTRY_ALLOW_LIST`. Both layers must be satisfied for injection to proceed.

## 0.10.1

* Fix false positive outcome in csi e2e test ([#2579](https://github.com/DataDog/helm-charts/pull/2579)).
* Bump CSI driver version to include bug fix ([#77](https://github.com/DataDog/datadog-csi-driver/pull/77)).

## 0.10.0

* Add `priorityClassName` support for CSI driver daemonset pods (default: `""`).

## 0.9.1

* Set csi driver image to `1.2.1`

## 0.9.0

* Set csi driver image to `1.2.0`

## 0.8.0

* Support configuring `NodeAffinity` and `NodeSelector` in datadog csi driver chart.

## 0.7.0

* [CONTP-1250] feat(csi_driver): Make updateStrategy configurable and increase default strategy. ([#2369](https://github.com/DataDog/helm-charts/pull/2369)).

## 0.6.0

* Add `apm.enabled` configuration option to enable/disable APM/SSI support (not yet supported on GKE Autopilot)

## 0.5.0

* [CONTP-719] Expose security context and annotation configurations ([#2317](https://github.com/DataDog/helm-charts/pull/2317)).

## 0.4.4

* Support the definition of tolerations

## 0.4.3

* Fix AllowlistSynchronizer helper

## 0.4.2

* Add gke AllowlistSynchronizer

## 0.4.1

* Mount `apm-socket` and `dsd-socket` to CSI node server container in readonly mode.
* Mount `plugins-dir` to node registrar container in readonly mode.

## 0.4.0

* Set node server image tag to `1.0.0`.

## 0.3.4

* Remove `hostNetwork: true` from csi driver daemonset.

## 0.3.3

* Fix bug that caused to pass the socket's parent directory to the start command arguments instead of the full socket path.

## 0.3.2

* Add option to configure CSI registrar image

## 0.3.1

* Fix image pull secrets of the CSI driver daemonset.

## 0.3.0

* Support configuring different host socket paths for apm and dogstatsd sockets.

## 0.2.0

* Support configuring apm and dogstatsd sockets hostpaths.

## 0.1.0

* Initial version
