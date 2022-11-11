{{/*
Expand the name of the chart.
*/}}
{{- define "vector.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "vector.fullname" -}}
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
{{- define "vector.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels.
*/}}
{{- define "vector.labels" -}}
helm.sh/chart: {{ include "vector.chart" . }}
{{ include "vector.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Values.image.tag | default .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{ with .Values.commonLabels }}
{{- toYaml . -}}
{{- end }}
{{- end }}

{{/*
Selector labels.
*/}}
{{- define "vector.selectorLabels" -}}
app.kubernetes.io/name: {{ include "vector.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if or (ne .Values.role "Agent") (ne .Values.role "Aggregator") (ne .Values.role "Stateless-Aggregator") }}
app.kubernetes.io/component: {{ .Values.role }}
{{- end }}
{{- end }}

{{/*
Create the name of the service account to use.
*/}}
{{- define "vector.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "vector.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

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
Generate an array of ServicePorts based on `.Values.customConfig`.
*/}}
{{- define "vector.ports" -}}
  {{- range $componentKind, $components := .Values.customConfig }}
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
Iterate over the components defined in `.Values.customConfig`.
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
Generate a single ServicePort based on a component configuration.
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
{{ fail "Component's `mode` is not a supported protocol, please raise a issue at https://github.com/vectordotdev/vector" }}
{{- end }}
{{- end }}

{{/*
Generate an array of ContainerPorts based on `.Values.customConfig`.
*/}}
{{- define "vector.containerPorts" -}}
  {{- range $componentKind, $components := .Values.customConfig }}
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
Generate a single ContainerPort based on a component configuration.
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
{{ fail "Component's `mode` is not a supported protocol, please raise a issue at https://github.com/vectordotdev/vector" }}
{{- end }}
{{- end }}
