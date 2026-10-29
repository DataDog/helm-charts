{{/*
Normalize a bare IP to a single-address CIDR, as the worker does for allowedIPRanges / blockedIPRanges.
Takes an IPv4 or IPv6 address or CIDR. Fails on any other format, including IPv4-embedded IPv6.
*/}}
{{- define "synthetics-private-location.cidr" -}}
{{- $ipv4 := `^((25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])\.){3}(25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])(/([0-9]|[12][0-9]|3[0-2]))?$` }}
{{- $ipv6 := `^(([0-9A-Fa-f]{1,4}:){7}[0-9A-Fa-f]{1,4}|([0-9A-Fa-f]{1,4}:){1,7}:|([0-9A-Fa-f]{1,4}:){1,6}:[0-9A-Fa-f]{1,4}|([0-9A-Fa-f]{1,4}:){1,5}(:[0-9A-Fa-f]{1,4}){1,2}|([0-9A-Fa-f]{1,4}:){1,4}(:[0-9A-Fa-f]{1,4}){1,3}|([0-9A-Fa-f]{1,4}:){1,3}(:[0-9A-Fa-f]{1,4}){1,4}|([0-9A-Fa-f]{1,4}:){1,2}(:[0-9A-Fa-f]{1,4}){1,5}|[0-9A-Fa-f]{1,4}:(:[0-9A-Fa-f]{1,4}){1,6}|:((:[0-9A-Fa-f]{1,4}){1,7}|:))(/([0-9]|[1-9][0-9]|1[01][0-9]|12[0-8]))?$` }}
{{- if regexMatch $ipv4 . }}{{ if contains "/" . }}{{ . }}{{ else }}{{ . }}/32{{ end }}
{{- else if regexMatch $ipv6 . }}{{ if contains "/" . }}{{ . }}{{ else }}{{ . }}/128{{ end }}
{{- else }}{{ fail (printf "multiContainer.networkPolicy IP range %q must be an IPv4 or IPv6 address or CIDR" .) }}
{{- end -}}
{{- end -}}

{{/*
Validate multi-container values. Fails on missing settings, missing or unrestricted DNS peers when the NetworkPolicy is created, or a token passed through env.
*/}}
{{- define "synthetics-private-location.multiContainer.validate" -}}
{{- $mc := .Values.multiContainer }}
{{- if not (kindIs "bool" $mc.networkPolicy.create) }}
{{- fail "multiContainer.networkPolicy.create must be set to true or false" }}
{{- end }}
{{- if $mc.networkPolicy.create }}
{{- if not $mc.networkPolicy.dns.to }}
{{- fail "multiContainer.networkPolicy.dns.to must specify DNS resolver peers when multiContainer.networkPolicy.create is true" }}
{{- end }}
{{- if not $mc.networkPolicy.dns.ports }}
{{- fail "multiContainer.networkPolicy.dns.ports must specify at least one port when multiContainer.networkPolicy.create is true" }}
{{- end }}
{{- range $mc.networkPolicy.dns.to }}
{{- if not (or (hasKey . "ipBlock") (hasKey . "namespaceSelector") (hasKey . "podSelector")) }}
{{- fail "each multiContainer.networkPolicy.dns.to entry must specify an ipBlock, namespaceSelector or podSelector" }}
{{- end }}
{{- end }}
{{- end }}
{{- $_ := required "multiContainer.browserPool.adminTokenSecret.name is required" $mc.browserPool.adminTokenSecret.name }}
{{- $_ := required "multiContainer.browserPool.adminTokenSecret.key is required" $mc.browserPool.adminTokenSecret.key }}
{{- range .Values.env }}
{{- if eq .name "BROWSER_POOL_ADMIN_TOKEN" }}
{{- fail "configure BROWSER_POOL_ADMIN_TOKEN through multiContainer.browserPool.adminTokenSecret, not env" }}
{{- end }}
{{- end }}
{{- end -}}

{{/*
Node preStop script for a daemon container, from files/prestop/wait-for-worker-exit.js. Waits for the worker to exit, unless the daemon itself is unhealthy (liveness restart).
Takes a dict with "root" (the root context) and "healthCheck" (the daemon's JS health check).
*/}}
{{- define "synthetics-private-location.multiContainer.waitForWorkerExit" -}}
{{- $js := .root.Files.Get "files/prestop/wait-for-worker-exit.js" | trim }}
{{- printf "%s waitForWorkerExit(%s)" $js .healthCheck }}
{{- end -}}

{{/*
JS health check for a daemon exposing /health on a local port, from files/prestop/http-healthy.js.
Takes a dict with "root" (the root context) and "port" (the port number).
*/}}
{{- define "synthetics-private-location.multiContainer.httpHealthy" -}}
{{- $js := .root.Files.Get "files/prestop/http-healthy.js" | trim }}
{{- replace "__PORT__" (toString .port) $js }}
{{- end -}}

{{/*
JS health check for the deno-executor socket, from files/prestop/deno-executor-healthy.js.
*/}}
{{- define "synthetics-private-location.multiContainer.denoExecutorHealthy" -}}
{{- .Files.Get "files/prestop/deno-executor-healthy.js" | trim }}
{{- end -}}

{{/*
Pod securityContext for the multi-container layout: hardened defaults merged with podSecurityContext.
Adds net.ipv4.ping_group_range for the worker group unless set, and fails on duplicate sysctl names.
*/}}
{{- define "synthetics-private-location.multiContainer.podSecurityContext" -}}
{{- $podSecurityContext := fromYaml `
runAsNonRoot: true
seccompProfile:
  type: RuntimeDefault
` }}
{{- $_ := set $podSecurityContext "fsGroup" (include "synthetics-private-location.multiContainer.denoRunnerId" . | int) }}
{{- $_ := set $podSecurityContext "runAsUser" (include "synthetics-private-location.multiContainer.dogUser" . | int) }}
{{- $pingGid := include "synthetics-private-location.multiContainer.dogUser" . | int }}
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
runAsNonRoot: true
allowPrivilegeEscalation: false
readOnlyRootFilesystem: true
capabilities:
  drop:
  - ALL
seccompProfile:
  type: RuntimeDefault
` }}
{{- $_ := set $workerSecurityContext "runAsUser" (include "synthetics-private-location.multiContainer.dogUser" . | int) }}
{{- $_ := set $workerSecurityContext "runAsGroup" (include "synthetics-private-location.multiContainer.dogUser" . | int) }}
{{- toYaml (mustMergeOverwrite $workerSecurityContext (deepCopy .Values.securityContext)) }}
{{- end -}}

{{/*
Blocked CIDRs for the multi-container NetworkPolicy: the default ranges when enableDefaultBlockedIpRanges is true, then blockedIPRanges, each normalized with synthetics-private-location.cidr.
Returns a JSON object with "IPv4" and "IPv6" lists.
*/}}
{{- define "synthetics-private-location.multiContainer.networkPolicy.blockedCIDRs" -}}
{{- $np := .Values.multiContainer.networkPolicy }}
{{- $rawIPv4 := $np.blockedIPRanges.IPv4 | default (list) }}
{{- $rawIPv6 := $np.blockedIPRanges.IPv6 | default (list) }}
{{- if $np.enableDefaultBlockedIpRanges }}
{{- $rawIPv4 = concat ($np.defaultBlockedIPRanges.IPv4 | default (list)) $rawIPv4 }}
{{- $rawIPv6 = concat ($np.defaultBlockedIPRanges.IPv6 | default (list)) $rawIPv6 }}
{{- end }}
{{- $blockedIPv4 := list }}
{{- range $rawIPv4 }}{{ $blockedIPv4 = append $blockedIPv4 (include "synthetics-private-location.cidr" .) }}{{ end }}
{{- $blockedIPv6 := list }}
{{- range $rawIPv6 }}{{ $blockedIPv6 = append $blockedIPv6 (include "synthetics-private-location.cidr" .) }}{{ end }}
{{- toJson (dict "IPv4" $blockedIPv4 "IPv6" $blockedIPv6) }}
{{- end -}}

{{/*
Allowed CIDRs for the multi-container NetworkPolicy: allowedIPRanges IPv4 then IPv6, each normalized with synthetics-private-location.cidr.
Returns a JSON list.
*/}}
{{- define "synthetics-private-location.multiContainer.networkPolicy.allowedCIDRs" -}}
{{- $np := .Values.multiContainer.networkPolicy }}
{{- $cidrs := list }}
{{- range (concat ($np.allowedIPRanges.IPv4 | default (list)) ($np.allowedIPRanges.IPv6 | default (list))) }}{{ $cidrs = append $cidrs (include "synthetics-private-location.cidr" .) }}{{ end }}
{{- toJson $cidrs }}
{{- end -}}

{{/*
Local port of the browser-pool HTTP server.
*/}}
{{- define "synthetics-private-location.multiContainer.browserPoolPort" -}}
9444
{{- end -}}

{{/*
Local port of the traceroute-server HTTP server.
*/}}
{{- define "synthetics-private-location.multiContainer.tracerouteServerPort" -}}
3765
{{- end -}}

{{/*
UID of the nobody user, used by the browser-pool and traceroute-server containers.
*/}}
{{- define "synthetics-private-location.multiContainer.nobodyUser" -}}
65534
{{- end -}}

{{/*
UID and GID of the deno-runner user in the image. deno-executor runs as it, and the pod fsGroup uses it so the worker can use the 0660 deno-executor socket.
*/}}
{{- define "synthetics-private-location.multiContainer.denoRunnerId" -}}
502
{{- end -}}

{{/*
UID and GID of the dog user in the image, used by the worker container.
*/}}
{{- define "synthetics-private-location.multiContainer.dogUser" -}}
501
{{- end -}}

{{/*
Port ranges of one protocol that the allow-all egress rule may use: every port from 1 to 65535 except the dns.ports of that protocol (protocol defaults to TCP).
Takes a dict with "ports" (multiContainer.networkPolicy.dns.ports) and "protocol" ("TCP" or "UDP"). Returns a JSON list of {"port": N} or {"port": N, "endPort": M}, sorted.
*/}}
{{- define "synthetics-private-location.multiContainer.networkPolicy.nonDNSPortRanges" -}}
{{- $protocol := .protocol }}
{{- $padded := list }}
{{- range .ports }}
{{- if not (kindIs "float64" .port) }}{{ fail (printf "multiContainer.networkPolicy.dns.ports port %v must be a number" .port) }}{{ end }}
{{- if eq (.protocol | default "TCP") $protocol }}{{ $padded = append $padded (printf "%05d" (.port | int)) }}{{ end }}
{{- end }}
{{- $ranges := list }}
{{- $start := 1 }}
{{- range ($padded | uniq | sortAlpha) }}
{{- $p := . | atoi }}
{{- if eq $start (sub $p 1) }}{{ $ranges = append $ranges (dict "port" $start) }}
{{- else if lt $start $p }}{{ $ranges = append $ranges (dict "port" $start "endPort" (sub $p 1)) }}{{ end }}
{{- $start = add $p 1 }}
{{- end }}
{{- if eq $start 65535 }}{{ $ranges = append $ranges (dict "port" $start) }}
{{- else if lt $start 65535 }}{{ $ranges = append $ranges (dict "port" $start "endPort" 65535) }}{{ end }}
{{- toJson $ranges }}
{{- end -}}

{{/*
Name of the cluster-scoped OpenShift SCC: the fullname plus the release namespace, so same-named releases in different namespaces do not collide. Truncated to 253 characters.
*/}}
{{- define "synthetics-private-location.multiContainer.sccName" -}}
{{- printf "%s-%s" (include "synthetics-private-location.fullname" .) .Release.Namespace | trunc 253 | trimSuffix "-" }}
{{- end -}}
