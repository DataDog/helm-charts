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
        username: {{ $c.username | quote }},
        password: {{ $c.password | quote }}
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
        tokenName: {{ $c.tokenName | quote }},
        tokenValue: {{ $c.tokenValue | quote }}
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
        username: {{ $c.username | quote }},
        token: {{ $c.token | quote }},
        domain: {{ $c.domain | quote }}
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
        "tokenName": "host",
        "tokenValue": {{ $c.host | quote }}
      },
      {
        "tokenName": "port",
        "tokenValue": {{ $c.port | quote }}
      },
      {
        "tokenName": "user",
        "tokenValue": {{ $c.user | quote }}
      },
      {
        "tokenName": "password",
        "tokenValue": {{ $c.password | quote }}
      },
      {
        "tokenName": "database",
        "tokenValue": {{ $c.database | quote }}
      },
      {
        "tokenName": "sslmode",
        "tokenValue": {{ $c.sslMode | quote }}
      },
    {{- if $c.applicationName }}
      {
        "tokenName": "applicationName",
        "tokenValue": {{ $c.applicationName | quote }}
      },
    {{ end }}
    {{- if $c.searchPath }}
      {
      {
        "tokenName": "searchPath",
        "tokenValue": {{ $c.searchPath | quote }}
      }
    {{ end }}
    ],
  }
{{- end -}}
{{- end -}}
{{- end -}}
