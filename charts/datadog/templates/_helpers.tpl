{{/* vim: set filetype=mustache: */}}

{{- define "check-version" -}}
{{- if not .Values.agents.image.doNotCheckTag -}}
{{- $version := .Values.agents.image.tag | toString | trimSuffix "-jmx" -}}
{{- $length := len (split "." $version) -}}
{{- if and (eq $length 1) (eq $version "6") -}}
{{- $version = "6.19.0" -}}
{{- end -}}
{{- if and (eq $length 1) (eq $version "7") -}}
{{- $version = "7.19.0" -}}
{{- end -}}
{{- if and (eq $length 1) (eq $version "latest") -}}
{{- $version = "7.19.0" -}}
{{- end -}}
{{- if not (semverCompare "^6.19.0-0 || ^7.19.0-0" $version) -}}
{{- fail "This version of the chart requires an agent image 7.19.0 or greater. If you want to force and skip this check, use `--set agents.image.doNotCheckTag=true`" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "check-cluster-name" }}
{{- $length := len .Values.datadog.clusterName -}}
{{- if (gt $length 80)}}
{{- fail "Your `clusterName` isn’t valid it has to be below 81 chars." -}}
{{- end}}
{{- if not (regexMatch "^([a-z]([a-z0-9\\-]*[a-z0-9])?\\.)*([a-z]([a-z0-9\\-]*[a-z0-9])?)$" .Values.datadog.clusterName) -}}
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
{{- if semverCompare "^1.8-0" .Capabilities.KubeVersion.GitVersion -}}
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
\\.\pipe\docker_engine
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
Return agent host mount root
*/}}
{{- define "datadog.hostMountRoot" -}}
{{- if .Values.providers.gke.autopilot -}}
/var/autopilot/addon/datadog
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
Return a remote image path based on `.Values` (passed as root) and `.` (any `.image` from `.Values` passed as parameter)
*/}}
{{- define "image-path" -}}
{{- if .image.repository -}}
{{- .image.repository -}}:{{ .image.tag }}
{{- else -}}
{{ .root.registry }}/{{ .image.name }}:{{ .image.tag }}
{{- end -}}
{{- end -}}

{{/*
Return true if a system-probe feature is enabled.
*/}}
{{- define "system-probe-feature" -}}
{{- if or .Values.datadog.securityAgent.runtime.enabled .Values.datadog.networkMonitoring.enabled .Values.datadog.systemProbe.enableTCPQueueLength .Values.datadog.systemProbe.enableOOMKill -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}

{{/*
Return true if the system-probe container should be created.
*/}}
{{- define "should-enable-system-probe" -}}
{{- if and (not .Values.providers.gke.autopilot) (eq (include "system-probe-feature" .) "true") -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}


{{/*
Return true if a security-agent feature is enabled.
*/}}
{{- define "security-agent-feature" -}}
{{- if or .Values.datadog.securityAgent.compliance.enabled .Values.datadog.securityAgent.runtime.enabled  -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}

{{/*
Return true if the security-agent container should be created.
*/}}
{{- define "should-enable-security-agent" -}}
{{- if and (not .Values.providers.gke.autopilot) (eq (include "security-agent-feature" .) "true") -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}

{{/*
Return true if the compliance features should be enabled.
*/}}
{{- define "should-enable-compliance" -}}
{{- if and (not .Values.providers.gke.autopilot) .Values.datadog.securityAgent.compliance.enabled -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}

{{/*
Return true if the runtime security features should be enabled.
*/}}
{{- define "should-enable-runtime-security" -}}
{{- if and (not .Values.providers.gke.autopilot) .Values.datadog.securityAgent.runtime.enabled -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}

{{/*
Return true if Kubernetes resource monitoring (orchestrator explorer) should be enabled.
*/}}
{{- define "should-enable-k8s-resource-monitoring" -}}
{{- if and .Values.datadog.orchestratorExplorer.enabled .Values.clusterAgent.enabled -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}
