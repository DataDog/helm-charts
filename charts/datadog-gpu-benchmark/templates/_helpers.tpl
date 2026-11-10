{{/*
Names that Datadog's benchmark worker uses as constants: changing them breaks runs.
*/}}
{{- define "gpu-benchmark.runServiceAccount" -}}datadog-gpu-benchmark-run{{- end -}}
{{- define "gpu-benchmark.podOverridesConfigMap" -}}datadog-gpu-benchmark-pod{{- end -}}
{{- define "gpu-benchmark.managedBy" -}}gpu-benchmark-worker{{- end -}}

{{- define "gpu-benchmark.namespace" -}}
{{- .Values.namespace.name -}}
{{- end -}}

{{- define "gpu-benchmark.fullname" -}}
{{- printf "%s-%s" .Release.Name .Chart.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "gpu-benchmark.labels" -}}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
app.kubernetes.io/name: {{ .Chart.Name }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Selects benchmark pods: the label Datadog's worker sets on every run's pod.
*/}}
{{- define "gpu-benchmark.podSelector" -}}
app.kubernetes.io/managed-by: {{ include "gpu-benchmark.managedBy" . }}
{{- end -}}

{{/*
The pod overrides the worker reads (overrides.json). Only non-empty fields: the worker rejects
unknown fields, and an empty one overrides nothing.
*/}}
{{- define "gpu-benchmark.podOverrides" -}}
{{- $o := dict -}}
{{- with .Values.podOverrides -}}
{{- if .labels }}{{ $_ := set $o "labels" .labels }}{{ end -}}
{{- if .annotations }}{{ $_ := set $o "annotations" .annotations }}{{ end -}}
{{- if .nodeSelector }}{{ $_ := set $o "nodeSelector" .nodeSelector }}{{ end -}}
{{- if .tolerations }}{{ $_ := set $o "tolerations" .tolerations }}{{ end -}}
{{- if .affinity }}{{ $_ := set $o "affinity" .affinity }}{{ end -}}
{{- if .topologySpreadConstraints }}{{ $_ := set $o "topologySpreadConstraints" .topologySpreadConstraints }}{{ end -}}
{{- if .imageRepository }}{{ $_ := set $o "imageRepository" .imageRepository }}{{ end -}}
{{- if .imagePullSecrets }}{{ $_ := set $o "imagePullSecrets" .imagePullSecrets }}{{ end -}}
{{- end -}}
{{- if .Values.podOverrides.priorityClassName -}}
{{- $_ := set $o "priorityClassName" .Values.podOverrides.priorityClassName -}}
{{- else if .Values.priorityClass.create -}}
{{- $_ := set $o "priorityClassName" .Values.priorityClass.name -}}
{{- end -}}
{{- toPrettyJson $o -}}
{{- end -}}

{{/*
Fails the render when a required value is missing.
*/}}
{{- define "gpu-benchmark.validate" -}}
{{- if .Values.rbac.create -}}
{{- if not .Values.clusterAgent.serviceAccount.name -}}
{{- fail "clusterAgent.serviceAccount.name is required: the Cluster Agent's ServiceAccount, e.g. <release>-cluster-agent with the datadog chart, or <DatadogAgent name>-cluster-agent with the Datadog Operator" -}}
{{- end -}}
{{- if not .Values.clusterAgent.serviceAccount.namespace -}}
{{- fail "clusterAgent.serviceAccount.namespace is required: the namespace of the Cluster Agent's ServiceAccount" -}}
{{- end -}}
{{- end -}}
{{- if ne .Values.namespace.name "datadog-gpu-benchmark" -}}
{{- fail "namespace.name must be datadog-gpu-benchmark: Datadog creates the benchmark Jobs there" -}}
{{- end -}}
{{- end -}}

{{/*
ValidatingAdmissionPolicy is GA (admissionregistration.k8s.io/v1) from Kubernetes 1.30. Both
checks: helm template advertises Helm's built-in API list whatever --kube-version says.
*/}}
{{- define "gpu-benchmark.admissionPolicySupported" -}}
{{- and (.Capabilities.APIVersions.Has "admissionregistration.k8s.io/v1/ValidatingAdmissionPolicy") (semverCompare ">=1.30.0-0" .Capabilities.KubeVersion.Version) -}}
{{- end -}}
