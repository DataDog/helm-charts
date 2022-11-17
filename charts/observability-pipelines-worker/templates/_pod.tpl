{{/*
Defines the PodSpec for Observability Pipelines Worker.
*/}}
{{- define "opw.pod" -}}
serviceAccountName: {{ include "opw.serviceAccountName" . }}
{{- with .Values.podHostNetwork }}
hostNetwork: {{ . }}
{{- end }}
{{- with .Values.podSecurityContext }}
securityContext:
{{ toYaml . | indent 2 }}
{{- end }}
{{- with .Values.podPriorityClassName }}
priorityClassName: {{ . }}
{{- end }}
{{- with .Values.dnsPolicy }}
dnsPolicy: {{ . }}
{{- end }}
{{- with .Values.dnsConfig }}
dnsConfig:
{{ toYaml . | indent 2 }}
{{- end }}
{{- with .Values.image.pullSecrets }}
imagePullSecrets:
{{ toYaml . | indent 2 }}
{{- end }}
{{- with .Values.initContainers }}
initContainers:
{{ toYaml . | indent 2 }}
{{- end }}
containers:
  - name: vector
{{- with .Values.securityContext }}
    securityContext:
{{ toYaml . | indent 6 }}
{{- end }}
{{- if .Values.image.sha }}
    image: "{{ .Values.image.repository }}:{{ .Values.image.tag | default .Chart.AppVersion }}@sha256:{{ .Values.image.sha }}"
{{- else }}
    image: "{{ .Values.image.repository }}:{{ .Values.image.tag | default .Chart.AppVersion }}"
{{- end }}
    imagePullPolicy: {{ .Values.image.pullPolicy }}
{{- with .Values.command }}
    command:
    {{- toYaml . | nindent 6 }}
{{- end }}
{{- with .Values.args }}
    args:
    {{- toYaml . | nindent 6 }}
{{- end }}
    env:
{{- if .Values.env }}
{{- with .Values.env }}
    {{- toYaml . | nindent 6 }}
{{- end }}
{{- end }}
{{- if .Values.envFrom }}
{{- with .Values.envFrom }}
    envFrom:
    {{- toYaml . | nindent 6 }}
{{- end }}
{{- end }}
    ports:
{{- if or .Values.containerPorts .Values.existingConfigMaps }}
    {{- toYaml .Values.containerPorts | nindent 6 }}
{{- else if .Values.customConfig }}
    {{- include "opw.containerPorts" . | indent 6 }}
{{- else }}
      - name: datadog-agent
        containerPort: 8282
        protocol: TCP
      - name: fluent
        containerPort: 24224
        protocol: TCP
      - name: logstash
        containerPort: 5044
        protocol: TCP
      - name: splunk-hec
        containerPort: 8080
        protocol: TCP
      - name: statsd
        containerPort: 8125
        protocol: TCP
      - name: syslog
        containerPort: 9000
        protocol: TCP
      - name: vector
        containerPort: 6000
        protocol: TCP
      - name: prom-exporter
        containerPort: 9090
        protocol: TCP
{{- end }}
{{- with .Values.livenessProbe }}
    livenessProbe:
      {{- toYaml . | trim | nindent 6 }}
{{- end }}
{{- with .Values.readinessProbe }}
    readinessProbe:
      {{- toYaml . | trim | nindent 6 }}
{{- end }}
{{- with .Values.resources }}
    resources:
{{- toYaml . | nindent 6 }}
{{- end }}
{{- with .Values.lifecycle }}
    lifecycle:
{{- toYaml . | nindent 6 }}
{{- end }}
    volumeMounts:
      - name: data
        {{- if .Values.existingConfigMaps }}
        mountPath: "{{ if .Values.dataDir }}{{ .Values.dataDir }}{{ else }}{{ fail "Specify `dataDir` if you're using `existingConfigMaps`" }}{{ end }}"
        {{- else }}
        mountPath: "{{ .Values.customConfig.data_dir | default "/vector-data-dir" }}"
        {{- end }}
      - name: config
        mountPath: "/etc/vector/"
        readOnly: true
{{- with .Values.extraVolumeMounts }}
{{- toYaml . | nindent 6 }}
{{- end }}
{{- with .Values.extraContainers }}
{{ toYaml . | indent 2 }}
{{- end }}
terminationGracePeriodSeconds: {{ .Values.terminationGracePeriodSeconds }}
{{- with .Values.nodeSelector }}
nodeSelector:
{{ toYaml . | indent 2 }}
{{- end }}
{{- with .Values.affinity }}
affinity:
{{ toYaml . | indent 2 }}
{{- end }}
{{- with .Values.tolerations }}
tolerations:
{{ toYaml . | indent 2 }}
{{- end }}
{{- with  .Values.topologySpreadConstraints }}
topologySpreadConstraints:
{{- toYaml . | nindent 2 }}
{{- end }}
volumes:
{{- if .Values.persistence.enabled }}
{{- with .Values.persistence.existingClaim }}
  - name: data
    persistentVolumeClaim:
      claimName: {{ . }}
{{- end }}
{{- else }}
  - name: data
    emptyDir: {}
{{- end }}
  - name: config
    projected:
      sources:
{{- if .Values.existingConfigMaps }}
  {{- range .Values.existingConfigMaps }}
        - configMap:
            name: {{ . }}
  {{- end }}
{{- else }}
        - configMap:
            name: {{ template "opw.fullname" . }}
{{- end }}
{{- with .Values.extraVolumes }}
{{- toYaml . | nindent 2 }}
{{- end }}
{{- end }}
