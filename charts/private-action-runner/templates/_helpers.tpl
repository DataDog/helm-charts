{{- define "chart.deploymentName" }} "private-action-runner-{{.}}" {{ end }}
{{- define "chart.serviceAccountName" }} "private-action-runner-{{.}}-serviceaccount" {{ end }}
{{- define "chart.roleName" }} "private-action-runner-{{.}}-role" {{ end }}
{{- define "chart.roleBindingName" }} "private-action-runner-{{.}}-rolebinding" {{ end }}
{{- define "chart.serviceName" }} "private-action-runner-{{.}}-service" {{ end }}
{{- define "chart.secretName" }} "private-action-runner-{{.}}-secrets" {{ end }}

{{- define "chart.credentialFiles" -}}
{{- if hasKey $.Values "credentialFiles" }}
{{- range $c := $.Values.credentialFiles }}
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
