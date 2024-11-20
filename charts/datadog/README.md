# Datadog

![Version: 3.81.0](https://img.shields.io/badge/Version-3.81.0-informational?style=flat-square) ![AppVersion: 7](https://img.shields.io/badge/AppVersion-7-informational?style=flat-square)

[Datadog](https://www.datadoghq.com/) is a hosted infrastructure monitoring platform. This chart adds the Datadog Agent to all nodes in your cluster via a DaemonSet. It also optionally depends on the [kube-state-metrics chart](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-state-metrics). For more information about monitoring Kubernetes with Datadog, please refer to the [Datadog documentation website](https://docs.datadoghq.com/agent/basic_agent_usage/kubernetes/).

Datadog [offers two variants](https://hub.docker.com/r/datadog/agent/tags/), switch to a `-jmx` tag if you need to run JMX/java integrations. The chart also supports running [the standalone dogstatsd image](https://hub.docker.com/r/datadog/dogstatsd/tags/).

See the [Datadog JMX integration](https://docs.datadoghq.com/integrations/java/) to learn more.

## How to use Datadog Helm repository

You need to add this repository to your Helm repositories:

```
helm repo add datadog https://helm.datadoghq.com
helm repo update
```

## Prerequisites

Kubernetes 1.10+ or OpenShift 3.10+, note that:

- the Datadog Agent supports Kubernetes 1.4+
- The Datadog chart's defaults are tailored to Kubernetes 1.10+, see [Datadog Agent legacy Kubernetes versions documentation](https://github.com/DataDog/datadog-agent/tree/main/Dockerfiles/agent#legacy-kubernetes-versions) for adjustments you might need to make for older versions

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://helm.datadoghq.com | datadog-crds | 1.7.2 |
| https://prometheus-community.github.io/helm-charts | kube-state-metrics | 2.13.2 |

## Quick start

By default, the Datadog Agent runs in a DaemonSet. It can alternatively run inside a Deployment for special use cases.

**Note:** simultaneous DaemonSet + Deployment installation within a single release will be deprecated in a future version, requiring two releases to achieve this.

### Installing the Datadog Chart

To install the chart with the release name `<RELEASE_NAME>`, retrieve your Datadog API key from your [Agent Installation Instructions](https://app.datadoghq.com/account/settings#agent/kubernetes) and run:

```bash
helm install <RELEASE_NAME> \
    --set datadog.apiKey=<DATADOG_API_KEY> datadog/datadog
```

By default, this Chart creates a Secret and puts an API key in that Secret.
However, you can use manually created secrets by setting the `datadog.apiKeyExistingSecret` and/or `datadog.appKeyExistingSecret` values (see [Creating a Secret](#create-and-provide-a-secret-that-contains-your-datadog-api-and-app-keys), below).

**Note:** When creating the secret(s), be sure to name the key fields `api-key` and `app-key`.

After a few minutes, you should see hosts and metrics being reported in Datadog.

**Note:** You can set your [Datadog site](https://docs.datadoghq.com/getting_started/site) using the `datadog.site` field.

```bash
helm install <RELEASE_NAME> \
    --set datadog.appKey=<DATADOG_APP_KEY> \
    --set datadog.site=<DATADOG_SITE> \
    datadog/datadog
```

#### Create and provide a secret that contains your Datadog API and APP Keys

To create a secret that contains your Datadog API key, replace the <DATADOG_API_KEY> below with the API key for your organization. This secret is used in the manifest to deploy the Datadog Agent.

```bash
DATADOG_API_SECRET_NAME=datadog-api-secret
kubectl create secret generic $DATADOG_API_SECRET_NAME --from-literal api-key="<DATADOG_API_KEY>"
```

**Note**: This creates a secret in the default namespace. If you are in a custom namespace, update the namespace parameter of the command before running it.

Now, the installation command contains the reference to the secret.

```bash
helm install <RELEASE_NAME> \
  --set datadog.apiKeyExistingSecret=$DATADOG_API_SECRET_NAME datadog/datadog
```

### Enabling the Datadog Cluster Agent

The Datadog Cluster Agent is now enabled by default.

Read about the Datadog Cluster Agent in the [official documentation](https://docs.datadoghq.com/agent/kubernetes/cluster/).

#### Custom Metrics Server

If you plan to use the [Custom Metrics Server](https://docs.datadoghq.com/agent/cluster_agent/external_metrics/?tab=helm) feature, provide a secret for the application key (AppKey) using the `datadog.appKeyExistingSecret` chart variable.

```bash
DATADOG_APP_SECRET_NAME=datadog-app-secret
kubectl create secret generic $DATADOG_APP_SECRET_NAME --from-literal app-key="<DATADOG_APP_KEY>"
```

**Note**: the same secret can store the API and APP keys

```bash
DATADOG_SECRET_NAME=datadog-secret
kubectl create secret generic $DATADOG_SECRET_NAME --from-literal api-key="<DATADOG_API_KEY>" --from-literal app-key="<DATADOG_APP_KEY>"
```

Run the following if you want to deploy the chart with the Custom Metrics Server enabled in the Cluster Agent:

```bash
helm install datadog-monitoring \
    --set datadog.apiKeyExistingSecret=$DATADOG_API_SECRET_NAME  \
    --set datadog.appKeyExistingSecret=$DATADOG_APP_SECRET_NAME \
    --set clusterAgent.enabled=true \
    --set clusterAgent.metricsProvider.enabled=true \
    datadog/datadog
```

If you want to learn to use this feature, you can check out this [Datadog Cluster Agent walkthrough](https://github.com/DataDog/datadog-agent/blob/main/docs/cluster-agent/CUSTOM_METRICS_SERVER.md).

The Leader Election is enabled by default in the chart for the Cluster Agent. Only the Cluster Agent(s) participate in the election, in case you have several replicas configured (using `clusterAgent.replicas`.

#### Cluster Agent Token

You can specify the Datadog Cluster Agent token used to secure the communication between the Cluster Agent(s) and the Agents with `clusterAgent.token`.

### Upgrading

#### From 2.x to 3.x

The migration from 2.x to 3.x does not require manual action.
As per the Changelog, we do not be guaranteeing support of Helm 2 moving forward.
If you already have the legacy Kubernetes State Metrics Check enabled, migrating will only show you the deprecation notice.

#### From 1.x to 2.x

⚠️ Migrating from 1.x to 2.x requires a manual action.

The `datadog` chart has been refactored to regroup the `values.yaml` parameters in a more logical way.
Please follow the [migration guide](https://github.com/DataDog/helm-charts/blob/main/charts/datadog/docs/Migration_1.x_to_2.x.md) to update your `values.yaml` file.

#### From 1.19.0 onwards

Version `1.19.0` introduces the use of release name as full name if it contains the chart name(`datadog` in this case).
E.g. with a release name of `datadog`, this renames the `DaemonSet` from `datadog-datadog` to `datadog`.
The suggested approach is to delete the release and reinstall it.

#### From 1.0.0 onwards

Starting with version 1.0.0, this chart does not support deploying Agent 5.x anymore. If you cannot upgrade to Agent 6.x or later, you can use a previous version of the chart by calling helm install with `--version 0.18.0`.

See [0.18.1's README](https://github.com/helm/charts/blob/847f737479bb78d89f8fb650db25627558fbe1f0/datadog/datadog/README.md) to see which options were supported at the time.

### Uninstalling the Chart

To uninstall/delete the `<RELEASE_NAME>` deployment:

```bash
helm uninstall <RELEASE_NAME>
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

As a best practice, a YAML file that specifies the values for the chart parameters should be used to configure the chart. Any parameters not specified in this file will default to those set in [values.yaml](values.yaml).

1. Create an empty `datadog-values.yaml` file.
2. Create a Kubernetes `secret` to store your [Datadog API key](https://app.datadoghq.com/organization-settings/api-keys) and [App key](https://app.datadoghq.com/organization-settings/application-keys)

```bash
kubectl create secret generic datadog-secret --from-literal api-key=$DD_API_KEY --from-literal app-key=$DD_APP_KEY
```

3. Set the following parameters in your `datadog-values.yaml` file to reference the secret:

```yaml
datadog:
  apiKeyExistingSecret: datadog-secret
  appKeyExistingSecret: datadog-secret
```

3. Install or upgrade the Datadog Helm chart with the new `datadog-values.yaml` file:

```bash
helm install -f datadog-values.yaml <RELEASE_NAME> datadog/datadog
```

OR

```bash
helm upgrade -f datadog-values.yaml <RELEASE_NAME> datadog/datadog
```

See the [All configuration options](#all-configuration-options) section to discover all configuration possibilities in the Datadog chart.

### Configuring Dogstatsd in the agent
<a name="dsd-config"></a>
The agent will start a server running Dogstatsd in order to process custom metrics sent from your applications. Check out the [official documentation on Dogstatsd](https://docs.datadoghq.com/developers/dogstatsd/?tab=hostagent) for more details.

By default the agent will create a unix domain socket to process the datagrams (not supported on Windows, see [below](#windows-config)).

To disable the socket in favor of the hostPort, use the following configuration:

```yaml
datadog:
  #(...)
  dogstatsd:
    useSocketVolume: false
    useHostPort: true
```

### Enabling APM and Tracing

APM is enabled by default using a socket for communication in the out-of-the-box [values.yaml](values.yaml) file; more details about application configuration are available on the [official documentation](https://docs.datadoghq.com/agent/kubernetes/apm/?tab=helm).
Update your `datadog-values.yaml` file with the following configration to enable TCP communication using a `hostPort`:

```yaml
datadog:
  # (...)
  apm:
    portEnabled: true
```

To disable APM, set `socketEnabled` to `false` in your `datadog-values.yaml` file (`portEnabled` is `false` by default):

```yaml
datadog:
  # (...)
  apm:
    socketEnabled: false
```

### Enabling APM Single Step Instrumentation (beta)

APM tracing libraries and configurations can be automatically injected in your application pods in the whole cluster or specific namespaces using Single Step Instrumentation.

Update your `datadog-values.yaml` file with the following configration to enable Single Step Instrumentation in the whole cluster:

```yaml
datadog:
  # (...)
  apm:
    instrumentation:
      enabled: true
```

Single Step Instrumentation can be disabled in specific namespaces using configuration option `disabledNamespaces`:

```yaml
datadog:
  # (...)
  apm:
    instrumentation:
      enabled: true
      disabledNamespaces:
        - namespaceA
        - namespaceB
```

Single Step Instrumentation can be enabled in specific namespaces using configuration option `enabledNamespaces`:

```yaml
datadog:
  # (...)
  apm:
    instrumentation:
      enabled: true
      enabledNamespaces:
        - namespaceC
```

To confiure the version of Tracing library that Single Step Instrumentation will instrument applications with, set the configuration `libVersions`:

```yaml
datadog:
  # (...)
  apm:
    instrumentation:
      enabled: true
      libVersions:
        java: v1.18.0
        python: v1.20.0
```

then upgrade your Datadog Helm chart:

```bash
helm upgrade -f datadog-values.yaml <RELEASE_NAME> datadog/datadog
```

### Enabling Log Collection

Update your `datadog-values.yaml` file with the following log collection configuration:

```yaml
datadog:
  # (...)
  logs:
    enabled: true
    containerCollectAll: true
```

then upgrade your Datadog Helm chart:

```bash
helm upgrade -f datadog-values.yaml <RELEASE_NAME> datadog/datadog
```

### Enabling Process Collection

Update your `datadog-values.yaml` file with the process collection configuration:

```yaml
datadog:
  # (...)
  processAgent:
    enabled: true
    processCollection: true
```

then upgrade your Datadog Helm chart:

```bash
helm upgrade -f datadog-values.yaml <RELEASE_NAME> datadog/datadog
```

### Enabling NPM Collection

The system-probe agent only runs in dedicated container environment. Update your `datadog-values.yaml` file with the NPM collection configuration:

```yaml
datadog:
  # (...)
  networkMonitoring:
    # (...)
    enabled: true

# (...)
```

then upgrade your Datadog Helm chart:

```bash
helm upgrade -f datadog-values.yaml <RELEASE_NAME> datadog/datadog
```

### Kubernetes event collection

Use the [Datadog Cluster Agent](#enabling-the-datadog-cluster-agent) to collect Kubernetes events. Please read [the official documentation](https://docs.datadoghq.com/agent/kubernetes/event_collection/) for more context.

Alternatively set the `datadog.leaderElection`, `datadog.collectEvents` and `rbac.create` options to `true` in order to enable Kubernetes event collection.

### conf.d and checks.d

The Datadog [entrypoint](https://github.com/DataDog/datadog-agent/blob/main/Dockerfiles/agent/entrypoint/89-copy-customfiles.sh) copies files with a `.yaml` extension found in `/conf.d` and files with `.py` extension in `/checks.d` to `/etc/datadog-agent/conf.d` and `/etc/datadog-agent/checks.d` respectively.

The keys for `datadog.confd` and `datadog.checksd` should mirror the content found in their respective ConfigMaps. Update your `datadog-values.yaml` file with the check configurations:

```yaml
datadog:
  confd:
    redisdb.yaml: |-
      ad_identifiers:
        - redis
        - bitnami/redis
      init_config:
      instances:
        - host: "%%host%%"
          port: "%%port%%"
    jmx.yaml: |-
      ad_identifiers:
        - openjdk
      instance_config:
      instances:
        - host: "%%host%%"
          port: "%%port_0%%"
    redisdb.yaml: |-
      init_config:
      instances:
        - host: "outside-k8s.example.com"
          port: 6379
```

then upgrade your Datadog Helm chart:

```bash
helm upgrade -f datadog-values.yaml <RELEASE_NAME> datadog/datadog
```

For more details, please refer to [the documentation](https://docs.datadoghq.com/agent/kubernetes/integrations/).

### Kubernetes Labels and Annotations

To map Kubernetes node labels and pod labels and annotations to Datadog tags, provide a dictionary with kubernetes labels/annotations as keys and Datadog tags key as values in your `datadog-values.yaml` file:

```yaml
nodeLabelsAsTags:
  beta.kubernetes.io/instance-type: aws_instance_type
  kubernetes.io/role: kube_role
```

```yaml
podAnnotationsAsTags:
  iam.amazonaws.com/role: kube_iamrole
```

```yaml
podLabelsAsTags:
  app: kube_app
  release: helm_release
```

then upgrade your Datadog Helm chart:

```bash
helm upgrade -f datadog-values.yaml <RELEASE_NAME> datadog/datadog
```

### CRI integration

As of the version 6.6.0, the Datadog Agent supports collecting metrics from any container runtime interface used in your cluster. Configure the location path of the socket with `datadog.criSocketPath`; default is the Docker container runtime socket. To deactivate this support, you just need to unset the `datadog.criSocketPath` setting.
Standard paths are:

- Docker socket: `/var/run/docker.sock`
- Containerd socket: `/var/run/containerd/containerd.sock`
- Cri-o socket: `/var/run/crio/crio.sock`

### Configuration required for Amazon Linux 2 based nodes

Amazon Linux 2 does not support apparmor profile enforcement.
Amazon Linux 2 is the default operating system for AWS Elastic Kubernetes Service (EKS) based clusters.
Update your `datadog-values.yaml` file to disable apparmor enforcement:

```yaml
agents:
  # (...)
  podSecurity:
    # (...)
    apparmor:
      # (...)
      enabled: false

# (...)
```

## Set an environment variable with the `--set` helm flag

You can set environment variables using the `--set` helm's flag  thanks to the `datadog.envDict` field.

For example, to set the `DD_ENV` environment variable:

```console
$ helm install --set datadog.envDict.DD_ENV=prod <release name> datadog/datadog
```

## All configuration options

The following table lists the configurable parameters of the Datadog chart and their default values. Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example,

```bash
helm install <RELEASE_NAME> \
  --set datadog.apiKey=<DATADOG_API_KEY>,datadog.logLevel=DEBUG \
  datadog/datadog
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| agents.additionalLabels | object | `{}` | Adds labels to the Agent daemonset and pods |
| agents.affinity | object | `{}` | Allow the DaemonSet to schedule using affinity rules |
| agents.containers.agent.env | list | `[]` | Additional environment variables for the agent container |
| agents.containers.agent.envDict | object | `{}` | Set environment variables specific to agent container defined in a dict |
| agents.containers.agent.envFrom | list | `[]` | Set environment variables specific to agent container from configMaps and/or secrets |
| agents.containers.agent.healthPort | int | `5555` | Port number to use in the node agent for the healthz endpoint |
| agents.containers.agent.livenessProbe | object | Every 15s / 6 KO / 1 OK | Override default agent liveness probe settings |
| agents.containers.agent.logLevel | string | `nil` | Set logging verbosity, valid log levels are: trace, debug, info, warn, error, critical, and off. If not set, fall back to the value of datadog.logLevel. |
| agents.containers.agent.ports | list | `[]` | Allows to specify extra ports (hostPorts for instance) for this container |
| agents.containers.agent.readinessProbe | object | Every 15s / 6 KO / 1 OK | Override default agent readiness probe settings |
| agents.containers.agent.resources | object | `{}` | Resource requests and limits for the agent container. |
| agents.containers.agent.securityContext | object | `{}` | Allows you to overwrite the default container SecurityContext for the agent container. |
| agents.containers.agent.startupProbe | object | Every 15s / 6 KO / 1 OK | Override default agent startup probe settings |
| agents.containers.initContainers.resources | object | `{}` | Resource requests and limits for the init containers |
| agents.containers.initContainers.securityContext | object | `{}` | Allows you to overwrite the default container SecurityContext for the init containers. |
| agents.containers.initContainers.volumeMounts | list | `[]` | Specify additional volumes to mount for the init containers |
| agents.containers.otelAgent.env | list | `[]` | Additional environment variables for the otel-agent container |
| agents.containers.otelAgent.envDict | object | `{}` | Set environment variables specific to otel-agent defined in a dict |
| agents.containers.otelAgent.envFrom | list | `[]` | Set environment variables specific to otel-agent from configMaps and/or secrets |
| agents.containers.otelAgent.ports | list | `[]` | Allows to specify extra ports (hostPorts for instance) for this container |
| agents.containers.otelAgent.resources | object | `{}` | Resource requests and limits for the otel-agent container |
| agents.containers.otelAgent.securityContext | object | `{}` | Allows you to overwrite the default container SecurityContext for the otel-agent container. |
| agents.containers.processAgent.env | list | `[]` | Additional environment variables for the process-agent container |
| agents.containers.processAgent.envDict | object | `{}` | Set environment variables specific to process-agent defined in a dict |
| agents.containers.processAgent.envFrom | list | `[]` | Set environment variables specific to process-agent from configMaps and/or secrets |
| agents.containers.processAgent.logLevel | string | `nil` | Set logging verbosity, valid log levels are: trace, debug, info, warn, error, critical, and off. If not set, fall back to the value of datadog.logLevel. |
| agents.containers.processAgent.ports | list | `[]` | Allows to specify extra ports (hostPorts for instance) for this container |
| agents.containers.processAgent.resources | object | `{}` | Resource requests and limits for the process-agent container |
| agents.containers.processAgent.securityContext | object | `{}` | Allows you to overwrite the default container SecurityContext for the process-agent container. |
| agents.containers.securityAgent.env | list | `[]` | Additional environment variables for the security-agent container |
| agents.containers.securityAgent.envDict | object | `{}` | Set environment variables specific to security-agent defined in a dict |
| agents.containers.securityAgent.envFrom | list | `[]` | Set environment variables specific to security-agent from configMaps and/or secrets |
| agents.containers.securityAgent.logLevel | string | `nil` | Set logging verbosity, valid log levels are: trace, debug, info, warn, error, critical, and off. If not set, fall back to the value of datadog.logLevel. |
| agents.containers.securityAgent.ports | list | `[]` | Allows to specify extra ports (hostPorts for instance) for this container |
| agents.containers.securityAgent.resources | object | `{}` | Resource requests and limits for the security-agent container |
| agents.containers.systemProbe.env | list | `[]` | Additional environment variables for the system-probe container |
| agents.containers.systemProbe.envDict | object | `{}` | Set environment variables specific to system-probe defined in a dict |
| agents.containers.systemProbe.envFrom | list | `[]` | Set environment variables specific to system-probe from configMaps and/or secrets |
| agents.containers.systemProbe.logLevel | string | `nil` | Set logging verbosity, valid log levels are: trace, debug, info, warn, error, critical, and off. If not set, fall back to the value of datadog.logLevel. |
| agents.containers.systemProbe.ports | list | `[]` | Allows to specify extra ports (hostPorts for instance) for this container |
| agents.containers.systemProbe.resources | object | `{}` | Resource requests and limits for the system-probe container |
| agents.containers.systemProbe.securityContext | object | `{"capabilities":{"add":["SYS_ADMIN","SYS_RESOURCE","SYS_PTRACE","NET_ADMIN","NET_BROADCAST","NET_RAW","IPC_LOCK","CHOWN","DAC_READ_SEARCH"]},"privileged":false}` | Allows you to overwrite the default container SecurityContext for the system-probe container. |
| agents.containers.traceAgent.env | list | `[]` | Additional environment variables for the trace-agent container |
| agents.containers.traceAgent.envDict | object | `{}` | Set environment variables specific to trace-agent defined in a dict |
| agents.containers.traceAgent.envFrom | list | `[]` | Set environment variables specific to trace-agent from configMaps and/or secrets |
| agents.containers.traceAgent.livenessProbe | object | Every 15s | Override default agent liveness probe settings |
| agents.containers.traceAgent.logLevel | string | `nil` | Set logging verbosity, valid log levels are: trace, debug, info, warn, error, critical, and off |
| agents.containers.traceAgent.ports | list | `[]` | Allows to specify extra ports (hostPorts for instance) for this container |
| agents.containers.traceAgent.resources | object | `{}` | Resource requests and limits for the trace-agent container |
| agents.containers.traceAgent.securityContext | object | `{}` | Allows you to overwrite the default container SecurityContext for the trace-agent container. |
| agents.customAgentConfig | object | `{}` | Specify custom contents for the datadog agent config (datadog.yaml) |
| agents.daemonsetAnnotations | object | `{}` | Annotations to add to the DaemonSet |
| agents.dnsConfig | object | `{}` | specify dns configuration options for datadog cluster agent containers e.g ndots |
| agents.enabled | bool | `true` | You should keep Datadog DaemonSet enabled! |
| agents.image.digest | string | `""` | Define Agent image digest to use, takes precedence over tag if specified |
| agents.image.doNotCheckTag | string | `nil` | Skip the version and chart compatibility check |
| agents.image.name | string | `"agent"` | Datadog Agent image name to use (relative to `registry`) |
| agents.image.pullPolicy | string | `"IfNotPresent"` | Datadog Agent image pull policy |
| agents.image.pullSecrets | list | `[]` | Datadog Agent repository pullSecret (ex: specify docker registry credentials) |
| agents.image.repository | string | `nil` | Override default registry + image.name for Agent |
| agents.image.tag | string | `"7.59.0"` | Define the Agent version to use |
| agents.image.tagSuffix | string | `""` | Suffix to append to Agent tag |
| agents.localService.forceLocalServiceEnabled | bool | `false` | Force the creation of the internal traffic policy service to target the agent running on the local node. By default, the internal traffic service is created only on Kubernetes 1.22+ where the feature became beta and enabled by default. This option allows to force the creation of the internal traffic service on kubernetes 1.21 where the feature was alpha and required a feature gate to be explicitly enabled. |
| agents.localService.overrideName | string | `""` | Name of the internal traffic service to target the agent running on the local node |
| agents.networkPolicy.create | bool | `false` | If true, create a NetworkPolicy for the agents. DEPRECATED. Use datadog.networkPolicy.create instead |
| agents.nodeSelector | object | `{}` | Allow the DaemonSet to schedule on selected nodes |
| agents.podAnnotations | object | `{}` | Annotations to add to the DaemonSet's Pods |
| agents.podLabels | object | `{}` | Sets podLabels if defined |
| agents.podSecurity.allowedUnsafeSysctls | list | `[]` | Allowed unsafe sysclts |
| agents.podSecurity.apparmor.enabled | bool | `true` | If true, enable apparmor enforcement |
| agents.podSecurity.apparmorProfiles | list | `["runtime/default","unconfined"]` | Allowed apparmor profiles |
| agents.podSecurity.capabilities | list | `["SYS_ADMIN","SYS_RESOURCE","SYS_PTRACE","NET_ADMIN","NET_BROADCAST","NET_RAW","IPC_LOCK","CHOWN","AUDIT_CONTROL","AUDIT_READ","DAC_READ_SEARCH"]` | Allowed capabilities |
| agents.podSecurity.defaultApparmor | string | `"runtime/default"` | Default AppArmor profile for all containers but system-probe |
| agents.podSecurity.podSecurityPolicy.create | bool | `false` | If true, create a PodSecurityPolicy resource for Agent pods |
| agents.podSecurity.privileged | bool | `false` | If true, Allow to run privileged containers |
| agents.podSecurity.seLinuxContext | object | Must run as spc_t | Provide seLinuxContext configuration for PSP/SCC |
| agents.podSecurity.seccompProfiles | list | `["runtime/default","localhost/system-probe"]` | Allowed seccomp profiles |
| agents.podSecurity.securityContextConstraints.create | bool | `false` | If true, create a SecurityContextConstraints resource for Agent pods |
| agents.podSecurity.volumes | list | `["configMap","downwardAPI","emptyDir","hostPath","secret"]` | Allowed volumes types |
| agents.priorityClassCreate | bool | `false` | Creates a priorityClass for the Datadog Agent's Daemonset pods. |
| agents.priorityClassName | string | `nil` | Sets PriorityClassName if defined |
| agents.priorityClassValue | int | `1000000000` | Value used to specify the priority of the scheduling of Datadog Agent's Daemonset pods. |
| agents.priorityPreemptionPolicyValue | string | `"PreemptLowerPriority"` | Set to "Never" to change the PriorityClass to non-preempting |
| agents.rbac.automountServiceAccountToken | bool | `true` | If true, automatically mount the ServiceAccount's API credentials if agents.rbac.create is true |
| agents.rbac.create | bool | `true` | If true, create & use RBAC resources |
| agents.rbac.serviceAccountAnnotations | object | `{}` | Annotations to add to the ServiceAccount if agents.rbac.create is true |
| agents.rbac.serviceAccountName | string | `"default"` | Specify a preexisting ServiceAccount to use if agents.rbac.create is false |
| agents.revisionHistoryLimit | int | `10` | The number of ControllerRevision to keep in this DaemonSet. |
| agents.shareProcessNamespace | bool | `false` | Set the process namespace sharing on the Datadog Daemonset |
| agents.tolerations | list | `[]` | Allow the DaemonSet to schedule on tainted nodes (requires Kubernetes >= 1.6) |
| agents.updateStrategy | object | `{"rollingUpdate":{"maxUnavailable":"10%"},"type":"RollingUpdate"}` | Allow the DaemonSet to perform a rolling update on helm update |
| agents.useConfigMap | string | `nil` | Configures a configmap to provide the agent configuration. Use this in combination with the `agents.customAgentConfig` parameter. |
| agents.useHostNetwork | bool | `false` | Bind ports on the hostNetwork |
| agents.volumeMounts | list | `[]` | Specify additional volumes to mount in all containers of the agent pod |
| agents.volumes | list | `[]` | Specify additional volumes to mount in the dd-agent container |
| clusterAgent.additionalLabels | object | `{}` | Adds labels to the Cluster Agent deployment and pods |
| clusterAgent.admissionController.agentSidecarInjection.clusterAgentCommunicationEnabled | bool | `true` | Enable communication between Agent sidecars and the Cluster Agent. |
| clusterAgent.admissionController.agentSidecarInjection.containerRegistry | string | `nil` | Override the default registry for the sidecar Agent. |
| clusterAgent.admissionController.agentSidecarInjection.enabled | bool | `false` | Enables Datadog Agent sidecar injection. |
| clusterAgent.admissionController.agentSidecarInjection.imageName | string | `nil` |  |
| clusterAgent.admissionController.agentSidecarInjection.imageTag | string | `nil` |  |
| clusterAgent.admissionController.agentSidecarInjection.profiles | list | `[]` | Defines the sidecar configuration override, currently only one profile is supported. |
| clusterAgent.admissionController.agentSidecarInjection.provider | string | `nil` | Used by the admission controller to add infrastructure provider-specific configurations to the Agent sidecar. |
| clusterAgent.admissionController.agentSidecarInjection.selectors | list | `[]` | Defines the pod selector for sidecar injection, currently only one rule is supported. |
| clusterAgent.admissionController.configMode | string | `nil` | The kind of configuration to be injected, it can be "hostip", "service", or "socket". |
| clusterAgent.admissionController.containerRegistry | string | `nil` | Override the default registry for the admission controller. |
| clusterAgent.admissionController.enabled | bool | `true` | Enable the admissionController to be able to inject APM/Dogstatsd config and standard tags (env, service, version) automatically into your pods |
| clusterAgent.admissionController.failurePolicy | string | `"Ignore"` | Set the failure policy for dynamic admission control.' |
| clusterAgent.admissionController.mutateUnlabelled | bool | `false` | Enable injecting config without having the pod label 'admission.datadoghq.com/enabled="true"' |
| clusterAgent.admissionController.mutation | object | `{"enabled":true}` | Mutation Webhook configuration options |
| clusterAgent.admissionController.mutation.enabled | bool | `true` | Enabled enables the Admission Controller mutation webhook. Default: true. (Requires Agent 7.59.0+). |
| clusterAgent.admissionController.port | int | `8000` | Set port of cluster-agent admission controller service |
| clusterAgent.admissionController.remoteInstrumentation.enabled | bool | `false` | Enable polling and applying library injection using Remote Config. # This feature is in beta, and enables Remote Config in the Cluster Agent. It also requires Cluster Agent version 7.43+. # Enabling this feature grants the Cluster Agent the permissions to patch Deployment objects in the cluster. |
| clusterAgent.admissionController.validation | object | `{"enabled":true}` | Validation Webhook configuration options |
| clusterAgent.admissionController.validation.enabled | bool | `true` | Enabled enables the Admission Controller validation webhook. Default: true. (Requires Agent 7.59.0+). |
| clusterAgent.admissionController.webhookName | string | `"datadog-webhook"` | Name of the validatingwebhookconfiguration and mutatingwebhookconfiguration created by the cluster-agent |
| clusterAgent.advancedConfd | object | `{}` | Provide additional cluster check configurations. Each key is an integration containing several config files. |
| clusterAgent.affinity | object | `{}` | Allow the Cluster Agent Deployment to schedule using affinity rules |
| clusterAgent.command | list | `[]` | Command to run in the Cluster Agent container as entrypoint |
| clusterAgent.confd | object | `{}` | Provide additional cluster check configurations. Each key will become a file in /conf.d. |
| clusterAgent.containerExclude | string | `nil` | Exclude containers from the Cluster Agent Autodiscovery, as a space-separated list. (Requires Agent/Cluster Agent 7.50.0+) |
| clusterAgent.containerInclude | string | `nil` | Include containers in the Cluster Agent Autodiscovery, as a space-separated list.  If a container matches an include rule, it’s always included in the Autodiscovery. (Requires Agent/Cluster Agent 7.50.0+) |
| clusterAgent.containers.clusterAgent.securityContext | object | `{"allowPrivilegeEscalation":false,"readOnlyRootFilesystem":true}` | Specify securityContext on the cluster-agent container. |
| clusterAgent.containers.initContainers.securityContext | object | `{}` |  |
| clusterAgent.createPodDisruptionBudget | bool | `false` | Create pod disruption budget for Cluster Agent deployments |
| clusterAgent.datadog_cluster_yaml | object | `{}` | Specify custom contents for the datadog cluster agent config (datadog-cluster.yaml) |
| clusterAgent.deploymentAnnotations | object | `{}` | Annotations to add to the cluster-agents's deployment |
| clusterAgent.dnsConfig | object | `{}` | Specify dns configuration options for datadog cluster agent containers e.g ndots |
| clusterAgent.enabled | bool | `true` | Set this to false to disable Datadog Cluster Agent |
| clusterAgent.env | list | `[]` | Set environment variables specific to Cluster Agent |
| clusterAgent.envDict | object | `{}` | Set environment variables specific to Cluster Agent defined in a dict |
| clusterAgent.envFrom | list | `[]` | Set environment variables specific to Cluster Agent from configMaps and/or secrets |
| clusterAgent.healthPort | int | `5556` | Port number to use in the Cluster Agent for the healthz endpoint |
| clusterAgent.image.digest | string | `""` | Cluster Agent image digest to use, takes precedence over tag if specified |
| clusterAgent.image.doNotCheckTag | string | `nil` | Skip the version and chart compatibility check |
| clusterAgent.image.name | string | `"cluster-agent"` | Cluster Agent image name to use (relative to `registry`) |
| clusterAgent.image.pullPolicy | string | `"IfNotPresent"` | Cluster Agent image pullPolicy |
| clusterAgent.image.pullSecrets | list | `[]` | Cluster Agent repository pullSecret (ex: specify docker registry credentials) |
| clusterAgent.image.repository | string | `nil` | Override default registry + image.name for Cluster Agent |
| clusterAgent.image.tag | string | `"7.59.0"` | Cluster Agent image tag to use |
| clusterAgent.livenessProbe | object | Every 15s / 6 KO / 1 OK | Override default Cluster Agent liveness probe settings |
| clusterAgent.metricsProvider.aggregator | string | `"avg"` | Define the aggregator the cluster agent will use to process the metrics. The options are (avg, min, max, sum) |
| clusterAgent.metricsProvider.createReaderRbac | bool | `true` | Create `external-metrics-reader` RBAC automatically (to allow HPA to read data from Cluster Agent) |
| clusterAgent.metricsProvider.enabled | bool | `false` | Set this to true to enable Metrics Provider |
| clusterAgent.metricsProvider.endpoint | string | `nil` | Override the external metrics provider endpoint. If not set, the cluster-agent defaults to `datadog.site` |
| clusterAgent.metricsProvider.registerAPIService | bool | `true` | Set this to false to disable external metrics registration as an APIService |
| clusterAgent.metricsProvider.service.port | int | `8443` | Set port of cluster-agent metrics server service (Kubernetes >= 1.15) |
| clusterAgent.metricsProvider.service.type | string | `"ClusterIP"` | Set type of cluster-agent metrics server service |
| clusterAgent.metricsProvider.useDatadogMetrics | bool | `false` | Enable usage of DatadogMetric CRD to autoscale on arbitrary Datadog queries |
| clusterAgent.metricsProvider.wpaController | bool | `false` | Enable informer and controller of the watermark pod autoscaler |
| clusterAgent.networkPolicy.create | bool | `false` | If true, create a NetworkPolicy for the cluster agent. DEPRECATED. Use datadog.networkPolicy.create instead |
| clusterAgent.nodeSelector | object | `{}` | Allow the Cluster Agent Deployment to be scheduled on selected nodes |
| clusterAgent.podAnnotations | object | `{}` | Annotations to add to the cluster-agents's pod(s) |
| clusterAgent.podSecurity.podSecurityPolicy.create | bool | `false` | If true, create a PodSecurityPolicy resource for Cluster Agent pods |
| clusterAgent.podSecurity.securityContextConstraints.create | bool | `false` | If true, create a SCC resource for Cluster Agent pods |
| clusterAgent.priorityClassName | string | `nil` | Name of the priorityClass to apply to the Cluster Agent |
| clusterAgent.rbac.automountServiceAccountToken | bool | `true` | If true, automatically mount the ServiceAccount's API credentials if clusterAgent.rbac.create is true |
| clusterAgent.rbac.create | bool | `true` | If true, create & use RBAC resources |
| clusterAgent.rbac.flareAdditionalPermissions | bool | `true` | If true, add Secrets and Configmaps get/list permissions to retrieve user Datadog Helm values from Cluster Agent namespace |
| clusterAgent.rbac.serviceAccountAnnotations | object | `{}` | Annotations to add to the ServiceAccount if clusterAgent.rbac.create is true |
| clusterAgent.rbac.serviceAccountName | string | `"default"` | Specify a preexisting ServiceAccount to use if clusterAgent.rbac.create is false |
| clusterAgent.readinessProbe | object | Every 15s / 6 KO / 1 OK | Override default Cluster Agent readiness probe settings |
| clusterAgent.replicas | int | `1` | Specify the of cluster agent replicas, if > 1 it allow the cluster agent to work in HA mode. |
| clusterAgent.resources | object | `{}` | Datadog cluster-agent resource requests and limits. |
| clusterAgent.revisionHistoryLimit | int | `10` | The number of old ReplicaSets to keep in this Deployment. |
| clusterAgent.securityContext | object | `{}` | Allows you to overwrite the default PodSecurityContext on the cluster-agent pods. |
| clusterAgent.shareProcessNamespace | bool | `false` | Set the process namespace sharing on the Datadog Cluster Agent |
| clusterAgent.startupProbe | object | Every 15s / 6 KO / 1 OK | Override default Cluster Agent startup probe settings |
| clusterAgent.strategy | object | `{"rollingUpdate":{"maxSurge":1,"maxUnavailable":0},"type":"RollingUpdate"}` | Allow the Cluster Agent deployment to perform a rolling update on helm update |
| clusterAgent.token | string | `""` | Cluster Agent token is a preshared key between node agents and cluster agent (autogenerated if empty, needs to be at least 32 characters a-zA-z) |
| clusterAgent.tokenExistingSecret | string | `""` | Existing secret name to use for Cluster Agent token. Put the Cluster Agent token in a key named `token` inside the Secret |
| clusterAgent.tolerations | list | `[]` | Allow the Cluster Agent Deployment to schedule on tainted nodes ((requires Kubernetes >= 1.6)) |
| clusterAgent.topologySpreadConstraints | list | `[]` | Allow the Cluster Agent Deployment to schedule using pod topology spreading |
| clusterAgent.useHostNetwork | bool | `false` | Bind ports on the hostNetwork |
| clusterAgent.volumeMounts | list | `[]` | Specify additional volumes to mount in the cluster-agent container |
| clusterAgent.volumes | list | `[]` | Specify additional volumes to mount in the cluster-agent container |
| clusterChecksRunner.additionalLabels | object | `{}` | Adds labels to the cluster checks runner deployment and pods |
| clusterChecksRunner.affinity | object | `{}` | Allow the ClusterChecks Deployment to schedule using affinity rules. |
| clusterChecksRunner.containers.agent.securityContext | object | `{}` | Specify securityContext on the agent container |
| clusterChecksRunner.containers.initContainers.securityContext | object | `{}` | Specify securityContext on the init containers |
| clusterChecksRunner.createPodDisruptionBudget | bool | `false` | Create the pod disruption budget to apply to the cluster checks agents |
| clusterChecksRunner.deploymentAnnotations | object | `{}` | Annotations to add to the cluster-checks-runner's Deployment |
| clusterChecksRunner.dnsConfig | object | `{}` | specify dns configuration options for datadog cluster agent containers e.g ndots |
| clusterChecksRunner.enabled | bool | `false` | If true, deploys agent dedicated for running the Cluster Checks instead of running in the Daemonset's agents. |
| clusterChecksRunner.env | list | `[]` | Environment variables specific to Cluster Checks Runner |
| clusterChecksRunner.envDict | object | `{}` | Set environment variables specific to Cluster Checks Runner defined in a dict |
| clusterChecksRunner.envFrom | list | `[]` | Set environment variables specific to Cluster Checks Runner from configMaps and/or secrets |
| clusterChecksRunner.healthPort | int | `5557` | Port number to use in the Cluster Checks Runner for the healthz endpoint |
| clusterChecksRunner.image.digest | string | `""` | Define Agent image digest to use, takes precedence over tag if specified |
| clusterChecksRunner.image.name | string | `"agent"` | Datadog Agent image name to use (relative to `registry`) |
| clusterChecksRunner.image.pullPolicy | string | `"IfNotPresent"` | Datadog Agent image pull policy |
| clusterChecksRunner.image.pullSecrets | list | `[]` | Datadog Agent repository pullSecret (ex: specify docker registry credentials) |
| clusterChecksRunner.image.repository | string | `nil` | Override default registry + image.name for Cluster Check Runners |
| clusterChecksRunner.image.tag | string | `"7.59.0"` | Define the Agent version to use |
| clusterChecksRunner.image.tagSuffix | string | `""` | Suffix to append to Agent tag |
| clusterChecksRunner.livenessProbe | object | Every 15s / 6 KO / 1 OK | Override default agent liveness probe settings |
| clusterChecksRunner.networkPolicy.create | bool | `false` | If true, create a NetworkPolicy for the cluster checks runners. DEPRECATED. Use datadog.networkPolicy.create instead |
| clusterChecksRunner.nodeSelector | object | `{}` | Allow the ClusterChecks Deployment to schedule on selected nodes |
| clusterChecksRunner.podAnnotations | object | `{}` | Annotations to add to the cluster-checks-runner's pod(s) |
| clusterChecksRunner.ports | list | `[]` | Allows to specify extra ports (hostPorts for instance) for this container |
| clusterChecksRunner.priorityClassName | string | `nil` | Name of the priorityClass to apply to the Cluster checks runners |
| clusterChecksRunner.rbac.automountServiceAccountToken | bool | `true` | If true, automatically mount the ServiceAccount's API credentials if clusterChecksRunner.rbac.create is true |
| clusterChecksRunner.rbac.create | bool | `true` | If true, create & use RBAC resources |
| clusterChecksRunner.rbac.dedicated | bool | `false` | If true, use a dedicated RBAC resource for the cluster checks agent(s) |
| clusterChecksRunner.rbac.serviceAccountAnnotations | object | `{}` | Annotations to add to the ServiceAccount if clusterChecksRunner.rbac.dedicated is true |
| clusterChecksRunner.rbac.serviceAccountName | string | `"default"` | Specify a preexisting ServiceAccount to use if clusterChecksRunner.rbac.create is false |
| clusterChecksRunner.readinessProbe | object | Every 15s / 6 KO / 1 OK | Override default agent readiness probe settings |
| clusterChecksRunner.replicas | int | `2` | Number of Cluster Checks Runner instances |
| clusterChecksRunner.resources | object | `{}` | Datadog clusterchecks-agent resource requests and limits. |
| clusterChecksRunner.revisionHistoryLimit | int | `10` | The number of old ReplicaSets to keep in this Deployment. |
| clusterChecksRunner.securityContext | object | `{}` | Allows you to overwrite the default PodSecurityContext on the clusterchecks pods. |
| clusterChecksRunner.startupProbe | object | Every 15s / 6 KO / 1 OK | Override default agent startup probe settings |
| clusterChecksRunner.strategy | object | `{"rollingUpdate":{"maxSurge":1,"maxUnavailable":0},"type":"RollingUpdate"}` | Allow the ClusterChecks deployment to perform a rolling update on helm update |
| clusterChecksRunner.tolerations | list | `[]` | Tolerations for pod assignment |
| clusterChecksRunner.topologySpreadConstraints | list | `[]` | Allow the ClusterChecks Deployment to schedule using pod topology spreading |
| clusterChecksRunner.volumeMounts | list | `[]` | Specify additional volumes to mount in the cluster checks container |
| clusterChecksRunner.volumes | list | `[]` | Specify additional volumes to mount in the cluster checks container |
| commonLabels | object | `{}` | Labels to apply to all resources |
| datadog-crds.crds.datadogMetrics | bool | `true` | Set to true to deploy the DatadogMetrics CRD |
| datadog-crds.crds.datadogPodAutoscalers | bool | `true` | Set to true to deploy the DatadogPodAutoscalers CRD |
| datadog.apiKey | string | `nil` | Your Datadog API key |
| datadog.apiKeyExistingSecret | string | `nil` | Use existing Secret which stores API key instead of creating a new one. The value should be set with the `api-key` key inside the secret. |
| datadog.apm.enabled | bool | `false` | Enable this to enable APM and tracing, on port 8126 DEPRECATED. Use datadog.apm.portEnabled instead |
| datadog.apm.hostSocketPath | string | `"/var/run/datadog/"` | Host path to the trace-agent socket |
| datadog.apm.instrumentation.disabledNamespaces | list | `[]` | Disable injecting the Datadog APM libraries into pods in specific namespaces (beta). |
| datadog.apm.instrumentation.enabled | bool | `false` | Enable injecting the Datadog APM libraries into all pods in the cluster (beta). |
| datadog.apm.instrumentation.enabledNamespaces | list | `[]` | Enable injecting the Datadog APM libraries into pods in specific namespaces (beta). |
| datadog.apm.instrumentation.language_detection.enabled | bool | `true` | Run language detection to automatically detect languages of user workloads (beta). |
| datadog.apm.instrumentation.libVersions | object | `{}` | Inject specific version of tracing libraries with Single Step Instrumentation (beta). |
| datadog.apm.instrumentation.skipKPITelemetry | bool | `false` | Disable generating Configmap for APM Instrumentation KPIs |
| datadog.apm.port | int | `8126` | Override the trace Agent port |
| datadog.apm.portEnabled | bool | `false` | Enable APM over TCP communication (hostPort 8126 by default) |
| datadog.apm.socketEnabled | bool | `true` | Enable APM over Socket (Unix Socket or windows named pipe) |
| datadog.apm.socketPath | string | `"/var/run/datadog/apm.socket"` | Path to the trace-agent socket |
| datadog.apm.useLocalService | bool | `false` | Enable APM over TCP communication to use the local service only (requires Kubernetes v1.22+) Note: The hostPort 8126 is disabled when this is enabled. |
| datadog.apm.useSocketVolume | bool | `false` | Enable APM over Unix Domain Socket DEPRECATED. Use datadog.apm.socketEnabled instead |
| datadog.appKey | string | `nil` | Datadog APP key required to use metricsProvider |
| datadog.appKeyExistingSecret | string | `nil` | Use existing Secret which stores APP key instead of creating a new one. The value should be set with the `app-key` key inside the secret. |
| datadog.asm.iast.enabled | bool | `false` | Enable Application Security Management Interactive Application Security Testing by injecting `DD_IAST_ENABLED=true` environment variable to all pods in the cluster |
| datadog.asm.sca.enabled | bool | `false` | Enable Application Security Management Software Composition Analysis by injecting `DD_APPSEC_SCA_ENABLED=true` environment variable to all pods in the cluster |
| datadog.asm.threats.enabled | bool | `false` | Enable Application Security Management Threats App & API Protection by injecting `DD_APPSEC_ENABLED=true` environment variable to all pods in the cluster |
| datadog.checksCardinality | string | `nil` | Sets the tag cardinality for the checks run by the Agent. |
| datadog.checksd | object | `{}` | Provide additional custom checks as python code |
| datadog.clusterChecks.enabled | bool | `true` | Enable the Cluster Checks feature on both the cluster-agents and the daemonset |
| datadog.clusterChecks.shareProcessNamespace | bool | `false` | Set the process namespace sharing on the cluster checks agent |
| datadog.clusterName | string | `nil` | Set a unique cluster name to allow scoping hosts and Cluster Checks easily |
| datadog.clusterTagger.collectKubernetesTags | bool | `false` | Enables Kubernetes resources tags collection. |
| datadog.collectEvents | bool | `true` | Enables this to start event collection from the kubernetes API |
| datadog.confd | object | `{}` | Provide additional check configurations (static and Autodiscovery) |
| datadog.containerExclude | string | `nil` | Exclude containers from Agent Autodiscovery, as a space-separated list |
| datadog.containerExcludeLogs | string | `nil` | Exclude logs from Agent Autodiscovery, as a space-separated list |
| datadog.containerExcludeMetrics | string | `nil` | Exclude metrics from Agent Autodiscovery, as a space-separated list |
| datadog.containerImageCollection.enabled | bool | `true` | Enable collection of container image metadata |
| datadog.containerInclude | string | `nil` | Include containers in Agent Autodiscovery, as a space-separated list. If a container matches an include rule, it’s always included in Autodiscovery |
| datadog.containerIncludeLogs | string | `nil` | Include logs in Agent Autodiscovery, as a space-separated list |
| datadog.containerIncludeMetrics | string | `nil` | Include metrics in Agent Autodiscovery, as a space-separated list |
| datadog.containerLifecycle.enabled | bool | `true` | Enable container lifecycle events collection |
| datadog.containerRuntimeSupport.enabled | bool | `true` | Set this to false to disable agent access to container runtime. |
| datadog.criSocketPath | string | `nil` | Path to the container runtime socket (if different from Docker) |
| datadog.dd_url | string | `nil` | The host of the Datadog intake server to send Agent data to, only set this option if you need the Agent to send data to a custom URL |
| datadog.dockerSocketPath | string | `nil` | Path to the docker socket |
| datadog.dogstatsd.hostSocketPath | string | `"/var/run/datadog/"` | Host path to the DogStatsD socket |
| datadog.dogstatsd.nonLocalTraffic | bool | `true` | Enable this to make each node accept non-local statsd traffic (from outside of the pod) |
| datadog.dogstatsd.originDetection | bool | `false` | Enable origin detection for container tagging |
| datadog.dogstatsd.port | int | `8125` | Override the Agent DogStatsD port |
| datadog.dogstatsd.socketPath | string | `"/var/run/datadog/dsd.socket"` | Path to the DogStatsD socket |
| datadog.dogstatsd.tagCardinality | string | `"low"` | Sets the tag cardinality relative to the origin detection |
| datadog.dogstatsd.tags | list | `[]` | List of static tags to attach to every custom metric, event and service check collected by Dogstatsd. |
| datadog.dogstatsd.useHostPID | bool | `false` | Run the agent in the host's PID namespace # DEPRECATED: use datadog.useHostPID instead. |
| datadog.dogstatsd.useHostPort | bool | `false` | Sets the hostPort to the same value of the container port |
| datadog.dogstatsd.useSocketVolume | bool | `true` | Enable dogstatsd over Unix Domain Socket with an HostVolume |
| datadog.env | list | `[]` | Set environment variables for all Agents |
| datadog.envDict | object | `{}` | Set environment variables for all Agents defined in a dict |
| datadog.envFrom | list | `[]` | Set environment variables for all Agents directly from configMaps and/or secrets |
| datadog.excludePauseContainer | bool | `true` | Exclude pause containers from Agent Autodiscovery. |
| datadog.expvarPort | int | `6000` | Specify the port to expose pprof and expvar to not interfere with the agent metrics port from the cluster-agent, which defaults to 5000 |
| datadog.helmCheck.collectEvents | bool | `false` | Set this to true to enable event collection in the Helm Check (Requires Agent 7.36.0+ and Cluster Agent 1.20.0+) This requires datadog.HelmCheck.enabled to be set to true |
| datadog.helmCheck.enabled | bool | `false` | Set this to true to enable the Helm check (Requires Agent 7.35.0+ and Cluster Agent 1.19.0+) This requires clusterAgent.enabled to be set to true |
| datadog.helmCheck.valuesAsTags | object | `{}` | Collects Helm values from a release and uses them as tags (Requires Agent and Cluster Agent 7.40.0+). This requires datadog.HelmCheck.enabled to be set to true |
| datadog.hostVolumeMountPropagation | string | `"None"` | Allow to specify the `mountPropagation` value on all volumeMounts using HostPath |
| datadog.ignoreAutoConfig | list | `[]` | List of integration to ignore auto_conf.yaml. |
| datadog.kubeStateMetricsCore.annotationsAsTags | object | `{}` | Extra annotations to collect from resources and to turn into datadog tag. |
| datadog.kubeStateMetricsCore.collectApiServicesMetrics | bool | `false` | Enable watching apiservices objects and collecting their corresponding metrics kubernetes_state.apiservice.* (Requires Cluster Agent 7.45.0+) |
| datadog.kubeStateMetricsCore.collectConfigMaps | bool | `true` | Enable watching configmap objects and collecting their corresponding metrics kubernetes_state.configmap.* |
| datadog.kubeStateMetricsCore.collectCrdMetrics | bool | `false` | Enable watching CRD objects and collecting their corresponding metrics kubernetes_state.crd.* |
| datadog.kubeStateMetricsCore.collectSecretMetrics | bool | `true` | Enable watching secret objects and collecting their corresponding metrics kubernetes_state.secret.* |
| datadog.kubeStateMetricsCore.collectVpaMetrics | bool | `false` | Enable watching VPA objects and collecting their corresponding metrics kubernetes_state.vpa.* |
| datadog.kubeStateMetricsCore.enabled | bool | `true` | Enable the kubernetes_state_core check in the Cluster Agent (Requires Cluster Agent 1.12.0+) |
| datadog.kubeStateMetricsCore.ignoreLegacyKSMCheck | bool | `true` | Disable the auto-configuration of legacy kubernetes_state check (taken into account only when datadog.kubeStateMetricsCore.enabled is true) |
| datadog.kubeStateMetricsCore.labelsAsTags | object | `{}` | Extra labels to collect from resources and to turn into datadog tag. |
| datadog.kubeStateMetricsCore.rbac.create | bool | `true` | If true, create & use RBAC resources |
| datadog.kubeStateMetricsCore.useClusterCheckRunners | bool | `false` | For large clusters where the Kubernetes State Metrics Check Core needs to be distributed on dedicated workers. |
| datadog.kubeStateMetricsEnabled | bool | `false` | If true, deploys the kube-state-metrics deployment |
| datadog.kubeStateMetricsNetworkPolicy.create | bool | `false` | If true, create a NetworkPolicy for kube state metrics |
| datadog.kubelet.agentCAPath | string | /var/run/host-kubelet-ca.crt if hostCAPath else /var/run/secrets/kubernetes.io/serviceaccount/ca.crt | Path (inside Agent containers) where the Kubelet CA certificate is stored |
| datadog.kubelet.coreCheckEnabled | bool | true | Toggle if kubelet core check should be used instead of Python check. (Requires Agent/Cluster Agent 7.53.0+) |
| datadog.kubelet.host | object | `{"valueFrom":{"fieldRef":{"fieldPath":"status.hostIP"}}}` | Override kubelet IP |
| datadog.kubelet.hostCAPath | string | None (no mount from host) | Path (on host) where the Kubelet CA certificate is stored |
| datadog.kubelet.podLogsPath | string | /var/log/pods on Linux, C:\var\log\pods on Windows | Path (on host) where the PODs logs are located |
| datadog.kubelet.tlsVerify | string | true | Toggle kubelet TLS verification |
| datadog.kubernetesEvents.collectedEventTypes | list | `[{"kind":"Pod","reasons":["Failed","BackOff","Unhealthy","FailedScheduling","FailedMount","FailedAttachVolume"]},{"kind":"Node","reasons":["TerminatingEvictedPod","NodeNotReady","Rebooted","HostPortConflict"]},{"kind":"CronJob","reasons":["SawCompletedJob"]}]` | Event types to be collected. This requires datadog.kubernetesEvents.unbundleEvents to be set to true. |
| datadog.kubernetesEvents.filteringEnabled | bool | `false` | Enable this to only include events that match the pre-defined allowed events. (Requires Cluster Agent 7.57.0+). |
| datadog.kubernetesEvents.sourceDetectionEnabled | bool | `false` | Enable this to map Kubernetes events to integration sources based on controller names. (Requires Cluster Agent 7.56.0+). |
| datadog.kubernetesEvents.unbundleEvents | bool | `false` | Allow unbundling kubernetes events, 1:1 mapping between Kubernetes and Datadog events. (Requires Cluster Agent 7.42.0+). |
| datadog.kubernetesResourcesAnnotationsAsTags | object | `{}` | Provide a mapping of Kubernetes Resources Annotations to Datadog Tags |
| datadog.kubernetesResourcesLabelsAsTags | object | `{}` | Provide a mapping of Kubernetes Resources Labels to Datadog Tags |
| datadog.leaderElection | bool | `true` | Enables leader election mechanism for event collection |
| datadog.leaderElectionResource | string | `"configmap"` | Selects the default resource to use for leader election. Can be: * "lease" / "leases". Only supported in agent 7.47+ * "configmap" / "configmaps". "" to automatically detect which one to use. |
| datadog.leaderLeaseDuration | string | `nil` | Set the lease time for leader election in second |
| datadog.logLevel | string | `"INFO"` | Set logging verbosity, valid log levels are: trace, debug, info, warn, error, critical, off |
| datadog.logs.autoMultiLineDetection | bool | `false` | Allows the Agent to detect common multi-line patterns automatically. |
| datadog.logs.containerCollectAll | bool | `false` | Enable this to allow log collection for all containers |
| datadog.logs.containerCollectUsingFiles | bool | `true` | Collect logs from files in /var/log/pods instead of using container runtime API |
| datadog.logs.enabled | bool | `false` | Enables this to activate Datadog Agent log collection |
| datadog.namespaceAnnotationsAsTags | object | `{}` | Provide a mapping of Kubernetes Namespace Annotations to Datadog Tags |
| datadog.namespaceLabelsAsTags | object | `{}` | Provide a mapping of Kubernetes Namespace Labels to Datadog Tags |
| datadog.networkMonitoring.enabled | bool | `false` | Enable network performance monitoring |
| datadog.networkPolicy.cilium.dnsSelector | object | kube-dns in namespace kube-system | Cilium selector of the DNS server entity |
| datadog.networkPolicy.create | bool | `false` | If true, create NetworkPolicy for all the components |
| datadog.networkPolicy.flavor | string | `"kubernetes"` | Flavor of the network policy to use. Can be: * kubernetes for networking.k8s.io/v1/NetworkPolicy * cilium     for cilium.io/v2/CiliumNetworkPolicy |
| datadog.nodeLabelsAsTags | object | `{}` | Provide a mapping of Kubernetes Node Labels to Datadog Tags |
| datadog.orchestratorExplorer.container_scrubbing | object | `{"enabled":true}` | Enable the scrubbing of containers in the kubernetes resource YAML for sensitive information |
| datadog.orchestratorExplorer.customResources | list | `[]` | Defines custom resources for the orchestrator explorer to collect |
| datadog.orchestratorExplorer.enabled | bool | `true` | Set this to false to disable the orchestrator explorer |
| datadog.originDetectionUnified.enabled | bool | `false` | Enabled enables unified mechanism for origin detection. Default: false. (Requires Agent 7.54.0+). |
| datadog.osReleasePath | string | `"/etc/os-release"` | Specify the path to your os-release file |
| datadog.otelCollector.config | string | `nil` | OTel collector configuration |
| datadog.otelCollector.enabled | bool | `false` | Enable the OTel Collector |
| datadog.otelCollector.ports | list | `[{"containerPort":"4317","name":"otel-grpc"},{"containerPort":"4318","name":"otel-http"}]` | Ports that OTel Collector is listening |
| datadog.otlp.logs.enabled | bool | `false` | Enable logs support in the OTLP ingest endpoint |
| datadog.otlp.receiver.protocols.grpc.enabled | bool | `false` | Enable the OTLP/gRPC endpoint |
| datadog.otlp.receiver.protocols.grpc.endpoint | string | `"0.0.0.0:4317"` | OTLP/gRPC endpoint |
| datadog.otlp.receiver.protocols.grpc.useHostPort | bool | `true` | Enable the Host Port for the OTLP/gRPC endpoint |
| datadog.otlp.receiver.protocols.http.enabled | bool | `false` | Enable the OTLP/HTTP endpoint |
| datadog.otlp.receiver.protocols.http.endpoint | string | `"0.0.0.0:4318"` | OTLP/HTTP endpoint |
| datadog.otlp.receiver.protocols.http.useHostPort | bool | `true` | Enable the Host Port for the OTLP/HTTP endpoint |
| datadog.podAnnotationsAsTags | object | `{}` | Provide a mapping of Kubernetes Annotations to Datadog Tags |
| datadog.podLabelsAsTags | object | `{}` | Provide a mapping of Kubernetes Labels to Datadog Tags |
| datadog.processAgent.containerCollection | bool | `true` | Set this to true to enable container collection # ref: https://docs.datadoghq.com/infrastructure/containers/?tab=helm |
| datadog.processAgent.enabled | bool | `true` | Set this to true to enable live process monitoring agent DEPRECATED. Set `datadog.processAgent.processCollection` or `datadog.processAgent.containerCollection` instead. # Note: /etc/passwd is automatically mounted when `processCollection`, `processDiscovery`, or `containerCollection` is enabled. # ref: https://docs.datadoghq.com/graphing/infrastructure/process/#kubernetes-daemonset |
| datadog.processAgent.processCollection | bool | `false` | Set this to true to enable process collection |
| datadog.processAgent.processDiscovery | bool | `true` | Enables or disables autodiscovery of integrations |
| datadog.processAgent.runInCoreAgent | bool | `false` | Set this to true to run the following features in the core agent: Live Processes, Live Containers, Process Discovery. # This requires Agent 7.57.0+ and Linux. |
| datadog.processAgent.stripProcessArguments | bool | `false` | Set this to scrub all arguments from collected processes # Requires datadog.processAgent.processCollection to be set to true to have any effect # ref: https://docs.datadoghq.com/infrastructure/process/?tab=linuxwindows#process-arguments-scrubbing |
| datadog.profiling.enabled | string | `nil` | Enable Continuous Profiler by injecting `DD_PROFILING_ENABLED` environment variable with the same value to all pods in the cluster Valid values are: - false: Profiler is turned off and can not be turned on by other means. - null: Profiler is turned off, but can be turned on by other means. - auto: Profiler is turned off, but the library will turn it on if the application is a good candidate for profiling. - true: Profiler is turned on. |
| datadog.prometheusScrape.additionalConfigs | list | `[]` | Allows adding advanced openmetrics check configurations with custom discovery rules. (Requires Agent version 7.27+) |
| datadog.prometheusScrape.enabled | bool | `false` | Enable autodiscovering pods and services exposing prometheus metrics. |
| datadog.prometheusScrape.serviceEndpoints | bool | `false` | Enable generating dedicated checks for service endpoints. |
| datadog.prometheusScrape.version | int | `2` | Version of the openmetrics check to schedule by default. |
| datadog.remoteConfiguration.enabled | bool | `true` | Set to true to enable remote configuration. DEPRECATED: Consider using remoteConfiguration.enabled instead |
| datadog.sbom.containerImage.enabled | bool | `false` | Enable SBOM collection for container images |
| datadog.sbom.containerImage.overlayFSDirectScan | bool | `false` | Use experimental overlayFS direct scan |
| datadog.sbom.containerImage.uncompressedLayersSupport | bool | `true` | Use container runtime snapshotter This should be set to true when using EKS, GKE or if containerd is configured to discard uncompressed layers. This feature will cause the SYS_ADMIN capability to be added to the Agent container. Setting this to false could cause a high error rate when generating SBOMs due to missing uncompressed layer. See https://docs.datadoghq.com/security/cloud_security_management/troubleshooting/vulnerabilities/#uncompressed-container-image-layers |
| datadog.sbom.host.enabled | bool | `false` | Enable SBOM collection for host filesystems |
| datadog.secretAnnotations | object | `{}` |  |
| datadog.secretBackend.arguments | string | `nil` | Configure the secret backend command arguments (space-separated strings). |
| datadog.secretBackend.command | string | `nil` | Configure the secret backend command, path to the secret backend binary. |
| datadog.secretBackend.enableGlobalPermissions | bool | `true` | Whether to create a global permission allowing Datadog agents to read all secrets when `datadog.secretBackend.command` is set to `"/readsecret_multiple_providers.sh"`. |
| datadog.secretBackend.roles | list | `[]` | Creates roles for Datadog to read the specified secrets - replacing `datadog.secretBackend.enableGlobalPermissions`. |
| datadog.secretBackend.timeout | string | `nil` | Configure the secret backend command timeout in seconds. |
| datadog.securityAgent.compliance.checkInterval | string | `"20m"` | Compliance check run interval |
| datadog.securityAgent.compliance.configMap | string | `nil` | Contains CSPM compliance benchmarks that will be used |
| datadog.securityAgent.compliance.enabled | bool | `false` | Set to true to enable Cloud Security Posture Management (CSPM) |
| datadog.securityAgent.compliance.host_benchmarks.enabled | bool | `true` | Set to false to disable host benchmarks. If enabled, this feature requires 160 MB extra memory for the `security-agent` container. (Requires Agent 7.47.0+) |
| datadog.securityAgent.compliance.xccdf.enabled | bool | `false` |  |
| datadog.securityAgent.runtime.activityDump.cgroupDumpTimeout | int | `20` | Set to the desired duration of a single container tracing (in minutes) |
| datadog.securityAgent.runtime.activityDump.cgroupWaitListSize | int | `0` | Set to the size of the wait list for already traced containers |
| datadog.securityAgent.runtime.activityDump.enabled | bool | `true` | Set to true to enable the collection of CWS activity dumps |
| datadog.securityAgent.runtime.activityDump.pathMerge.enabled | bool | `false` | Set to true to enable the merging of similar paths |
| datadog.securityAgent.runtime.activityDump.tracedCgroupsCount | int | `3` | Set to the number of containers that should be traced concurrently |
| datadog.securityAgent.runtime.enabled | bool | `false` | Set to true to enable Cloud Workload Security (CWS) |
| datadog.securityAgent.runtime.fimEnabled | bool | `false` | Set to true to enable Cloud Workload Security (CWS) File Integrity Monitoring |
| datadog.securityAgent.runtime.network.enabled | bool | `true` | Set to true to enable the collection of CWS network events |
| datadog.securityAgent.runtime.policies.configMap | string | `nil` | Contains CWS policies that will be used |
| datadog.securityAgent.runtime.securityProfile.anomalyDetection.enabled | bool | `true` | Set to true to enable CWS runtime drift events |
| datadog.securityAgent.runtime.securityProfile.autoSuppression.enabled | bool | `true` | Set to true to enable CWS runtime auto suppression |
| datadog.securityAgent.runtime.securityProfile.enabled | bool | `true` | Set to true to enable CWS runtime security profiles |
| datadog.securityAgent.runtime.syscallMonitor.enabled | bool | `false` | Set to true to enable the Syscall monitoring (recommended for troubleshooting only) |
| datadog.securityAgent.runtime.useSecruntimeTrack | bool | `true` | Set to true to send Cloud Workload Security (CWS) events directly to the Agent events explorer |
| datadog.securityContext | object | `{"runAsUser":0}` | Allows you to overwrite the default PodSecurityContext on the Daemonset or Deployment |
| datadog.serviceMonitoring.enabled | bool | `false` | Enable Universal Service Monitoring |
| datadog.site | string | `nil` | The site of the Datadog intake to send Agent data to. (documentation: https://docs.datadoghq.com/getting_started/site/) |
| datadog.systemProbe.apparmor | string | `"unconfined"` | Specify a apparmor profile for system-probe |
| datadog.systemProbe.bpfDebug | bool | `false` | Enable logging for kernel debug |
| datadog.systemProbe.btfPath | string | `""` | Specify the path to a BTF file for your kernel |
| datadog.systemProbe.collectDNSStats | bool | `true` | Enable DNS stat collection |
| datadog.systemProbe.conntrackInitTimeout | string | `"10s"` | the time to wait for conntrack to initialize before failing |
| datadog.systemProbe.conntrackMaxStateSize | int | `131072` | the maximum size of the userspace conntrack cache |
| datadog.systemProbe.debugPort | int | `0` | Specify the port to expose pprof and expvar for system-probe agent |
| datadog.systemProbe.enableConntrack | bool | `true` | Enable the system-probe agent to connect to the netlink/conntrack subsystem to add NAT information to connection data |
| datadog.systemProbe.enableDefaultKernelHeadersPaths | bool | `true` | Enable mount of default paths where kernel headers are stored |
| datadog.systemProbe.enableDefaultOsReleasePaths | bool | `true` | enable default os-release files mount |
| datadog.systemProbe.enableOOMKill | bool | `false` | Enable the OOM kill eBPF-based check |
| datadog.systemProbe.enableTCPQueueLength | bool | `false` | Enable the TCP queue length eBPF-based check |
| datadog.systemProbe.maxTrackedConnections | int | `131072` | the maximum number of tracked connections |
| datadog.systemProbe.mountPackageManagementDirs | list | `[]` | Enables mounting of specific package management directories when runtime compilation is enabled |
| datadog.systemProbe.runtimeCompilationAssetDir | string | `"/var/tmp/datadog-agent/system-probe"` | Specify a directory for runtime compilation assets to live in |
| datadog.systemProbe.seccomp | string | `"localhost/system-probe"` | Apply an ad-hoc seccomp profile to the system-probe agent to restrict its privileges |
| datadog.systemProbe.seccompRoot | string | `"/var/lib/kubelet/seccomp"` | Specify the seccomp profile root directory |
| datadog.tags | list | `[]` | List of static tags to attach to every metric, event and service check collected by this Agent. |
| datadog.useHostPID | bool | `true` | Run the agent in the host's PID namespace, required for origin detection / unified service tagging |
| existingClusterAgent.clusterchecksEnabled | bool | `true` | set this to false if you don’t want the agents to run the cluster checks of the joined external cluster agent |
| existingClusterAgent.join | bool | `false` | set this to true if you want the agents deployed by this chart to connect to a Cluster Agent deployed independently |
| existingClusterAgent.serviceName | string | `nil` | Existing service name to use for reaching the external Cluster Agent |
| existingClusterAgent.tokenSecretName | string | `nil` | Existing secret name to use for external Cluster Agent token |
| fips.customFipsConfig | object | `{}` | Configure a custom configMap to provide the FIPS configuration. Specify custom contents for the FIPS proxy sidecar container config (/etc/datadog-fips-proxy/datadog-fips-proxy.cfg). If empty, the default FIPS proxy sidecar container config is used. |
| fips.enabled | bool | `false` | Enable fips sidecar |
| fips.image.digest | string | `""` | Define the FIPS sidecar image digest to use, takes precedence over `fips.image.tag` if specified. |
| fips.image.name | string | `"fips-proxy"` |  |
| fips.image.pullPolicy | string | `"IfNotPresent"` | Datadog the FIPS sidecar image pull policy |
| fips.image.repository | string | `nil` | Override default registry + image.name for the FIPS sidecar container. |
| fips.image.tag | string | `"1.1.5"` | Define the FIPS sidecar container version to use. |
| fips.local_address | string | `"127.0.0.1"` | Set local IP address |
| fips.port | int | `9803` | Specifies which port is used by the containers to communicate to the FIPS sidecar. |
| fips.portRange | int | `15` | Specifies the number of ports used, defaults to 13 https://github.com/DataDog/datadog-agent/blob/7.44.x/pkg/config/config.go#L1564-L1577 |
| fips.resources | object | `{}` | Resource requests and limits for the FIPS sidecar container. |
| fips.use_https | bool | `false` | Option to enable https |
| fullnameOverride | string | `nil` | Override the full qualified app name |
| kube-state-metrics.image.repository | string | `"registry.k8s.io/kube-state-metrics/kube-state-metrics"` | Default kube-state-metrics image repository. |
| kube-state-metrics.nodeSelector | object | `{"kubernetes.io/os":"linux"}` | Node selector for KSM. KSM only supports Linux. |
| kube-state-metrics.rbac.create | bool | `true` | If true, create & use RBAC resources |
| kube-state-metrics.resources | object | `{}` | Resource requests and limits for the kube-state-metrics container. |
| kube-state-metrics.serviceAccount.create | bool | `true` | If true, create ServiceAccount, require rbac kube-state-metrics.rbac.create true |
| kube-state-metrics.serviceAccount.name | string | `nil` | The name of the ServiceAccount to use. |
| nameOverride | string | `nil` | Override name of app |
| providers.aks.enabled | bool | `false` | Activate all specificities related to AKS configuration. Required as currently we cannot auto-detect AKS. |
| providers.eks.ec2.useHostnameFromFile | bool | `false` | Use hostname from EC2 filesystem instead of fetching from metadata endpoint. |
| providers.gke.autopilot | bool | `false` | Enables Datadog Agent deployment on GKE Autopilot |
| providers.gke.cos | bool | `false` | Enables Datadog Agent deployment on GKE with Container-Optimized OS (COS) |
| providers.gke.gdc | bool | `false` | Enables Datadog Agent deployment on GKE on Google Distributed Cloud (GDC) |
| registry | string | `nil` | Registry to use for all Agent images (default to [gcr.io | eu.gcr.io | asia.gcr.io | datadoghq.azurecr.io | public.ecr.aws/datadog] depending on datadog.site value) |
| remoteConfiguration.enabled | bool | `true` | Set to true to enable remote configuration on the Cluster Agent (if set) and the node agent. Can be overridden if `datadog.remoteConfiguration.enabled` Preferred way to enable Remote Configuration. |
| targetSystem | string | `"linux"` | Target OS for this deployment (possible values: linux, windows) |

## Configuration options for Windows deployments
<a name="windows-config"></a>
Some options above are not working/not available on Windows, here is the list of **unsupported** options:

| Parameter                                | Reason                                           |
|------------------------------------------|--------------------------------------------------|
| `datadog.dogstatsd.useHostPID`           | Host PID not supported by Windows Containers     |
| `datadog.useHostPID`                     | Host PID not supported by Windows Containers     |
| `datadog.dogstatsd.useSocketVolume`      | Unix sockets not supported on Windows            |
| `datadog.dogstatsd.socketPath`           | Unix sockets not supported on Windows            |
| `datadog.processAgent.processCollection` | Unable to access host/other containers processes |
| `datadog.systemProbe.seccomp`            | System probe is not available for Windows        |
| `datadog.systemProbe.seccompRoot`        | System probe is not available for Windows        |
| `datadog.systemProbe.debugPort`          | System probe is not available for Windows        |
| `datadog.systemProbe.enableConntrack`    | System probe is not available for Windows        |
| `datadog.systemProbe.bpfDebug`           | System probe is not available for Windows        |
| `datadog.systemProbe.apparmor`           | System probe is not available for Windows        |
| `agents.useHostNetwork`                  | Host network not supported by Windows Containers |

### How to join a Cluster Agent from another helm chart deployment (Linux)

Because the Cluster Agent can only be deployed on Linux Node, the communication between
the Agents deployed on the Windows nodes with the a Cluster Agent need to be configured.

The following `datadog-values.yaml` file contains all the parameters needed to configure this communication.

```yaml
targetSystem: windows

existingClusterAgent:
  join: true
  serviceName: "<EXISTING_DCA_SERVICE_NAME>" # from the other datadog helm chart release
  tokenSecretName: "<EXISTING_DCA_SECRET_NAME>" # from the other datadog helm chart release

# Disabled datadogMetrics deployment since it should have been already deployed with the other chart release.
datadog-crds:
  crds:
    datadogMetrics: false

# Disable kube-state-metrics deployment
datadog:
  kubeStateMetricsEnabled: false
```
