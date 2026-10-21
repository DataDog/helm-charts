# Datadog Synthetics Private Location

![Version: 1.0.0](https://img.shields.io/badge/Version-1.0.0-informational?style=flat-square) ![AppVersion: 1.74.0](https://img.shields.io/badge/AppVersion-1.74.0-informational?style=flat-square)

[Datadog](https://www.datadoghq.com/) is a hosted infrastructure monitoring platform. This chart adds a Datadog Synthetics Private Location Deployment. For more information about synthetics monitoring with Datadog, please refer to the [Datadog documentation website](https://docs.datadoghq.com/synthetics/private_locations/?tab=helmchart).

## How to use Datadog Helm repository

You need to add this repository to your Helm repositories:

```
helm repo add datadog https://helm.datadoghq.com
helm repo update
```

## Quick start

To install the chart with the release name `<RELEASE_NAME>`, retrieve your Private Location configuration file from your [Synthetics Private Location settings page](https://app.datadoghq.com/synthetics/settings/private-locations/) and save it under `config.json` then run:

```bash
helm install <RELEASE_NAME> datadog/synthetics-private-location --set-file configFile=config.json
```

## Multi-container deployment (preview)

Set `multiContainer.enabled=true` to run the private location without `sudo`, `NET_ADMIN` or privilege escalation. The worker, `deno-executor`, `browser-pool` and `traceroute-server` run as separate containers in the same pod, from the same image. This layout is compatible with clusters that enforce `allowPrivilegeEscalation: false` and `readOnlyRootFilesystem: true`.

This mode requires a private location image version >= 1.77.0. The chart rejects older versions, prerelease tags and custom tags when this mode is enabled. Version 1.77.0 is not published yet; the examples below are for use once a compatible image is available.

```bash
helm install <RELEASE_NAME> datadog/synthetics-private-location \
  --set-file configFile=config.json \
  --set image.tag=1.77.0 \
  --set multiContainer.browserPool.adminTokenSecret.name=pl-browser-pool \
  --set multiContainer.networkPolicy.mode=external \
  --set multiContainer.enabled=true
```

Before installing, provision a Secret named `pl-browser-pool` in the release namespace with a `token` key containing at least 32 cryptographically random bytes (for example, hex-encoded). Use your secret manager or create it outside Helm. The chart references this Secret only in the worker and browser-pool containers; it does not generate or store the token in Helm values. Replicas share this credential. Restrict access to the Secret and restart the Deployment after rotating it, because environment variables do not reload automatically. The browser launcher must keep the admin token out of browser child-process environments.

In this layout the worker cannot apply its `iptables` firewall. You must explicitly choose `multiContainer.networkPolicy.mode`:

- `external`: the chart creates no policy. The operator is responsible for egress enforcement. The commands above use this mode; they do not restrict egress by themselves.
- `managed`: the chart creates a pod-wide egress policy, blocking reserved ranges by default. Configure DNS resolver destinations and any necessary exceptions. The policy also restricts worker traffic, including upstream proxies.

Example managed-mode values (verify the DNS labels match your cluster):

```yaml
image:
  tag: "1.77.0"
multiContainer:
  enabled: true
  browserPool:
    adminTokenSecret:
      name: pl-browser-pool
      key: token
  networkPolicy:
    mode: managed
    dns:
      to:
        - namespaceSelector:
            matchLabels:
              kubernetes.io/metadata.name: kube-system
          podSelector:
            matchLabels:
              k8s-app: kube-dns
    # Optional: allow an upstream proxy without opening the whole private subnet.
    additionalEgress:
      - to:
          - ipBlock:
              cidr: 10.20.30.40/32
        ports:
          - protocol: TCP
            port: 3128
```

For NodeLocal or custom DNS, replace `dns.to` with resolver `ipBlock` entries, for example `cidr: 169.254.20.10/32`; use your cluster's actual resolver addresses. DNS exceptions allow TCP/UDP port 53 only. Namespace and pod selectors in the same peer are combined, not alternatives.

Set `allowedIPRanges` and `blockedIPRanges` under `multiContainer.networkPolicy` to match the intended worker configuration; the chart cannot read configuration from an external Secret. `allowedIPRanges`, DNS destinations and `additionalEgress` grant exceptions even inside blocked ranges. Set `enableDefaultBlockedIpRanges=false` only if you intend to allow reserved ranges except for your explicit blocks.

This policy is not a complete replacement for the in-container firewall: it requires a CNI that enforces NetworkPolicies, other policies can broaden access, and ICMP filtering depends on the CNI. It cannot isolate containers within a pod or block a pod from itself. Validate actual DNS, proxy, ICMP/traceroute and blocked-destination behavior on your CNI before relying on it. See [Kubernetes NetworkPolicy semantics](https://kubernetes.io/docs/concepts/services-networking/network-policies/).

The `traceroute-server` container needs the `NET_RAW` capability. On clusters that enforce the `restricted` Pod Security Standard, set `multiContainer.tracerouteServer.enabled=false`. Network path (traceroute) features are then not available.

`podSecurityContext` applies to the pod and `securityContext` to the worker container. Both are merged over the hardened defaults, and your values take precedence. If `podSecurityContext.sysctls` supplies `net.ipv4.ping_group_range`, that value replaces the generated range; otherwise it defaults to the worker `securityContext.runAsGroup` (501). Ensure an explicit range includes the worker group. Duplicate sysctl names are rejected.

### OpenShift

The multi-container pod runs fixed non-root UIDs (501, 502 and 65534) with `fsGroup` 502, and `traceroute-server` needs `NET_RAW`. The default `restricted-v2` SCC allows neither. Set `multiContainer.podSecurity.securityContextConstraints.create=true` to create an SCC that allows them and nothing more: no root, no privilege escalation, no privileged containers, a read-only root filesystem and all other capabilities dropped. Creating an SCC requires cluster-admin permissions.

```bash
helm install <RELEASE_NAME> datadog/synthetics-private-location \
  --set-file configFile=config.json \
  --set image.tag=1.77.0 \
  --set multiContainer.browserPool.adminTokenSecret.name=pl-browser-pool \
  --set multiContainer.networkPolicy.mode=external \
  --set multiContainer.enabled=true \
  --set multiContainer.podSecurity.securityContextConstraints.create=true
```

If `extraVolumes` uses other volume types, add them to `multiContainer.podSecurity.securityContextConstraints.volumes`.

`env`, `envFrom` and `extraVolumeMounts` apply to the worker container only. Custom CA certificates mounted in the worker under `/etc/datadog/certs` are shared with `browser-pool`, and browsers send their traffic through the worker's proxy.

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Allows to specify affinity for Datadog Synthetics Private Location PODs |
| commonLabels | object | `{}` | Labels to apply to all resources |
| configConfigMap | string | `""` | Config Map that stores the configuration of the private location worker for the deployment |
| configFile | string | `"{}"` | JSON string containing the configuration of the private location worker |
| configSecret | string | `""` | Name of the secret that stores the configuration of the private location worker for the deployment. Use it only if you want to manage the secret outside of the Helm chart as using `configFile` will create a secret. The `data` inside the secret needs to have the key `synthetics-check-runner.json`. |
| dnsConfig | object | `{}` | DNS Config to set to the Datadog Synthetics Private Location PODs |
| dnsPolicy | string | `"ClusterFirst"` | DNS Policy to set to the Datadog Synthetics Private Location PODs |
| enableStatusProbes | bool | `false` | Enable both liveness and readiness probes (minimal private location image version required: 1.12.0) |
| env | list | `[]` | Set environment variables |
| envFrom | list | `[]` | Set environment variables from configMaps and/or secrets |
| extraVolumeMounts | list | `[]` | Optionally specify extra list of additional volumeMounts for container |
| extraVolumes | list | `[]` | Optionally specify extra list of additional volumes to mount into the pod |
| fullnameOverride | string | `""` | Override the full qualified app name |
| hostAliases | list | `[]` | Add entries to Datadog Synthetics Private Location PODs' /etc/hosts |
| image.pullPolicy | string | `"IfNotPresent"` | Define the pullPolicy for Datadog Synthetics Private Location image |
| image.repository | string | `"gcr.io/datadoghq/synthetics-private-location-worker"` | Repository to use for Datadog Synthetics Private Location image |
| image.tag | string | `"1.74.0"` | Define the Datadog Synthetics Private Location version to use |
| imagePullSecrets | list | `[]` | Datadog Synthetics Private Location repository pullSecret (ex: specify docker registry credentials) |
| multiContainer.browserPool.adminTokenSecret.key | string | `"token"` | Key containing the admin token in the existing Secret. Restart the Deployment after rotating the token. |
| multiContainer.browserPool.adminTokenSecret.name | string | `""` | Required in multi-container mode. Existing Secret containing a cryptographically random browser-pool admin token (at least 32 random bytes), shared by the worker and browser-pool. Created and rotated outside this chart. |
| multiContainer.browserPool.maxBrowserIdleMs | int | `300000` | Reclaim a browser after this long without a request from the worker. Not a cap on test duration. |
| multiContainer.browserPool.maxConcurrent | int | `20` | Concurrent browsers per replica. Past it the worker waits for a slot. Keep it at about twice the worker `concurrency` (default 10) and size `browserPool.resources` to match. |
| multiContainer.browserPool.resources | object | `{"limits":{"cpu":"2","memory":"6Gi"},"requests":{"cpu":"2","memory":"6Gi"}}` | Resource requests/limits for the browser-pool container |
| multiContainer.browserPool.shmSizeLimit | string | `"2Gi"` | Size of the in-memory `/dev/shm` volume. The runtime default (64Mi) crashes renderers on heavier pages. |
| multiContainer.denoExecutor.maxConcurrent | int | `10` | Concurrent `deno run` processes per replica. Past it the worker retries the run. |
| multiContainer.denoExecutor.resources | object | `{"limits":{"cpu":"200m","memory":"1Gi"},"requests":{"cpu":"200m","memory":"1Gi"}}` | Resource requests/limits for the deno-executor container. Each JS-using test step uses ~30-50Mi. |
| multiContainer.enabled | bool | `false` | Preview. Run the worker, deno-executor, browser-pool and traceroute-server as separate containers in one pod, none using `sudo`, `NET_ADMIN` or `allowPrivilegeEscalation`. Requires a private location image version >= 1.77.0; prerelease and custom tags are not supported. `podSecurityContext` applies to the pod and `securityContext` to the worker container, merged over hardened defaults. |
| multiContainer.networkPolicy.additionalEgress | list | `[]` | Additional standard NetworkPolicy egress rules, for example access to an upstream proxy. These exceptions take precedence over blocked ranges. |
| multiContainer.networkPolicy.allowedIPRanges | object | `{"IPv4":[],"IPv6":[]}` | CIDRs that stay reachable even when inside a blocked range |
| multiContainer.networkPolicy.blockedIPRanges | object | `{"IPv4":[],"IPv6":[]}` | Additional CIDRs to block, added to `defaultBlockedIPRanges` |
| multiContainer.networkPolicy.defaultBlockedIPRanges | object | `{"IPv4":["0.0.0.0/8","10.0.0.0/8","100.64.0.0/10","127.0.0.0/8","169.254.0.0/16","172.16.0.0/12","192.0.0.0/24","192.0.0.0/29","192.0.0.8/32","192.0.0.9/32","192.0.0.10/32","192.0.0.170/32","192.0.0.171/32","192.0.2.0/24","192.31.196.0/24","192.52.193.0/24","192.88.99.0/24","192.168.0.0/16","192.175.48.0/24","198.18.0.0/15","198.51.100.0/24","203.0.113.0/24","240.0.0.0/4","224.0.0.0/4","255.255.255.255/32"],"IPv6":["::1/128","::/128","64:ff9b::/96","64:ff9b:1::/48","100::/64","2001::/23","2001::/32","2001:1::1/128","2001:1::2/128","2001:2::/48","2001:3::/32","2001:4:112::/48","2001:5::/32","2001:10::/28","2001:20::/28","2001:db8::/32","2002::/16","2620:4f:8000::/48","fc00::/7","fe80::/10"]}` | IANA reserved ranges blocked when `enableDefaultBlockedIpRanges` is true. Same list as the worker, without `::ffff:0:0/96` (rejected in a NetworkPolicy, and covered by the IPv4 list). |
| multiContainer.networkPolicy.dns.to | list | `[]` | Required in managed mode. NetworkPolicy peers for DNS (TCP/UDP 53). Use namespace/pod selectors for cluster DNS or ipBlock CIDRs for NodeLocal/custom resolvers. There is no allow-all DNS exception. |
| multiContainer.networkPolicy.enableDefaultBlockedIpRanges | bool | `true` | Block the ranges in defaultBlockedIPRanges in managed mode. Explicit allowed ranges, DNS peers and additionalEgress can grant exceptions. |
| multiContainer.networkPolicy.mode | string | `""` | Required in multi-container mode: `managed` creates a pod-wide egress NetworkPolicy; `external` creates none and leaves egress enforcement to the operator. Requires a CNI that enforces NetworkPolicies. Other policies can broaden access; ICMP filtering is CNI-dependent. |
| multiContainer.podSecurity.securityContextConstraints.create | bool | `false` | If true, create an OpenShift SecurityContextConstraints resource that allows the multi-container pod (fixed non-root UIDs, `NET_RAW` for traceroute-server). Requires `multiContainer.enabled`. |
| multiContainer.podSecurity.securityContextConstraints.volumes | list | `["configMap","downwardAPI","emptyDir","projected","secret"]` | Allowed volume types. Add the types used by `extraVolumes`. |
| multiContainer.terminationGracePeriodSeconds | int | `620` | Must stay above the browser-pool drain timeout (10 minutes) so in-flight browser tests can finish |
| multiContainer.tracerouteServer.enabled | bool | `true` | Run the traceroute-server container, which needs the `NET_RAW` capability. Required for network path (traceroute) features. Disable on clusters enforcing the `restricted` Pod Security Standard. |
| multiContainer.tracerouteServer.logLevel | string | `"info"` | Log level of the traceroute-server container (`debug`, `info`, `warn`, `error`). Set the worker log level with the worker `verbosity` option or `DATADOG_WORKER_LOG_VERBOSITY`. |
| multiContainer.tracerouteServer.resources | object | `{"limits":{"cpu":"100m","memory":"128Mi"},"requests":{"cpu":"100m","memory":"128Mi"}}` | Resource requests/limits for the traceroute-server container |
| multiContainer.worker.resources | object | `{"limits":{"cpu":"500m","memory":"1Gi"},"requests":{"cpu":"500m","memory":"1Gi"}}` | Resource requests/limits for the worker container. Browsers run in the browser-pool container, so the worker needs less than in the single-container layout. `resources` is not used in the multi-container layout. |
| nameOverride | string | `""` | Override name of app |
| nodeSelector | object | `{}` | Allows to schedule Datadog Synthetics Private Location on specific nodes |
| podAnnotations | object | `{}` | Annotations to set to Datadog Synthetics Private Location PODs |
| podDisruptionBudget | object | `{"enabled":false,"minAvailable":1}` | Allows to create and configure PodDisruptionBudget for Datadog Synthetics Private Location deployment |
| podLabels | object | `{}` | Labels to be placed on pods managed by the deployment |
| podSecurityContext | object | `{}` | Security context to set to Datadog Synthetics Private Location PODs |
| priorityClassName | string | `""` | Allows to specify PriorityClass for Datadog Synthetics Private Location PODs |
| replicaCount | int | `1` | Number of instances of Datadog Synthetics Private Location |
| resources | object | `{}` | Set resources requests/limits for Datadog Synthetics Private Location PODs |
| securityContext | object | `{}` | Security context to set to the Datadog Synthetics Private Location container |
| serviceAccount.annotations | object | `{}` | Annotations for the service account |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `""` | The name of the service account to use. If not set name is generated using the fullname template |
| tolerations | list | `[]` | Allows to schedule Datadog Synthetics Private Location on tainted nodes |
