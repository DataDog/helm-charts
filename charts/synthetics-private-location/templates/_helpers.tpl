{{/*
Expand the name of the chart.
*/}}
{{- define "synthetics-private-location.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "synthetics-private-location.fullname" -}}
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
{{- define "synthetics-private-location.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "synthetics-private-location.labels" -}}
helm.sh/chart: {{ include "synthetics-private-location.chart" . }}
{{ include "synthetics-private-location.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "synthetics-private-location.selectorLabels" -}}
app.kubernetes.io/name: {{ include "synthetics-private-location.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Values.commonLabels}}
{{ toYaml .Values.commonLabels }}
{{- end }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "synthetics-private-location.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "synthetics-private-location.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Return the appropriate apiVersion for PodDisruptionBudget policy APIs.
*/}}
{{- define "policy.poddisruptionbudget.apiVersion" -}}
{{- if or (.Capabilities.APIVersions.Has "policy/v1/PodDisruptionBudget") (semverCompare ">=1.21" .Capabilities.KubeVersion.Version) -}}
"policy/v1"
{{- else -}}
"policy/v1beta1"
{{- end -}}
{{- end -}}

{{/*
Normalize a bare IP to a single-address CIDR, as the worker does for allowedIPRanges / blockedIPRanges.
*/}}
{{- define "synthetics-private-location.cidr" -}}
{{- if contains "/" . }}{{ . }}{{ else if contains ":" . }}{{ . }}/128{{ else }}{{ . }}/32{{ end -}}
{{- end -}}

{{/*
Validate multi-container values. Fails on removed or missing settings, a token passed through env, or an unsupported image tag.
*/}}
{{- define "synthetics-private-location.multiContainer.validate" -}}
{{- $mc := .Values.multiContainer }}
{{- if hasKey $mc.networkPolicy "enabled" }}
{{- fail "multiContainer.networkPolicy.enabled has been replaced by mode: managed or external" }}
{{- end }}
{{- if not (has $mc.networkPolicy.mode (list "managed" "external")) }}
{{- fail "multiContainer.networkPolicy.mode must be explicitly set to managed or external" }}
{{- end }}
{{- $_ := required "multiContainer.browserPool.adminTokenSecret.name is required" $mc.browserPool.adminTokenSecret.name }}
{{- $_ := required "multiContainer.browserPool.adminTokenSecret.key is required" $mc.browserPool.adminTokenSecret.key }}
{{- range .Values.env }}
{{- if eq .name "BROWSER_POOL_ADMIN_TOKEN" }}
{{- fail "configure BROWSER_POOL_ADMIN_TOKEN through multiContainer.browserPool.adminTokenSecret, not env" }}
{{- end }}
{{- end }}
{{- $imageTag := include "synthetics-private-location.multiContainer.imageTag" . }}
{{- if not (regexMatch `^v?[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$` $imageTag) }}
{{- fail "multiContainer.enabled requires image.tag to be a semantic version >= 1.77.0 (prerelease and custom tags are not supported)" }}
{{- end }}
{{- if not (semverCompare ">=1.77.0" $imageTag) }}
{{- fail "multiContainer.enabled requires a private location image version >= 1.77.0 (prerelease tags are not supported)" }}
{{- end }}
{{- end -}}

{{/*
Image tag used by all multi-container containers.
*/}}
{{- define "synthetics-private-location.multiContainer.imageTag" -}}
{{- .Values.image.tag | default .Chart.AppVersion | toString }}
{{- end -}}

{{/*
Full image reference used by all multi-container containers.
*/}}
{{- define "synthetics-private-location.multiContainer.image" -}}
{{- printf "%s:%s" .Values.image.repository (include "synthetics-private-location.multiContainer.imageTag" .) }}
{{- end -}}

{{/*
Node preStop script for a daemon container. Waits for the worker to exit, unless the daemon itself is unhealthy (liveness restart).
Takes the JS health check of the daemon as argument.
*/}}
{{- define "synthetics-private-location.multiContainer.waitForWorkerExit" -}}
{{- printf "%s waitForWorkerExit(%s)" `const workerUp = () => fetch('http://127.0.0.1:8080/liveness', {signal: AbortSignal.timeout(2000)}).then(() => true, (error) => error?.cause?.code !== 'ECONNREFUSED'); const waitForWorkerExit = async (healthy) => { if (!(await healthy())) return; while (await workerUp()) await new Promise((resolve) => setTimeout(resolve, 2000)) };` . }}
{{- end -}}

{{/*
JS health check for a daemon exposing /health on the given local port.
*/}}
{{- define "synthetics-private-location.multiContainer.httpHealthy" -}}
{{- printf `() => fetch('http://127.0.0.1:%d/health', {signal: AbortSignal.timeout(2000)}).then((response) => response.ok, () => false)` . }}
{{- end -}}

{{/*
JS health check for the deno-executor socket.
*/}}
{{- define "synthetics-private-location.multiContainer.denoExecutorHealthy" -}}
async () => require('node:child_process').spawnSync('node', ['--no-warnings', '/home/dog/packages/runtime/deno-executor/dist/cli/healthcheck.js', '/run/deno-executor/deno-executor.sock']).status === 0
{{- end -}}

{{/*
Pod securityContext for the multi-container layout: hardened defaults merged with podSecurityContext.
Adds net.ipv4.ping_group_range for the worker group unless set, and fails on duplicate sysctl names.
*/}}
{{- define "synthetics-private-location.multiContainer.podSecurityContext" -}}
{{- $podSecurityContext := fromYaml `
fsGroup: 502
runAsUser: 501
runAsNonRoot: true
seccompProfile:
  type: RuntimeDefault
` }}
{{- $pingGid := 501 }}
{{- if hasKey .Values.securityContext "runAsGroup" }}{{ $pingGid = .Values.securityContext.runAsGroup | int }}{{ end }}
{{- $sysctls := list }}
{{- $sysctlNames := dict }}
{{- range (.Values.podSecurityContext.sysctls | default (list)) }}
{{- if hasKey $sysctlNames .name }}{{ fail (printf "podSecurityContext.sysctls contains duplicate name %s" .name) }}{{ end }}
{{- $_ := set $sysctlNames .name true }}
{{- $sysctls = append $sysctls (dict "name" .name "value" (toString .value)) }}
{{- end }}
{{- if not (hasKey $sysctlNames "net.ipv4.ping_group_range") }}
{{- $sysctls = append $sysctls (dict "name" "net.ipv4.ping_group_range" "value" (printf "%d %d" $pingGid $pingGid)) }}
{{- end }}
{{- $_ := set $podSecurityContext "sysctls" $sysctls }}
{{- toYaml (mustMergeOverwrite $podSecurityContext (omit (deepCopy .Values.podSecurityContext) "sysctls")) }}
{{- end -}}

{{/*
Worker container securityContext for the multi-container layout: hardened defaults merged with securityContext.
*/}}
{{- define "synthetics-private-location.multiContainer.workerSecurityContext" -}}
{{- $workerSecurityContext := fromYaml `
runAsUser: 501
runAsGroup: 501
runAsNonRoot: true
allowPrivilegeEscalation: false
readOnlyRootFilesystem: true
capabilities:
  drop:
  - ALL
seccompProfile:
  type: RuntimeDefault
` }}
{{- toYaml (mustMergeOverwrite $workerSecurityContext (deepCopy .Values.securityContext)) }}
{{- end -}}
