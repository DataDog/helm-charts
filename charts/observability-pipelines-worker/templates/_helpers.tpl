{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "opw.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate strings at 63 characters because some Kubernetes name fields are limited to this (by the DNS naming spec).
If the release name contains a chart name it will be used as a full name.
*/}}
{{- define "opw.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "opw.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Return the API key Secret name to be used based on provided values.
*/}}
{{- define "opw.apiSecretName" -}}
{{- $fullName := printf "%s-apikey" (include "opw.fullname" .) -}}
{{- default $fullName .Values.datadog.apiKeyExistingSecret | quote -}}
{{- end -}}

{{/*
Return the configuration key Secret name to be used based on provided values.
*/}}
{{- define "opw.configKeySecretName" -}}
{{- $fullName := printf "%s-configkey" (include "opw.fullname" .) -}}
{{- default $fullName .Values.datadog.configKeyExistingSecret | quote -}}
{{- end -}}

{{/*
Common template labels.
*/}}
{{- define "opw.template-labels" -}}
app.kubernetes.io/name: {{ include "opw.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Common labels.
*/}}
{{- define "opw.labels" -}}
helm.sh/chart: {{ include "opw.chart" . }}
{{ include "opw.template-labels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Values.image.tag | quote }}
{{- end }}
{{- if .Values.commonLabels }}
{{ toYaml .Values.commonLabels }}
{{- end }}
{{- end -}}

{{/*
Return the ServiceAccount name
*/}}
{{- define "opw.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- template "opw.fullname" . }}
{{- else }}
{{- .Values.serviceAccount.name }}
{{- end -}}
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
Return the appropriate apiVersion for HPA autoscaling APIs.
*/}}
{{- define "autoscaling.apiVersion" -}}
{{- if or (.Capabilities.APIVersions.Has "autoscaling/v2/HorizontalPodAutoscaler") (semverCompare ">=1.23" .Capabilities.KubeVersion.Version) -}}
"autoscaling/v2"
{{- else -}}
"autoscaling/v2beta2"
{{- end -}}
{{- end -}}

{{/*
The helpers below are used to attempt to parse the configuration passed into the `config` option and construct
the Container and Service Ports without manual specification.

Being limited to just what is available to Go Templates and Sprig functions, the code is rather complex and hard to
follow. If the auto-generation of these is critical, it may suggest a need to prioritize an operator to handle this in a
more powerful language. Thankfully this behavior is non-critical as all Ports can be defined by hand, so issues with our
attempt to generate them can be side-stepped by users.
*/}}

{{/*
Generate an array of Service.Ports based on `.Values.pipelineConfig`.
*/}}
{{- define "opw.ports" -}}
  {{- range $componentKind, $components := .Values.pipelineConfig }}
    {{- if eq $componentKind "sources" }}
      {{- tuple $components "_helper.generatePort" | include "_helper.componentIter" }}
    {{- else if eq $componentKind "sinks" }}
      {{- tuple $components "_helper.generatePort" | include "_helper.componentIter" }}
    {{- else if eq $componentKind "api" }}
      {{- if $components.enabled }}
- name: api
  port: {{ mustRegexFind "[0-9]+$" (get $components "address") }}
  protocol: TCP
  targetPort: {{ mustRegexFind "[0-9]+$" (get $components "address") }}
      {{- end }}
    {{- end }}
  {{- end }}
{{- end }}

{{/*
Iterate over the components defined in `.Values.pipelineConfig`.
*/}}
{{- define "_helper.componentIter" -}}
{{- $components := index . 0 }}
{{- $helper := index . 1 }}
  {{- range $id, $options := $components }}
    {{- if (hasKey $options "address") }}
      {{- tuple $id $options | include $helper -}}
    {{- end }}
  {{- end }}
{{- end }}

{{/*
Generate a single Service.Port based on a component configuration.
*/}}
{{- define "_helper.generatePort" -}}
{{- $name := index . 0 | kebabcase -}}
{{- $config := index . 1 -}}
{{- $port := mustRegexFind "[0-9]+$" (get $config "address") -}}
{{- $protocol := default "TCP" (get $config "mode" | upper) }}
- name: {{ $name }}
  port: {{ $port }}
  protocol: {{ $protocol }}
  targetPort: {{ $port }}
{{- if not (mustHas $protocol (list "TCP" "UDP")) }}
{{ fail "Component's `mode` is not a supported protocol" }}
{{- end }}
{{- end }}

{{/*
Generate an array of Container.Ports based on `.Values.pipelineConfig`.
*/}}
{{- define "opw.containerPorts" -}}
  {{- range $componentKind, $components := .Values.pipelineConfig }}
    {{- if eq $componentKind "sources" }}
      {{- tuple $components "_helper.generateContainerPort" | include "_helper.componentIter" }}
    {{- else if eq $componentKind "sinks" }}
      {{- tuple $components "_helper.generateContainerPort" | include "_helper.componentIter" }}
    {{- else if eq $componentKind "api" }}
      {{- if $components.enabled }}
- name: api
  containerPort: {{ mustRegexFind "[0-9]+$" (get $components "address") }}
  protocol: TCP
      {{- end }}
    {{- end }}
  {{- end }}
{{- end }}

{{/*
Generate a single Container.Port based on a component configuration.
*/}}
{{- define "_helper.generateContainerPort" -}}
{{- $name := index . 0 | kebabcase -}}
{{- $config := index . 1 -}}
{{- $port := mustRegexFind "[0-9]+$" (get $config "address") -}}
{{- $protocol := default "TCP" (get $config "mode" | upper) }}
- name: {{ $name | trunc 15 | trimSuffix "-" }}
  containerPort: {{ $port }}
  protocol: {{ $protocol }}
{{- if not (mustHas $protocol (list "TCP" "UDP")) }}
{{ fail "Component's `mode` is not a supported protocol" }}
{{- end }}
{{- end }}
