{{- define "chart.deploymentName" }} "private-action-runner-{{.}}" {{ end }}
{{- define "chart.serviceAccountName" }} "private-action-runner-{{.}}-serviceaccount" {{ end }}
{{- define "chart.roleName" }} "private-action-runner-{{.}}-role" {{ end }}
{{- define "chart.roleBindingName" }} "private-action-runner-{{.}}-rolebinding" {{ end }}
{{- define "chart.serviceName" }} "private-action-runner-{{.}}-service" {{ end }}
{{- define "chart.secretName" }} "private-action-runner-{{.}}-secrets" {{ end }}
