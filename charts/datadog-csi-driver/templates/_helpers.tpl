{{/*
Expand the name of the chart.
*/}}
{{- define "datadog-csi-driver.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 32 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "datadog-csi-driver.fullname" -}}
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
Generate the DaemonSet name by appending "-node-server" to the name and truncating to 63 chars.
*/}}
{{- define "datadog-csi-driver.daemonsetName" -}}
{{- printf "%s-node-server" (include "datadog-csi-driver.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end }}
    

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "datadog-csi-driver.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "datadog-csi-driver.labels" -}}
helm.sh/chart: {{ include "datadog-csi-driver.chart" . }}
{{ include "datadog-csi-driver.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "datadog-csi-driver.selectorLabels" -}}
app.kubernetes.io/name: {{ include "datadog-csi-driver.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "datadog-csi-driver.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "datadog-csi-driver.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Check if target cluster supports GKE Autopilot WorkloadAllowlists.
GKE Autopilot WorkloadAllowlists are supported in GKE versions >= 1.32.1-gke.1729000.
*/}}
{{- define "csi.gke-autopilot-workloadallowlists-enabled" -}}
{{- if and (.Capabilities.APIVersions.Has "auto.gke.io/v1/AllowlistSynchronizer") (.Capabilities.APIVersions.Has "auto.gke.io/v1/WorkloadAllowlist") (semverCompare ">=v1.32.1-gke.1729000" .Capabilities.KubeVersion.Version) -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}
