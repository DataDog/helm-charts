# datadog Chart — PR Review Guide for AI Agents

Actionable guidance for reviewing PRs to `charts/datadog`. Covers what CI cannot catch automatically.

---

## 1. Helm Chart Upgrade Compatibility — Breaking Value Changes

Flag any PR that makes these changes without a deprecation notice, migration path, or major version bump.

### Value key changes

| Change type | Why it breaks |
|---|---|
| Renaming an existing values key | Users' existing `values.yaml` files break silently — old key ignored, feature may silently disable |
| Removing a previously supported values key | Same as above |
| Changing the type of a values key (e.g., bool → string) | Helm will error or silently misinterpret the value |
| Adding a new required value with no default | Users who do not set the value get a render error on upgrade |
| Changing a previously stable default (e.g., enabling a feature that was off) | Opt-out behavior changes silently for all existing installs |

### Kubernetes resource changes

| Change type | Why it breaks |
|---|---|
| Renaming a Kubernetes resource | Helm tries to create the new resource while the old one still exists — conflict on upgrade |
| Changing a ClusterRole, ClusterRoleBinding, or ServiceAccount name | Breaks RBAC for existing installs; old bindings become orphaned |
| Changing `spec.selector` or `spec.selector.matchLabels` on a DaemonSet, Deployment, or StatefulSet | These fields are immutable in Kubernetes — `helm upgrade` will fail with an immutable field error |
| Adding, removing, or renaming labels referenced by a selector | Same as above — the selector and pod template labels must stay in sync and cannot change after initial creation |

**Reviewer action:** Flag renames, removals, type changes, default changes, Kubernetes resource renames, and any modifications to `spec.selector` or pod template labels on workload resources as potential breaking changes.

---

## 2. Kubernetes Resource Naming

Resource names are derived from the Helm release name via `{{ template "datadog.fullname" . }}` in `charts/datadog/templates/_helpers.tpl`. These cannot change without breaking upgrades.

**Reviewer action:** If a PR modifies `datadog.fullname` usage or hardcodes a new resource name suffix that differs from the existing pattern, flag it — it will cause a resource conflict on `helm upgrade`.

---

## 3. CHANGELOG and Version Bump Requirements

Every chart change requires all of the following. Flag PRs missing any:

1. Version bump in `charts/datadog/Chart.yaml` (patch for fixes/docs, minor for new features)
2. Entry in `charts/datadog/CHANGELOG.md`
3. Updated README via `.github/helm-docs.sh`
4. Updated baseline manifests via `make update-test-baselines-datadog-agent`
5. All commits signed and showing as "Verified" on GitHub (GPG, SSH, or S/MIME)

---

## 4. CI Test Notes

- Unit tests: `make unit-test-datadog` — must pass before merge.
- Baseline manifests in `test/datadog/baseline/manifests/` are golden files. Unexpected diffs signal unintended side effects.

---

## 5. GKE Autopilot and GDC Constraints

If the PR touches DaemonSet volumes, hostPaths, capabilities, containers, or securityContext fields, also consult [gke-constraints-review-guide.md](gke-constraints-review-guide.md).
