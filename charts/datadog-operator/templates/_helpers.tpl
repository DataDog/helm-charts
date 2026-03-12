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
Name label value for agent-install resources, truncated to 63 chars.
The base name is truncated to 49 chars first so the "-agent-install"
suffix (14 chars) is always preserved and never collides with the
unsuffixed name.
*/}}
{{- define "datadog-operator.agentInstallName" -}}
{{- printf "%s-agent-install" (include "datadog-operator.name" . | trunc 49 | trimSuffix "-") -}}
{{- end -}}

{{/*
Fully qualified name for agent-install resources, truncated to 63 chars.
The base fullname is truncated to 49 chars first so the "-agent-install"
suffix (14 chars) is always preserved and never collides with the
unsuffixed fullname.
*/}}
{{- define "datadog-operator.agentInstallFullname" -}}
{{- printf "%s-agent-install" (include "datadog-operator.fullname" . | trunc 49 | trimSuffix "-") -}}
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
Tries <releaseName>-endpoint-config by name first (standard installs), then falls
back to label-based discovery (aliased installs). If multiple ConfigMaps match,
picks the alphabetically first by name for deterministic behavior.
TODO: make the target aliased endpoint-config configMap user-configurable.
*/}}
{{- define "get-endpoint-config-data-key" -}}
{{- $ctx := index . 0 }}
{{- $key := index . 1 }}
{{- $ns := $ctx.Release.Namespace -}}
{{- $cm := lookup "v1" "ConfigMap" $ns (printf "%s-endpoint-config" $ctx.Release.Name) -}}
{{- if not $cm -}}
  {{- $matchingCMs := dict -}}
  {{- $matchingNames := list -}}
  {{- $allCMs := lookup "v1" "ConfigMap" $ns "" -}}
  {{- if $allCMs -}}
    {{- range $item := $allCMs.items -}}
      {{- $labels := default dict $item.metadata.labels -}}
      {{- if and (eq (default "" (get $labels "datadoghq.com/component")) "endpoint-config") (eq (default "" (get $labels "app.kubernetes.io/instance")) $ctx.Release.Name) -}}
        {{- $matchingNames = append $matchingNames $item.metadata.name -}}
        {{- $_ := set $matchingCMs $item.metadata.name $item -}}
      {{- end -}}
    {{- end -}}
  {{- end -}}
  {{- if $matchingNames -}}
    {{- $sorted := sortAlpha $matchingNames -}}
    {{- $winner := index $sorted 0 -}}
    {{- $cm = get $matchingCMs $winner -}}
  {{- end -}}
{{- end -}}
{{- if $cm -}}
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
Return the registry migration mode.
*/}}
{{- define "datadog-registry-mode" -}}
{{- $mode := .Values.registryMigrationMode -}}
{{- if and $mode (not (has $mode (list "auto" "all"))) -}}
  {{- fail (printf "registryMigrationMode must be \"auto\" or \"all\". Got: %q" $mode) -}}
{{- end -}}
{{- $mode -}}
{{- end -}}

{{/*
Return the namespace where the agent-install Job should create the DatadogAgent.
Always returns .Release.Namespace because chart-managed credential Secrets
(secret_api_key.yaml, secret_application_key.yaml) are created there and the
DatadogAgent credential schema references secrets by name only (no cross-namespace).
Fails at template time if the operator is configured to not watch the release
namespace for DatadogAgent resources, since the CR would never be reconciled.
*/}}
{{- define "datadog-operator.agentInstallNamespace" -}}
{{- if .Values.watchNamespacesAgent -}}
  {{- if not (or (has "" .Values.watchNamespacesAgent) (has .Release.Namespace .Values.watchNamespacesAgent)) -}}
    {{- fail (printf "installAgents is true but watchNamespacesAgent %v does not include the release namespace %q where credential Secrets are created. Add %q to watchNamespacesAgent or remove installAgents." .Values.watchNamespacesAgent .Release.Namespace .Release.Namespace) -}}
  {{- end -}}
{{- else if .Values.watchNamespaces -}}
  {{- if not (or (has "" .Values.watchNamespaces) (has .Release.Namespace .Values.watchNamespaces)) -}}
    {{- fail (printf "installAgents is true but watchNamespaces %v does not include the release namespace %q where credential Secrets are created. Add %q to watchNamespaces or remove installAgents." .Values.watchNamespaces .Release.Namespace .Release.Namespace) -}}
  {{- end -}}
{{- end -}}
{{- .Release.Namespace -}}
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
{{ "1.25.0-rc.1" }}
{{- end -}}
{{- end -}}
