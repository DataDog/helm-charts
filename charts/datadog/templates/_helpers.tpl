{{/* vim: set filetype=mustache: */}}

{{/*
  Returns node agent version based on image tag. This assumes `agents.image.doNotCheckTag` is false.
*/}}
{{- define "get-agent-version" -}}
{{- $version := .Values.agents.image.tag | toString | trimSuffix "-jmx" -}}
{{- $length := len (split "." $version) -}}
{{- if and (eq $length 1) (eq $version "6") -}}
{{- $version = "6.55.1" -}}
{{- end -}}
{{- if and (eq $length 1) (or (eq $version "7") (eq $version "latest")) -}}
{{- $version = "7.58.1" -}}
{{- end -}}
{{- $version -}}
{{- end -}}


{{- define "check-version" -}}
{{- if not .Values.agents.image.doNotCheckTag -}}
{{- $version := (include "get-agent-version" .) -}}
{{- if not (semverCompare "^6.36.0-0 || ^7.36.0-0" $version) -}}
{{- fail "This version of the chart requires an agent image 7.36.0 or greater. If you want to force and skip this check, use `--set agents.image.doNotCheckTag=true`" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "check-dca-version" -}}
{{- if not .Values.clusterAgent.image.doNotCheckTag -}}
{{- $version := .Values.clusterAgent.image.tag | toString -}}
{{- $length := len (split "." $version) -}}
{{- if and (eq $length 1) (eq $version "latest") -}}
{{- $version = "1.20.0" -}}
{{- end -}}
{{- if not (semverCompare ">=1.20.0-0" $version) -}}
{{- fail "This version of the chart requires a cluster agent image 1.20.0 or greater. If you want to force and skip this check, use `--set clusterAgent.image.doNotCheckTag=true`" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Check if target cluster is running OpenShift.
*/}}
{{- define "is-openshift" -}}
{{- if .Capabilities.APIVersions.Has "quota.openshift.io/v1/ClusterResourceQuota" -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}

{{- define "agent-has-env-ad" -}}
{{- if not .Values.agents.image.doNotCheckTag -}}
{{- $version := (include "get-agent-version" .) -}}
{{- if semverCompare "^6.27.0-0 || ^7.27.0-0" $version -}}
true
{{- else -}}
false
{{- end -}}
{{- else -}}
true
{{- end -}}
{{- end -}}

{{- define "check-cluster-name" }}
{{- $clusterName := tpl .Values.datadog.clusterName . -}}
{{- $length := len $clusterName -}}
{{- if (gt $length 80)}}
{{- fail "Your `clusterName` isn’t valid it has to be below 81 chars." -}}
{{- end}}
{{- if not (regexMatch "^([a-z]([a-z0-9\\-]*[a-z0-9])?\\.)*([a-z]([a-z0-9\\-]*[a-z0-9])?)$" $clusterName) -}}
{{- fail "Your `clusterName` isn’t valid. It must be dot-separated tokens where a token start with a lowercase letter followed by lowercase letters, numbers, or hyphens, can only end with a with [a-z0-9] and has to be below 80 chars." -}}
{{- end -}}
{{- end -}}

{{/*
Expand the name of the chart.
*/}}
{{- define "datadog.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
And depending on the resources the name is completed with an extension.
If release name contains chart name it will be used as a full name.
*/}}
{{- define "datadog.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "datadog.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Return true if the OTelAgent needs to be deployed
*/}}
{{- define "should-enable-otel-agent" -}}
{{- if and .Values.datadog.otelCollector.enabled  (not .Values.providers.gke.gdc) -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}



{{/*
Return secret name to be used based on provided values.
*/}}
{{- define "datadog.apiSecretName" -}}
{{- $fullName := include "datadog.fullname" . -}}
{{- default $fullName .Values.datadog.apiKeyExistingSecret | quote -}}
{{- end -}}

{{/*
Return secret name to be used based on provided values.
*/}}
{{- define "datadog.appKeySecretName" -}}
{{- $fullName := printf "%s-appkey" (include "datadog.fullname" .) -}}
{{- default $fullName .Values.datadog.appKeyExistingSecret | quote -}}
{{- end -}}

{{/*
Return secret name to be used based on provided values.
*/}}
{{- define "clusterAgent.tokenSecretName" -}}
{{- if not .Values.clusterAgent.tokenExistingSecret -}}
{{- include "datadog.fullname" . -}}-cluster-agent
{{- else -}}
{{- .Values.clusterAgent.tokenExistingSecret -}}
{{- end -}}
{{- end -}}

{{/*
Return the appropriate apiVersion for RBAC APIs.
*/}}
{{- define "rbac.apiVersion" -}}
{{- if .Capabilities.APIVersions.Has "rbac.authorization.k8s.io/v1" -}}
"rbac.authorization.k8s.io/v1"
{{- else -}}
"rbac.authorization.k8s.io/v1beta1"
{{- end -}}
{{- end -}}

{{/*
Return the appropriate os label
*/}}
{{- define "label.os" -}}
{{- if semverCompare "^1.14-0" .Capabilities.KubeVersion.GitVersion -}}
kubernetes.io/os
{{- else -}}
beta.kubernetes.io/os
{{- end -}}
{{- end -}}

{{/*
Correct `clusterAgent.metricsProvider.service.port` if Kubernetes <= 1.15
*/}}
{{- define "clusterAgent.metricsProvider.port" -}}
{{- if semverCompare "^1.15-0" .Capabilities.KubeVersion.GitVersion -}}
{{- .Values.clusterAgent.metricsProvider.service.port -}}
{{- else -}}
443
{{- end -}}
{{- end -}}

{{/*
Return the container runtime socket
*/}}
{{- define "datadog.dockerOrCriSocketPath" -}}
{{- if eq .Values.targetSystem "linux" -}}
{{- if .Values.providers.gke.autopilot -}}
/var/run/containerd/containerd.sock
{{- else -}}
{{- .Values.datadog.dockerSocketPath | default .Values.datadog.criSocketPath | default "/var/run/docker.sock" -}}
{{- end -}}
{{- end -}}
{{- if eq .Values.targetSystem "windows" -}}
{{- .Values.datadog.dockerSocketPath | default .Values.datadog.criSocketPath | default `\\.\pipe\docker_engine` -}}
{{- end -}}
{{- end -}}

{{/*
Return agent log directory path
*/}}
{{- define "datadog.logDirectoryPath" -}}
{{- if eq .Values.targetSystem "linux" -}}
/var/log/datadog
{{- end -}}
{{- if eq .Values.targetSystem "windows" -}}
C:/ProgramData/Datadog/logs
{{- end -}}
{{- end -}}

{{/*
Return agent config path
*/}}
{{- define "datadog.confPath" -}}
{{- if eq .Values.targetSystem "linux" -}}
/etc/datadog-agent
{{- end -}}
{{- if eq .Values.targetSystem "windows" -}}
C:/ProgramData/Datadog
{{- end -}}
{{- end -}}

{{/*
Return agent config path
*/}}
{{- define "datadog.otelconfPath" -}}
{{- if eq .Values.targetSystem "linux" -}}
/etc/otel-agent
{{- end -}}
{{- if eq .Values.targetSystem "windows" -}}
C:/ProgramData/Datadog
{{- end -}}
{{- end -}}

{{/*
Return agent host mount root
*/}}
{{- define "datadog.hostMountRoot" -}}
{{- if .Values.providers.gke.autopilot -}}
/var/autopilot/addon/datadog
{{- else if .Values.providers.gke.gdc -}}
/var/datadog
{{- else -}}
/var/lib/datadog-agent
{{- end -}}
{{- end -}}

{{/*
Return true if we are installing on a GKE cluster without RBAC setup (versions older than GKE R26)
*/}}
{{- define "is-gke-without-external-metrics" -}}
{{- if contains "-gke." .Capabilities.KubeVersion.GitVersion -}}
{{- if semverCompare ">=1.17.9-gke.600 || >=1.16.13-gke.1" .Capabilities.KubeVersion.GitVersion -}}
false
{{- else -}}
true
{{- end -}}
{{- else -}}
false
{{- end -}}
{{- end -}}

{{/*
Returns probe definition based on user settings and default HTTP port.
Accepts a map with `port` (default port), `path` (probe handler URI) and `settings` (probe settings).
*/}}
{{- define "probe.http" -}}
{{- if or .settings.httpGet .settings.tcpSocket .settings.exec -}}
{{ toYaml .settings }}
{{- else -}}
{{- $handler := dict "httpGet" (dict "port" .port "path" .path "scheme" "HTTP") -}}
{{ toYaml (merge $handler .settings) }}
{{- end -}}
{{- end -}}

{{/*
Returns probe definition based on user settings and default TCP socket port.
Accepts a map with `port` (default port) and `settings` (probe settings).
*/}}
{{- define "probe.tcp" -}}
{{- if or .settings.httpGet .settings.tcpSocket .settings.exec -}}
{{ toYaml .settings }}
{{- else -}}
{{- $handler := dict "tcpSocket" (dict "port" .port) -}}
{{- toYaml (merge $handler .settings) -}}
{{- end -}}
{{- end -}}

{{/*
Return the proper registry based on datadog.site (requires .Values to be passed as .)
*/}}
{{- define "registry" -}}
{{- if .registry -}}
{{- .registry -}}
{{- else if eq .datadog.site "datadoghq.eu" -}}
eu.gcr.io/datadoghq
{{- else if eq .datadog.site "ddog-gov.com" -}}
public.ecr.aws/datadog
{{- else if eq .datadog.site "ap1.datadoghq.com" -}}
asia.gcr.io/datadoghq
{{- else if eq .datadog.site "us3.datadoghq.com" -}}
datadoghq.azurecr.io
{{- else -}}
gcr.io/datadoghq
{{- end -}}
{{- end -}}

{{/*
Return a remote image path based on `.Values` (passed as root) and `.` (any `.image` from `.Values` passed as parameter)
*/}}
{{- define "image-path" -}}
{{- if .image.digest -}}
{{- if .image.repository -}}
{{- .image.repository -}}@{{ .image.digest }}
{{- else -}}
{{ include "registry" .root }}/{{ .image.name }}@{{ .image.digest }}
{{- end -}}
{{- else -}}
{{- $tagSuffix := "" -}}
{{- if .image.tagSuffix -}}
{{- $tagSuffix = printf "-%s" .image.tagSuffix -}}
{{- end -}}
{{- if .image.repository -}}
{{- .image.repository -}}:{{ .image.tag }}{{ $tagSuffix }}
{{- else -}}
{{ include "registry" .root }}/{{ .image.name }}:{{ .image.tag }}{{ $tagSuffix }}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Return true if a system-probe feature is enabled.
*/}}
{{- define "system-probe-feature" -}}
{{- if or .Values.datadog.securityAgent.runtime.enabled .Values.datadog.securityAgent.runtime.fimEnabled .Values.datadog.networkMonitoring.enabled .Values.datadog.systemProbe.enableTCPQueueLength .Values.datadog.systemProbe.enableOOMKill .Values.datadog.serviceMonitoring.enabled -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}

{{/*
Return true if the system-probe container should be created.
*/}}
{{- define "should-enable-system-probe" -}}
{{- if and (not (or .Values.providers.gke.autopilot .Values.providers.gke.gdc )) (eq (include "system-probe-feature" .) "true") (eq .Values.targetSystem "linux") -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}


{{/*
Return true if a security-agent feature is enabled.
*/}}
{{- define "security-agent-feature" -}}
{{- if or .Values.datadog.securityAgent.compliance.enabled .Values.datadog.securityAgent.runtime.enabled .Values.datadog.securityAgent.runtime.fimEnabled -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}

{{/*
Return true if the fips side car container should be created.
*/}}
{{- define "should-enable-fips" -}}
{{- if and (not (or .Values.providers.gke.autopilot .Values.providers.gke.gdc )) (eq .Values.targetSystem "linux") .Values.fips.enabled -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}

{{/*
Return true if the fips side car configMap should be mounted.
*/}}
{{- define "should-mount-fips-configmap" -}}
{{- if and (eq (include "should-enable-fips" .) "true") (not (empty .Values.fips.customFipsConfig)) -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}

{{/*
Return true if the security-agent container should be created.
*/}}
{{- define "should-enable-security-agent" -}}
{{- if and (not (or .Values.providers.gke.autopilot .Values.providers.gke.gdc )) (eq .Values.targetSystem "linux") (eq (include "security-agent-feature" .) "true") -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}

{{/*
Return true if the compliance features should be enabled.
*/}}
{{- define "should-enable-compliance" -}}
{{- if and (not (or .Values.providers.gke.autopilot .Values.providers.gke.gdc )) (eq .Values.targetSystem "linux") .Values.datadog.securityAgent.compliance.enabled -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}

{{/*
Return true if the runtime security features should be enabled.
*/}}
{{- define "should-enable-runtime-security" -}}
{{- if and (not (or .Values.providers.gke.autopilot .Values.providers.gke.gdc)) (or .Values.datadog.securityAgent.runtime.enabled .Values.datadog.securityAgent.runtime.fimEnabled) -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}

{{/*
Return true if the hostPid features should be enabled for the Agent pod.
*/}}
{{- define "should-enable-host-pid" -}}
{{- if eq .Values.targetSystem "windows" -}}
false
{{- else if and (not (or .Values.providers.gke.autopilot .Values.providers.gke.gdc)) (or (eq  (include "should-enable-compliance" .) "true") .Values.datadog.dogstatsd.useHostPID .Values.datadog.useHostPID) -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}

{{/*
Return true if .Values.existingClusterAgent is fully configured
*/}}
{{- define "existingClusterAgent-configured" -}}
{{- if and .Values.existingClusterAgent.join .Values.existingClusterAgent.serviceName .Values.existingClusterAgent.tokenSecretName -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}

{{/*
Return true if the ClusterAgent is enabled
*/}}
{{- define "cluster-agent-enabled" -}}
{{- if or (eq (include "existingClusterAgent-configured" .) "true") .Values.clusterAgent.enabled -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}


{{/*
Return true if the ClusterAgent needs to be deployed
*/}}
{{- define "should-deploy-cluster-agent" -}}
{{- if and .Values.clusterAgent.enabled (not .Values.existingClusterAgent.join) -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}


{{/*
Return true if a trace-agent needs to be deployed.
*/}}
{{- define "should-enable-trace-agent" -}}
{{- if or (eq  (include "trace-agent-use-tcp-port" .) "true") (eq  (include "trace-agent-use-uds" .) "true") (eq (include "trace-agent-use-local-service" .) "true") -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}

{{/*
Return true hostPath should be use for DSD socket. Return always false on GKE autopilot or GDC.
*/}}
{{- define "should-mount-hostPath-for-dsd-socket" -}}
{{- if or .Values.providers.gke.autopilot .Values.providers.gke.gdc (eq .Values.targetSystem "windows") -}}
false
{{- end -}}
{{- if .Values.datadog.dogstatsd.useSocketVolume -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}

{{/*
Return true if a APM over UDS is configured. Return always false on GKE Autopilot or Google Distributed Cloud.
*/}}
{{- define "trace-agent-use-uds" -}}
{{- if or .Values.providers.gke.autopilot .Values.providers.gke.gdc (eq .Values.targetSystem "windows") -}}
false
{{- end -}}
{{- if and (or .Values.datadog.apm.socketEnabled .Values.datadog.apm.useSocketVolume) (not .Values.providers.gke.gdc) -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}

{{/*
Return true if APM is configured to only use local service via the trace-agent's containerPort otherwise matches datadog.apm.portEnabled.
*/}}
{{- define "trace-agent-use-local-service" -}}
{{- default (include "trace-agent-use-host-port" .) .Values.datadog.apm.useLocalService -}}
{{- end -}}


{{/*
Return true if a host port is desired for APM.
*/}}
{{- define "trace-agent-use-host-port" -}}
{{- if or .Values.datadog.apm.portEnabled .Values.datadog.apm.enabled -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}

{{/*
Return true if a traffic over TCP is configured for APM.
*/}}
{{- define "trace-agent-use-tcp-port" -}}
{{- if or (eq  (include "trace-agent-use-host-port" .) "true") (eq  (include "trace-agent-use-local-service" .) "true") -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}

{{/*
Return true if Kubernetes resource monitoring (orchestrator explorer) should be enabled.
*/}}
{{- define "should-enable-k8s-resource-monitoring" -}}
{{- if and .Values.datadog.orchestratorExplorer.enabled (or .Values.clusterAgent.enabled (eq (include "existingClusterAgent-configured" .) "true")) -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}

{{/*
Return true if the Cluster Check Workers have to be deployed
*/}}
{{- define "should-enable-cluster-check-workers" -}}
{{- if or .Values.datadog.kubeStateMetricsCore.useClusterCheckRunners (and .Values.datadog.clusterChecks.enabled .Values.clusterChecksRunner.enabled) -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}

{{/*
Returns provider kind
*/}}
{{- define "provider-kind" -}}
{{- if .Values.providers.gke.autopilot -}}
gke-autopilot
{{- end -}}
{{- if .Values.providers.gke.gdc -}}
gke-gdc
{{- end -}}
{{- end -}}

{{/*
Return the service account name
*/}}
{{- define "agents.serviceAccountName" -}}
{{- if .Values.providers.gke.autopilot -}}
datadog-agent
{{- else if .Values.agents.rbac.create -}}
{{ template "datadog.fullname" . }}
{{- else -}}
{{ .Values.agents.rbac.serviceAccountName }}
{{- end -}}
{{- end -}}

{{- define "agents-useConfigMap-configmap-name" -}}
{{- if .Values.providers.gke.autopilot -}}
datadog-agent-datadog-yaml
{{- else -}}
{{ template "datadog.fullname" . }}-datadog-yaml
{{- end -}}
{{- end -}}

{{- define "agents-install-info-configmap-name" -}}
{{- if .Values.providers.gke.autopilot -}}
datadog-agent-installinfo
{{- else -}}
{{ template "datadog.fullname" . }}-installinfo
{{- end -}}
{{- end -}}

{{- define "agents.confd-configmap-name" -}}
{{- if .Values.providers.gke.autopilot -}}
datadog-agent-confd
{{- else -}}
{{ template "datadog.fullname" . }}-confd
{{- end -}}
{{- end -}}

{{- define "datadog-checksd-configmap-name" -}}
{{- if .Values.providers.gke.autopilot -}}
datadog-agent-checksd
{{- else -}}
{{ template "datadog.fullname" . }}-checksd
{{- end -}}
{{- end -}}

{{- define "fips-useConfigMap-configmap-name" -}}
{{- if .Values.providers.gke.autopilot -}}
datadog-agent-fips-config
{{- else -}}
{{ template "datadog.fullname" . }}-fips-config
{{- end -}}
{{- end -}}

{{- define "agents-install-otel-configmap-name" -}}
{{ template "datadog.fullname" . }}-otel-config
{{- end -}}

{{/*
Common template labels
*/}}
{{- define "datadog.template-labels" -}}
app.kubernetes.io/name: "{{ template "datadog.fullname" . }}"
app.kubernetes.io/instance: {{ .Release.Name | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "datadog.labels" -}}
helm.sh/chart: '{{ include "datadog.chart" . }}'
{{ include "datadog.template-labels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
{{- if .Values.commonLabels}}
{{ toYaml .Values.commonLabels }}
{{- end }}
{{- end -}}

{{/*
Returns provider-specific labels if any
*/}}
{{- define "provider-labels" -}}
{{- if include "provider-kind" . -}}
env.datadoghq.com/kind: {{ include "provider-kind" . }}
{{- end -}}
{{- end -}}

{{/*
Returns provider-specific env vars if any
*/}}
{{- define "provider-env" -}}
{{- if include "provider-kind" . -}}
- name: DD_PROVIDER_KIND
  value: {{ include "provider-kind" . }}
{{- end -}}
{{- end -}}

{{/*
Return Kubelet CA path inside Agent containers
*/}}
{{- define "datadog.kubelet.mountPath" -}}
{{- if .Values.datadog.kubelet.agentCAPath -}}
{{- .Values.datadog.kubelet.agentCAPath -}}
{{- else if .Values.datadog.kubelet.hostCAPath -}}
{{- if eq .Values.targetSystem "windows" -}}
C:/var/kubelet-ca/{{ base .Values.datadog.kubelet.hostCAPath }}
{{- else -}}
/var/run/kubelet-ca/{{ base .Values.datadog.kubelet.hostCAPath }}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Return Kubelet volumeMount
*/}}
{{- define "datadog.kubelet.volumeMount" -}}
- name: kubelet-ca
  {{- if eq .Values.targetSystem "linux" }}
  mountPath: {{ include "datadog.kubelet.mountPath" . }}
  {{- end }}
  {{- if eq .Values.targetSystem "windows" }}
  mountPath: {{ dir (include "datadog.kubelet.mountPath" .) }}
  {{- end }}
  readOnly: true
{{- end -}}

{{/*
Return true if the Cluster Agent needs a confd configmap
*/}}
{{- define "need-cluster-agent-confd" -}}
{{- if (or (.Values.clusterAgent.confd) (.Values.datadog.kubeStateMetricsCore.enabled) (.Values.clusterAgent.advancedConfd) (.Values.datadog.helmCheck.enabled)) -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}

{{/*
Return true if we can enable Service Internal Traffic Policy
*/}}
{{- define "enable-service-internal-traffic-policy" -}}
{{- if or (semverCompare "^1.22-0" .Capabilities.KubeVersion.GitVersion) .Values.agents.localService.forceLocalServiceEnabled -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}

{{/*
Return the local service name
*/}}
{{- define "localService.name" -}}
{{- if ne .Values.agents.localService.overrideName "" }}
{{- .Values.agents.localService.overrideName -}}
{{- else -}}
{{ template "datadog.fullname" . }}
{{- end -}}
{{- end -}}

{{/*
Return true if runtime compilation is enabled in the system-probe
*/}}
{{- define "runtime-compilation-enabled" -}}
{{- if or .Values.datadog.systemProbe.enableTCPQueueLength .Values.datadog.systemProbe.enableOOMKill .Values.datadog.serviceMonitoring.enabled -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}

{{/*
Return true if secret RBACs are needed for secret backend.
*/}}
{{- define "need-secret-permissions" -}}
{{- if .Values.datadog.secretBackend.command -}}
{{- if and .Values.datadog.secretBackend.enableGlobalPermissions (eq .Values.datadog.secretBackend.command "/readsecret_multiple_providers.sh") -}}
true
{{- end -}}
{{- else -}}
false
{{- end -}}
{{- end -}}

Returns env vars correctly quoted and valueFrom respected
*/}}
{{- define "additional-env-entries" -}}
{{- if . -}}
{{- range . }}
- name: {{ .name }}
{{- if .value }}
  value: {{ .value | quote }}
{{- else }}
  valueFrom:
{{ toYaml .valueFrom | indent 4 }}
{{- end }}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Returns env vars correctly quoted and valueFrom respected, defined in a dict
*/}}
{{- define "additional-env-dict-entries" -}}
{{- range $key, $value := . }}
- name: {{ $key }}
{{- if kindIs "map" $value }}
{{ toYaml $value | indent 2 }}
{{- else }}
  value: {{ $value | quote }}
{{- end }}
{{- end }}
{{- end -}}

{{/*
Return the appropriate apiVersion for PodDisruptionBudget policy APIs.
*/}}
{{- define "policy.poddisruptionbudget.apiVersion" -}}
{{- if or (.Capabilities.APIVersions.Has "policy/v1/PodDisruptionBudget") (semverCompare ">=1.21" .Capabilities.KubeVersion.Version) -}}
"policy/v1"
{{- else -}}
"policy/v1beta1"
{{- end -}}
{{- end -}}

{{/*
Returns securityContext depending of the OS
*/}}
{{- define "generate-security-context" -}}
{{- if .securityContext -}}
{{- if eq .targetSystem "windows" -}}
  {{- if .securityContext.windowsOptions }}
securityContext:
  windowsOptions:
    {{ toYaml .securityContext.windowsOptions }}
  {{- end -}}
{{- else }}
securityContext:
{{- if .sysAdmin }}
{{- $capabilities := dict "capabilities" (dict "add" (list "SYS_ADMIN")) }}
{{ toYaml (merge $capabilities .securityContext) | indent 2 }}
{{- else }}
{{ toYaml .securityContext | indent 2 }}
{{- end -}}
{{- if and .seccomp .kubeversion (semverCompare ">=1.19.0" .kubeversion) }}
  seccompProfile:
    {{- if hasPrefix "localhost/" .seccomp }}
    type: Localhost
    {{- else if eq "runtime/default" .seccomp }}
    type: RuntimeDefault
    {{- else }}
    type: Unconfined
    {{- end -}}
    {{- if hasPrefix "localhost/" .seccomp }}
    localhostProfile: {{ trimPrefix "localhost/" .seccomp }}
    {{- end }}
{{- end -}}
{{- end -}}
{{- else if .sysAdmin }}
securityContext:
{{ toYaml (dict "capabilities" (dict "add" (list "SYS_ADMIN"))) | indent 2 }}
{{- end -}}
{{- end -}}

{{/*
Verifies the OTLP/gRPC endpoint prefix.
gRPC supports several naming schemes: https://github.com/grpc/grpc/blob/master/doc/naming.md
The Datadog Agent Helm Chart currently only supports 'host:port' (usually '0.0.0.0:port').
*/}}
{{- define "verify-otlp-grpc-endpoint-prefix" -}}
{{- if hasPrefix "unix:" . }}
{{ fail "'unix' protocol is not currently supported on OTLP/gRPC endpoint" }}
{{- end }}
{{- if hasPrefix "unix-abstract:" . }}
{{ fail "'unix-abstract' protocol is not currently supported on OTLP/gRPC endpoint" }}
{{- end }}
{{- end -}}

{{/*
Verifies that an OTLP endpoint has a port explicitly set.
*/}}
{{- define "verify-otlp-endpoint-port" -}}
{{- if not ( regexMatch ":[0-9]+$" . ) }}
{{ fail "port must be set explicitly on OTLP endpoints" }}
{{- end }}
{{- end -}}

{{/*
Returns the flag used to specify the config file for the process-agent.
In 7.36, `--config` was deprecated and `--cfgpath` should be used instead.
*/}}
{{- define "process-agent-config-file-flag" -}}
{{- if  .Values.providers.gke.autopilot -}}
-config
{{- else if not .Values.agents.image.doNotCheckTag -}}
{{- $version := .Values.agents.image.tag | toString | trimSuffix "-jmx" -}}
{{- $length := len (split "." $version ) -}}
{{- if and (gt $length 1) (not (semverCompare "^6.36.0 || ^7.36.0" $version)) -}}
--config
{{- else -}}
--cfgpath
{{- end -}}
{{- else -}}
--config
{{- end -}}
{{- end -}}

{{/*
Returns whether or not the underlying OS is Google Container-Optimized-OS
Note: GKE Autopilot only use COS (see https://cloud.google.com/kubernetes-engine/docs/concepts/node-images)
*/}}
{{- define "can-mount-host-usr-src" -}}
{{- if or .Values.providers.gke.autopilot .Values.providers.gke.cos -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}

{{/*
Returns whether Remote Configuration should be enabled in the agent
*/}}
{{- define "datadog-remoteConfiguration-enabled" -}}
{{- if and (.Values.remoteConfiguration.enabled) (.Values.datadog.remoteConfiguration.enabled) (not .Values.providers.gke.gdc ) -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}

{{/*
Returns whether Remote Configuration should be enabled in the cluster agent
*/}}
{{- define "clusterAgent-remoteConfiguration-enabled" -}}
{{- if and .Values.remoteConfiguration.enabled (or .Values.clusterAgent.admissionController.remoteInstrumentation.enabled (((.Values.datadog.autoscaling).workload).enabled)) (not .Values.providers.gke.gdc ) -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}

{{/*
Create RBACs for custom resources
*/}}
{{- define "orchestratorExplorer-config-crs" -}}
{{- range $cr := .Values.datadog.orchestratorExplorer.customResources }}
- apiGroups:
  - {{ (splitList "/" $cr) | first | quote }}
  resources:
  - {{ (splitList "/" $cr) | last | quote }}
  verbs:
  - get
  - list
  - watch
{{- end }}
{{- end }}

{{/*
  Return true if Container Runtime Support is enabled
*/}}
{{- define "container-runtime-support-enabled" -}}
  {{- if and .Values.datadog.containerRuntimeSupport.enabled (not .Values.providers.gke.gdc) -}}
    true
  {{- else -}}
    false
  {{- end -}}
{{- end -}}

{{/*
  Return true if container image collection is enabled
*/}}
{{- define "should-enable-container-image-collection" -}}
  {{- if and (not (include "container-runtime-support-enabled" .)) (or .Values.datadog.containerImageCollection.enabled .Values.datadog.sbom.containerImage.enabled) -}}
    {{- fail "Container runtime support has to be enabled for container image collection to work. Please enable it using `datadog.containerRuntimeSupport.enabled`." -}}
  {{- end -}}
  {{- if or .Values.datadog.containerImageCollection.enabled .Values.datadog.sbom.containerImage.enabled -}}
    true
  {{- else -}}
    false
  {{- end -}}
{{- end -}}

{{/*
  Return true if SBOM collection for container image is enabled
*/}}
{{- define "should-enable-sbom-container-image-collection" -}}
  {{- if .Values.datadog.sbom.containerImage.enabled -}}
    {{- if not (eq (include "should-enable-container-image-collection" .) "true") -}}
      {{- fail "Container runtime support has to be enabled for SBOM collection to work. Please enable it using `datadog.containerRuntimeSupport.enabled`." -}}
    {{- end -}}
    true
  {{- else -}}
    false
  {{- end -}}
{{- end -}}

{{/*
  Return true if language detection feature is enabled
*/}}
{{- define "language-detection-enabled" -}}
  {{- if and .Values.datadog.apm.instrumentation.enabled .Values.datadog.apm.instrumentation.language_detection.enabled -}}
    true
  {{- else -}}
    false
  {{- end -}}
{{- end -}}

{{/*
  Return true if any process-related check is enabled
*/}}
{{- define "process-checks-enabled" -}}
  {{- if .Values.providers.gke.gdc }}
    false
  {{- end -}}
  {{- if or .Values.datadog.processAgent.containerCollection .Values.datadog.processAgent.processCollection .Values.datadog.processAgent.processDiscovery (eq (include "language-detection-enabled" .) "true") -}}
    true
  {{- else -}}
    false
  {{- end -}}
{{- end -}}

{{/*
  Return value of "DD_PROCESS_CONFIG_RUN_IN_CORE_AGENT_ENABLED" env var in core agent container.
*/}}
{{- define "get-process-checks-in-core-agent-envvar" -}}
  {{- range .Values.agents.containers.agent.env -}}
    {{- if eq .name "DD_PROCESS_CONFIG_RUN_IN_CORE_AGENT_ENABLED" -}}
      {{- .value -}}
    {{- end -}}
  {{- end -}}
{{- end -}}

{{/*
  Returns true if process-related checks should run on the core agent.
*/}}
{{- define "should-run-process-checks-on-core-agent" -}}
  {{- if .Values.providers.gke.gdc -}}
    false
  {{- end -}}
  {{- if ne .Values.targetSystem "linux" -}}
    false
  {{- else if (ne (include "get-process-checks-in-core-agent-envvar" .) "") -}}
    {{- include "get-process-checks-in-core-agent-envvar" . -}}
  {{- else if and (not .Values.agents.image.doNotCheckTag) .Values.datadog.processAgent.runInCoreAgent (semverCompare ">=7.57.0-0" (include "get-agent-version" .)) -}}
      true
  {{- else -}}
    false
  {{- end -}}
{{- end -}}

{{/*
  Returns true if the process-agent container should be created.
*/}}
{{- define "should-enable-process-agent" -}}
  {{- if .Values.providers.gke.gdc -}}
    false
  {{- end -}}
  {{- if or .Values.datadog.networkMonitoring.enabled .Values.datadog.serviceMonitoring.enabled -}}
    true
  {{- else if and (not .Values.agents.image.doNotCheckTag) (eq (include "should-enable-k8s-resource-monitoring" .) "true") (semverCompare "<=7.51.0-0" (include "get-agent-version" .)) -}}
    true
  {{- else if (eq (include "should-run-process-checks-on-core-agent" .) "true") -}}
    false
  {{- else -}}
    {{- include "process-checks-enabled" . -}}
  {{- end -}}
{{- end -}}


{{- define "get-port-number-from-name" -}}
{{- $portName := .portName -}}
{{- range .ports -}}
  {{- if eq .name $portName -}}
    {{ .containerPort }}
  {{- end -}}
{{- end -}}
{{- end -}}
