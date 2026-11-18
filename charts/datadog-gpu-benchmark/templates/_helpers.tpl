{{/*
The namespace Datadog always creates benchmark Jobs in. The chart is installed there.
*/}}
{{- define "gpu-benchmark.namespace" -}}datadog-gpu-benchmark{{- end -}}

{{- define "gpu-benchmark.labels" -}}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
app.kubernetes.io/name: {{ .Chart.Name }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Fails the render without the Cluster Agent's ServiceAccount. (The namespace is only warned about in NOTES.txt: CI renders and installs charts in other namespaces.)
*/}}
{{- define "gpu-benchmark.validate" -}}
{{- if not .Values.clusterAgent.serviceAccount -}}
{{- fail "clusterAgent.serviceAccount is required: the Cluster Agent's ServiceAccount, e.g. <release>-cluster-agent with the datadog chart, or <DatadogAgent name>-cluster-agent with the Datadog Operator" -}}
{{- end -}}
{{- if not .Values.clusterAgent.namespace -}}
{{- fail "clusterAgent.namespace is required: the namespace of the Cluster Agent's ServiceAccount" -}}
{{- end -}}
{{- end -}}
