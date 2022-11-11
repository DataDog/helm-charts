{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "haproxy.fullname" -}}
{{- printf "%s-haproxy" (include "vector.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "haproxy.labels" -}}
helm.sh/chart: {{ include "vector.chart" . }}
{{ include "haproxy.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Values.haproxy.image.tag | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "haproxy.selectorLabels" -}}
app.kubernetes.io/name: {{ include "vector.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: load-balancer
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "haproxy.serviceAccountName" -}}
{{- if .Values.haproxy.serviceAccount.create }}
{{- default (include "haproxy.fullname" .) .Values.haproxy.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.haproxy.serviceAccount.name }}
{{- end }}
{{- end }}

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
