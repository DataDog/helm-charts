{{/*
Defines the PodSpec for Observability Pipelines Worker.
*/}}
{{- define "opw.pod" -}}
serviceAccountName: {{ include "opw.serviceAccountName" . }}
{{- if .Values.podHostNetwork }}
hostNetwork: {{ .Values.podHostNetwork }}
{{- end }}
{{- if .Values.podSecurityContext }}
securityContext: {{ toYaml .Values.podSecurityContext | nindent 2 }}
{{- end }}
{{- if .Values.podPriorityClassName }}
priorityClassName: {{ .Values.podPriorityClassName }}
{{- end }}
{{- if .Values.dnsPolicy }}
dnsPolicy: {{ .Values.dnsPolicy }}
{{- end }}
{{- if .Values.dnsConfig }}
dnsConfig: {{ toYaml .Values.dnsConfig | nindent 2 }}
{{- end }}
{{- if .Values.image.pullSecrets }}
imagePullSecrets: {{ toYaml .Values.image.pullSecrets | nindent 2 }}
{{- end }}
{{- if .Values.initContainers }}
initContainers: {{ toYaml .Values.initContainers | nindent 2 }}
{{- end }}
containers:
  - name: worker
{{- if .Values.securityContext }}
    securityContext: {{ toYaml .Values.securityContext | nindent 6 }}
{{- end }}
{{- if .Values.image.digest }}
    image: "{{ .Values.image.repository }}/{{ .Values.image.name }}@{{ .Values.image.digest }}"
{{- else }}
    image: "{{ .Values.image.repository }}/{{ .Values.image.name }}:{{ .Values.image.tag | default .Chart.AppVersion }}"
{{- end }}
    imagePullPolicy: {{ .Values.image.pullPolicy }}
{{- if .Values.command }}
    command: {{ toYaml .Values.command | nindent 6 }}
{{- end }}
{{- if .Values.args }}
    args: {{ toYaml .Values.args | nindent 6 }}
{{- end }}
    env:
      - name: DD_API_KEY
        valueFrom:
          secretKeyRef:
            name: {{ template "opw.apiSecretName" . }}
            key: api-key
      - name: DD_OP_PIPELINE_ID
      {{- if or .Values.datadog.configKey .Values.datadog.configKeyExistingSecret }}
        valueFrom:
          secretKeyRef:
            name: {{ template "opw.configKeySecretName" . }}
            key: config-key
      {{- else }}
        value: {{ .Values.datadog.pipelineId | quote }}
      {{- end }}
      {{- with .Values.datadog.site }}
      - name: DD_SITE
        value: {{ . | quote }}
      {{- end }}
      {{- with .Values.datadog.dataDir }}
      - name: DD_OP_DATA_DIR
        value: {{ . | quote }}
      {{- end }}
      - name: DD_OP_REMOTE_CONFIGURATION_ENABLED
        value: {{ .Values.datadog.remoteConfigurationEnabled | quote }}
{{- if .Values.env }}
{{ toYaml .Values.env | indent 6 }}
{{- end }}
{{- if .Values.envFrom }}
    envFrom: {{ toYaml .Values.envFrom | nindent 6 }}
{{- end }}
    ports:
{{- if .Values.containerPorts }}
{{ toYaml .Values.containerPorts | indent 6 }}
{{- else if .Values.pipelineConfig }}
{{- include "opw.containerPorts" . | indent 6 }}
{{- end }}
{{- if .Values.livenessProbe }}
    livenessProbe: {{ toYaml .Values.livenessProbe | trim | nindent 6 }}
{{- end }}
{{- if .Values.readinessProbe }}
    readinessProbe: {{ toYaml .Values.readinessProbe | trim | nindent 6 }}
{{- end }}
{{- if .Values.resources }}
    resources: {{ toYaml .Values.resources | nindent 6 }}
{{- end }}
{{- if .Values.lifecycle }}
    lifecycle: {{ toYaml .Values.lifecycle | nindent 6 }}
{{- end }}
    volumeMounts:
      - name: data
        mountPath: "{{ .Values.datadog.dataDir | default "/var/lib/observability-pipelines-worker" }}"
      {{- if not .Values.datadog.remoteConfigurationEnabled }}
      - name: config
        mountPath: "/etc/observability-pipelines-worker/"
        readOnly: true
      {{- end }}
{{- if .Values.extraVolumeMounts }}
{{ toYaml .Values.extraVolumeMounts | indent 6 }}
{{- end }}
{{- if .Values.extraContainers }}
{{ toYaml .Values.extraContainers | indent 2 }}
{{- end }}
terminationGracePeriodSeconds: {{ .Values.terminationGracePeriodSeconds }}
{{- if .Values.nodeSelector }}
nodeSelector: {{ toYaml .Values.nodeSelector | nindent 2 }}
{{- end }}
{{- if .Values.affinity }}
affinity: {{ toYaml .Values.affinity | nindent 2 }}
{{- end }}
{{- if .Values.tolerations }}
tolerations: {{ toYaml .Values.tolerations | nindent 2 }}
{{- end }}
{{- if  .Values.topologySpreadConstraints }}
topologySpreadConstraints: {{ toYaml .Values.topologySpreadConstraints | nindent 2 }}
{{- end }}
volumes:
{{- if .Values.persistence.enabled }}
{{- if .Values.persistence.existingClaim }}
  - name: data
    persistentVolumeClaim:
      claimName: {{ .Values.persistence.existingClaim }}
{{- end }}
{{- else }}
  - name: data
    emptyDir: {}
{{- end }}
{{- if not .Values.datadog.remoteConfigurationEnabled }}
  - name: config
    projected:
      sources:
        - configMap:
            name: {{ template "opw.fullname" . }}
{{- end }}
{{- if .Values.extraVolumes }}
{{ toYaml .Values.extraVolumes | indent 2 }}
{{- end }}
{{- end }}
