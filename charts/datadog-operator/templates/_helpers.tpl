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
Uses label-based discovery to find the endpoint-config ConfigMap,
supporting both non-aliased and aliased datadog chart installations.

The lookup matches ConfigMaps by two labels:
  - datadoghq.com/component: endpoint-config — identifies the ConfigMap as an
    endpoint-config resource (set by the datadog chart's configmap template).
  - app.kubernetes.io/instance: <releaseName> — scopes the match to the current
    Helm release, preventing cross-release matches when multiple releases exist
    in the same namespace.

The endpoint-config ConfigMap also carries a datadoghq.com/chart-name label
(set to .Chart.Name by the datadog chart) for informational/debugging purposes,
but the operator does NOT filter on it. This is intentional: in wrapper chart
setups where ALL datadog dependencies are aliased, none have chart-name "datadog",
so filtering on it would cause the operator to find nothing.

If multiple aliased datadog instances exist in the same release (e.g. for
dual-shipping with different API keys), the operator deterministically picks
the alphabetically first matching ConfigMap (by name) using sortAlpha. This
ensures consistent behavior across renders and upgrades.
*/}}
{{- define "get-endpoint-config-data-key" -}}
{{- $ctx := index . 0 }}
{{- $key := index . 1 }}
{{- $ns := $ctx.Release.Namespace -}}
{{- $matchingCMs := dict -}}
{{- $matchingNames := list -}}
{{- $allCMs := lookup "v1" "ConfigMap" $ns "" -}}
{{- if $allCMs -}}
  {{- range $cm := $allCMs.items -}}
    {{- $labels := default dict $cm.metadata.labels -}}
    {{- if and (eq (default "" (get $labels "datadoghq.com/component")) "endpoint-config") (eq (default "" (get $labels "app.kubernetes.io/instance")) $ctx.Release.Name) -}}
      {{- $matchingNames = append $matchingNames $cm.metadata.name -}}
      {{- $_ := set $matchingCMs $cm.metadata.name $cm -}}
    {{- end -}}
  {{- end -}}
{{- end -}}
{{- if $matchingNames -}}
  {{- $sorted := sortAlpha $matchingNames -}}
  {{- $winner := index $sorted 0 -}}
  {{- $cm := get $matchingCMs $winner -}}
  {{- default "" (get $cm.data $key) -}}
{{- end -}}
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
*/}}
{{- define "should-set-dd-api-key" -}}
{{- if or .Values.apiKey .Values.apiKeyExistingSecret (eq (include "is-valid-endpoint-config-data" ( list . "api-key-secret-name")) "true") -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}

{{/*
Return true if DD_APP_KEY env var should be set.
*/}}
{{- define "should-set-dd-app-key" -}}
{{- if or .Values.appKey .Values.appKeyExistingSecret (eq (include "is-valid-endpoint-config-data" ( list . "app-key-secret-name")) "true") -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}

{{/*
Return apiKey secret name to be used based on provided values.
Priority for determining secret name:
1. .Values.apiKey
2. .Values.apiKeyExistingSecret
3. api-key-secret-name from endpoint-config configMap
*/}}
{{- define "datadog-operator.apiKeySecretName" -}}
{{- if and (eq (include "is-valid-endpoint-config-data" (list . "api-key-secret-name")) "true") (not .Values.apiKey) (not .Values.apiKeyExistingSecret) }}
{{- (include "get-endpoint-config-data-key" (list . "api-key-secret-name")) }}
{{- else }}
{{- $fullName := printf "%s-apikey" (include "datadog-operator.fullname" .) -}}
{{- default $fullName .Values.apiKeyExistingSecret -}}
{{- end -}}
{{- end -}}

{{/*
Return appKey secret name to be used based on provided values.
Priority for determining secret name:
1. .Values.appKey
2. .Values.appKeyExistingSecret
3. app-key-secret-name from endpoint-config configMap
*/}}
{{- define "datadog-operator.appKeySecretName" -}}
{{- if and (eq (include "is-valid-endpoint-config-data" (list . "app-key-secret-name")) "true") (not .Values.appKey) (not .Values.appKeyExistingSecret) }}
{{- (include "get-endpoint-config-data-key" (list . "app-key-secret-name")) }}
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
