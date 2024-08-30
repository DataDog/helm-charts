{{- define "chart.deploymentName" }} "private-action-runner-{{.}}" {{ end }}
{{- define "chart.serviceAccountName" }} "private-action-runner-{{.}}-serviceaccount" {{ end }}
{{- define "chart.roleName" }} "private-action-runner-{{.}}-role" {{ end }}
{{- define "chart.roleBindingName" }} "private-action-runner-{{.}}-rolebinding" {{ end }}
{{- define "chart.serviceName" }} "private-action-runner-{{.}}-service" {{ end }}
{{- define "chart.secretName" }} "private-action-runner-{{.}}-secrets" {{ end }}

{{- define "chart.basicAuth" -}}
{{- if hasKey $.Values.connectionCredentials.basicAuth "credentials" }}
{{- range $c := $.Values.connectionCredentials.basicAuth.credentials }}
{{ $c.fileName }}: |
  {
     auth_type: 'Basic Auth',
     credentials: [
        {
           username: {{ $c.username }},
           password: {{ $c.password }},
        },
     ],
  }
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "chart.tokenAuth" -}}
{{- if hasKey $.Values.connectionCredentials.tokenAuth "credentials" }}
{{- range $c := $.Values.connectionCredentials.tokenAuth.credentials }}
{{ $c.fileName }}: |
  {
     auth_type: 'Token Auth',
     credentials: [
        {
           tokenName: {{ $c.tokenName }},
           tokenValue: {{ $c.tokenValue }},
        },
     ],
  }
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "chart.jenkinsAuth" -}}
{{- if hasKey $.Values.connectionCredentials.jenkinsAuth "credentials" }}
{{- range $c := $.Values.connectionCredentials.jenkinsAuth.credentials }}
{{ $c.fileName }}: |
  {
     auth_type: 'Token Auth',
     credentials: [
        {
           username: {{ $c.username }},
           token: {{ $c.token }},
           domain: {{ $c.domain }},
        },
     ],
  }
{{- end -}}
{{- end -}}
{{- end -}}


{{- define "chart.postgresAuth" -}}
{{- if hasKey $.Values.connectionCredentials.postgresAuth "credentials" }}
{{- range $c := $.Values.connectionCredentials.postgresAuth.credentials }}
{{ $c.fileName }}: |
  {
     auth_type: 'Token Auth',
     credentials: [
        {
           host: {{ $c.host }}
           port: {{ $c.port }}
           user: {{ $c.user }}
           password: {{ $c.password }}
           database: {{ $c.database }}
           sslMode: {{ $c.sslMode }}
        },
     ],
  }
{{- end -}}
{{- end -}}
{{- end -}}
