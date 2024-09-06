{{- define "chart.deploymentName" }} "private-action-runner-{{.}}" {{ end }}
{{- define "chart.serviceAccountName" }} "private-action-runner-{{.}}-serviceaccount" {{ end }}
{{- define "chart.roleName" }} "private-action-runner-{{.}}-role" {{ end }}
{{- define "chart.roleBindingName" }} "private-action-runner-{{.}}-rolebinding" {{ end }}
{{- define "chart.serviceName" }} "private-action-runner-{{.}}-service" {{ end }}
{{- define "chart.secretName" }} "private-action-runner-{{.}}-secrets" {{ end }}

{{/*
Defines an RBAC rule for provided apiGroup, resource type and allowed verbs
*/}}
{{- define "rbacRule" }}
- apiGroups:
  - {{ .apiGroup }}
  resources:
  - {{ .resource }}
  verbs:
{{- range $_, $verb := .verbs }}
  - {{ $verb }}
{{- end }}
{{- end }}

{{/*
Defines an RBAC "get" rule for provided apiGroup and resource type
*/}}
{{- define "rbacGetRule" }}
{{- include "rbacRule" (dict "apiGroup" .apiGroup "resource" .resource "verbs" (list "get"))}}
{{- end }}

{{/*
Defines an RBAC "list" rule for provided apiGroup and resource type
*/}}
{{- define "rbacListRule" }}
{{- include "rbacRule" (dict "apiGroup" .apiGroup "resource" .resource "verbs" (list "list"))}}
{{- end }}

{{/*
Defines an RBAC "update" rule for provided apiGroup and resource type
*/}}
{{- define "rbacUpdateRule" }}
{{- include "rbacRule" (dict "apiGroup" .apiGroup "resource" .resource "verbs" (list "update"))}}
{{- end }}

{{/*
Defines an RBAC "patch" rule for provided apiGroup and resource type
*/}}
{{- define "rbacPatchRule" }}
{{- include "rbacRule" (dict "apiGroup" .apiGroup "resource" .resource "verbs" (list "patch"))}}
{{- end }}
