# Datadog

![Version: 2.4.22](https://img.shields.io/badge/Version-2.4.22-informational?style=flat-square) ![AppVersion: 7](https://img.shields.io/badge/AppVersion-7-informational?style=flat-square)

[Datadog](https://www.datadoghq.com/) is a hosted infrastructure monitoring platform. This chart adds the Datadog Agent to all nodes in your cluster via a DaemonSet. It also optionally depends on the [kube-state-metrics chart](https://github.com/kubernetes/charts/tree/master/stable/kube-state-metrics). For more information about monitoring Kubernetes with Datadog, please refer to the [Datadog documentation website](https://docs.datadoghq.com/agent/basic_agent_usage/kubernetes/).

Datadog [offers two variants](https://hub.docker.com/r/datadog/agent/tags/), switch to a `-jmx` tag if you need to run JMX/java integrations. The chart also supports running [the standalone dogstatsd image](https://hub.docker.com/r/datadog/dogstatsd/tags/).

See the [Datadog JMX integration](https://docs.datadoghq.com/integrations/java/) to learn more.

## How to use Datadog Helm repository

You need to add this repository to your Helm repositories:

```
helm repo add datadog https://helm.datadoghq.com
helm repo update
```

## Prerequisites

Kubernetes 1.4+ or OpenShift 3.4+, note that:

- the Datadog Agent supports Kubernetes 1.4+
- The Datadog chart's defaults are tailored to Kubernetes 1.7.6+, see [Datadog Agent legacy Kubernetes versions documentation](https://github.com/DataDog/datadog-agent/tree/master/Dockerfiles/agent#legacy-kubernetes-versions) for adjustments you might need to make for older versions

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://kubernetes-charts.storage.googleapis.com/ | kube-state-metrics | =2.8.11 |

## Quick start

By default, the Datadog Agent runs in a DaemonSet. It can alternatively run inside a Deployment for special use cases.

**Note:** simultaneous DaemonSet + Deployment installation within a single release will be deprecated in a future version, requiring two releases to achieve this.

### Installing the Datadog Chart

To install the chart with the release name `<RELEASE_NAME>`, retrieve your Datadog API key from your [Agent Installation Instructions](https://app.datadoghq.com/account/settings#agent/kubernetes) and run:

```bash
helm install --name <RELEASE_NAME> \
  --set datadog.apiKey=<DATADOG_API_KEY> datadog/datadog
```

By default, this Chart creates a Secret and puts an API key in that Secret.
However, you can use manually created secret by setting the `datadog.apiKeyExistingSecret` value. After a few minutes, you should see hosts and metrics being reported in Datadog.

#### Create and provide a secret that contains your Datadog API Key

To create a secret that contains your Datadog API key, replace the <DATADOG_API_KEY> below with the API key for your organization. This secret is used in the manifest to deploy the Datadog Agent.

```bash
DATADOG_SECRET_NAME=datadog-secret
kubectl create secret generic $DATADOG_SECRET_NAME --from-literal api-key="<DATADOG_API_KEY>" --namespace="default"
```

**Note**: This creates a secret in the default namespace. If you are in a custom namespace, update the namespace parameter of the command before running it.

Now, the installation command contains the reference to the secret.

```bash
helm install --name <RELEASE_NAME> \
  --set datadog.apiKeyExistingSecret=$DATADOG_SECRET_NAME datadog/datadog
```

**Note**: Provide a secret for the application key (AppKey) using the `datadog.appKeyExistingSecret` chart variable.

### Enabling the Datadog Cluster Agent

Read about the Datadog Cluster Agent in the [official documentation](https://docs.datadoghq.com/agent/kubernetes/cluster/).

Run the following if you want to deploy the chart with the Datadog Cluster Agent:

```bash
helm install --name datadog-monitoring \
    --set datadog.apiKey=<DATADOG_API_KEY> \
    --set datadog.appKey=<DATADOG_APP_KEY> \
    --set clusterAgent.enabled=true \
    --set clusterAgent.metricsProvider.enabled=true \
    datadog/datadog
```

**Note**: Specifying `clusterAgent.metricsProvider.enabled=true` enables the External Metrics Server.
If you want to learn to use this feature, you can check out this [Datadog Cluster Agent walkthrough](https://github.com/DataDog/datadog-agent/blob/master/docs/cluster-agent/CUSTOM_METRICS_SERVER.md).

The Leader Election is enabled by default in the chart for the Cluster Agent. Only the Cluster Agent(s) participate in the election, in case you have several replicas configured (using `clusterAgent.replicas`.

#### Cluster Agent Token

You can specify the Datadog Cluster Agent token used to secure the communication between the Cluster Agent(s) and the Agents with `clusterAgent.token`.

**If you don't specify a token, a random one is generated at each deployment so you must use `--recreate-pods` to ensure all pod use the same token.** see[Datadog Chart notes](https://github.com/helm/charts/blob/57d3030941ad2ec2d6f97c86afdf36666658a884/datadog/datadog/templates/NOTES.txt#L49-L59) to learn more.

### Upgrading

#### From 1.x to 2.x

⚠️ Migrating from 1.x to 2.x requires a manual action.

The `datadog` chart has been refactored to regroup the `values.yaml` parameters in a more logical way.
Please follow the [migration guide](https://github.com/DataDog/helm-charts/blob/master/charts/datadog/docs/Migration_1.x_to_2.x.md) to update you `values.yaml` file.

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
helm delete <RELEASE_NAME> --purge
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

As a best practice, a YAML file that specifies the values for the chart parameters should be provided to configure the chart:

1. **Copy the default [`datadog-values.yaml`](values.yaml) value file.**
2. Set the `apiKey` parameter with your [Datadog API key](https://app.datadoghq.com/account/settings#api).
3. Upgrade the Datadog Helm chart with the new `datadog-values.yaml` file:

```bash
helm upgrade -f datadog-values.yaml <RELEASE_NAME> datadog/datadog --recreate-pods
```

See the [All configuration options](#all-configuration-options) section to discover all possibilities offered by the Datadog chart.

### Enabling Log Collection

Update your [datadog-values.yaml](values.yaml) file with the following log collection configuration:

```yaml
datadog:
  # (...)
  logs:
    enabled: true
    containerCollectAll: true
```

then upgrade your Datadog Helm chart:

```bash
helm upgrade -f datadog-values.yaml <RELEASE_NAME> datadog/datadog --recreate-pods
```

### Enabling Process Collection

Update your [datadog-values.yaml](values.yaml) file with the process collection configuration:

```yaml
datadog:
  # (...)
  processAgent:
    enabled: true
    processCollection: true
```

then upgrade your Datadog Helm chart:

```bash
helm upgrade -f datadog-values.yaml <RELEASE_NAME> datadog/datadog --recreate-pods
```

### Enabling System Probe Collection

The system-probe agent only runs in dedicated container environment. Update your [datadog-values.yaml](values.yaml) file with the system-probe collection configuration:

```yaml
datadog:
  # (...)
  systemProbe:
    # (...)
    enabled: true

# (...)
```

then upgrade your Datadog Helm chart:

```bash
helm upgrade -f datadog-values.yaml <RELEASE_NAME> datadog/datadog --recreate-pods
```

### Kubernetes event collection

Use the [Datadog Cluster Agent](#enabling-the-datadog-cluster-agent) to collect Kubernetes events. Please read [the official documentation](https://docs.datadoghq.com/agent/kubernetes/event_collection/) for more context.

Alternatively set the `datadog.leaderElection`, `datadog.collectEvents` and `rbac.create` options to `true` in order to enable Kubernetes event collection.

### conf.d and checks.d

The Datadog [entrypoint](https://github.com/DataDog/datadog-agent/blob/master/Dockerfiles/agent/entrypoint/89-copy-customfiles.sh) copies files with a `.yaml` extension found in `/conf.d` and files with `.py` extension in `/checks.d` to `/etc/datadog-agent/conf.d` and `/etc/datadog-agent/checks.d` respectively.

The keys for `datadog.confd` and `datadog.checksd` should mirror the content found in their respective ConfigMaps. Update your [datadog-values.yaml](values.yaml) file with the check configurations:

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
helm upgrade -f datadog-values.yaml <RELEASE_NAME> datadog/datadog --recreate-pods
```

For more details, please refer to [the documentation](https://docs.datadoghq.com/agent/kubernetes/integrations/).

### Kubernetes Labels and Annotations

To map Kubernetes node labels and pod labels and annotations to Datadog tags, provide a dictionary with kubernetes labels/annotations as keys and Datadog tags key as values in your [datadog-values.yaml](values.yaml) file:

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
helm upgrade -f datadog-values.yaml <RELEASE_NAME> datadog/datadog --recreate-pods
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
Update your [datadog-values.yaml](values.yaml) file to disable apparmor enforcement:

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

## All configuration options

The following table lists the configurable parameters of the Datadog chart and their default values. Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example,

```bash
helm install --name <RELEASE_NAME> \
  --set datadog.apiKey=<DATADOG_API_KEY>,datadog.logLevel=DEBUG \
  datadog/datadog
```

<<<<<<< HEAD
## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| agents.affinity | object | `{}` | Allow the DaemonSet to schedule using affinity rules |
| agents.containers.agent.env | list | `[]` | Additional environment variables for the agent container |
| agents.containers.agent.livenessProbe | object | Every 15s / 6 KO / 1 OK | Override default agent liveness probe settings |
| agents.containers.agent.logLevel | string | `nil` | Set logging verbosity, valid log levels are: trace, debug, info, warn, error, critical, and off |
| agents.containers.agent.readinessProbe | object | Every 15s / 6 KO / 1 OK | Override default agent readiness probe settings |
| agents.containers.agent.resources | object | `{}` | Resource requests and limits for the agent container. |
| agents.containers.initContainers.resources | object | `{}` | Resource requests and limits for the init containers |
| agents.containers.processAgent.env | list | `[]` | Additional environment variables for the process-agent container |
| agents.containers.processAgent.logLevel | string | `nil` | Set logging verbosity, valid log levels are: trace, debug, info, warn, error, critical, and off |
| agents.containers.processAgent.resources | object | `{}` | Resource requests and limits for the process-agent container |
| agents.containers.securityAgent.env | string | `nil` | Additional environment variables for the security-agent container |
| agents.containers.securityAgent.logLevel | string | `nil` | Set logging verbosity, valid log levels are: trace, debug, info, warn, error, critical, and off |
| agents.containers.securityAgent.resources | object | `{}` | Resource requests and limits for the security-agent container |
| agents.containers.systemProbe.env | list | `[]` | Additional environment variables for the system-probe container |
| agents.containers.systemProbe.logLevel | string | `nil` | Set logging verbosity, valid log levels are: trace, debug, info, warn, error, critical, and off. |
| agents.containers.systemProbe.resources | object | `{}` | Resource requests and limits for the system-probe container |
| agents.containers.traceAgent.env | string | `nil` | Additional environment variables for the trace-agent container |
| agents.containers.traceAgent.livenessProbe | object | Every 15s | Override default agent liveness probe settings |
| agents.containers.traceAgent.logLevel | string | `nil` | Set logging verbosity, valid log levels are: trace, debug, info, warn, error, critical, and off |
| agents.containers.traceAgent.resources | object | `{}` | Resource requests and limits for the trace-agent container |
| agents.customAgentConfig | object | `{}` | Specify custom contents for the datadog agent config (datadog.yaml) |
| agents.dnsConfig | object | `{}` | specify dns configuration options for datadog cluster agent containers e.g ndots |
| agents.enabled | bool | `true` | You should keep Datadog DaemonSet enabled! |
| agents.image.doNotCheckTag | string | `nil` | Skip the version<>chart compatibility check |
| agents.image.pullPolicy | string | `"IfNotPresent"` | Datadog Agent image pull policy |
| agents.image.pullSecrets | list | `[]` | Datadog Agent repository pullSecret (ex: specify docker registry credentials) |
| agents.image.repository | string | `"datadog/agent"` | Datadog Agent image repository to use |
| agents.image.tag | string | `"7.22.0"` | Define the Agent version to use |
| agents.networkPolicy.create | bool | `false` | If true, create a NetworkPolicy for the agents |
| agents.nodeSelector | object | `{}` | Allow the DaemonSet to schedule on selected nodes |
| agents.podAnnotations | object | `{}` | Annotations to add to the DaemonSet's Pods |
| agents.podLabels | object | `{}` | Sets podLabels if defined |
| agents.podSecurity.apparmor.enabled | bool | `true` | If true, enable apparmor enforcement |
| agents.podSecurity.apparmorProfiles | list | `["runtime/default"]` | Allowed apparmor profiles |
| agents.podSecurity.capabilites | list | `["SYS_ADMIN","SYS_RESOURCE","SYS_PTRACE","NET_ADMIN","NET_BROADCAST","IPC_LOCK","AUDIT_CONTROL","AUDIT_READ"]` | Allowed capabilites |
| agents.podSecurity.podSecurityPolicy.create | bool | `false` | If true, create a PodSecurityPolicy resource for Agent pods |
| agents.podSecurity.privileged | bool | `false` | If true, Allow to run privileged containers |
| agents.podSecurity.seLinuxContext | object | Must run as spc_t | Provide seLinuxContext configuration for PSP/SCC |
| agents.podSecurity.seccompProfiles | list | `["runtime/default","localhost/system-probe"]` | Allowed seccomp profiles |
| agents.podSecurity.securityContextConstraints.create | bool | `false` | If true, create a SecurityContextConstraints resource for Agent pods |
| agents.podSecurity.volumes | list | `["configMap","downwardAPI","emptyDir","hostPath","secret"]` | Allowed volumes types |
| agents.priorityClassName | string | `nil` | Sets PriorityClassName if defineds |
| agents.rbac.create | bool | `true` | If true, create & use RBAC resources |
| agents.rbac.serviceAccountName | string | `"default"` | Specify service account name to use (usually pre-existing, created if create is true) |
| agents.tolerations | list | `[]` | Allow the DaemonSet to schedule on tainted nodes (requires Kubernetes >= 1.6) |
| agents.updateStrategy | object | `{"rollingUpdate":{"maxUnavailable":"10%"},"type":"RollingUpdate"}` | Allow the DaemonSet to perform a rolling update on helm update |
| agents.useConfigMap | string | `nil` | Configures a configmap to provide the agent configuration. Use this in combination with the `agents.customAgentConfig` parameter. |
| agents.useHostNetwork | bool | `false` | Bind ports on the hostNetwork |
| agents.volumeMounts | list | `[]` | Specify additional volumes to mount in the dd-agent container |
| agents.volumes | list | `[]` | Specify additional volumes to mount in the dd-agent container |
| clusterAgent.admissionController.enabled | bool | `false` | Enable the admissionController to be able to inject APM/Dogstatsd config and standard tags (env, service, version) automatically into your pods |
| clusterAgent.admissionController.mutateUnlabelled | bool | `false` | Enable injecting config without having the pod label 'admission.datadoghq.com/enabled="true"' |
| clusterAgent.affinity | object | `{}` | Allow the Cluster Agent Deployment to schedule using affinity rules |
| clusterAgent.command | list | `[]` | Command to run in the Cluster Agent container as entrypoint |
| clusterAgent.confd | object | `{}` | Provide additional cluster check configurations |
| clusterAgent.createPodDisruptionBudget | bool | `false` | Create pod disruption budget for Cluster Agent deployments |
| clusterAgent.datadog_cluster_yaml | object | `{}` | Specify custom contents for the datadog cluster agent config (datadog-cluster.yaml) |
| clusterAgent.dnsConfig | object | `{}` | Specify dns configuration options for datadog cluster agent containers e.g ndots |
| clusterAgent.enabled | bool | `false` | Set this to true to enable Datadog Cluster Agent |
| clusterAgent.env | list | `[]` | Set environment variables specific to Cluster Agent |
| clusterAgent.healthPort | int | `5555` | Port number use the cluster-agent to server healthz endpoint |
| clusterAgent.image.pullPolicy | string | `"IfNotPresent"` | Cluster Agent image pullPolicy |
| clusterAgent.image.pullSecrets | list | `[]` | Cluster Agent repository pullSecret (ex: specify docker registry credentials) |
| clusterAgent.image.repository | string | `"datadog/cluster-agent"` | Cluster Agent image repository to use |
| clusterAgent.image.tag | string | `"1.8.0"` | Cluster Agent image tag to use |
| clusterAgent.livenessProbe | object | Every 15s / 6 KO / 1 OK | Override default agent liveness probe settings |
| clusterAgent.metricsProvider.aggregator | string | `"avg"` | Define the aggregator the cluster agent will use to process the metrics. The options are (avg, min, max, sum) |
| clusterAgent.metricsProvider.createReaderRbac | bool | `true` | Create `external-metrics-reader` RBAC automatically (to allow HPA to read data from Cluster Agent) |
| clusterAgent.metricsProvider.enabled | bool | `false` | Set this to true to enable Metrics Provider |
| clusterAgent.metricsProvider.service.port | int | `8443` | Set port of cluster-agent metrics server service (Kubernetes >= 1.15) |
| clusterAgent.metricsProvider.service.type | string | `"ClusterIP"` | Set type of cluster-agent metrics server service |
| clusterAgent.metricsProvider.useDatadogMetrics | bool | `false` | Enable usage of DatadogMetric CRD to autoscale on arbitrary Datadog queries |
| clusterAgent.metricsProvider.wpaController | bool | `false` | Enable informer and controller of the watermark pod autoscaler |
| clusterAgent.networkPolicy.create | bool | `false` | If true, create a NetworkPolicy for the cluster agent |
| clusterAgent.nodeSelector | object | `{}` | Allow the Cluster Agent Deployment to be scheduled on selected nodes |
| clusterAgent.podAnnotations | object | `{}` | Annotations to add to the cluster-agents's pod(s) |
| clusterAgent.priorityClassName | string | `nil` | Name of the priorityClass to apply to the Cluster Agent |
| clusterAgent.rbac.create | bool | `true` | If true, create & use RBAC resources |
| clusterAgent.rbac.serviceAccountName | string | `"default"` | Specify service account name to use (usually pre-existing, created if create is true) |
| clusterAgent.readinessProbe | object | Every 15s / 6 KO / 1 OK | Override default cluster-agent readiness probe settings |
| clusterAgent.replicas | int | `1` | Specify the of cluster agent replicas, if > 1 it allow the cluster agent to work in HA mode. |
| clusterAgent.resources | object | `{}` | Datadog cluster-agent resource requests and limits. |
| clusterAgent.strategy | object | `{"rollingUpdate":{"maxSurge":1,"maxUnavailable":0},"type":"RollingUpdate"}` | Allow the Cluster Agent deployment to perform a rolling update on helm update |
| clusterAgent.token | string | `""` | Cluster Agent token is a preshared key between node agents and cluster agent (autogenerated if empty, needs to be at least 32 characters a-zA-z) |
| clusterAgent.useHostNetwork | string | `nil` | Bind ports on the hostNetwork |
| clusterAgent.volumeMounts | list | `[]` | Specify additional volumes to mount in the cluster-agent container |
| clusterAgent.volumes | list | `[]` | Specify additional volumes to mount in the cluster-agent container |
| clusterChecksRunner.affinity | object | `{}` | Allow the ClusterChecks Deployment to schedule using affinity rules. |
| clusterChecksRunner.createPodDisruptionBudget | bool | `false` | Create the pod disruption budget to apply to the cluster checks agents |
| clusterChecksRunner.dnsConfig | object | `{}` | specify dns configuration options for datadog cluster agent containers e.g ndots |
| clusterChecksRunner.enabled | bool | `false` | If true, deploys agent dedicated for running the Cluster Checks instead of running in the Daemonset's agents. |
| clusterChecksRunner.env | list | `[]` | Environment variables specific to Cluster Checks Runner |
| clusterChecksRunner.image.pullPolicy | string | `"IfNotPresent"` | Datadog Agent image pull policy |
| clusterChecksRunner.image.pullSecrets | list | `[]` | Datadog Agent repository pullSecret (ex: specify docker registry credentials) |
| clusterChecksRunner.image.repository | string | `"datadog/agent"` | Datadog Agent image repository to use |
| clusterChecksRunner.image.tag | string | `"7.22.0"` | Define the Agent version to use |
| clusterChecksRunner.livenessProbe | object | Every 15s / 6 KO / 1 OK | Override default agent liveness probe settings |
| clusterChecksRunner.networkPolicy.create | bool | `false` | If true, create a NetworkPolicy for the cluster checks runners |
| clusterChecksRunner.nodeSelector | object | `{}` | Allow the ClusterChecks Deployment to schedule on selected nodes |
| clusterChecksRunner.rbac.create | bool | `true` | If true, create & use RBAC resources |
| clusterChecksRunner.rbac.dedicated | bool | `false` | If true, use a dedicated RBAC resource for the cluster checks agent(s) |
| clusterChecksRunner.rbac.serviceAccountAnnotations | object | `{}` | Annotations to add to the ServiceAccount if clusterChecksRunner.rbac.dedicated is true |
| clusterChecksRunner.rbac.serviceAccountName | string | `"default"` | Specify service account name to use (usually pre-existing, created if create is true) |
| clusterChecksRunner.readinessProbe | object | Every 15s / 6 KO / 1 OK | Override default agent readiness probe settings |
| clusterChecksRunner.replicas | int | `2` | Number of Cluster Checks Runner instances |
| clusterChecksRunner.resources | object | `{}` | Datadog clusterchecks-agent resource requests and limits. |
| clusterChecksRunner.strategy | object | `{"rollingUpdate":{"maxSurge":1,"maxUnavailable":0},"type":"RollingUpdate"}` | Allow the ClusterChecks deployment to perform a rolling update on helm update |
| clusterChecksRunner.tolerations | list | `[]` | Tolerations for pod assignment |
| clusterChecksRunner.volumeMounts | list | `[]` | Specify additional volumes to mount in the cluster checks container |
| clusterChecksRunner.volumes | list | `[]` | Specify additional volumes to mount in the cluster checks container |
| datadog.apiKey | string | `"<DATADOG_API_KEY>"` | Your Datadog API key ref: https://app.datadoghq.com/account/settings#agent/kubernetes |
| datadog.apiKeyExistingSecret | string | `nil` | Use existing Secret which stores API key instead of creating a new one |
| datadog.apm.enabled | bool | `false` | Enable this to enable APM and tracing, on port 8126 |
| datadog.apm.hostSocketPath | string | `"/var/run/datadog/"` | Host path to the trace-agent socket |
| datadog.apm.port | int | `8126` | Override the trace Agent port |
| datadog.apm.socketPath | string | `"/var/run/datadog/apm.socket"` | Path to the trace-agent socket |
| datadog.apm.useSocketVolume | bool | `false` | Enable APM over Unix Domain Socket |
| datadog.appKey | string | `nil` | Datadog APP key required to use metricsProvider |
| datadog.appKeyExistingSecret | string | `nil` | Use existing Secret which stores APP key instead of creating a new one |
| datadog.checksd | object | `{}` | Provide additional custom checks as python code |
| datadog.clusterChecks.enabled | bool | `false` | Enable the Cluster Checks feature on both the cluster-agents and the daemonset |
| datadog.clusterName | string | `nil` | Set a unique cluster name to allow scoping hosts and Cluster Checks easily |
| datadog.collectEvents | bool | `false` | Enables this to start event collection from the kubernetes API |
| datadog.confd | object | `{}` | Provide additional check configurations (static and Autodiscovery) |
| datadog.criSocketPath | string | `nil` | Path to the container runtime socket (if different from Docker) |
| datadog.dd_url | string | `nil` | The host of the Datadog intake server to send Agent data to, only set this option if you need the Agent to send data to a custom URL |
| datadog.dockerSocketPath | string | `nil` | Path to the docker socket |
| datadog.dogstatsd.hostSocketPath | string | `"/var/run/datadog/"` | Host path to the DogStatsD socket |
| datadog.dogstatsd.nonLocalTraffic | bool | `false` | Enable this to make each node accept non-local statsd traffic |
| datadog.dogstatsd.originDetection | bool | `false` | Enable origin detection for container tagging |
| datadog.dogstatsd.port | int | `8125` | Override the Agent DogStatsD port |
| datadog.dogstatsd.socketPath | string | `"/var/run/datadog/dsd.socket"` | Path to the DogStatsD socket |
| datadog.dogstatsd.tagCardinality | string | `"low"` | Sets the tag cardinality relative to the origin detection |
| datadog.dogstatsd.useHostPID | bool | `false` | Run the agent in the host's PID namespace |
| datadog.dogstatsd.useHostPort | bool | `false` | Sets the hostPort to the same value of the container port |
| datadog.dogstatsd.useSocketVolume | bool | `false` | Enable dogstatsd over Unix Domain Socket |
| datadog.env | list | `[]` | Set environment variables for all Agents |
| datadog.hostVolumeMountPropagation | string | `"None"` | Allow to specify the `mountPropagation` value on all volumeMounts using HostPath |
| datadog.kubeStateMetricsEnabled | bool | `true` | If true, deploys the kube-state-metrics deployment |
| datadog.leaderElection | bool | `false` | Enables leader election mechanism for event collection |
| datadog.leaderLeaseDuration | string | `nil` | Set the lease time for leader election in second |
| datadog.logLevel | string | `"INFO"` | Set logging verbosity, valid log levels are: trace, debug, info, warn, error, critical, off |
| datadog.logs.containerCollectAll | bool | `false` | Enable this to allow log collection for all containers |
| datadog.logs.containerCollectUsingFiles | bool | `true` | Collect logs from files in /var/log/pods instead of using container runtime API |
| datadog.logs.enabled | bool | `false` | Enables this to activate Datadog Agent log collection |
| datadog.nodeLabelsAsTags | object | `{}` | Provide a mapping of Kubernetes Node Labels to Datadog Tags |
| datadog.orchestratorExplorer.container_scrubbing | object | `{"enabled":true}` | Enable the scrubbing of containers in the kubernetes resource YAML for sensitive information |
| datadog.orchestratorExplorer.enabled | bool | `false` | Set this to true to enable the orchestrator explorer |
| datadog.podAnnotationsAsTags | object | `{}` | Provide a mapping of Kubernetes Annotations to Datadog Tags |
| datadog.podLabelsAsTags | object | `{}` | Provide a mapping of Kubernetes Labels to Datadog Tags |
| datadog.processAgent.enabled | bool | `true` | Set this to true to enable live process monitoring agent |
| datadog.processAgent.processCollection | bool | `false` | Set this to true to enable process collection in process monitoring agent |
| datadog.securityAgent.compliance.checkInterval | string | `"20m"` | Compliance check run interval |
| datadog.securityAgent.compliance.configMap | string | `nil` | Contains compliance benchmarks that will be used |
| datadog.securityAgent.compliance.enabled | bool | `false` | Set this to true to enable compliance checks |
| datadog.securityAgent.runtime.enabled | bool | `false` | Set to true to enable the Security Runtime Module |
| datadog.securityAgent.runtime.policies.configMap | string | `nil` | Contains policies that will be used |
| datadog.securityAgent.runtime.syscallMonitor.enabled | bool | `false` | Set to true to enable the Syscall monitoring. |
| datadog.securityContext | object | `{}` | Allows you to overwrite the default securityContext applied to the container |
| datadog.site | string | `nil` | The site of the Datadog intake to send Agent data to |
| datadog.systemProbe.apparmor | string | `"unconfined"` | Specify a apparmor profile for system-probe |
| datadog.systemProbe.bpfDebug | bool | `false` | Enable logging for kernel debug |
| datadog.systemProbe.collectDNSStats | bool | `false` | Enable DNS stat collection |
| datadog.systemProbe.debugPort | int | `0` | Specify the port to expose pprof and expvar for system-probe agent |
| datadog.systemProbe.enableConntrack | bool | `true` | Enable the system-probe agent to connect to the netlink/conntrack subsystem to add NAT information to connection data |
| datadog.systemProbe.enableOOMKill | bool | `false` | Enable the OOM kill eBPF-based check |
| datadog.systemProbe.enableTCPQueueLength | bool | `false` | Enable the TCP queue length eBPF-based check |
| datadog.systemProbe.enabled | bool | `false` | Set this to true to enable system-probe agent |
| datadog.systemProbe.seccomp | string | `"localhost/system-probe"` | Apply an ad-hoc seccomp profile to the system-probe agent to restrict its privileges |
| datadog.systemProbe.seccompRoot | string | `"/var/lib/kubelet/seccomp"` | Specify the seccomp profile root directory |
| datadog.tags | list | `[]` | List of statoc tags to attach to every metric, event and service check collected by this Agent. |
| fullnameOverride | string | `nil` | Override the full qualified app name |
| kube-state-metrics.rbac.create | bool | `true` | If true, create & use RBAC resources |
| kube-state-metrics.resources | object | `{}` | Resource requests and limits for the kube-state-metrics container. |
| kube-state-metrics.serviceAccount.create | bool | `true` | If true, create ServiceAccount, require rbac kube-state-metrics.rbac.create true |
| kube-state-metrics.serviceAccount.name | string | `nil` | The name of the ServiceAccount to use. |
| nameOverride | string | `nil` | Override name of app |
| targetSystem | string | `"linux"` | Target OS for this deployment (possible values: linux, windows) |
=======
| Parameter                                                    | Description                                                                                                                                                                                                                                                                                         | Default                                           |
|--------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------|
| `targetSystem`                                               | Target OS of this installation (supported: `linux`, `windows`)                                                                                                                                                                                                                                      | `linux`                                           |
| `datadog.apiKey`                                             | Your Datadog API key                                                                                                                                                                                                                                                                                | `nil` You must provide your own key               |
| `datadog.apiKeyExistingSecret`                               | If set, use the secret with a provided name instead of creating a new one                                                                                                                                                                                                                           | `nil`                                             |
| `datadog.appKey`                                             | Datadog APP key required to use metricsProvider                                                                                                                                                                                                                                                     | `nil` You must provide your own key               |
| `datadog.appKeyExistingSecret`                               | If set, use the secret with a provided name instead of creating a new one                                                                                                                                                                                                                           | `nil`                                             |
| `agents.image.repository`                                    | The image repository to pull from                                                                                                                                                                                                                                                                   | `datadog/agent`                                   |
| `agents.image.tag`                                           | The image tag to pull                                                                                                                                                                                                                                                                               | `7.19.0`                                          |
| `agents.image.doNotCheckTag`                                 | By default, the helm chart will check that the version provided in `agents.image.tag` is superior to the minimal version requested by the chart. If `doNotCheckTag` is explicitly set to `true`, this check is skipped. This is useful for custom tags that are not respecting semantic versioning. | `false`                                           |
| `agents.image.pullPolicy`                                    | Image pull policy                                                                                                                                                                                                                                                                                   | `IfNotPresent`                                    |
| `agents.image.pullSecrets`                                   | Image pull secrets                                                                                                                                                                                                                                                                                  | `nil`                                             |
| `nameOverride`                                               | Override name of app                                                                                                                                                                                                                                                                                | `""`                                              |
| `fullnameOverride`                                           | Override full name of app                                                                                                                                                                                                                                                                           | `""`                                              |
| `agents.rbac.create`                                         | If true, create & use RBAC resources                                                                                                                                                                                                                                                                | `true`                                            |
| `agents.rbac.serviceAccountName`                             | existing ServiceAccount to use (ignored if rbac.create=true)                                                                                                                                                                                                                                        | `default`                                         |
| `datadog.site`                                               | Site ('datadoghq.com' or 'datadoghq.eu')                                                                                                                                                                                                                                                            | `nil`                                             |
| `datadog.dd_url`                                             | Datadog intake server                                                                                                                                                                                                                                                                               | `nil`                                             |
| `datadog.envFrom`                                            | References to Configmaps and/or Secrets for additional Datadog environment variables                                                                                                                                                                                                                | `nil`                                             |
| `datadog.env`                                                | Additional Datadog environment variables                                                                                                                                                                                                                                                            | `nil`                                             |
| `datadog.logLevel`                                           | Agent log verbosity (possible values: trace, debug, info, warn, error, critical, and off)                                                                                                                                                                                                           | `INFO`                                            |
| `datadog.logs.enabled`                                       | Enable log collection                                                                                                                                                                                                                                                                               | `nil`                                             |
| `datadog.logs.containerCollectAll`                           | Collect logs from all containers                                                                                                                                                                                                                                                                    | `nil`                                             |
| `datadog.logs.containerCollectUsingFiles`                    | Collect container logs from files on disk instead of container runtime API                                                                                                                                                                                                                          | `true`                                            |
| `datadog.apm.enabled`                                        | Enable tracing from the host                                                                                                                                                                                                                                                                        | `false`                                           |
| `datadog.apm.port`                                           | Used to override the default agent APM Port                                                                                                                                                                                                                                                         | `8126`                                            |
| `datadog.apm.useSocketVolume`                                | Enable APM over Unix Domain Socket                                                                                                                                                                                                                                                                  | `False`                                           |
| `datadog.apm.socketPath`                                     | Custom path to the socket, has to be located in the `/var/run/datadog/` folder path                                                                                                                                                                                                                 | `/var/run/datadog/apm.socket`                     |
| `datadog.apm.hostPath`                                       | host directory that contains the trace-agent socket path                                                                                                                                                                                                                                            | `/var/run/datadog/`                               |
| `datadog.clusterChecks.enabled`                              | Enable Cluster Checks on both the Cluster Agent and the Agent daemonset                                                                                                                                                                                                                             | `false`                                           |
| `datadog.processAgent.enabled`                               | Enable live process and container monitoring agent. Possible values: `true` enable process-agent, `false` disable process-agent                                                                                                                                                                     | `true`                                            |
| `datadog.processAgent.processCollection`                     | Enable live process collection. Possible values: `true` enable process collection, `false` disable process collection                                                                                                                                                                               | `false`                                           |
| `datadog.checksd`                                            | Additional custom checks as python code                                                                                                                                                                                                                                                             | `nil`                                             |
| `datadog.confd`                                              | Additional check configurations (static and Autodiscovery)                                                                                                                                                                                                                                          | `nil`                                             |
| `datadog.dockerSocketPath`                                   | Path to the docker socket                                                                                                                                                                                                                                                                           | `/var/run/docker.sock`                            |
| `datadog.criSocketPath`                                      | Path to the container runtime socket (default is Docker runtime)                                                                                                                                                                                                                                    | `nil`                                             |
| `datadog.tags`                                               | Set host tags                                                                                                                                                                                                                                                                                       | `nil`                                             |
| `datadog.dogstatsd.originDetection`                          | Enable origin detection for container tagging                                                                                                                                                                                                                                                       | `False`                                           |
| `datadog.dogstatsd.tagCardinality`                           | Set the tag cardinality of the origin detection for container tagging                                                                                                                                                                                                                               | `low`                                             |
| `datadog.dogstatsd.port`                                     | Used to override the default agent DogStatsD Port                                                                                                                                                                                                                                                   | `8125`                                            |
| `datadog.dogstatsd.useHostPID`                               | If true, use the host's PID namespace                                                                                                                                                                                                                                                               | `nil`                                             |
| `datadog.dogstatsd.useHostPort`                              | If true, use the same ports for both host and container                                                                                                                                                                                                                                             | `nil`                                             |
| `datadog.dogstatsd.nonLocalTraffic`                          | Enable statsd reporting from any external ip                                                                                                                                                                                                                                                        | `False`                                           |
| `datadog.dogstatsd.useSocketVolume`                          | Enable dogstatsd over Unix Domain Socket                                                                                                                                                                                                                                                            | `False`                                           |
| `datadog.dogstatsd.socketPath`                               | Custom path to the socket, has to be located in the `/var/run/datadog/` folder path                                                                                                                                                                                                                 | `/var/run/datadog/dsd.socket`                     |
| `datadog.dogstatsd.hostPath`                                 | host directory that contains the dogstatsd socket                                                                                                                                                                                                                                                   | `/var/run/datadog/`                               |
| `datadog.nodeLabelsAsTags`                                   | Kubernetes Node Labels to Datadog Tags mapping                                                                                                                                                                                                                                                      | `nil`                                             |
| `datadog.podAnnotationsAsTags`                               | Kubernetes Annotations to Datadog Tags mapping                                                                                                                                                                                                                                                      | `nil`                                             |
| `datadog.podLabelsAsTags`                                    | Kubernetes Labels to Datadog Tags mapping                                                                                                                                                                                                                                                           | `nil`                                             |
| `datadog.securityContext`                                    | Allows you to overwrite the default securityContext applied to the container                                                                                                                                                                                                                        | `nil`                                             |
| `datadog.hostVolumeMountPropagation`                         | Allow to specify the `mountPropagation` value on all volumeMounts using HostPath                                                                                                                                                                                                                    | `None`                                            |
| `datadog.acInclude`                                          | (Deprecated) Include containers based on image name                                                                                                                                                                                                                                                 | `nil`                                             |
| `datadog.acExclude`                                          | (Deprecated) Exclude containers based on image name                                                                                                                                                                                                                                                 | `nil`                                             |
| `datadog.containerInclude`                                   | Include containers based on image name, container name or kubernetes namespace                                                                                                                                                                                                                      | `nil`                                             |
| `datadog.containerExclude`                                   | Exclude containers based on image name, container name or kubernetes namespace                                                                                                                                                                                                                      | `nil`                                             |
| `datadog.containerIncludeMetrics`                            | Include containers for metrics collection based on image name, container name or kubernetes namespace                                                                                                                                                                                               | `nil`                                             |
| `datadog.containerExcludeMetrics`                            | Exclude containers from metrics based on image name, container name or kubernetes namespace                                                                                                                                                                                                         | `nil`                                             |
| `datadog.containerIncludeLogs`                               | Include containers for logs collection based on image name, container name or kubernetes namespace                                                                                                                                                                                                  | `nil`                                             |
| `datadog.containerExcludeLogs`                               | Exclude containers from logs based on image name, container name or kubernetes namespace                                                                                                                                                                                                            | `nil`                                             |
| `datadog.systemProbe.enabled`                                | enable system probe collection                                                                                                                                                                                                                                                                      | `false`                                           |
| `datadog.systemProbe.seccomp`                                | Apply an ad-hoc seccomp profile to system-probe to restrict its privileges                                                                                                                                                                                                                          | `localhost/system-probe`                          |
| `datadog.systemProbe.seccompRoot`                            | Seccomp root directory for system-probe                                                                                                                                                                                                                                                             | `/var/lib/kubelet/seccomp`                        |
| `datadog.systemProbe.debugPort`                              | The port to expose pprof and expvar for system-probe agent, it is not enabled if the value is set to 0                                                                                                                                                                                              | `0`                                               |
| `datadog.systemProbe.enableConntrack`                        | If true, system-probe connects to the netlink/conntrack subsystem to add NAT information to connection data. Ref: http://conntrack-tools.netfilter.org/                                                                                                                                             | `true`                                            |
| `datadog.systemProbe.bpfDebug`                               | If true, system-probe writes debug logs to /sys/kernel/debug/tracing/trace_pipe                                                                                                                                                                                                                     | `false`                                           |
| `datadog.systemProbe.apparmor`                               | Apparmor profile for system-probe                                                                                                                                                                                                                                                                   | `unconfined`                                      |
| `datadog.systemProbe.enableTCPQueueLength`                   | Enable the TCP queue length eBPF-based check                                                                                                                                                                                                                                                        | `false`                                           |
| `datadog.systemProbe.enableOOMKill`                          | Enable the OOM kill eBPF-based check                                                                                                                                                                                                                                                                | `false`                                           |
| `datadog.orchestratorExplorer.enabled`                       | Enable the Orchestrator Explorer data collection                                                                                                                                                                                                                                                    | `false`                                           |
| `datadog.orchestratorExplorer.container_scrubbing.enabled`   | Enable the scrubbing of containers in the kubernetes resource YAML for sensitive information                                                                                                                                                                                                        | `true`                                            |
| `datadog.securityAgent.compliance.enabled`                   | Enable compliance checks in the security agent                                                                                                                                                                                                                                                      | `false`                                           |
| `datadog.securityAgent.compliance.checkInterval`             | Compliance check interval                                                                                                                                                                                                                                                                           | `20m`                                             |
| `datadog.securityAgent.compliance.configMap`                 | Optional ConfigMap with compliance checks                                                                                                                                                                                                                                                           | `""`                                              |
| `datadog.securityAgent.runtime.enabled`                      | Enable runtime security module in the security agent                                                                                                                                                                                                                                                | `false`                                           |
| `datadog.securityAgent.runtime.policies.configMap`           | Optional ConfigMap with runtime security policies                                                                                                                                                                                                                                                   | `""`                                              |
| `datadog.securityAgent.runtime.syscallMonitor.enabled`       | Enable syscall monitoring                                                                                                                                                                                                                                                                           | `false`                                           |
| `agents.podAnnotations`                                      | Annotations to add to the DaemonSet's Pods                                                                                                                                                                                                                                                          | `nil`                                             |
| `agents.podLabels`                                           | labels to add to each pod                                                                                                                                                                                                                                                                           | `nil`                                             |
| `agents.tolerations`                                         | List of node taints to tolerate (requires Kubernetes >= 1.6)                                                                                                                                                                                                                                        | `nil`                                             |
| `agents.nodeSelector`                                        | Node selectors                                                                                                                                                                                                                                                                                      | `nil`                                             |
| `agents.affinity`                                            | Node affinities                                                                                                                                                                                                                                                                                     | `nil`                                             |
| `agents.useHostNetwork`                                      | If true, use the host's network                                                                                                                                                                                                                                                                     | `nil`                                             |
| `agents.dnsConfig`                                           | If set, configure dnsConfig options in datadog agent containers                                                                                                                                                                                                                                     | `nil`                                             |
| `agents.containers.agent.env`                                | Additional list of environment variables to use in the agent container                                                                                                                                                                                                                              | `nil`                                             |
| `agents.containers.agent.logLevel`                           | Agent log verbosity                                                                                                                                                                                                                                                                                 | `INFO`                                            |
| `agents.containers.agent.resources.limits.cpu`               | CPU resource limits for the agent container                                                                                                                                                                                                                                                         | not set                                           |
| `agents.containers.agent.resources.requests.cpu`             | CPU resource requests for the agent container                                                                                                                                                                                                                                                       | not set                                           |
| `agents.containers.agent.resources.limits.memory`            | Memory resource limits for the agent container                                                                                                                                                                                                                                                      | not set                                           |
| `agents.containers.agent.resources.requests.memory`          | Memory resource requests for the agent container                                                                                                                                                                                                                                                    | not set                                           |
| `agents.containers.agent.livenessProbe`                      | Overrides the default liveness probe                                                                                                                                                                                                                                                                | http check on /live with port 5555                |
| `agents.containers.agent.readinessProbe`                     | Overrides the default readiness probe                                                                                                                                                                                                                                                               | http check on /ready with port 5555               |
| `agents.containers.processAgent.env`                         | Additional list of environment variables to use in the process-agent container                                                                                                                                                                                                                      | `nil`                                             |
| `agents.containers.processAgent.logLevel`                    | Process agent log verbosity                                                                                                                                                                                                                                                                         | `INFO`                                            |
| `agents.containers.processAgent.resources.limits.cpu`        | CPU resource limits for the process-agent container                                                                                                                                                                                                                                                 | `100m`                                            |
| `agents.containers.processAgent.resources.requests.cpu`      | CPU resource requests for the process-agent container                                                                                                                                                                                                                                               | `100m`                                            |
| `agents.containers.processAgent.resources.limits.memory`     | Memory resource limits for the process-agent container                                                                                                                                                                                                                                              | `200Mi`                                           |
| `agents.containers.processAgent.resources.requests.memory`   | Memory resource requests for the process-agent container                                                                                                                                                                                                                                            | `200Mi`                                           |
| `agents.containers.traceAgent.env`                           | Additional list of environment variables to use in the trace-agent container                                                                                                                                                                                                                        | `nil`                                             |
| `agents.containers.traceAgent.logLevel`                      | Trace agent log verbosity                                                                                                                                                                                                                                                                           | `INFO`                                            |
| `agents.containers.traceAgent.resources.limits.cpu`          | CPU resource limits for the trace-agent container                                                                                                                                                                                                                                                   | `100m`                                            |
| `agents.containers.traceAgent.resources.requests.cpu`        | CPU resource requests for the trace-agent container                                                                                                                                                                                                                                                 | `100m`                                            |
| `agents.containers.traceAgent.resources.limits.memory`       | Memory resource limits for the trace-agent container                                                                                                                                                                                                                                                | `200Mi`                                           |
| `agents.containers.traceAgent.resources.requests.memory`     | Memory resource requests for the trace-agent container                                                                                                                                                                                                                                              | `200Mi`                                           |
| `agents.containers.systemProbe.env`                          | Additional list of environment variables to use in the system-probe container                                                                                                                                                                                                                       | `nil`                                             |
| `agents.containers.systemProbe.logLevel`                     | System probe log verbosity                                                                                                                                                                                                                                                                          | `INFO`                                            |
| `agents.containers.systemProbe.resources.limits.cpu`         | CPU resource limits for the system-probe container                                                                                                                                                                                                                                                  | `100m`                                            |
| `agents.containers.systemProbe.resources.requests.cpu`       | CPU resource requests for the system-probe container                                                                                                                                                                                                                                                | `100m`                                            |
| `agents.containers.systemProbe.resources.limits.memory`      | Memory resource limits for the system-probe container                                                                                                                                                                                                                                               | `200Mi`                                           |
| `agents.containers.systemProbe.resources.requests.memory`    | Memory resource requests for the system-probe container                                                                                                                                                                                                                                             | `200Mi`                                           |
| `agents.containers.initContainers.resources.limits.cpu`      | CPU resource limits for the init containers container                                                                                                                                                                                                                                               | not set                                           |
| `agents.containers.initContainers.resources.requests.cpu`    | CPU resource requests for the init containers container                                                                                                                                                                                                                                             | not set                                           |
| `agents.containers.initContainers.resources.limits.memory`   | Memory resource limits for the init containers container                                                                                                                                                                                                                                            | not set                                           |
| `agents.containers.initContainers.resources.requests.memory` | Memory resource requests for the init containers container                                                                                                                                                                                                                                          | not set                                           |
| `agents.priorityClassName`                                   | Which Priority Class to associate with the daemonset                                                                                                                                                                                                                                                | `nil`                                             |
| `agents.useConfigMap`                                        | Configures a configmap to provide the agent configuration. Use this in combination with the `agent.customAgentConfig` parameter.                                                                                                                                                                    | `false`                                           |
| `agents.customAgentConfig`                                   | Specify custom contents for the datadog agent config (datadog.yaml). Note the `agents.useConfigMap` parameter needs to be set to `true` for this parameter to be taken into account.                                                                                                                | `{}`                                              |
| `agents.updateStrategy`                                      | Which update strategy to deploy the daemonset                                                                                                                                                                                                                                                       | RollingUpdate with 10% maxUnavailable             |
| `agents.volumes`                                             | Additional volumes for the daemonset or deployment                                                                                                                                                                                                                                                  | `nil`                                             |
| `agents.volumeMounts`                                        | Additional volumeMounts for the daemonset or deployment                                                                                                                                                                                                                                             | `nil`                                             |
| `agents.podSecurity.podSecurityPolicy.create`                | If true, create a PodSecurityPolicy resource for the Agent's Pods. Supported only for Linux agent's daemonset.                                                                                                                                                                                      | `False`                                           |
| `agents.podSecurity.securityContextConstraints.create`       | If true, create a SecurityContextConstraints resource for the Agent's Pods. Supported only for Linux agent's daemonset.                                                                                                                                                                             | `False`                                           |
| `agents.networkPolicy.create`                                | If true, create a NetworkPolicy for the agents                                                                                                                                                                                                                                                      | `false`                                           |
| `agents.podSecurity.seLinuxContext`                          | Allows you to overwrite the seLinux context in PSP/SCC                                                                                                                                                                                                                                              | default security context configuration            |
| `agents.podSecurity.privileged`                              | If true, allowed privileged containers                                                                                                                                                                                                                                                              | `False`                                           |
| `agents.podSecurity.capabilites`                             | list of allowed capabilities                                                                                                                                                                                                                                                                        | `[SYS_ADMIN, SYS_RESOURCE, SYS_ADMIN, IPC_LOCK]`  |
| `agents.podSecurity.volumes`                                 | list of allowed volumes types                                                                                                                                                                                                                                                                       | `[configMap,downwardAPI,emptyDir,ostPath,secret]` |
| `agents.podSecurity.seccompProfiles`                         | List of allowed seccomp profiles                                                                                                                                                                                                                                                                    | `["*"]`                                           |
| `agents.podSecurity.apparmorProfiles`                        | List of allowed apparmor profiles                                                                                                                                                                                                                                                                   | `["*"]`                                           |
| `agents.podSecurity.apparmor.enabled`                        | Whether to enable apparmor profile enforcement                                                                                                                                                                                                                                                      | `true`                                            |
| `datadog.leaderElection`                                     | Enable the leader Election feature                                                                                                                                                                                                                                                                  | `false`                                           |
| `datadog.leaderLeaseDuration`                                | The duration for which a leader stays elected.                                                                                                                                                                                                                                                      | 60 sec, 15 if Cluster Checks enabled              |
| `datadog.collectEvents`                                      | Enable Kubernetes event collection. Requires leader election.                                                                                                                                                                                                                                       | `false`                                           |
| `datadog.kubeStateMetricsEnabled`                            | If true, create kube-state-metrics                                                                                                                                                                                                                                                                  | `true`                                            |
| `clusterAgent.enabled`                                       | Use the cluster-agent for cluster metrics (Kubernetes 1.10+ only)                                                                                                                                                                                                                                   | `false`                                           |
| `clusterAgent.token`                                         | A cluster-internal secret for agent-to-agent communication. Must be 32+ characters a-zA-Z                                                                                                                                                                                                           | Generates a random value                          |
| `clusterAgent.tokenExistingSecret`                           | If set, use the secret with a provided name instead of creating a new one                                                                                                                                                                                                                           | `nil`                                             |
| `clusterAgent.image.repository`                              | The image repository for the cluster-agent                                                                                                                                                                                                                                                          | `datadog/cluster-agent`                           |
| `clusterAgent.image.tag`                                     | The image tag to pull                                                                                                                                                                                                                                                                               | `1.2.0`                                           |
| `clusterAgent.image.pullPolicy`                              | Image pull policy                                                                                                                                                                                                                                                                                   | `IfNotPresent`                                    |
| `clusterAgent.image.pullSecrets`                             | Image pull secrets                                                                                                                                                                                                                                                                                  | `nil`                                             |
| `clusterAgent.command`                                       | Override the default command to run in the container                                                                                                                                                                                                                                                | `nil`                                             |
| `clusterAgent.rbac.create`                                   | If true, create & use RBAC resources for cluster agent's pods                                                                                                                                                                                                                                       | `true`                                            |
| `clusterAgent.rbac.serviceAccount`                           | existing ServiceAccount to use (ignored if rbac.create=true) for cluster agent's pods                                                                                                                                                                                                               | `default`                                         |
| `clusterAgent.metricsProvider.enabled`                       | Enable Datadog metrics as a source for HPA scaling                                                                                                                                                                                                                                                  | `false`                                           |
| `clusterAgent.metricsProvider.service.type`                  | The type of service to use for the clusterAgent metrics server                                                                                                                                                                                                                                      | `ClusterIP`                                       |
| `clusterAgent.metricsProvider.service.port`                  | The port for service to use for the clusterAgent metrics server (Kubernetes >= 1.15)                                                                                                                                                                                                                | `8443`                                            |
| `clusterAgent.metricsProvider.wpaController`                 | Allows metricsProvider to work with WatermarkPodAutoscaler (requires WPA CRD)                                                                                                                                                                                                                       | `false`                                           |
| `clusterAgent.metricsProvider.useDatadogMetrics`             | Enable usage of DatadogMetric CRD to autoscale on arbitrary Datadog queries (requires DatadogMetric CRD)                                                                                                                                                                                            | `false`                                           |
| `clusterAgent.metricsProvider.createReaderRbac`              | Create `external-metrics-reader` RBAC automatically (to allow HPA to read data from Cluster Agent)                                                                                                                                                                                                  | `true`                                            |
| `clusterAgent.metricsProvider.aggregator`                    | Define the metric aggregator Cluster Agent will use to integrate with HPA                                                                                                                                                                                                                           | `avg`                                             |
| `clusterAgent.env`                                           | Additional Datadog environment variables for the cluster-agent                                                                                                                                                                                                                                      | `nil`                                             |
| `clusterAgent.confd`                                         | Additional check configurations (static and Autodiscovery)                                                                                                                                                                                                                                          | `nil`                                             |
| `clusterAgent.podAnnotations`                                | Annotations to add to the Cluster Agent Pod(s)                                                                                                                                                                                                                                                      | `nil`                                             |
| `clusterAgent.podLabels`                                     | Labels to add to the Cluster Agent Pod(s)                                                                                                                                                                                                                                                           | `nil`                                             |
| `clusterAgent.createPodDisruptionBudget`                     | Enable a pod disruption budget to apply to the Cluster Agent pods                                                                                                                                                                                                                                   | `false`                                           |
| `clusterAgent.priorityClassName`                             | Name of the priorityClass to apply to the Cluster Agent                                                                                                                                                                                                                                             | `nil`                                             |
| `clusterAgent.nodeSelector`                                  | Node selectors to apply to the Cluster Agent deployment                                                                                                                                                                                                                                             | `nil`                                             |
| `clusterAgent.affinity`                                      | Node affinities to apply to the Cluster Agent deployment                                                                                                                                                                                                                                            | `nil`                                             |
| `clusterAgent.resources.requests.cpu`                        | CPU resource requests                                                                                                                                                                                                                                                                               | not set                                           |
| `clusterAgent.resources.limits.cpu`                          | CPU resource limits                                                                                                                                                                                                                                                                                 | not set                                           |
| `clusterAgent.resources.requests.memory`                     | Memory resource requests                                                                                                                                                                                                                                                                            | not set                                           |
| `clusterAgent.resources.limits.memory`                       | Memory resource limits                                                                                                                                                                                                                                                                              | not set                                           |
| `clusterAgent.tolerations`                                   | List of node taints to tolerate                                                                                                                                                                                                                                                                     | `[]`                                              |
| `clusterAgent.healthPort`                                    | Overrides the default health port used by the liveness and readiness endpoint                                                                                                                                                                                                                       | `5555`                                            |
| `clusterAgent.livenessProbe`                                 | Overrides the default liveness probe                                                                                                                                                                                                                                                                | `http check on /live with port 5555`              |
| `clusterAgent.readinessProbe`                                | Overrides the default readiness probe                                                                                                                                                                                                                                                               | `http check on /ready with port 5555`             |
| `clusterAgent.strategy`                                      | Which update strategy to deploy the cluster-agent                                                                                                                                                                                                                                                   | RollingUpdate with 0 maxUnavailable, 1 maxSurge   |
| `clusterAgent.useHostNetwork`                                | If true, use the host's network                                                                                                                                                                                                                                                                     | `nil`                                             |
| `clusterAgent.dnsConfig`                                     | If set, configure dnsConfig options in datadog cluster agent containers                                                                                                                                                                                                                             | `nil`                                             |
| `clusterAgent.volumes`                                       | Additional volumes for the cluster-agent deployment                                                                                                                                                                                                                                                 | `nil`                                             |
| `clusterAgent.volumeMounts`                                  | Additional volumeMounts for the cluster-agent deployment                                                                                                                                                                                                                                            | `nil`                                             |
| `clusterAgent.networkPolicy.create`                          | If true, create a NetworkPolicy for the cluster agent                                                                                                                                                                                                                                               | `false`                                           |
| `clusterChecksRunner.enabled`                                | Enable Datadog agent deployment dedicated for running Cluster Checks. It allows having different resources (Request/Limit) for Cluster Checks agent pods.                                                                                                                                           | `false`                                           |
| `clusterChecksRunner.env`                                    | Additional Datadog environment variables for Cluster Checks Deployment                                                                                                                                                                                                                              | `nil`                                             |
| `clusterChecksRunner.createPodDisruptionBudget`              | Enable a pod disruption budget to apply to the Cluster Checks pods                                                                                                                                                                                                                                  | `false`                                           |
| `clusterChecksRunner.resources.requests.cpu`                 | CPU resource requests                                                                                                                                                                                                                                                                               | not set                                           |
| `clusterChecksRunner.resources.limits.cpu`                   | CPU resource limits                                                                                                                                                                                                                                                                                 | not set                                           |
| `clusterChecksRunner.resources.requests.memory`              | Memory resource requests                                                                                                                                                                                                                                                                            | not set                                           |
| `clusterChecksRunner.resources.limits.memory`                | Memory resource limits                                                                                                                                                                                                                                                                              | not set                                           |
| `clusterChecksRunner.nodeSelector`                           | Node selectors                                                                                                                                                                                                                                                                                      | `nil`                                             |
| `clusterChecksRunner.tolerations`                            | List of node taints to tolerate                                                                                                                                                                                                                                                                     | `nil`                                             |
| `clusterChecksRunner.affinity`                               | Node affinities                                                                                                                                                                                                                                                                                     | avoid running pods on the same node               |
| `clusterChecksRunner.livenessProbe`                          | Overrides the default liveness probe                                                                                                                                                                                                                                                                | http check on /live with port 5555                |
| `clusterChecksRunner.readinessProbe`                         | Overrides the default readiness probe                                                                                                                                                                                                                                                               | http check on /ready with port 5555               |
| `clusterChecksRunner.rbac.create`                            | If true, create & use RBAC resources for clusterchecks agent's pods                                                                                                                                                                                                                                 | `true`                                            |
| `clusterChecksRunner.rbac.dedicated`                         | If true, use dedicated RBAC resources for clusterchecks agent's pods                                                                                                                                                                                                                                | `false`                                           |
| `clusterChecksRunner.rbac.serviceAccount`                    | existing ServiceAccount to use (ignored if rbac.create=true) for clusterchecks agent's pods                                                                                                                                                                                                         | `default`                                         |
| `clusterChecksRunner.rbac.serviceAccountAnnotations`         | Annotations to add to the ServiceAccount if `clusterChecksRunner.rbac.dedicated` is `true`                                                                                                                                                                                                          | `{}`                                              |
| `clusterChecksRunner.strategy`                               | Which update strategy to deploy the Cluster Checks Deployment                                                                                                                                                                                                                                       | RollingUpdate with 0 maxUnavailable, 1 maxSurge   |
| `clusterChecksRunner.volumes`                                | Additional volumes for the Cluster Checks deployment                                                                                                                                                                                                                                                | `nil`                                             |
| `clusterChecksRunner.volumeMounts`                           | Additional volumeMounts for the Cluster Checks deployment                                                                                                                                                                                                                                           | `nil`                                             |
| `clusterChecksRunner.dnsConfig`                              | If set, configure dnsConfig options in datadog cluster agent clusterChecks containers                                                                                                                                                                                                               | `nil`                                             |
| `clusterChecksRunner.networkPolicy.create`                   | If true, create a NetworkPolicy for the cluster check runners                                                                                                                                                                                                                                       | `false`                                           |
| `kube-state-metrics.rbac.create`                             | If true, create & use RBAC resources for kube-state-metrics                                                                                                                                                                                                                                         | `true`                                            |
| `kube-state-metrics.serviceAccount.create`                   | If true, create & use serviceAccount                                                                                                                                                                                                                                                                | `true`                                            |
| `kube-state-metrics.serviceAccount.name`                     | If not set & create is true, use template fullname                                                                                                                                                                                                                                                  |                                                   |
| `kube-state-metrics.resources`                               | Overwrite the default kube-state-metrics container resources (Optional)                                                                                                                                                                                                                             |                                                   |
>>>>>>> b1cf2e1... add suport for env vars from references to configmaps and secrets

## Configuration options for Windows deployments

Some options above are not working/not available on Windows, here is the list of **unsupported** options:

| Parameter                                | Reason                                           |
|------------------------------------------|--------------------------------------------------|
| `datadog.dogstatsd.useHostPID`           | Host PID not supported by Windows Containers     |
| `datadog.dogstatsd.useSocketVolume`      | Unix sockets not supported on Windows            |
| `datadog.dogstatsd.socketPath`           | Unix sockets not supported on Windows            |
| `datadog.processAgent.processCollection` | Unable to access host/other containers processes |
| `datadog.systemProbe.enabled`            | System probe is not available for Windows        |
| `datadog.systemProbe.seccomp`            | System probe is not available for Windows        |
| `datadog.systemProbe.seccompRoot`        | System probe is not available for Windows        |
| `datadog.systemProbe.debugPort`          | System probe is not available for Windows        |
| `datadog.systemProbe.enableConntrack`    | System probe is not available for Windows        |
| `datadog.systemProbe.bpfDebug`           | System probe is not available for Windows        |
| `datadog.systemProbe.apparmor`           | System probe is not available for Windows        |
| `agents.useHostNetwork`                  | Host network not supported by Windows Containers |
