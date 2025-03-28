{{/*
Expand the name of the chart.
*/}}
{{- define "chart.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "chart.fullname" -}}
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
{{- define "chart.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "chart.labels" -}}
helm.sh/chart: {{ include "chart.chart" . }}
{{ include "chart.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "chart.selectorLabels" -}}
app.kubernetes.io/name: {{ include "chart.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}


{{- define "chart.credentialFiles" -}}
{{- if hasKey $.Values "runner.credentialFiles" }}
{{- range $c := $.Values.runner.credentialFiles }}
{{ $c.fileName }}: |
{{ $c.data | indent 2 }}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Defines an RBAC rule for provided apiGroup, resource type and allowed verbs
*/}}
{{- define "rbacRule" }}
- apiGroups:
  - {{ .apiGroup }}
  resources:
  - {{ .resource }}
  verbs:
{{- range $_, $verb := (.verbs | uniq) }}
  - {{ $verb }}
{{- end }}
{{- end }}

{{/*
Map from plural(resourceName) to actionBundle
*/}}
{{- define "chart.k8sResourceMap" -}}
{{- $resourceMap := dict
    "customResourceDefinitions" "apiextensions"
    "controllerRevisions" "apps"
    "daemonSets" "apps"
    "deployments" "apps"
    "replicaSets" "apps"
    "statefulSets" "apps"
    "cronJobs" "batch"
    "jobs" "batch"
    "configMaps" "core"
    "endpoints" "core"
    "events" "core"
    "limitRanges" "core"
    "namespaces" "core"
    "nodes" "core"
    "persistentVolumes" "core"
    "persistentVolumeClaims" "core"
    "pods" "core"
    "podTemplates" "core"
    "replicationControllers" "core"
    "resourceQuotas" "core"
    "services" "core"
    "serviceAccounts" "core"
}}
{{- toYaml $resourceMap -}}
{{- end -}}

{{/*
Turns a plural(resourceName) into a singular(resourceName)
*/}}
{{- define "chart.k8sResourceSingular" -}}
{{- $resource := . -}}
{{- if eq $resource "endpoints" -}}
  {{- $resource -}}
{{- else -}}
  {{- printf "%s" (trimSuffix "s" $resource) -}}
{{- end -}}
{{- end -}}

{{/*
Returns the kubernetes apiGroup for the plural(resourceName)
*/}}
{{- define "chart.k8sApiGroup" -}}
{{- $bundle := . -}}
{{- if eq $bundle "apiextensions" -}}
apiextensions.k8s.io
{{- else if eq $bundle "core" -}}
""
{{- else -}}
  {{- $bundle -}}
{{- end -}}
{{- end -}}

{{/*
Transform a list of actions into the list of k8s verbs that are required to perform those actions
*/}}
{{- define "chart.k8sVerbs" -}}
{{- $actions := . -}}
{{- $allVerbs := list -}}
{{- range $action := $actions }}
  {{- if eq $action "deleteMultiple" -}}
    {{- $allVerbs = concat $allVerbs (list "delete" "list") -}}
  {{- else if eq $action "restart" -}}
    {{- $allVerbs = append $allVerbs "patch" -}}
  {{- else -}}
    {{- $allVerbs = append $allVerbs $action -}}
  {{- end -}}
{{- end -}}
{{- $allVerbs | toJson -}}
{{- end -}}
