# Datadog changelog

## 3.81.0

* Add a new option to disable hostPorts for the trace-agent with `datadog.apm.useLocalService`. This option enables K8s clusters with hostPort and hostPath volumes restrictions to use the K8s local service to send traces.

## 3.80.0

* Add `datadog.admissionController.validation` and `datadog.admissionController.mutation` to enable/disable the admission controller validation and mutation webhooks.

## 3.79.1

* Document how to use `datadog.envDict` option with the `--set` helm's flag.

## 3.79.0

* Add Logs Collection support for Google GKE on GDC

## 3.78.0

* Set default `Agent` and `Cluster-Agent` version to `7.59.0`.

## 3.77.3

* Update version required for datadog.processAgent.runInCoreAgent and remove experimental status.

## 3.77.2

* Add the ability to include Security Contexts at the container level for Cluster Checks Runners.

## 3.77.1

* Modify command that removes the default conf.d directory from the Cluster Checks Runners and only removes the default YAML files.

## 3.77.0

* Add experimental support for overlayfs direct scan for SBOMs

## 3.76.3

* Add `podisruptionbudgets` RBAC to the Cluster Agent.

## 3.76.2

* Fix warning message displayed when installing/upgrading the Agent with OTel collector.
* Add preview message in values.yaml file.

## 3.76.1

* Gate `datadog.sbom.containerImage.uncompressedLayersSupport` feature behind `datadog.sbom.containerImage.enabled`: if the latter is not enabled (default), do not modify template based on `datadog.sbom.containerImage.uncompressedLayersSupport`.

## 3.76.0

* Set `datadog.sbom.containerImage.uncompressedLayersSupport` to `true` by default.

## 3.75.0

* Set default `Agent` and `Cluster-Agent` version to `7.58.0`.

## 3.74.6

* Fix error message for when System Probe is enabled on GKE Autopilot

## 3.74.5

* Add configuration option for `datadog.KubernetesEvents.sourceDetectionEnabled` to map Kubernetes events to integration sources based on controller names. Disabled by default.

## 3.74.4

* Define `admission_controller.container_registry` regardless of `clusterAgent.admissionController.agentSidecarInjection` feature status.

## 3.74.3

* Do not mount `/usr/lib/sysimage/rpm` (reverts https://github.com/DataDog/helm-charts/pull/1541): in some operating systems such as Bottlerocket, `/usr` is `read-only`, preventing the Agent from being deployed when `datadog.sbom.host.enabled` is set to `true` as kubelet cannot create the directory at this location if it does not exist.

## 3.74.2

* Mount `/usr/lib/sysimage/rpm` in the Agent DaemonSet when using host SBOM feature (required on hosts running Amazon Linux distributions).

## 3.74.1

* Pass components env variables to the cluster checks runner deployment pod spec.

## 3.74.0

* Simplify OTel Agent OOTB pipelines:
  * Remove `traces/otlp` pipeline from the default OTel Agent config
  * Add `infaattributes` processor and `datadog` exporter to the `traces` pipeline.

## 3.73.3

* Fix a few typos on OTel Agent configs.

## 3.73.2

* Add `admissionregistration.k8s.io/v1/validatingwebhookconfigurations` RBACs to the Cluster Agent.

## 3.73.1

* Add role-based access control rules to Datadog Cluster Agent to read k8s resources annotations and labels to create tags.

## 3.73.0

* Add Azure Container Registry, enabled automatically when targeting `us3.datadoghq.com`.

## 3.72.1

* Add configuration option for `datadog.KubernetesEvents.filteringEnabled` to only include pre-defined allowed events. Disabled by default.

## 3.72.0

* Set default `Agent` and `Cluster-Agent` version to `7.57.2`.

## 3.71.2

* Add `datadog.kubernetesResourcesLabelsAsTags` to assign Kubernetes Resources Labels as tags in the tagger
* Add `datadog.kubernetesResourcesAnnotationsAsTags` to assign Kuberenetes Resources Annotations as tags in the tagger

## 3.71.1

* Update `fips.image.tag` to `1.1.5` updating openSSL version to 3.0.15

## 3.71.0

* Add `datadog.profiling` section to configure Continuous Profiler. Disabled by default.

## 3.70.7

* Set default `Agent` and `Cluster-Agent` version to `7.56.2`.

## 3.70.6

* Add private beta note for OTel Collector.

## 3.70.5

* Set default `Agent` and `Cluster-Agent` version to `7.56.1`.

## 3.70.4

* Improve support for `processAgent.runInCoreAgent` feature.

## 3.70.3

* Update `fips.image.tag` to `1.1.4`

## 3.70.2

* Add admission controller port to cilium network policy for the cluster agent

## 3.70.1

* Fix datadog.kubelet.coreCheckEnabled conditional statement to accept false value

## 3.70.0

* Set default `Agent` and `Cluster-Agent` version to `7.56.0`.

## 3.69.3

* Update `datadog-crds` dependency to `1.7.2`.

## 3.69.2

* Allow activation of autoscaling.

## 3.69.1

* Set default `Agent` and `Cluster-Agent` version to `7.55.2`.

## 3.69.0

* Add support OTel Agent container. OTel Agent is Datadog's distribution of OTel collector.

## 3.68.2

* Fix datadog.containerLifecycle.enabled conditional statement to accept false value

## 3.68.1

* Add automatic detection for enablement of process agent container.

## 3.68.0

* Set default `Agent` and `Cluster-Agent` version to `7.55.1`.

## 3.67.5

* Add support for `processAgent.runInCoreAgent` as an experimental feature.

## 3.67.4

* Overwrite the securityContext for the `seccomp-setup` initContainer with `agents.containers.initContainers.securityContext`.

## 3.67.3

* Make sure that disabling CSPM host benchmarks is propagated to the agent.

## 3.67.2

* Remove startup probe for `Agent` in GKE AutoPilot due to deployment restrictions

## 3.67.1

* Update `fips.image.tag` to `1.1.3`

## 3.67.0

* Add startup probe for `Agent`, `Cluster-Agent` and `Cluster-Check-Runner`.

## 3.66.1

* Add 'datadog.namespaceAnnotationsAsTags' to assign namespace annotations as tags on pod entities in the tagger.

## 3.66.0

* Set default `Agent` and `Cluster-Agent` version to `7.54.0`.

## 3.65.3

* Add RBAC rules for collection of StorageClass and LimitRange resources in the Orchestrator Explorer.

## 3.65.2

* Do not enable live process collection by default when language detection is enabled for `APM SSI`.

## 3.65.1

* Make sure the security agent is aware of `datadog.securityAgent.runtime.useSecruntimeTrack`.

## 3.65.0

* Default `datadog.securityAgent.runtime.useSecruntimeTrack` to `true`, sending CWS events directly to the new secruntime track (and to the new agent events explorer).

## 3.64.1

* Add `datadog.securityAgent.runtime.useSecruntimeTrack` config to start sending CWS events directly to the new secruntime track (and to the new agent events explorer).

## 3.64.0

* Add `datadog.originDetectionUnified.enabled` setting to enable unified origin detection for container tagging. Disabled by default

## 3.63.0

* Set kubelet core check to be enabled by default

## 3.62.1

* Update `fips.image.tag` to `1.1.2`

## 3.62.0

* Add `datadog.asm` section to configure various features of the ASM Security Product. Disabled by default

## 3.61.0

* Add `datadog.kubelet.core_check` option to configure whether the kubelet core check should be used
  Note: this requires agent/cluster agent version 7.53.0+

## 3.60.0

* Set default `Agent` and `Cluster-Agent` version to `7.53.0`

## 3.59.7

* Add configuration option to specify clusterAgent.admissionController.containerRegistry, which defaults to registry
* No longer set `DD_ADMISSION_CONTROLLER_AGENT_SIDECAR_CONTAINER_REGISTRY` to registry as a fallback,
  that option is implicit from us now setting the higher level `clusterAgent.admissionController.containerRegistry`.

## 3.59.6

* Add configuration option datadog.apm.instrumentation.skipKPITelemetry.

## 3.59.5

* Set default `Agent` and `Cluster-Agent` version to `7.52.1`.

## 3.59.4

* Add language detection enable option for `APM` instrumentation.

## 3.59.3

* Add `contimage-intake.datadoghq.com` & `contlcycle-intake.datadoghq.com` endpoints to the `Agent` cilium network policy.

## 3.59.2

* Disable language detection reporting by default in Cluster Agent with Agent 7.52+.

## 3.59.1

* Add support for configuring Agent sidecar injection using Admission Controller.

## 3.59.0

* Set default `Agent` and `Cluster-Agent` version to `7.52.0`.

## 3.58.1

* Fix typo in PodSecurityPolicy warning note.

## 3.58.0

* Change configuration options for APM Instrumentation. Starting from Agent and Cluster-Agent version `7.51.0` APM Instrumentation needs to be configured using the following configuration options:
* `datadog.apm.instrumentation.enabled` - set to `true` to enable automatic instrumentation.
* `datadog.apm.instrumentation.enabledNamespaces` - optional; list of namespaces to enable automatic instrumentation in. If not provided, every namespace in the cluster will be instrumented.
* `datadog.apm.instrumentation.disabledNamespaces` - optional; list of namespaces to disable automatic instrumentation in.

## 3.57.3

* Exclude agent, cluster agent and agent clusterchecks pods from injection from the admission controller.

## 3.57.2

* Add `networkpolicies` default permission for the cluster agent.

## 3.57.1

* Allow configuring CWS security profile based auto suppression feature and enable it by default.

## 3.57.0

* Set default `Agent` and `Cluster-Agent` version to `7.51.0`.

## 3.56.0

* Allow templating of `datadog.clusterName`.

## 3.55.0

* Modify `datadog.dogstatsd.originDetection` to also support container tagging for origin detection enabled clients.

## 3.54.2

* Set `DD_APM_ENABLED` value in the core agent container to properly report its value.

## 3.54.1

* Migrate from `kubeval` to `kubeconform` for ci chart validation.

## 3.53.3

* Update `fips.image.tag` to `1.1.1`

## 3.53.2

* Exclude agent pod from labels injection from the admission controller.

## 3.53.1

* Update `fips.image.tag` to `1.1.0`

## 3.53.0

* Add `otlp.logs.enabled` option to datadog agent to set the `DD_OTLP_CONFIG_LOGS_ENABLED` env variable.

## 3.52.0

* Allow configuring CWS security profile features and enable drift events by default

## 3.51.2

* Use correct kpi-telemetry-configmap in Cluster Agent and Trace Agent.

## 3.51.1

* Parametrize the name of kpi-telemetry-configmap.

## 3.51.0

* Add `DD_INSTRUMENTATION_INSTALL_TIME`, `DD_INSTRUMENTATION_INSTALL_ID`, `DD_INSTRUMENTATION_INSTALL_TYPE` env variables to the Trace and Cluster agents to support APM Telemetry KPIs.

## 3.50.5

* Add option to use containerd snapshotter to generate SBOMs.

## 3.50.4

* Mount host files for proper OS detection in SBOMs.

## 3.50.3

* Set default `Agent` and `Cluster-Agent` version to `7.50.3`.

## 3.50.2

* Support automatic registry selection based on `datadog.site` on GKE Autopilot.

## 3.50.1

* Set default `Agent` and `Cluster-Agent` version to `7.50.2`.

## 3.50.0

* Set default `Agent` and `Cluster-Agent` version to `7.50.1`.

## 3.49.9

* Update `fips.image.tag` to `1.0.1`

## 3.49.8

* Mount host package manager database when host SBOM is enabled.

## 3.49.7

Fix NOTES warning for APM Instrumentation

## 3.49.6

Get rid of the old GODEBUG=x509ignoreCN=0 hack that is not effective anymore in lastest versions of the agent.

## 3.49.5

* Fix registry selection with GKE Autopilot until new registries are allowed.

## 3.49.4

* Exclude a namespace with Datadog resources from APM Single Step Instrumentation

## 3.49.3

* Fix NOTES warning for APM Instrumentation when apm.intrumentation.disabledNamespaces is set

## 3.49.2

* Fix check for APM Instrumentation when apm.intrumentation.disabledNamespaces is set

## 3.49.1

* Update `fips.image.tag` to `1.0.0`

## 3.49.0

* Beta: Add `datadog.apm.instrumentation` section to configure APM Single Step Instrumentation

## 3.48.0

* Set default `Agent` and `Cluster-Agent` version to `7.49.1`.

## 3.47.2

* Fix CI following enabling container image collection by default.

## 3.47.1

* Fix `registry` being ignored even if set.

## 3.47.0

* `registry` is now set automatically adapted based on `datadog.site` value. Still default to `gcr.io/datadoghq` if not set.

## 3.46.0

* Enable container image collection by default.

## 3.45.0

* Separate values for `DD_CONTAINER_INCLUDE` and `DD_CONTAINER_EXCLUDE` in `Agent` and `Cluster-Agent`
  Note: this requires agent/cluster agent version 7.50.0+

## 3.44.1

* Fix local agent Kubernetes service to include APM traceport

## 3.44.0

* Remove buggy `chmod` directive in the init container of the cluster agent.

## 3.43.2

* Remove line break in helpers tpl file that prevents the chart from rendering in older Helm versions.

## 3.43.1

* Fix docstring typos and remove unneeded lines.

## 3.43.0

* Default `Agent` and `Cluster-Agent` to `7.49.0` version.

## 3.42.1

* Bump FIPS proxy OpenSSL version to 3.0.12

## 3.42.0

* Allow enabling SBOM collection for host and container images.

## 3.41.0

* Enable container lifecycle events collection by default.

## 3.40.4

* Add the option `clusterAgent.metricsProvider.registerAPIService` to allow user to disable registering external-metrics server as an `APIService`

## 3.40.3

* Default `Agent` and `Cluster-Agent` to `7.48.1` version.

## 3.40.2

* Gate `PodSecurityPolicy` RBAC for k8s versions which no longer support this deprecated API.

## 3.40.1

* Add support for initContainer volume mounts

## 3.40.0

* Default `Agent` and `Cluster-Agent` to `7.48.0` version.

## 3.39.3

* Omit cluster check and leader election in orchestrator check configuration if custom resources are provided

## 3.39.2

* Support custom resources and custom resource definitions collection in orchestrator explorer

## 3.39.1

* Add `kubeStateMetricsCore.collectConfigMaps` config field to the Agent

## 3.39.0

* Add a new parameter `datadog.leaderElectionResource` to select which resource lock to use in the leader election. Can be `leases(s)` in agent 7.47+, `configmap(s)`, or empty for auto detection.

## 3.38.4

* Add `orchestrator_explorer.enabled` for the Agent

## 3.38.3

* Update `fips.image.tag` to `0.6.0`

## 3.38.2

* Skip references to PodSecurityPolicy where the support of this API has been dropped.

## 3.38.1

* Enable Remote Config by default on the host agent only

## 3.38.0

* Default `Agent` and `Cluster-Agent` to `7.47.1` version.

## 3.37.1

* Temporarily revert enabling Remote Config by default

## 3.37.0

* Rename `datadog.securityAgent.compliance.xccdf.enabled` parameter to `datadog.securityAgent.compliance.host_benchmarks.enabled`.

## 3.36.4

* Disable Remote Config on the cluster checks runner

## 3.36.3

* Mount `/etc/passwd` in process agent only if `datadog.processAgent.processCollection` or `datadog.processAgent.processDiscovery` is enabled.

## 3.36.2

* Update `fips.image.tag` to `0.5.5` which upgrades HAProxy to 2.4.24 and zlib to 1.3

## 3.36.1

* Add option to enable CWS security profiles (runtime anomaly detection)

## 3.36.0

* Enable Remote Config by default

## 3.35.2

* Fix Agent Service Account Name used in `RoleBinding` for Secret Backend permissions when in GKE Autopliot

## 3.35.1

* Add permissions to curl `/metrics/slis` to agent cluster role.

## 3.35.0

* Default `Agent` and `Cluster-Agent` to `7.47.0` version.

## 3.34.3

* Fix extra empty line in helmchecks, issue [#953](https://github.com/DataDog/helm-charts/issues/953).

## 3.34.2

* Add containerPort 8000/TCP to `cluster-agent` deployment for Admission Controller.

## 3.34.1

* Fix `clusterAgent.admissionController.webhookName` RBAC to avoid restricting `create` by resource name.

## 3.34.0

* Introduced a new parameter `clusterAgent.admissionController.webhookName` for selecting the name of the mutating webhook.
* Narrowed the admission controller's RBAC scope in the cluster agent to only include a single resourceName, specifically `clusterAgent.admissionController.webhookName`.

## 3.33.10

* Avoid creating the `DD_PROVIDER_KIND` environment variable twice for containers.

## 3.33.9

* Add `fips.customFipsConfig` parameter to allow configuring FIPS proxy sidecar `datadog-fips-proxy.cfg` using a ConfigMap.

## 3.33.8

* Remove `mountPropagation` for `/etc/os-release` files.

## 3.33.7

* Add additional intakes into `CiliumNetworkPolicy` for node Agent and Cluster Check Runner for profiling, network monitoring, dbm, and remote config

## 3.33.6

* Ensure the core agent is aware that CSPM is enabled (for inventories purposes).

## 3.33.5

* Daemonset includes `logdatadog` volume when rendered for `targetSystem: "windows"`

## 3.33.4

* Update `fips.image.tag` to `0.5.4` increasing the health checks interval from 2 to 10 seconds in the FIPS compliant side car container

## 3.33.3

* Remove `datadog.dataStreamsMonitoring.enabled` parameter.

## 3.33.2

* Add emptyDir and volumeMounts for Agent log files in Windows containers to fix log file access

# 3.33.0

* Default `Agent` and `Cluster-Agent` to `7.46.0` version.

## 3.32.8

* Always set the Remote Configuration environment variable

## 3.32.7

* Update the cluster agent network policy to allow telemetry submission.

## 3.32.6

* Fix cluster agent pod failing to start when securityContext is set.

## 3.32.5

* Fix comment for datadog.kubernetesEvents.collectedEventTypes in values.yaml.

## 3.32.4

* Add futimens, utime, utimes and utimensat syscalls to system-probe seccomp.

## 3.32.3

* Allows configuration of `dogstatsd.tagCardinality` independent of `dogstatsd.originDetection`.

## 3.32.2

* Set the `priority` field of the OpenShift’s SCC to `null` in order to not have a higher priority than the OpenShift 4.11+ default `restricted-v2` SCC.

## 3.32.1

* Add AP1 Site Comment at `value.yaml`.
* Fix CVE in the FIPS compliant side car container

## 3.32.0

* Add a new preferred parameter to enable Remote Configuration on both the agent and the cluster agent.

## 3.31.0

* Default `Agent` and `Cluster-Agent` to `7.45.0` version.

## 3.30.10

* Updated pointerdir mountPath for Windows deployments.

## 3.30.9

* Pass its pod name to the cluster-agent. This is used by cluster agent 7.46+ to make leader election work when using host network.

## 3.30.8

* Update `fips.image.tag` to `0.5.2` version

## 3.30.7

* Fix Windows support of `agents.customAgentConfig` to avoid bind mount of a file.

## 3.30.6

* Adds `datadog.kubeStateMetricsCore.collectApiServicesMetrics` (`false` by default) to collect apiservices metrics in Kube State Metrics Core.
  Note: APIServices metrics collection requires Cluster Agent 7.45.0+.

## 3.30.5

* Add `list` and `watch` permissions of `apiservices` resources for the `kubernetes_state_core` check.

## 3.30.4

* Remove USM private beta comments.

## 3.30.3

* Remove resourceName field from `create` permission of `leases` in `cluster-agent-rbac`.

## 3.30.2

* Add `get`, `create`, `update` permissions of `leases` to `cluster-agent-rbac`.

## 3.30.1

* Remove guidance that users must manually convert tag syntax for `labelsAsTags`

## 3.30.0

* Add `datadog.dataStreamsMonitoring.enabled` parameter to enable Data Stream Monitoring.

## 3.29.3

* Add `inotify_add_watch`, `inotify_init`, `inotify_init1`, and `inotify_rm_watch` to the default seccomp profile of system-probe.

## 3.29.2

* Default `Agent` and `Cluster-Agent` to `7.44.1` version.

## 3.29.1

* Add `customresourcedefinitions` option to enable CRD metrics collection in KSM Core.

## 3.29.0

* Add `datadog.securityAgent.compliance.xccdf.enabled` parameter to enable XCCDF feature in CSPM.

## 3.28.1

* Add `memfd_create` syscall to seccomp profile for system-probe.

## 3.28.0

* Adding support to use a FIPS compliant side car container in the Datadog Cluster Agent, the Datadog Agent, and the Datadog Cluster Check Runners pods.

## 3.27.0

* Default `Agent` and `Cluster-Agent` to `7.44.0` version.

## 3.26.2

* Adds statx syscall to seccomp for system-probe

## 3.26.1

* Add support for `topologySpreadConstraints` in pod templates

## 3.26.0

* Default `Agent` and `Cluster-Agent` to `7.43.2` version.

## 3.25.5

* Adds securityContext and resource annotations for initContainers in cluster agent

## 3.25.4

* Add `list` and `watch` permissions of `customresourcedefinitions` to `kube-state-metrics-core-rbac`.

## 3.25.3

* Remote Config is now enabled even if the Cluster Agent is disabled.

## 3.25.2

* Fix a bug with `datadog.remoteConfiguration.enabled` where Remote Config was only enabled for the main agent container but not other containers such as the trace-agent.

## 3.25.1

* Fix CI to unblock release of charts

## 3.25.0

* Automatically collect Security Profiles when CWS is enabled.

## 3.24.0

* Move `kube-state-metrics` default image registry from k8s.gcr.io to registry.k8s.io.

## 3.23.0

* Injects additional environment variables in the Cluster Agent
* Add `clusterAgent.rbac.flareAdditionalPermissions` parameter to enable user Helm values retrieval in DCA flare (`true` by default)

## 3.22.0

* Auto-configure `clusterAgent.admissionController.configMode` based on `datadog.apm.socketEnabled|portEnabled`.

## 3.21.0

* Add `datadog.remoteConfiguration.enabled` parameter to enable remote configuration.

## 3.20.3

* Fix command script in linux init container to prevent blocking deployment in GKE Autopilot on Rapid release channel.
* Only mount DogStatsD socket in non-Autopilot environments.

## 3.20.2

* Fix R/W volume mounts for CRI on Windows

## 3.20.1

* Fix command args in linux init container to prevent blocking deployment in GKE Autopilot.

## 3.20.0

* Enable CWS network detections by default.

## 3.19.2

* Fix R/W volume mounts in init containers on Windows

## 3.19.1

* Mount emptyDir volumes in `/etc/datadog-agent` and `/tmp` to allow the cluster-agent to write files in those
  locations with read-only root filesystem.

## 3.19.0

* Declare `readOnly` in volumeMounts.

## 3.18.0

* Default `Agent` and `Cluster-Agent` image tags to `7.43.1`.

## 3.17.1

* Fix Cilium egress rules to kube-apiserver entities.

## 3.17.0

* Add the following configurations which allow environment variables to be defined in a dictionary:
  * `agents.containers.agent.envDict`
  * `agents.containers.processAgent.envDict`
  * `agents.containers.securityAgent.envDict`
  * `agents.containers.systemProbe.envDict`
  * `agents.containers.traceAgent.envDict`
  * `clusterAgent.envDict`
  * `clusterChecksRunner.envDict`
  * `datadog.envDict`

## 3.16.2

* Mount an emptyDir volume in `/opt/datadog-agent/run` to allow the cluster-agent to write files in that location
  with read-only root filesystem.

## 3.16.1

* Fix `cluster-agent` deployment to allow the cluster-agent to write file in `/var/log/datadog` when it runs with
  read-only root filesystem.

## 3.16.0

* Add new checksum to cluster agent deployment base on all cluster-agent configmap configuration.

## 3.15.0

* Beta: Enable remote configuration if `clusterAgent.admissionController.remoteInstrumentation` is enabled.

## 3.14.0

* Make the root filesystem of the cluster agent container read only by default

## 3.13.0

* Beta: Support APM library injection with Remote Configuration.

## 3.12.0

* Add `automountServiceAccountToken` option to configure automatic mounting of ServiceAccount's API credentials

## 3.11.0

* Default `Agent` and `Cluster-Agent` image tags to `7.43.0`.

## 3.10.9

* Default `Agent` and `Cluster-Agent` image tags to `7.42.2`.

## 3.10.8

* Fix `cluster-agent` SCC, remove duplicate `users` field.

## 3.10.7

* Default `Agent` and `Cluster-Agent` image tags to `7.42.1`.

## 3.10.6

* Includes the imagePullPolicy key for the seccomp-setup container template

## 3.10.5

* Only expose the shared volume for the auth-token in non autopilot environments.

## 3.10.4

* Fix documentation for `agents.containers.traceAgent.env` and `agents.containers.securityAgent.env`

## 3.10.3

* Fix default `hostPid` value set to true on Windows.
* Fix auth token path value on Windows.

## 3.10.1

* Fix: add missing `DAC_READ_SEARCH` capability in agent PSP and SCC (openshift)

## 3.10.0

* Default `Agent` and `Cluster-Agent` image tags to `7.42.0`.

## 3.9.0

* Set processDiscovery to be true by default

## 3.8.1

* Update docs for `datadog.otlp.receiver.protocols.grpc.endpoint`

## 3.8.0

* Add `providers.gke.cos` option to prevent `/usr/src` from being mounted on COS

## 3.7.3

* Add support for Secret Annotations using `datadog.SecretAnnotations` helm value

## 3.7.2

* Rename dogstatsd port on the Agent Service to match the name of the dogstatsd port in the Agent pod (`dogstatsd -> dogstatsdport`).

## 3.7.1

* Add required capability to system-probe in order to make the `auth_token` file readable.

## 3.7.0

* Add `datadog.kubernetesEvents.*` options to configure new Kubernetes unbundling events feature.
  (This parameter exists only in agent 7.42.0 and above and cluster-agent 7.42.0 and above.)
* Add `datadog.clusterTagger.*` options to configure the Kubernetes cluster-tagger feature.
  (This parameter exists only in agent 7.42.0 and above and cluster-agent 7.42.0 and above.)
* Create `components-common-env` to define shared environment variable between "agent" and "cluster-agent" containers, and refactor `containers-common-env`.

## 3.6.9

* Add `auth_token` to all the containers.

## 3.6.8

* Add missing RBAC rules for collection of Vertical Pod Autoscaler resources in the Orchestrator Explorer.

## 3.6.7

* Default `Agent` and `Cluster-Agent` image tags to `7.41.1`.

## 3.6.6

* Fix missing volumeMount in `security-agent` container when `datadog.kubelet.hostCAPath` is provided.

## 3.6.5

* Fix missing Cluster Agent configuration in `security-agent` if CSPM is not actived.

## 3.6.4

* Change nesting for `providers.aks.enabled` parameter in Helm template.

## 3.6.3

* Add `datadog.kubeStateMetricsCore.annotationsAsTags` that expose the `annotations_as_tags` parameter of the KSM core check.
  This parameter exists only in agent 7.42.0 and above and cluster-agent 7.42.0 and above.

# 3.6.2

* Add CRDs to the cluster agent RBAC to be able to collect them using the Orchestrator Explorer.

## 3.6.1

* Add `providers.aks.enabled` parameter to activate specific configuration options for AKS.

## 3.6.0

* Update "Agent" and "Cluster-Agent" versions to `7.41.0` by default.

## 3.5.2

* Fix API Key check in NOTES.txt following change of default value for `datadog.apiKey`.
* Fix failure if PSP activated in Kubernetes 1.25 (PSP have been removed).

## 3.5.1

* Removing default value placeholder for the API Key in the values.yaml.

## 3.5.0

* Remove runtime compilation-related config values `enableKernelHeaderDownload` and `enableRuntimeCompiler` in the system-probe.

## 3.4.0

* Add `datadog.systemProbe.btfPath` for mounting user-provided BTF files (see datadog-agent PRs #13962 and #14096 for more context).

## 3.3.3

* Add a warning note to alert users about suboptimal configuration of Cluster Checks Runner.

## 3.3.2

* Fix GKE Autopilot mounts in the `trace-agent` container and `hostPid` setting for the Agent pods

## 3.3.1

* Remove `mountPropagation` for `*-release` files in `/etc`. It is not needed for individual files.

## 3.3.0

* Add datadog.hostPID option and deprecate datadog.dogstatsd.hostPID.

## 3.2.2

* Mount `/host/proc` and `/host/sys/fs/cgroup` in trace-agent container for better support of container tagging

## 3.2.1

* Default "Agent" and "Cluster-Agent" image tag to `7.40.1`.

## 3.2.0

* Default "Agent" and "Cluster-Agent" image tag to `7.40.0`.

## 3.1.11

* Allow disabling use of the Host Port when enabling OTLP Ingest for Agent
* Add OTLP Ingest ports to Agent Service, to be used when Host Port is disabled

## 3.1.10

* Default "Agent" and "Cluster-Agent" image tag to `7.39.2`.

## 3.1.9

* Add `faccessat` to system-probe seccomp profile.

## 3.1.8

* Add `clone3` and `rseq` to system-probe seccomp profile.

## 3.1.7

* Fix the configuration of the default seccomp profile for system-probe

## 3.1.6

* Fix usage of `generate-security-context` helper.

## 3.1.5

* Use `securityContext.seccompProfile` instead of annotations for system-probe on kubernetes 1.19+.

## 3.1.4

* Default "Agent" and "Cluster-Agent" image tag to `7.39.1`.

## 3.1.3

* Add `datadog.helmCheck.valuesAsTags` option to collect helm values and use them as tags.

## 3.1.2

* Add `datadog.securityAgent.runtime.activityDump.enabled` configuration to enable CWS activity dumps.

## 3.1.1

* Set default value for `datadog.systemProbe.enableKernelHeaderDownload` to `true`

## 3.1.0

* Default Agent image to `7.39.0`.
* Default Cluster-Agent image to `7.39.0`. Cluster-Agent versioning is now aligned with the Agent.

## 3.0.4

* Fix preventing mounting os-release in GKE autopilot for all containers.

## 3.0.3

* Add `faccessat2` to allowed actions in system-probe seccomp profile.

## 3.0.2

* Allow disabling kubeStateMetricsCore rbac creation.

## 3.0.1

* Add `datadog.systemProbe.enableDefaultKernelHeadersPaths` option that allows
  to choose whether to mount the default kernel headers paths.

## 3.0.0

* Minimum version of the Agent supported is 7.36.0 and minimum version of the Cluster Agent supported is 1.20.0.
* Disable the legacy KSM check and enable the KSM core check by default.
* Drop support for Helm 2.

## 2.37.9

* Add `DD_PROMETHEUS_SCRAPE_VERSION` to Cluster Agent to match Agent version

## 2.37.8

* Fix the volumeMount duplication in `system-probe` container if `datadog.osReleasePath` value
  corresponds to one of the default os-release-paths automatically mounted.
* Add the option to disable the default os-release path mount linked to `system-probe` container.

## 2.37.7

* Fix Windows nodes deployment: do not mount `container-host-release-volumemounts` if
  the `targetSystem` is "Windows".

## 2.37.6

* Add `chmod` to allowed actions in system-probe seccomp profile

## 2.37.5

* Mount host release files for proper host OS detection

## 2.37.4

* Add `digest` as a configurable value for all datadog images used

## 2.37.3

* Update default agent image version tag to `7.38.2`.
* Rename view CI values.yaml files to be executed by the CI.

## 2.37.2

* Set traced_cgroups_count default value to 0 in the system-config file for CWS.

## 2.37.1

* Default Datadog Agent image to `7.38.1`.

## 2.37.0

* Default Datadog Agent image to `7.38.0`.
* Default Datadog Cluster Agent image to `1.22.0`.

## 2.36.9

* Add `/etc/dnf/vars` and `/etc/yum/vars` to the default package management directories mounted for kernel header downloading.

## 2.36.8

* Add `datadog.clusterName` on clusterCheckRunner pods

## 2.36.7

* Add `priorityPreemptionPolicyValue` as a configurable value on the Agent charts

## 2.36.6

* Fix GKE Autopilot installation. The `process-agent` command must
  use the `-config` argument to be compliant with the Datadog Agent's
  GKE Autopilot security profile.

## 2.36.5

* Use `regexFind` in favor of `mustRegexFind` to support helm2.

## 2.36.4

* Support `commonlabels` configuration to be able to add common labels on all resources created by the chart.

## 2.36.3

* Fix usage of deprecated command flags in the process-agent.

## 2.36.2

* Documentation updates to comments in some agent templates

## 2.36.1

* Add `datadog.otlp` section to configure OTLP ingest.

## 2.36.0

* Default Datadog Agent image to `7.37.1`.
* Default Datadog Cluster Agent image to `1.21.0`.

## 2.35.6

* Fix `include` in clusterchecks deployment template.

## 2.35.5

* Allow cross-DCA communication in DCA `NetworkPolicy` and `CiliumNetworkPolicy`

## 2.35.4

* Fix comments in `values.yaml` to allow a seamless `helm-docs` update.

## 2.35.3

* Add `openat2` to system-probe seccomp profile to fix issues with opening files.

## 2.35.2

* Update RBACs and the default check configuration to collect ingress metrics in Kube State Metrics Core.
  Note: Ingress metrics collection requires Cluster Agent 1.21+.

## 2.35.1

* Fix Cluster-Agent SCC creation on openshift 3.x.

## 2.35.0

* The Admission Controller is now enabled by default.

## 2.34.6

* Avoid the error `<eq .Values.clusterAgent.admissionController.configMode "service">: error calling eq: incompatible types for comparison` that can happen in older helm versions.

## 2.34.5

* Add `datadog.securityAgent.runtime.fimEnabled` configuration to enable CWS File Integrity Monitoring.

## 2.34.4

* Add `clusterAgent.admissionController.failurePolicy` configuration to set the failure policy for dynamic admission control

## 2.34.3

* Introduce `clusterAgent.admissionController.configMode` (requires Cluster Agent `1.20+`). It allows choosing the kind of configuration to be injected ("hostip", "service", or "socket").

## 2.34.2

* Default Cluster Agent image to `1.20.0`.

## 2.34.1

* Add the `datadog.secretBackend.enableGlobalPermissions` value, which when set to `false`, does not allow Datadog agents to read all secrets in all clusters. Defaults to `true`.
* Add the `datadog.secretBackend.roles` value,  which creates `Role` and `RoleBinding` for each namespace defined. Allows for opt-in read permissions for secrets in those namespaces.

## 2.34.0

* Default Datadog Agent image to `7.36.1`.

## 2.33.8

* Add `datadog.securityAgent.runtime.network.enabled` configuration to enable CWS network events.

## 2.33.7

* Fix inaccurate documentation example for `datadog.kubeStateMetricsCore.labelsAsTags`.

## 2.33.6

* Add `renameat2` to system-probe seccomp profile to fix issues with renaming files.

## 2.33.5

* Make the DCA leader election ConfigMap name depend on the Helm release name. (Requires DCA 1.21+)

## 2.33.4

* Improves help message when only `.datadog.containerInclude` is defined but no `.datadog.containerExclude`

## 2.33.3

* Add enableKernelHeaderDownload configuration option to system-probe.

## 2.33.2

* Add `revisionHistoryLimit` to set the number of old ReplicaSets in the Deployment.

## 2.33.1

* Default Datadog Agent image to `7.35.2`.

## 2.33.0

***Warning:*** From this version onwards, on GKE Autopilot, only one "datadog" Helm chart release is allowed by Kubernetes namespace due to the following new constraints:

* On GKE Autopilot, hardcode the "Agent" DaemonSet serviceAccountName.
* On GKE Autopilot, hardcode the "Install Info" ConfigMap name.

## 2.32.6

* Add `verticalpodautoscalers` in `kubernetes_state_core.yaml.default` to enable collection in KSM Core by default

## 2.32.5

* Fix process detection, by adding `kill` syscall with signal `0` to system-probe seccomp profile.

## 2.32.4

* Update `cluster-agent` image to the latest stable version: `1.19.0`

## 2.32.3

* Fix Go CPU profiling, by adding `setitimer` to system-probe seccomp profile.

## 2.32.2

* Fix scheduling of Helm check due to missing `helm.yaml` in Cluster Agent `confd`.

## 2.32.1

* Remove usage of `concat` to restore compatibility with Helm2.

## 2.32.0

* Default Datadog Agent image to `7.35.0`.

## 2.31.1

* Improves how securityContext are set depending on the `targetSystem` option (fix #590).

## 2.31.0

* Add `datadog.prometheusScrape.version` parameter to choose the version of the openmetrics check that the Prometheus auto-discovery should instantiate by default.
  It now defaults to `2`, which requires an agent 7.34+.
  It can be explicitely set to `1` to restore the behaviour of previous versions.

## 2.30.21

* Add `datadog.kubelet.podLogsPath` to customize hostPath mounted in to get Kubernetes PODs logs.

## 2.30.20

* Update "agents are spinning up" message to point towards the new Events Explorer

## 2.30.19

* Update documentation for enabling NPM.

## 2.30.18

* Enforce use of `root` user for the node agent.

## 2.30.17

* Add `datadog.helmCheck.collectEvents` to enable event collection in the Helm check.

## 2.30.16

* Default Datadog CRD chart to `0.4.7`.

## 2.30.15

* Default Datadog Agent image to `7.34.0`.
* Default Datadog Cluster-Agent image to `1.18.0`.

## 2.30.14

* Default Datadog Agent image to `7.33.1`.

## 2.30.13

* Feat: Add `shareProcessNamespace` parameter.

## 2.30.12

* Add an option to remove the container runtime socket access.

## 2.30.11

* Fix CiliumNetworkPolicy: Allow sending support flares.

## 2.30.10

* Fix scheduling of Helm check. It's no longer scheduled on a daemonset agent.

## 2.30.9

* Add RBAC rules for Roles, RoleBindings, ClusterRoles, ClusterRoleBindings and ServiceAccounts in order to collect them in the Orchestrator Explorer from the Cluster-agent.

## 2.30.8

* Add option to enable Helm Check (requires Agent 7.35.0+ and Cluster Agent 1.19.0+).

## 2.30.7

* Add ingress RBAC rules for the Cluster Agent to collect ingress resources in the Orchestrator Explorer. (Feature available starting Cluster Agent v1.19)

## 2.30.6

* Fix syntax of agents.podAnnotations to be aligned with other podAnnotations setting.

## 2.30.5

* Add a new note to recommand to the Cluster Agent in HA mode when the `admission-controller` or the `metrics provider` are enabled.

## 2.30.4

* Add PV and PVC RBAC rules for the Cluster Agent in order to collect new resources in the Orchestrator Explorer.

## 2.30.3

* Add `datadog.logs.autoMultiLineDetection` parameter to setup automatic multi-line log detection
  See <https://docs.datadoghq.com/agent/logs/advanced_log_collection/?tab=configurationfile#automatic-multi-line-aggregation>
  This new option requires an agent 7.32+.

## 2.30.2

* rename the APM port in the local traffic policy service from `apm` to `traceport`

## 2.30.1

* clusterAgent.tolerations documented in values.yaml

## 2.30.0

* Default Datadog Agent image to `7.33.0`.
* Default Datadog Cluster-Agent image to `1.17.0`.

## 2.29.0

* Add `agents.podSecurity.allowedUnsafeSysctls` parameter

## 2.28.15

* Remove unused configuration option from system_probe.yaml to address error message: `Unknown key in config file: runtime_security_config.debug`

## 2.28.14

* Update cluster-agent's podAntiAffinity from required to preferred

## 2.28.13

* Do not declare the volumes for `/etc/*-release` if there is no `system-probe`.
  Only the `system-probe` container mounts them.

## 2.28.12

* Fix some typos in comments

## 2.28.11

* Fix deprecation warning in examples caused by the `datadog.apm.enabled` parameter

## 2.28.10

* Update confd examples for the mysql integration

## 2.28.9

* Fix Cluster-Agent SCC creation on openshift 3.x. : remove unset parameters.

## 2.28.8

* Fix `PodDisruptionBudget` api version definition when using `helm template`.

## 2.28.7

* Fix environment variables to be quoted correct with a loop and `quote` instead of `toYaml`.

## 2.28.6

* Update `PodDisruptionBudget` api version to get rid of `policy/v1beta1 PodDisruptionBudget is deprecated in v1.21+, unavailable in v1.25+; use policy/v1 PodDisruptionBudget` warning.

## 2.28.5

* Default Datadog Agent image to `7.32.4`.

## 2.28.4

* Add a new configuration section `datadog.secretBackend`.
* Configuring `datadog.secretBackend.command="/readsecret_multiple_providers.sh"` will add the secret permissions required by the `/readsecret_multiple_providers.sh` helper.

## 2.28.3

* Update `agents.podSecurity.capabilities` to contain all `agents.containers.systemProbe.securityContext.capabilities`.

## 2.28.2

* Fix conflict between `clusterAgent.confd` and `clusterAgent.advancedConfd`: merge the 2 ConfigMaps.

## 2.28.1

* Fix `CAP_CHOWN` capability configuration for system-probe.

## 2.28.0

* Create priority Class to better support environments such as GKE Autopilot.

## 2.27.10

* Add `CAP_CHOWN` to the list of capabilities for system-probe.

## 2.27.9

* Adds `systemProbe.enableRuntimeCompiler`, `systemProbe.mountPackageManagementDirs` and `systemprobe.runtimeCompilationAssetDir` to configure eBPF runtime compiler in the system-probe.
* Adds `systemProbe.mountPackageManagementDirs` to configure what volumes are mounted in the system-probe for runtime compilation.
* Adds `systemProbe.osReleasePath` to configure what volume is mounted in the system-probe for host OS detection.
* Adds renameat, symlinkat and flock to the allow syscalls in the system-probe's seccomp profile.

## 2.27.8

* Default Datadog Agent image to `7.32.3`.

## 2.27.7

* Nothing

## 2.27.6

* Default Datadog Agent image to `7.32.2`.

## 2.27.5

* Fix bugs that prevented running the ksm core check as a cluster check.

## 2.27.4

* Do not allow unsupported configs with the security agent in windows environments.
* Ensure autoconf/extra config files are mounted in windows environments.

## 2.27.3

* Fix CiliumNetworkPolicy: Update toFQDNs policy to include `agent-http-intake` endpoint.
* Fix CiliumNetworkPolicy: Update toFQDNs to include `api` endpoint.

## 2.27.2

* Expose the `labels_as_tags` parameter of the KSM core check.
  This parameter exists only in agent 7.32.0 and above and cluster-agent 1.16.0 and above.

# 2.27.1

* Update README.md to clarify Helm 2 vs. Helm 3 instructions.
* Fix typos in README.md in `How to join a Cluster Agent from another helm chart deployment (Linux)`.
* Fixes a port number typo for the `datadog.apm.portEnabled` option from 8216 to 8126.

# 2.27.0

* Introduce `processAgent.processDiscovery` to configure `DD_PROCESS_AGENT_DISCOVERY_ENABLED`

## 2.26.5

* Add `verticalpodautoscalers` RBACs when `datadog.kubeStateMetricsCore.enabled` is `true`

## 2.26.4

* Update API/APP keys secret management documentation.

## 2.26.3

* Update CRDs version to `0.4.5` (reduced size)

## 2.26.2

* Add support for Universal Service Monitoring (currently under private Beta)

## 2.26.1

* Update CRDs version to `0.4.4`

## 2.26.0

* Default Datadog Agent image to `7.32.1`.

## 2.25.0

* Adding the following `agents.daemonsetAnnotations`, `clusterAgent.deploymentAnnotation` and `clusterChecksRunner.deploymentAnnotations` parameters to allow custom annotations on the agent's deployments/daemonsets to be setup

## 2.24.1

* Fix typo in variable name : `agents.localService.forceLocalServiceEnabled`

## 2.24.0

* Default Datadog Agent image to `7.32.0`.
* Default Datadog Cluster Agent image to `1.16.0`.

## 2.23.6

* Add `datadog.expvarPort` parameter to customize the default expvar default port to not conflict with the default clusteragent metrics port if running in hostNetwork mode.
* Defined cluster-agent containerPort `agentmetrics` to expose the default port, which is set to 5000 and already defined in the `NetworkPolicy` for the cluster-agent.

## 2.23.5

Change OpenShift SCC priorities from 10 to 8 to avoid conflicts with OpenShift Auth operator.

## 2.23.4

* Add a new configuration field `datadog.providers.eks.ec2.useHostnameFromFile` to allow use of host's `/var/lib/cloud/data/instance-id` for hostname detection.

## 2.23.3

* Add `agents.localService` parameters to customize the internal traffic policy service name and force its creation of Kubernetes 1.21.

## 2.23.2

* Add an `agents.podSecurity.defaultApparmor` setting to allow customizing the default AppArmor profile used by all containers but `system-probe`.

## 2.23.1

* Fix APM reporting via `trace-agent` hostPort if `datadog.apm.enabled: true`.

## 2.23.0

* Add new option to the Kubernetes State Metrics Core feature to run the Cluster Check on Cluster Check Workers. This option is meant to be leveraged in large clusters.

## 2.22.18

* Do not configure `trace-agent` hostPort if `datadog.apm.portEnabled: false`.

## 2.22.17

* Update general installation documentation and add how to disable APM.

## 2.22.16

* Support containerd on windows node with logs enabled.

## 2.22.15

* Add a new configuration field `datadog.kubeStateMetricsCore.collectSecretMetrics` to allow disabling the collection of `kubernetes_state.secret.*` metrics by the `kubernetes_state_core` check.

## 2.22.14

* Apply security context capabilities to security-agent only if compliance is enabled.

## 2.22.13

* Add configurable conntrack_init_timeout to sysprobe config.

## 2.22.12

* Replace the `prometheus` check targetting the Datadog Cluster Agent by the new `datadog_cluster_agent` integration. (Requires Datadog Agent 7.31+)

## 2.22.11

* Adds missing configuration option `DD_STRIP_PROCESS_ARGS` for the process agent.

## 2.22.10

* Default Datadog Agent image to `7.31.1`.
* Default Datadog Cluster Agent image to `1.15.1`.

## 2.22.9

* Makes the runtime socket configurable when running on Windows instead of defaulting to `\\.\pipe\docker_engine`.

## 2.22.8

* Add a service with local [internal traffic policy](https://kubernetes.io/docs/concepts/services-networking/service-traffic-policy/) for traces and dogstatsd.
  This works only on Kubernetes 1.22 or more recent.

## 2.22.7

* Add a default required pod anti-affinity for the cluster agent.

## 2.22.6

* Adds missing configuration option for `DD_KUBERNETES_NAMESPACE_LABELS_AS_TAGS`.

## 2.22.5

* Add support for using `envFrom` on all container definitions.

## 2.22.4

* Cluster Agent: `DD_TAGS` are included even when Datadog is not set as metrics provider.

## 2.22.3

* CiliumNetworkPolicy: Grant access to the agent to ECS container agent via localhost.

## 2.22.2

* Bind mount host /etc/os-release in system probe container.

## 2.22.1

* Fix CiliumNetworkPolicy `port` field.

## 2.22.0

* Default Datadog Agent image to 7.31.0.
* Default Datadog Cluster Agent image to 1.15.0.

## 2.21.5

* Update descriptions for securityAgent configuration.

## 2.21.4

* Fix condition for including `sysprobe-socket-dir` and `sysprobe-config` volume mounts for `agent`.

## 2.21.3

* Default Datadog Agent image to 7.30.1.

## 2.21.2

* Fix Dogstatsd UDS socket configuration with a HostVolume when `useSocketVolume: true`.

## 2.21.1

* Disable by default UDS socket for dogstastd and apm on GKE autopilot.

## 2.21.0

* Enable APM by default with using a Unix Domain socket for communication.

## 2.20.4

* Skip KSM network policy creation when KSM creation is disabled.

## 2.20.3

* Add `agents.image.tagSuffix` and `clusterChecksRunner.image.tagSuffix` to be able to request JMX or Windows servercore images without having to explicitly specify the full version.

## 2.20.2

* Add an additional way to configure cluster check allowing multiple configs for the same check.

## 2.20.1

* Add Statefulsets RBAC rules for the Cluster Agent in order to collect new resources in the Orchestrator Explorer.

## 2.20.0

* Update default Agent image tag to `7.30.0`
* Update default Cluster-Agent image tag to `1.14.0`

## 2.19.9

* Print a configuration notice to clarify the containers filtering behavior when a misconfiguration is detected.

## 2.19.8

* Update `datadog-crds` to `0.3.2`.

## 2.19.7

* Fix test value files in datadog/ci directory.

## 2.19.6

* Update `agent` image tag to `7.29.1`.
* Update `clusterChecksRunner` image tag to `7.29.1`.

## 2.19.5

* Update link toe `kube-state-metrics` in README.md.

## 2.19.4

* Fix `runtimesocket` volumeMount for the `trace-agent` on windows deployment.

## 2.19.3

* Fix condition defining `should-enable-k8s-resource-monitoring`, which toggles the orchestrator explorer feature.

## 2.19.2

* Fix `dsdsocket` volumeMount for the `trace-agent` on windows deployment.

## 2.19.1

* Fix chart release process after updating the `kube-state-metrics` chart registry.

## 2.19.0

* Move to the new `kube-state-metrics` chart registry, but keep the version `2.13.2`.

## 2.18.2

* Update `kube-state-metrics` requirement chart documentation.
* Add missing `DD_TAGS` envvar in `cluster-agent` deployment (Fix #304).

## 2.18.1

* Honor `doNotCheckTag` in Env AD detection, preventing install failures with custom images using non semver tags.

## 2.18.0

* Configure and activate the Dogstatsd UDS socket in an "emptyDir" volume by default. It will allow JMX-Fetch to use UDS by default.

## 2.17.1

* Update `cluster-agent` image tag to `1.13.1`.

## 2.17.0

* Update `agent` image tag to `7.29.0`.
* Update `cluster-agent` image tag to `1.13.0`.

## 2.16.6

* Support template expansion for `clusterAgent.podAnnotations`
* Support template expansion for `clusterAgent.rbac.serviceAccountAnnotations`

## 2.16.5

* Remove other way of detecting OpenShift cluster as it's not supported by Helm2.

## 2.16.4

* Rename the `Role` and `RoleBinding` of the Datadog Cluster Agent to avoid edge cases where `helm upgrade` can fail because of object name conflict.

## 2.16.3

* Add Daemonsets RBAC rules for the Cluster Agent in order to collect new resources in the Orchestrator Explorer.

## 2.16.2

* Document Autodiscovery management parameters: `datadog.containerExclude`, `datadog.containerInclude`, `datadog.containerExcludeMetrics`, `datadog.containerIncludeMetrics`, `datadog.containerExcludeLogs` and `datadog.containerIncludeLogs`.
* Introduce `datadog.includePauseContainer` to control autodiscovery of pause containers.
* Introduce a deprecation noticed for the undocumented and long deprecated `datadog.acInclude` and `datadog.acExclude`.

## 2.16.1

* Use the pod name as cluster check runner ID to allow deploying multiple cluster check runners on the same node. (Requires agent 7.27.0+)

## 2.16.0

* Always mount `/var/log/containers` for the Datadog Agent to better handle logs file scanning with short-lived containers. (See [datadog-agent#8143](https://github.com/DataDog/datadog-agent/pull/8143))

## 2.15.6

* Set `GODEBUG=x509ignoreCN=0` to revert Agent SSL certificates validation to behaviour to Golang <= 1.14. Notably it fixes issues with Kubelet certificates on AKS with Agent >= 7.28.

## 2.15.5

* Add RBAC rules for the Cluster Agent in order to collect new resources in the Orchestrator Explorer.

## 2.15.4

* Bump Agent version to `7.28.1`.

## 2.15.3

* Fix Cilium network policies.

## 2.15.2

* OpenShift: Automatically use built-in SCCs instead of failing if create SCC option is not used

## 2.15.1

* Add parameter `clusterAgent.rbac.serviceAccountAnnotations` for specifying annotations for dedicated ServiceAccount for Cluster Agent.
* Add parameter `agents.rbac.serviceAccountAnnotations` for specifying annotations for dedicated ServiceAccount for Agents.
* Support template expansion for `agents.podAnnotations`

## 2.15.0

* Bump Agent version to `7.28.0`.

## 2.14.0

* Improve resources labels with kubermetes/helm standard labels.

## 2.13.3

* Add `datadog.checksCardinality` field to configure `DD_CHECKS_TAG_CARDINALITY`.
* Add a reminder to set the `datadog.site` field if needed.

## 2.13.2

* Fix `YAML parse error on datadog/templates/daemonset.yaml` when autopilot is enabled.
* Fix "README.md" generation.

## 2.13.1

* Fix Kubelet connection on GKE-autopilot environment: force `http` endpoint to retrieves pods information.

## 2.13.0

* Update `kube-state-metrics` chart version to `2.13.2` that include `kubernetes/kube-state-metrics#1442` fix for `helm2`.

## 2.12.4

* Fix missing namespaces in chart templates

## 2.12.3

* Added `datadog.ignoreAutoConfig` config option to ignore `auto_conf.yaml` configurations.

## 2.12.2

* The Datadog Cluster Agent's Admission Controller now uses a `Role` to watch secrets instead of a `ClusterRole`. (Requires Datadog Cluster Agent v1.12+)

## 2.12.1

* Add more kube-state-metrics core check documentation

## 2.12.0

* Update the Cluster Agent version to `1.12.0`
* Support kube-state-metrics core check (Requires Datadog Cluster Agent v1.12+)

## 2.11.6

* Improve support for environment autodiscovery by removing explicit setting of `DOCKER_HOST` by default with Agent 7.27+.
Starting Agent 7.27, the recommended setup is to never set `datadog.dockerSocketPath` or `datadog.criSocketPath`, except if your setup is using non-standard paths.

## 2.11.5

* Remove comment in the `seccomp` json profile, which is break the json parsing.

## 2.11.4

* Add missing system calls to system-probe `seccomp` profile.

## 2.11.3

* Update the documentation with the new path of the `kube-state-metrics` chart

## 2.11.2

* Update `agent.customAgentConfig` config example in the `values.yaml`: removes reference to APM configuration.

## 2.11.1

* Enable `collectDNSStats` by default

## 2.11.0

* Bump Agent version to `7.27.0`.
* Support configuring advanced openmetrics check parameters via `datadog.prometheusScrape.additionalConfigs`.

## 2.10.14

* Add Kubelet `hostCAPath` and `agentCAPath` parameters to automatically mount and use CA cert from host filesystem for Kubelet connection.
* Fix default value for DCA hostNetwork

## 2.10.13

* Fix `security-agent-feature` helper function to support `helm2`.
* Fix `provider-labels` helper function to support `helm2`.
* Fix `provider-env` helper function to support `helm2`.

## 2.10.12

* Add the possibility to specify securityContext for cluster-agent containers

## 2.10.11

* Fix RBAC needed for the external metrics provider for the future release of the DCA.

## 2.10.10

* Fix system-probe version check when using `datadog.networkMonitoring.enabled`

## 2.10.9

* Add the possibility to specify a priority class name for the cluster checks runner pods.

## 2.10.8

* When node agents are joining an existing DCA managed by another Helm release, we must control if they should be eligible to cluster checks dispatch or not depending on whether CLC have been deployed with the external DCA.

## 2.10.7

* Fix bug regarding using "Metric collection with Prometheus annotations".

## 2.10.6

* Add provider labels on pods, warning on dogstatsd with UDS on GKE Autopilot.

## 2.10.5

* Increase default `datadog.systemProbe.maxTrackedConnections` to 131072.

## 2.10.4

* Fix several bugs with OpenShift SCC and hostNetwork.

## 2.10.3

* Bump version of KSM chart to get rid of `rbac.authorization.k8s.io/v1beta1 ClusterRole is deprecated in v1.17+, unavailable in v1.22+; use rbac.authorization.k8s.io/v1` warnings

## 2.10.2

* Use an EmptyDir volume shared between all the agents for logs so that `agent flare` can gather the logs of all of them.

## 2.10.1

* Remove the cluster-id configmap mount for process-agent. (Requires Datadog Agent 7.25+ and Datadog Cluster Agent 1.11+, otherwise collection of pods for the Kubernetes Resources page will fail).

## 2.10.0

* Remove the cluster-id configmap mount for process-agent. (Requires Datadog Agent 7.26+ and Datadog Cluster Agent 1.11+, otherwise collection of pods for the Kubernetes Resources page will fail).

## 2.9.11

* Allow system-probe container to send flares by adding main agent config file to container.

## 2.9.10

* Support configuring Prometheus Autodiscovery. (Requires Datadog Agent 7/6.26+ and Datadog Cluster Agent 1.11+).

## 2.9.9

* Update "agent" image tag to `7.26.0` and "cluster-agent" to `1.11.0`.
* Fix nit comments

## 2.9.8

* Make pod collection for the Kubernetes Explorer work with an external Cluster Agent deployment.

## 2.9.7

* Allow cluster-agent to override metrics provider endpoint with `clusterAgent.metricsProvider.endpoint`.

## 2.9.6

* Add missing `NET_RAW` capability to `System-probe` to support `CVE-2020-14386` mitigation.

## 2.9.5

* Fix typo in variable name. `agents.podSecurity.capabilities` replaces `agents.podSecurity.capabilites`.

## 2.9.4

* Remove uses of `systemProbe.enabled`.

## 2.9.3

* Enable support for GKE Autopilot.

## 2.9.2

* Fixed a bug where `datadog.leaderElection` would not configure the cluster-agent environment variable `DD_LEADER_ELECTION` correctly.

## 2.9.1

* add `datadog.systemProbe.conntrackMaxStateSize` and  `datadog.systemProbe.maxTrackedConnections`.

## 2.9.0

* Remove `systemProbe.enabled` config param in favor of `networkMonitoring.enabled`, `securityAgent.runtime.enabled`, `systemProbe.enableOOMKill`, and `systemProbe.enableTCPQueueLength`.
* Fix bug preventing network monitoring to be disabled by setting `datadog.networkMonitoring.enabled` to `false`.

## 2.8.6

* Add support for Service Topology to target the Datadog Agent via a kubernetes service instead of host ports. This will allow sending traces and custom metrics without using host ports. Note: Service Topology is a new Kubernetes feature, it's still in alpha and disabled by default.

## 2.8.5

* Allow `namespaces` in RBAC for `kubernetes_namespace_labels_as_tags`.

## 2.8.4

* Grant access to the `Lease` objects.
  `Lease` objects can be read by the `kube_scheduler` and `kube_controller_manager` checks on agent 7.27+ on Kubernetes clusters 1.14+.

## 2.8.3

* Fix potential duplicate `DD_KUBERNETES_KUBELET_TLS_VERIFY` env var due to new parameter `kubelet.tlsVerify`. Parameter has now 3 states and env var won't be added if not set, improving backward compatibility.
* Fix activation of Cluster Checks while Cluster Agent is disabled.
* Change default value for `clusterAgent.metricsProvider.useDatadogMetrics` from `true` to `false` as it may trigger CRD ownership issues in several situations.

## 2.8.2

* Open port 5000/TCP for ingress on cluster agent for Prometheus check from the agent.

## 2.8.1

* Fix `datadog.kubelet.tlsVerify` value when set to `false`

## 2.8.0

* Enable the orchestrator explorer by default.

## 2.7.2

* Add a new fields `datadog.kubelet.host` (to override `DD_KUBERNETES_KUBELET_HOST`) and `datadog.kubelet.tlsVerify` (to toggle kubelet TLS verification)

## 2.7.1

* Open port 8000/TCP for ingress on cluster agent for Admission Controller communication.

## 2.7.0

* Changes default values to activate a maximum of built-in features to ease configuration.
  Notable changes:
  * Cluster Agent, cluster checks and event collection are activated by default
  * DatadogMetrics CRD usage is activated by default if ExternalMetrics are used
  * Dogstatsd non-local traffic is activated by default (hostPort usage is not)
* Bump Agent version to `7.25.0` and Cluster Agent version to `1.10.0`
* Introduce `.registry` parameter to quickly change registry for all Datadog images. Image name is retrieved from `.image.name`, however setting `.image.repository` still allows to override per image, ensuring backward compatibility

## 2.6.15

* Add `ports` options to all Agent containers to allow users to add any binding they'd like for integrations

## 2.6.14

* Opens port 6443/TCP on kube-state-metrics netpol.

## 2.6.13

* Opens ports 6443/TCP and 53/UDP for egress on cluster agent.
* Adds PodSecurityPolicy support for Cluster Agents.

## 2.6.12

* Mount `/etc/passwd` as `readOnly` in the `process-agent`.

## 2.6.11

* Adds `unconfined` as a default value for `agents.podSecurity.apparmorProfiles`. It now aligns with `datadog.systemProbe.apparmor` default value.
* Updates `hostPID` for PodSecurityPolicy, bringing it in line with SCC.

## 2.6.10

* Allow cluster-agent to access apps/daemonsets when admissionController is enabled.

## 2.6.9

* Add `/tmp` in Agent POD as an emptyDir to allow VOLUME removal from Agent Dockerfile
* Clarify documentation of `datadog.dogstatsd.nonLocalTraffic`

## 2.6.8

* Fix `helm lint` by renaming YAML files lacking metadata info.

## 2.6.7

* Change the default agent version to `7.24.1`

## 2.6.6

* Add `agents.containers.systemProbe.securityContext` option.

## 2.6.5

* Make sure all agents are rolled out on API key update and the Cluster agents on Application key update.

## 2.6.4

* Fix agent container volumeMounts when oom kill check or tcp queue length check is enabled.

## 2.6.3

* Add a new field `datadog.dogstatsd.tags` to configure `DD_DOGSTATSD_TAGS`.

## 2.6.2

* Make sure KSM deploys on Linux nodes

## 2.6.1

* Fix `process-agent` and `trace-agent` communication with the `cluster-agent`: When the `cluster-agent` is activated,
  the agents should communicated with the `cluster-agent` to retrived tags like `kube_service` instead of communicating
  directly with the Kubernetes API-Server.

## 2.6.0

* deprecates `systemProbe.enabled` in favor of `networkMonitoring.enabled`, `securityAgent.runtime.enabled`, `systemProbe.enableOOMKill`, and `systemProbe.enableTCPQueueLength`.
* fixes a bug where network performance monitoring would be enabled if any systemProbe feature was enabled.

## 2.5.5

* Add CiliumNetworkPolicy

## 2.5.4

* Supports `clusterChecksRunner` pod annotations

## 2.5.3

* Add "datadog-crds" chart as dependency. It is used to install the `DatadogMetrics` CRD if needed.

## 2.5.2

* Change `datadog.tags` to a `tpl` value

## 2.5.0

* Use `gcr.io` instead of Dockerhub
* Change the default agent version `7.23.1`
* Change the default cluster agent version `1.9.1`
* Change the default cluster checks runner version `7.23.1`

## 2.4.39

* Fixed a bug where `networkMonitoring.enabled` would not configure the process-agent correctly, causing network data to not be reported.

## 2.4.38

* Move the kube-state-metrics subchart from google's helm registry to charts.helm.sh/stable.

## 2.4.37

* Fix incorrect link for Event Collection in `values.yaml`.

## 2.4.36

* Fix `should-enable-system-probe` helper function to support `helm2`.

## 2.4.35

* Add options to set pod and container securityContext

## 2.4.34

* Add `datadog.networkMonitoring` section to allow the system-probe to be run without network performance monitoring. Deprecates `systemProbe.enabled`.

## 2.4.33

* Introduce overall cluster-name limit of 80
* Remove character limit of single parts of the cluster-name

## 2.4.32

* The `agents.volumeMounts` option is now properly propagated to all agent containers.

## 2.4.31

* Support adding labels to the Agent pods and daemonset via `agents.additionalLabels`.
* Support adding labels to the Cluster Agent pods and deployment via `clusterAgent.additionalLabels`.
* Support adding labels to the Cluster Checks Runner pods and deployment via `clusterChecksRunner.additionalLabels`.

## 2.4.30

* Refactor liveness and readiness probes with helpers to allow user overrides with other types of probes or disabling
  probes entirely.
* Introduce `clusterChecksRunner.healthPort` default setting.
* Use health port defaults instead of hardcoded values.

## 2.4.29

* Add `common-env-vars` to `system-probe` container

## 2.4.28

* Make sure we rollout Agent/CLC/DCA when an upgrade is done (thus triggering a change in token secret)

## 2.4.27

* Remove port defaults from liveness/readiness probes and show error notices on misconfiguration if user overrides are supplying custom node settings.

## 2.4.26

* Revert to Helm2 hash in `requirements.yaml` to retain compatibility with Helm 2

## 2.4.25

* Update default `datadog/agent` image tag to `7.23.0`
* Update default `datadog/cluster-agent` image tag to `1.9.0`

## 2.4.24

* Fix the Cluster Agent's network policy (allow ingress from node Agents)
* Add kube-state-metrics network policy

## 2.4.23

* Add `datadog.envFrom` parameter to support passing references to secrets and/or configmaps for environment
variables, instead of passing one by one.

## 2.4.22

* Add automatic README.md generation from `Values.yaml`

## 2.4.21

* Change `securityContext` variable name to `seLinuxContext` allow setting the PSP/SCC seLinux `type` or `rule`. Backward compatible.

## 2.4.20

* Add NetworkPolicy ingress rules for dogstatsd and APM

## 2.4.19

* Add NetworkPolicy
  Add the following parameters to control the creation of NetworkPolicy:
  * `agents.networkPolicy.create`
  * `clusterAgent.networkPolicy.create`
  * `clusterChecksRunner.networkPolicy.create`
  The NetworkPolicy managed by the Helm chart are designed to work out-of-the-box on most setups.
  In particular, the agents need to connect to the datadog intakes. NetworkPolicy can be restricted
  by IP but the datadog intake IP cannot be guaranteed to be stable.
  The agents are also susceptible to connect to any pod, on any port, depending on the "auto-discovery" annotations
  that can be dynamically added to them.

## 2.4.18

* Fix `config` volume not being mounted in clusterChecksRunner pods.

## 2.4.17

* Update default `Agent` and `Cluster-Agent` image tags: `7.22` and `1.18`.

## 2.4.16

* Add `External Metric` Aggregator config on Chart.

## 2.4.15

* Add `agents.podSecurity.apparmor.enabled` flag (defaulted to `true`).

## 2.4.14

* Fix external metrics on GKE due to Google fix on recent versions (introduced in 2.4.1).

## 2.4.13

* fix Agent `PodSecurityPolicy` with `hostPorts` definition, and missing RBAC.

## 2.4.12

* Add `compliance` and `runtime` `security-agent` support.

## 2.4.11

* Add `NET_BROADCAST` capability for `system-probe`.

## 2.4.10

* Add `scrubbing` option for helm charts to "Orchestrator Explorer" support.

## 2.4.9

* Add `DD_DOGSTATSD_TAG_CARDINALITY` capability.

## 2.4.8

* Fix, Only try to mount `/lib/modules` and `/usr/src` when needed.

## 2.4.7

* Add `eventfd` and `eventfd2` to allowed syscalls for `system-probe`.

## 2.4.6

* Fix Windows deployment support (fixes #15).

## 2.4.5

* Add mount propagation option for `hostVolumes`.

## 2.4.4

* Fix typo in `allowHostPorts`.
* Add support of `MustRunAs` in Agent `PodSecurityPolicy` and `SecurityContextConstraints`.

## 2.4.3

* Fix `Cluster-Agent` RBAC to collect new resources for the "Orchestrator Explorer" support.

## 2.4.2

* Add `install_info` file.

## 2.4.1

* Fix MetricsProvider RBAC setup on GKE clusters

## 2.4.0

* First release on github.com/datadog/helm-charts

## 2.3.41

* Fix issue with Kubernetes <= 1.14 and Cluster Agent's External Metrics Provider (must be 443)

## 2.3.40

* Update documentation for resource requests & limits default values.

## 2.3.39

* Propagate `datadog.checksd` to the clusterchecks runner to support custom checks there.

## 2.3.38

* Add support of DD\_CONTAINER\_{INCLUDE,EXCLUDE}\_{METRICS,LOGS}

## 2.3.37

* Add NET\_BROADCAST capability

## 2.3.36

* Bump default Agent version to `7.21.1`

## 2.3.35

* Add support for configuring the Datadog Admission Controller

## 2.3.34

* Add support for scaling based on `DatadogMetric` CRD

## 2.3.33

* Create new `datadog.podSecurity.securityContext` field to fix windows agent daemonset config.

## 2.3.32

* Always add os in nodeSelector based on `targetSystem`

## 2.3.31

* Fixed daemonset template for go 1.14

## 2.3.29

* Change the default port for the Cluster Agent's External Metrics Provider
  from 443 to 8443.
* Document usage of `clusterAgent.env`

## 2.3.28

* fix daemonset template generation if `datadog.securityContext` is set to `nil`

## 2.3.27

* add systemProbe.collectDNSStats option

## 2.3.26

* fix PodSecurityContext configuration

## 2.3.25

* Use directly .env var YAML block for all agents (was already the case for Cluster Agent)

## 2.3.24

* Allow enabling Orchestrator Explorer data collection from the process-agent

## 2.3.23

* Add the possibility to create a `PodSecurityPolicy` or a `SecurityContextConstraints` (Openshift) for the Agent's Daemonset Pods.

## 2.3.22

* Remove duplicate imagePullSecrets
* Fix DataDog location to useConfigMap in docs
* Adding explanation for metricsProvider.enabled

## 2.3.21

* Fix additional default values in `values.yaml` to prevent errors with Helm 2.x

## 2.3.20

* Fix process-agent <> system-probe communication

## 2.3.19

* Fix the container-trace-agent.yaml template creates invalid yaml when  `useSocketVolume` is enabled.

## 2.3.18

* Support arguments in the cluster-agent container `command` value

## 2.3.17

* grammar edits to datadog helm docs!
* Typo in log config

## 2.3.16

* Add parameter `clusterChecksRunner.rbac.serviceAccountAnnotations` for specifying annotations for dedicated ServiceAccount for Cluster Checks runners.
* Add parameters `clusterChecksRunner.volumes` and `clusterChecksRunner.volumeMounts` that can be used for providing a secret backend to Cluster Checks runners.

## 2.3.15

* Mount kernel headers in system-probe container
* Fix the mount of the `system-probe` socket in core agent
* Add parameters to enable eBPF based checks

## 2.3.14

* Allow overriding the `command` to run in the cluster-agent container

## 2.3.13

* Use two distinct health endpoints for liveness and readiness probes.

## 2.3.12

* Fix endpoints checks scheduling between agent and cluster check runners
* Cluster Check Runner now runs without s6 (similar to other agents)

## 2.3.11

* Bump the default version of the agent docker images

## 2.3.10

* Add dnsConfig options to all containers

## 2.3.9

* Add `clusterAgent.podLabels` variable to add labels to the Cluster Agent Pod(s)

## 2.3.8

* Fix templating errors when `clusterAgent.datadog_cluster_yaml` is being used.

## 2.3.7

* Fix an agent warning at startup because of a deprecated parameter

## 2.3.6

* Add `affinity` parameter in `values.yaml` for cluster agent deployment

## 2.3.5

* Add `DD_AC_INCLUDE` and `DD_AC_EXCLUDE` to all containers
* Add "Unix Domain Socket" support in trace-agent
* Add new parameter to specify the dogstatsd socket path on the host
* Fix typos in values.yaml
* Update "tags:" example in values.yaml
* Add "rate_limit_queries_*" in the datadog.cluster-agent prometheus check configuration

## 2.3.4

* Fix default values in `values.yaml` to prevent warnings with Helm 2.x

## 2.3.3

* Allow pre-release versions as docker image tag

## 2.3.2

* Update the DCA RBAC to allow it to create events in the HPA

## 2.3.1

* Update the example for `datadog.securityContext`

## 2.3.0

* Mount the directory containing the CRI socket instead of the socket itself
  This is to handle the cases where the docker daemon is restarted.
  In this case, the docker daemon will recreate its docker socket and,
  if the container bind-mounted directly the socket, the container would
  still have access to the old socket instead of the one of the new docker
  daemon.
  ⚠ This version of the chart requires an agent image 7.19.0 or more recent

## 2.2.12

* Adding resources for `system-probe` init container

## 2.2.11

* Add documentations around secret management in the datadog helm chart. It is to upstream
  requested changes in the IBM charts repository: <https://github.com/IBM/charts/pull/690#discussion_r411702458>
* update `kube-state-metrics` dependency
* uncomment every values.yaml parameters for IBM chart compliancy

## 2.2.10

* Remove `kubeStateMetrics` section from `values.yaml` as not used anymore

## 2.2.9

* Fixing variables description in README and Migration documentation (#22031)
* Avoid volumes mount conflict between `system-probe` and `logs` volumes in the `agent`.

## 2.2.8

* Mount `system-probe` socket in `agent` container when system-probe is enabled

## 2.2.7

* Add "Cluster-Agent" `Event` `create` RBAC permission

## 2.2.6

* Ensure the `trace-agent` computes the same hostname as the core `agent`.
  by giving it access to all the elements that might be used to compute the hostname:
  the `DD_CLUSTER_NAME` environment variable and the docker socket.

## 2.2.5

* Fix RBAC

## 2.2.4

* Move several EnvVars to `common-env-vars` to be accessible by the `trace-agent` #21991.
* Fix discrepancies migration-guide and readme reporded in #21806 and #21920.
* Fix EnvVars with integer value due to yaml. serialization, reported by #21853.
* Fix .Values.datadog.tags encoding, reported by #21663.
* Add Checksum to `xxx-cluster-agent-config` config map, reported by #21622 and contribution #21656.

## 2.2.3

* Fix `datadog.dockerOrCriSocketPath` helper #21992

## 2.2.2

* Fix indentation for `clusterAgent.volumes`.

## 2.2.1

* Updating `agents.useConfigMap` and `agents.customAgentConfig` parameter descriptions in the chart and main readme.

## 2.2.0

* Add Windows support
* Update documentation to reflect some changes that were made default
* Enable endpoint checks by default in DCA/Agent

## 2.1.2

* Fixed a bug where `DD_LEADER_ELECTION` was not set in the config init container, leading to a failure to adapt
config to this environment variable.

## 2.1.1

* Add option to enable WPA in the Cluster Agent.

## 2.1.0

* Changed the default for `processAgent.enabled` to `true`.

## 2.0.14

* Fixed a bug where the `trace-agent` runs in the same container as `dd-agent`

## 2.0.13

* Fix `system-probe` startup on latest versions of containerd.
  Here is the error that this change fixes:

  ```    State:          Waiting
      Reason:       CrashLoopBackOff
    Last State:     Terminated
      Reason:       StartError
      Message:      failed to create containerd task: OCI runtime create failed: container_linux.go:349: starting container process caused "close exec fds: ensure /proc/self/fd is on procfs: operation not permitted": unknown
      Exit Code:    128
   ```

## 2.0.11

* Add missing syscalls in the `system-probe` seccomp profile

## 2.0.10

* Do not enable the `cri` check when running on a `docker` setup.

## 2.0.7

* Pass expected `DD_DOGSTATSD_PORT` to datadog-agent rather than invalid `DD_DOGSTATD_PORT`

## 2.0.6

* Introduces `procesAgent.processCollection` to correctly configure `DD_PROCESS_AGENT_ENABLED` for the process agent.

## 2.0.5

* Honor the `datadog.env` parameter in all containers.

## 2.0.4

* Honor the image pull policy in init containers.
* Pass the `DD_CRI_SOCKET_PATH` environment variable to the config init container so that it can adapt the agent config based on the CRI.

## 2.0.3

* Fix templating error when `agents.useConfigMap` is set to true.
* Add DD\_APM\_ENABLED environment variable to trace agent container.

## 2.0.2

* Revert the docker socket path inside the agent container to its standard location to fix #21223.

## 2.0.1

* Add parameters `datadog.logs.enabled` and `datadog.logs.containerCollectAll` to replace `datadog.logsEnabled` and `datadog.logsConfigContainerCollectAll`.
* Update the migration document link in the `Readme.md`.

### 2.0.0

* Remove Datadog agent deployment configuration.
* Cleanup resources labels, to fit with recommended labels.
* Cleanup useless or unused values parameters.
* each component have its own RBAC configuration (create,configuration).
* container runtime socket update values configuration simplification.
* `nameOverride` `fullnameOverride` is now optional in values.yaml.
