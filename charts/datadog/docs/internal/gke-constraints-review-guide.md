# datadog Chart — GKE Autopilot and GDC Constraint Review Guide

Reference for reviewing PRs that touch DaemonSet volumes, hostPaths, capabilities, containers, or securityContext fields. Changes in these areas can silently break installs on GKE Autopilot and GKE Distributed Cloud (GDC).

---

## 1. GKE Autopilot — WorkloadAllowlist (clusters >= 1.32.1-gke.1729000)

The Datadog WorkloadAllowlist grants exemptions for the Datadog DaemonSet on GKE Autopilot. The Warden admission webhook enforces it at install and upgrade time. A mismatch produces:

```
Workload Mismatches Found for Allowlist
```

### securityContext restrictions

The WorkloadAllowlist only evaluates **three** securityContext fields: `capabilities`, `privileged`, `appArmorProfile`.
`readOnlyRootFilesystem` is **not** evaluated — it is allowed generally by Autopilot.

### Allowed hostPaths

Any hostPath not in this list triggers a Warden rejection:

```
/var/run/datadog           /var/lib/docker/containers   /var/run/containerd
/sys/fs/cgroup             /var/log/containers          /proc
/etc/passwd                /var/autopilot/addon/datadog/logs
/var/log/pods              /etc/os-release              /sys/kernel/debug
/var/tmp/datadog-agent/system-probe/build
/var/tmp/datadog-agent/system-probe/kernel-headers
/var/lib/kubelet/seccomp   /                            /lib/modules
/sys/fs/bpf                /etc/apt                     /etc/yum.repos.d
/etc/zypp                  /etc/pki                     /etc/yum/vars
/etc/dnf/vars              /etc/rhsm
```

**Reviewer action:** If a PR adds a new hostPath volume to the DaemonSet that is not in this list, flag it — it will break GKE Autopilot installs unless gated.

### Allowed capabilities (system-probe container only)

```
BPF, CHOWN, DAC_READ_SEARCH, IPC_LOCK, NET_ADMIN, NET_BROADCAST, NET_RAW, SYS_ADMIN, SYS_PTRACE, SYS_RESOURCE
```

**Reviewer action:** If a PR adds a capability not in this list, or adds any capability to a container other than `system-probe`, flag it.

### Volume constraints

`datadogrun` emptyDir is **not** allowed. The WorkloadAllowlist only permits `pointerdir` (hostPath) at `/opt/datadog-agent/run`.

**Reviewer action:** Flag any PR that introduces `datadogrun` emptyDir for Autopilot/GDC environments.

### The gating pattern

New features that add hostPaths, capabilities, or volumes not yet in the WorkloadAllowlist must be gated:

```
{{- if not (or .Values.providers.gke.autopilot .Values.providers.gke.gdc) }}
```

**Reviewer action:** Flag any PR that adds hostPaths, capabilities, or volumes to the unguarded DaemonSet spec without this gate — it will break Autopilot installs until the WorkloadAllowlist is updated.

### HELM_FORCE_RENDER

`datadog.envDict.HELM_FORCE_RENDER=true` is used in unit tests and CI (Kind clusters) to simulate a cluster with WorkloadAllowlist CRDs. It must **not** appear in production values files.

---

## 2. GKE Autopilot — AllowlistedV2Workload (clusters < 1.32.1-gke.1729000)

Legacy mode (`datadog-daemonset-dec2023`). The allowlist was written for an older chart version that ran `process-agent` and `trace-agent` as separate sidecar containers. The current chart runs process collection inside the core `agent` container by default, so only 1 container is rendered in this mode.

### What the allowlist permits vs. what the chart currently renders

| Container | Allowed by allowlist | Currently rendered by chart |
|---|---|---|
| `agent` | ✅ | ✅ |
| `process-agent` | ✅ | ❌ (runs in-process inside `agent`) |
| `trace-agent` | ✅ | ❌ (runs in-process inside `agent`) |
| `system-probe` | ❌ | ❌ (gated out) |
| `otel-agent` | ❌ | ❌ (disabled by default) |

**Reviewer action:** Any PR that adds `system-probe` or `otel-agent` to the unguarded Autopilot path will break installs on legacy clusters. The allowlist also permits `process-agent` and `trace-agent` as additional containers, but the current chart runs these in-process inside the core `agent` container.

### Allowed hostPaths

```
/var/lib/docker/containers   /var/run/containerd
/sys/fs/cgroup               /var/log/containers
/proc                        /etc/passwd
/var/autopilot/addon/datadog/logs   /var/log/pods
```

`pointerdir` (hostPath at `/var/autopilot/addon/datadog/logs`) is required — `datadogrun` emptyDir is not in the allowlist.

### No capabilities

No Linux capability exemptions are granted — `system-probe` (which requires `BPF`, `NET_ADMIN`, etc.) is not supported in this mode.

### Exemptions granted

- `autogke-no-write-mode-hostpath` — allows write-mode hostPath mounts
- `autogke-no-host-port` — allows host ports

Test file: `test/datadog/gke_autopilot_allowlistedv2workload_test.go`

---

## 3. GKE Distributed Cloud (GDC)

GDC is more restricted than GKE Autopilot. Only 1 container (core agent). Allowed hostPaths:

```
/var/datadog/logs   /var/log/pods   /var/log/containers
```

`/proc`, `/sys/fs/cgroup`, and other system-level paths are not allowed. Use `pointerdir` (hostPath at `/var/datadog/logs`), not `datadogrun` emptyDir.

**Reviewer action:** Any PR adding containers, hostPaths, or volumes must gate GDC with `{{- if not (or .Values.providers.gke.autopilot .Values.providers.gke.gdc) }}`.
Test file: `test/datadog/gke_gdc_test.go`

---

## 4. Test files to check

If a PR touches DaemonSet volumes, containers, or securityContext and does not update these tests, flag it as incomplete:

| Test file | What it covers |
|---|---|
| `test/datadog/gke_autopilot_workloadallowlist_test.go` | WorkloadAllowlist (section 1) |
| `test/datadog/gke_autopilot_allowlistedv2workload_test.go` | Legacy AllowlistedV2Workload (section 2) |
| `test/datadog/gke_gdc_test.go` | GDC constraints (section 3) |

---

## 5. E2E tests (internal DataDog developers)

If any of the above unit test files are updated by a PR, the corresponding E2E tests should be run against a real GKE Autopilot cluster before merge.

To trigger E2E tests:
1. Go to https://gitlab.ddbuild.io/DataDog/helm-charts/-/pipelines/
2. Find the pipeline corresponding to your commit
3. Manually trigger the relevant E2E job(s) (e.g. `e2e_autopilot`)
