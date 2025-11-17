{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "datadog-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "datadog-operator.fullname" -}}
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
{{- define "datadog-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "datadog-operator.labels" -}}
app.kubernetes.io/name: {{ include "datadog-operator.name" . }}
helm.sh/chart: {{ include "datadog-operator.chart" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "datadog-operator.serviceAccountName" -}}
{{ default (include "datadog-operator.fullname" .) .Values.serviceAccount.name }}
{{- end -}}

{{/*
Return secret name to be used based on provided values.
Priority:
1. Local apiKeyExistingSecret (if set)
2. Local apiKey (creates own secret)
3. Fallback to parent's default secret name
Note: If parent uses apiKeyExistingSecret, user must also set it in subchart values
*/}}
{{- define "datadog-operator.apiKeySecretName" -}}
{{- if .Values.apiKeyExistingSecret -}}
{{- .Values.apiKeyExistingSecret | quote -}}
{{- else if .Values.apiKey -}}
{{- printf "%s-apikey" (include "datadog-operator.fullname" .) | quote -}}
{{- else -}}
{{- include "datadog-operator.parentApiSecretName" . | quote -}}
{{- end -}}
{{- end -}}

{{/*
Return secret name to be used based on provided values.
Priority:
1. Local appKeyExistingSecret (if set)
2. Local appKey (creates own secret)
3. Fallback to parent's default secret name
Note: If parent uses appKeyExistingSecret, user must also set it in subchart values
*/}}
{{- define "datadog-operator.appKeySecretName" -}}
{{- if .Values.appKeyExistingSecret -}}
{{- .Values.appKeyExistingSecret | quote -}}
{{- else if .Values.appKey -}}
{{- printf "%s-appkey" (include "datadog-operator.fullname" .) | quote -}}
{{- else -}}
{{- include "datadog-operator.parentAppKeySecretName" . | quote -}}
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
Check operator image tag version.
*/}}
{{- define "check-image-tag" -}}
{{- if not .Values.image.doNotCheckTag -}}
{{- $tag := .Values.image.tag -}}
{{- $parts := split "@" $tag -}}
{{- index $parts "_0"}}
{{- else -}}
{{ "1.20.0-rc.4" }}
{{- end -}}
{{- end -}}

{{/*
Return the parent datadog chart's fullname (when this is installed as a subchart)
*/}}
{{- define "datadog-operator.parentFullname" -}}
{{- .Release.Name }}-datadog
{{- end -}}

{{/*
Return the parent datadog chart's API secret name
This mimics the parent chart's datadog.apiSecretName template
*/}}
{{- define "datadog-operator.parentApiSecretName" -}}
{{- include "datadog-operator.parentFullname" . -}}
{{- end -}}

{{/*
Return the parent datadog chart's APP key secret name
This mimics the parent chart's datadog.appKeySecretName template
*/}}
{{- define "datadog-operator.parentAppKeySecretName" -}}
{{- printf "%s-appkey" (include "datadog-operator.parentFullname" .) -}}
{{- end -}}
