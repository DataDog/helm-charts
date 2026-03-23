{{/*
Expand the name of the chart.
*/}}
{{- define "quickwit.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "quickwit.fullname" -}}
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
{{- define "quickwit.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Custom labels
*/}}
{{- define "quickwit.additionalLabels" -}}
{{- if .Values.additionalLabels }}
{{ toYaml .Values.additionalLabels }}
{{- end }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "quickwit.labels" -}}
helm.sh/chart: {{ include "quickwit.chart" . }}
{{ include "quickwit.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- include "quickwit.additionalLabels" . }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "quickwit.selectorLabels" -}}
app.kubernetes.io/name: {{ include "quickwit.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Searcher Selector labels
*/}}
{{- define "quickwit.searcher.selectorLabels" -}}
{{ include "quickwit.selectorLabels" . }}
app.kubernetes.io/component: searcher
{{- end }}

{{/*
Janitor Selector labels
*/}}
{{- define "quickwit.janitor.selectorLabels" -}}
{{ include "quickwit.selectorLabels" . }}
app.kubernetes.io/component: janitor
{{- end }}

{{/*
Metastore Selector labels
*/}}
{{- define "quickwit.metastore.selectorLabels" -}}
{{ include "quickwit.selectorLabels" . }}
app.kubernetes.io/component: metastore
{{- end }}

{{/*
Control Plane Selector labels
*/}}
{{- define "quickwit.control_plane.selectorLabels" -}}
{{ include "quickwit.selectorLabels" . }}
app.kubernetes.io/component: control-plane
{{- end }}

{{/*
Indexer Selector labels
*/}}
{{- define "quickwit.indexer.selectorLabels" -}}
{{ include "quickwit.selectorLabels" . }}
app.kubernetes.io/component: indexer
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "quickwit.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "quickwit.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Quickwit ports
*/}}
{{- define "quickwit.ports" -}}
- name: rest
  containerPort: 7280
  protocol: TCP
- name: grpc
  containerPort: 7281
  protocol: TCP
- name: discovery
  containerPort: 7282
  protocol: UDP
- name: cloudprem
  containerPort: 7283
  protocol:  TCP
{{- end }}


{{/*
Quickwit environment
*/}}
{{- define "quickwit.environment" -}}
- name: KUBERNETES_NAMESPACE
  valueFrom:
    fieldRef:
      fieldPath: metadata.namespace
- name: KUBERNETES_COMPONENT
  valueFrom:
    fieldRef:
      fieldPath: metadata.labels['app.kubernetes.io/component']
- name: KUBERNETES_POD_NAME
  valueFrom:
    fieldRef:
      fieldPath: metadata.name
- name: KUBERNETES_POD_IP
  valueFrom:
    fieldRef:
      fieldPath: metadata.name
- name: KUBERNETES_LIMITS_CPU
  valueFrom:
    resourceFieldRef:
      containerName: {{ .Chart.Name }}
      resource: limits.cpu
- name: KUBERNETES_LIMITS_MEMORY
  valueFrom:
    resourceFieldRef:
      containerName: {{ .Chart.Name }}
      resource: limits.memory
- name: KUBERNETES_REQUESTS_CPU
  valueFrom:
    resourceFieldRef:
      containerName: {{ .Chart.Name }}
      resource: requests.cpu
- name: KUBERNETES_REQUESTS_MEMORY
  valueFrom:
    resourceFieldRef:
      containerName: {{ .Chart.Name }}
      resource: requests.memory
- name: QW_CONFIG
  value: {{ .Values.configLocation }}
{{- if not .Values.config.cluster_id }}
- name: QW_CLUSTER_ID
  value: {{ .Release.Namespace }}-{{ include "quickwit.fullname" . }}
{{- end }}
- name: QW_NODE_ID
  value: "$(KUBERNETES_POD_NAME)"
{{ if semverCompare ">=1.33.0" .Capabilities.KubeVersion.Version }}
- name: QW_AVAILABILITY_ZONE
  valueFrom:
    fieldRef:
      fieldPath: metadata.labels['topology.kubernetes.io/zone']
{{- end }}
- name: QW_PEER_SEEDS
  value: {{ include "quickwit.fullname" . }}-headless
- name: QW_ADVERTISE_ADDRESS
  value: "$(KUBERNETES_POD_IP)"
- name: QW_CLUSTER_ENDPOINT
  value: http://{{ include "quickwit.fullname" $ }}-metastore.{{ $.Release.Namespace }}.svc.{{ .Values.clusterDomain }}:7280
{{- if .Values.azure.tenantId }}
- name: AZURE_TENANT_ID
  value: {{ .Values.azure.tenantId | quote }}
{{- end }}
{{- if .Values.azure.clientId }}
- name: AZURE_CLIENT_ID
  value: {{ .Values.azure.clientId | quote }}
{{- end }}
{{- if .Values.azure.clientSecretRef }}
- name: AZURE_CLIENT_SECRET
  valueFrom:
    secretKeyRef:
      name: {{ .Values.azure.clientSecretRef.name }}
      key: {{ .Values.azure.clientSecretRef.key }}
{{- end }}
{{- if .Values.azure.storageAccount.name }}
- name: QW_AZURE_STORAGE_ACCOUNT
  value: {{ .Values.azure.storageAccount.name | quote }}
{{- end }}
{{- if .Values.azure.storageAccount.accessKeySecretRef }}
- name: QW_AZURE_STORAGE_ACCESS_KEY
  valueFrom:
    secretKeyRef:
      name: {{ .Values.azure.storageAccount.accessKeySecretRef.name }}
      key: {{ .Values.azure.storageAccount.accessKeySecretRef.key }}
{{- end}}
- name: CP_DOGSTATSD_SERVER_HOST
{{- if .Values.dogstatsdServer.host.value }}
  value: {{ .Values.dogstatsdServer.host.value | quote }}
{{- else if .Values.dogstatsdServer.host.valueFrom }}
  valueFrom:
      {{- toYaml .Values.dogstatsdServer.host.valueFrom | nindent 4 }}
{{- end }}
- name: CP_DOGSTATSD_SERVER_PORT
  value: {{ .Values.dogstatsdServer.port | quote }}
- name: CP_ENABLE_REVERSE_CONNECTION
  value: {{ .Values.cloudprem.reverseConnection.enabled | quote }}
- name: CP_MIN_SHARDS
  value: {{ .Values.cloudprem.index.minShards | quote }}
- name: CP_RETENTION_PERIOD
  value: {{ .Values.cloudprem.index.retention | quote }}
- name: DD_SITE
  value: {{ .Values.datadog.site | quote }}
{{- if or .Values.datadog.apiKey .Values.datadog.apiKeyExistingSecret }}
- name: DD_API_KEY
  valueFrom:
    secretKeyRef:
      {{- if .Values.datadog.apiKeyExistingSecret }}
      name: {{ .Values.datadog.apiKeyExistingSecret }}
      {{- else }}
      name: {{ include "quickwit.fullname" . }}-api-key-secret
      {{- end }}
      key: api-key
{{- end }}
{{- if .Values.tracingEnabled }}
- name: QW_ENABLE_OPENTELEMETRY_OTLP_EXPORTER
  value: "true"
- name: OTEL_EXPORTER_OTLP_ENDPOINT
  value: http://{{ include "quickwit.fullname" $ }}-indexer:7281
- name: OTEL_EXPORTER_OTLP_PROTOCOL
  value: "grpc"
- name: OTEL_EXPORTER_OTLP_TIMEOUT
  value: "10"
- name: IMAGE_NAME
  value: {{ .Values.image.repository }}
- name: IMAGE_TAG
  value: {{ .Values.image.tag }}
{{- end }}
{{- with (include "quickwit.environmentDefaults" .Values.environment) }}
{{ . }}
{{- end }}
{{- end }}

{{/*
Merge default environment variables (NO_COLOR, QW_DISABLE_TELEMETRY, QW_LOG_FORMAT) with
user-provided values. Supports both legacy map and list formats. User-provided values
take precedence over defaults.
Defaults are stored as a list (not a dict) to guarantee deterministic rendering order
and avoid spurious rollouts from manifest drift.
*/}}
{{- define "quickwit.environmentDefaults" -}}
{{- $defaults := list (dict "name" "NO_COLOR" "value" "true") (dict "name" "QW_DISABLE_TELEMETRY" "value" "true") (dict "name" "QW_LOG_FORMAT" "value" "json") -}}
{{- $envs := list -}}
{{- $keys := list -}}
{{- if kindIs "map" . -}}
{{- range $key, $value := . -}}
{{- $envs = append $envs (dict "name" $key "value" ($value | toString)) -}}
{{- $keys = append $keys $key -}}
{{- end -}}
{{- else -}}
{{- range . -}}
{{- $envs = append $envs . -}}
{{- $keys = append $keys .name -}}
{{- end -}}
{{- end -}}
{{- range $defaults -}}
{{- if not (has .name $keys) -}}
{{- $envs = append $envs . -}}
{{- end -}}
{{- end -}}
{{- with $envs -}}
{{- toYaml . -}}
{{- end -}}
{{- end }}

{{/*
Render extra environment variables supporting both map and list formats.
Map format (legacy): { KEY: VALUE }
List format (recommended): [{ name: KEY, value: VALUE, valueFrom: ... }]
*/}}
{{- define "quickwit.extraEnv" -}}
{{- if kindIs "map" . -}}
{{- $envList := list -}}
{{- range $key, $value := . -}}
{{- $envList = append $envList (dict "name" $key "value" ($value | toString)) -}}
{{- end -}}
{{- if $envList -}}
{{- toYaml $envList -}}
{{- end -}}
{{- else -}}
{{- with . -}}
{{- toYaml . -}}
{{- end -}}
{{- end -}}
{{- end }}
