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
Return the value for a given data key in the datadog endpoint-config ConfigMap.
Looks up the ConfigMap by exact name based on the release name to avoid
concatenating values from multiple ConfigMaps when multiple Datadog releases
exist in the same namespace.
*/}}
{{- define "get-endpoint-config-data-key" -}}
{{- $ctx := index . 0 }}
{{- $key := index . 1 }}
{{- $ns := $ctx.Release.Namespace -}}
{{- $cmName := printf "%s-endpoint-config" $ctx.Release.Name -}}
{{- $cm := lookup "v1" "ConfigMap" $ns $cmName -}}
{{- if $cm }}
  {{- get $cm.data $key -}}
{{- end }}
{{- end -}}

{{/*
Return true if value for a given key in the datadog endpoint-config ConfigMap is valid.
*/}}
{{- define "is-valid-endpoint-config-data" -}}
{{- $ctx := index . 0 }}
{{- $key := index . 1 }}
{{- $val := include "get-endpoint-config-data-key" (list $ctx $key) -}}
{{- if gt (len $val) 0 -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}

{{/*
Return true if DD_API_KEY env var should be set.
Checks:
1. operator.apiKey
2. operator.apiKeyExistingSecret
3. datadog.apiKey (parent chart)
4. datadog.apiKeyExistingSecret (parent chart)
5. endpoint-config ConfigMap
*/}}
{{- define "should-set-dd-api-key" -}}
{{- $parentApiKey := "" -}}
{{- $parentApiKeyExistingSecret := "" -}}
{{- if .Values.datadog -}}
  {{- $parentApiKey = .Values.datadog.apiKey -}}
  {{- $parentApiKeyExistingSecret = .Values.datadog.apiKeyExistingSecret -}}
{{- end -}}
{{- if or .Values.apiKey .Values.apiKeyExistingSecret $parentApiKey $parentApiKeyExistingSecret (eq (include "is-valid-endpoint-config-data" ( list . "api-key-secret-name")) "true") -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}

{{/*
Return true if DD_APP_KEY env var should be set.
Checks:
1. operator.appKey
2. operator.appKeyExistingSecret
3. datadog.appKey (parent chart)
4. datadog.appKeyExistingSecret (parent chart)
5. endpoint-config ConfigMap
*/}}
{{- define "should-set-dd-app-key" -}}
{{- $parentAppKey := "" -}}
{{- $parentAppKeyExistingSecret := "" -}}
{{- if .Values.datadog -}}
  {{- $parentAppKey = .Values.datadog.appKey -}}
  {{- $parentAppKeyExistingSecret = .Values.datadog.appKeyExistingSecret -}}
{{- end -}}
{{- if or .Values.appKey .Values.appKeyExistingSecret $parentAppKey $parentAppKeyExistingSecret (eq (include "is-valid-endpoint-config-data" ( list . "app-key-secret-name")) "true") -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}

{{/*
Return apiKey secret name to be used based on provided values.
Priority for determining secret name:
1. .Values.apiKey (operator's own value)
2. .Values.apiKeyExistingSecret (operator's own value)
3. .Values.datadog.apiKeyExistingSecret (parent chart value)
4. api-key-secret-name from endpoint-config configMap
*/}}
{{- define "datadog-operator.apiKeySecretName" -}}
{{- $parentApiKeyExistingSecret := "" -}}
{{- if .Values.datadog -}}
  {{- $parentApiKeyExistingSecret = .Values.datadog.apiKeyExistingSecret -}}
{{- end -}}
{{- if and (eq (include "is-valid-endpoint-config-data" (list . "api-key-secret-name")) "true") (not .Values.apiKey) (not .Values.apiKeyExistingSecret) (not $parentApiKeyExistingSecret) }}
{{- (include "get-endpoint-config-data-key" (list . "api-key-secret-name")) }}
{{- else if and (not .Values.apiKey) (not .Values.apiKeyExistingSecret) $parentApiKeyExistingSecret }}
{{- $parentApiKeyExistingSecret | quote -}}
{{- else }}
{{- $fullName := printf "%s-apikey" (include "datadog-operator.fullname" .) -}}
{{- default $fullName .Values.apiKeyExistingSecret -}}
{{- end -}}
{{- end -}}

{{/*
Return appKey secret name to be used based on provided values.
Priority for determining secret name:
1. .Values.appKey (operator's own value)
2. .Values.appKeyExistingSecret (operator's own value)
3. .Values.datadog.appKeyExistingSecret (parent chart value)
4. app-key-secret-name from endpoint-config configMap
*/}}
{{- define "datadog-operator.appKeySecretName" -}}
{{- $parentAppKeyExistingSecret := "" -}}
{{- if .Values.datadog -}}
  {{- $parentAppKeyExistingSecret = .Values.datadog.appKeyExistingSecret -}}
{{- end -}}
{{- if and (eq (include "is-valid-endpoint-config-data" (list . "app-key-secret-name")) "true") (not .Values.appKey) (not .Values.appKeyExistingSecret) (not $parentAppKeyExistingSecret) }}
{{- (include "get-endpoint-config-data-key" (list . "app-key-secret-name")) }}
{{- else if and (not .Values.appKey) (not .Values.appKeyExistingSecret) $parentAppKeyExistingSecret }}
{{- $parentAppKeyExistingSecret | quote -}}
{{- else }}
{{- $fullName := printf "%s-appkey" (include "datadog-operator.fullname" .) -}}
{{- default $fullName .Values.appKeyExistingSecret -}}
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
{{ "1.23.0-rc.2" }}
{{- end -}}
{{- end -}}
