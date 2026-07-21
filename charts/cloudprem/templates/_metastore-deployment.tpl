{{/*
Shared Deployment template for the primary and read-only metastore components.

Expected context:
  root: the chart root context
  component: the Kubernetes component/name suffix
  selectorLabels: the named selector-label template
  values: the component-specific values
  service: the Quickwit service passed to `quickwit run`
*/}}
{{- define "quickwit.metastore.deployment" -}}
{{- $root := .root -}}
{{- $component := .component -}}
{{- $selectorLabels := .selectorLabels -}}
{{- $values := .values -}}
{{- $service := .service -}}
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "quickwit.fullname" $root }}-{{ $component }}
  labels:
    {{- include "quickwit.labels" $root | nindent 4 }}
    {{- if $root.Values.azure.clientId }}
    azure.workload.identity/use: "true"
    {{- end }}
  annotations:
    {{- with $root.Values.annotations }}
    {{- toYaml . | nindent 4 }}
    {{- end }}
    {{- with $values.annotations }}
    {{- toYaml . | nindent 4 }}
    {{- end }}
spec:
  replicas: {{ $values.replicaCount }}
  selector:
    matchLabels:
      {{- include $selectorLabels $root | nindent 6 }}
  strategy: {{- toYaml $values.strategy | nindent 4 }}
  template:
    metadata:
      annotations:
        checksum/config: {{ include (print $root.Template.BasePath "/configmap.yaml") $root | sha256sum }}
      {{- with $root.Values.podAnnotations }}
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with $values.podAnnotations }}
        {{- toYaml . | nindent 8 }}
      {{- end }}
      labels:
        {{- include "quickwit.additionalLabels" $root | nindent 8 }}
        {{- include $selectorLabels $root | nindent 8 }}
    spec:
      {{- with $root.Values.imagePullSecrets }}
      imagePullSecrets:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      serviceAccountName: {{ include "quickwit.serviceAccountName" $root }}
      securityContext:
        {{- toYaml $root.Values.podSecurityContext | nindent 8 }}
      {{- with $values.initContainers }}
      initContainers:
        {{- toYaml . | nindent 8 }}
      {{ end }}
      containers:
        - name: {{ $root.Chart.Name }}
          securityContext:
            {{- toYaml $root.Values.securityContext | nindent 12 }}
          image: "{{ $root.Values.image.repository }}:{{ $root.Values.image.tag | default $root.Chart.AppVersion }}"
          imagePullPolicy: {{ $root.Values.image.pullPolicy }}
          {{- if $root.Values.signals.metrics.enabled }}
          command: ["quickwit-metrics"]
          {{- end }}
          {{- if $values.args }}
          args: {{- toYaml $values.args | nindent 10 }}
          {{- else }}
          args: ["run", "--service", {{ $service | quote }}]
          {{- end }}
          env:
            {{- include "quickwit.environment" $root | nindent 12 }}
            {{- with (include "quickwit.extraEnv" $values.extraEnv) }}
            {{- . | nindent 12 }}
            {{- end }}
          {{- if or ($root.Values.environmentFrom) ($values.extraEnvFrom) }}
          envFrom:
          {{- with $root.Values.environmentFrom }}
            {{- toYaml . | nindent 12 }}
          {{- end }}
          {{- with $values.extraEnvFrom }}
            {{- toYaml . | nindent 12 }}
          {{- end }}
          {{- end }}
          ports:
            {{- include "quickwit.ports" $root | nindent 12 }}
          startupProbe:
            {{- toYaml $values.startupProbe | nindent 12 }}
          livenessProbe:
            {{- toYaml $values.livenessProbe | nindent 12 }}
          readinessProbe:
            {{- toYaml $values.readinessProbe | nindent 12 }}
          volumeMounts:
            - name: config
              mountPath: /quickwit/node.yaml
              subPath: node.yaml
            - name: data
              mountPath: /quickwit/qwdata
            {{- range $root.Values.configMaps }}
            - name: {{ .name }}
              mountPath: {{ .mountPath }}
            {{- end }}
            {{- with concat ($root.Values.volumeMounts | default list) ($values.extraVolumeMounts | default list) }}
              {{- toYaml . | nindent 12 }}
            {{- end }}
          resources:
            {{- toYaml $values.resources | nindent 12 }}
      volumes:
        - name: config
          configMap:
            name: {{ template "quickwit.fullname" $root }}
            items:
              - key: node.yaml
                path: node.yaml
        - name: data
          emptyDir: {}
        {{- range $root.Values.configMaps }}
        - name: {{ .name }}
          configMap:
            name: {{ .name }}
        {{- end }}
        {{- with concat ($root.Values.volumes | default list) ($values.extraVolumes | default list) }}
          {{- toYaml . | nindent 8 }}
        {{- end }}
      {{- with $values.nodeSelector }}
      nodeSelector:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with merge (dict) ($values.affinity | default dict) ($root.Values.affinity | default dict) }}
      affinity:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- $tolerations := concat ($root.Values.tolerations | default list) ($values.tolerations | default list) | compact | uniq }}
      tolerations:
        {{- toYaml $tolerations | nindent 8 }}
      {{- if $values.runtimeClassName }}
      runtimeClassName: {{ $values.runtimeClassName | quote }}
      {{- end }}
      {{- $tsc := concat ($root.Values.topologySpreadConstraints | default list) ($values.topologySpreadConstraints | default list) | compact }}
      {{- if $tsc }}
      topologySpreadConstraints:
        {{- toYaml $tsc | nindent 8 }}
      {{- end }}
{{- end }}
