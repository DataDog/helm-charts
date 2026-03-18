# datadog Chart — PR Review Guide for AI Agents

Actionable guidance for reviewing PRs to `charts/datadog`. Covers what CI cannot catch automatically.

> **Scope rule:** Focus only on the changes introduced by the PR under review. Do not flag pre-existing issues in unchanged code — those are out of scope and distract from the actual review.

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

PRs that modify chart templates, values, or chart behaviour require:

1. Version bump in `charts/datadog/Chart.yaml` (patch for fixes, minor for new features) — use label `datadog/patch-version` or `datadog/minor-version`
2. Entry in `charts/datadog/CHANGELOG.md`
3. Updated README via `.github/helm-docs.sh`
4. Updated baseline manifests via `make update-test-baselines-datadog-agent`

PRs that only change CI/tooling, tests, or documentation (no chart template or values changes) may use `datadog/no-version-bump` and do not require a `Chart.yaml` or `CHANGELOG.md` update. Do not flag these as missing a version bump.

All PRs require:

5. All commits signed and showing as "Verified" on GitHub (GPG, SSH, or S/MIME)

---

## 4. CI Test Notes

- Unit tests: `make unit-test-datadog` — must pass before merge.
- Baseline manifests in `test/datadog/baseline/manifests/` are golden files. Unexpected diffs signal unintended side effects.

**Reviewer action:** When baseline manifest diffs are present, spot-check the rendered YAML for Kubernetes correctness — missing required fields (e.g., `name`, `containers`, `selector`), invalid field types, mismatched label selectors between a workload's `spec.selector.matchLabels` and its pod template `metadata.labels`, or malformed volume/mount definitions. These errors pass Helm rendering but fail at apply time.

---

## 5. CODEOWNERS — add new team-owned templates

If a PR introduces a new team-specific template (e.g. `_container-<feature>.yaml`, `<feature>-configmap.yaml`), the author should add it to `.github/CODEOWNERS` under their team. This ensures correct ownership is recorded for future review requests.

Example: if `@DataDog/some-team` adds `charts/datadog/templates/_container-some-feature.yaml`, add:
```
charts/datadog/templates/_container-some-feature.yaml  @DataDog/some-team
```

**Reviewer action:** If a PR adds new `charts/datadog/templates/` files with a clear team owner but does not update CODEOWNERS, flag it.

---

## 6. Avoid Redundant Additions — Prefer Template Simplification

When a PR adds new logic, conditionals, or helpers, check whether the same outcome could be achieved by simplifying or reusing existing template code.

| Pattern to flag | Preferred alternative |
|---|---|
| New helper that duplicates logic already in an existing helper | Extend or parameterise the existing helper |
| Duplicated `if`/`else` blocks across multiple templates | Extract to a shared named template in `_helpers.tpl` |
| New template file for a feature that could be a conditional block in an existing file | Add the block to the existing file with a feature gate |
| Copy-pasted value mappings (e.g., env vars, volume mounts) that already exist in another container template | Refactor into a shared partial |

**Reviewer action:** Before approving new template code, verify that no existing helper or partial already covers the same logic. Suggest the approach that is simplest to read, maintain, and reuse.

---

## 7. GKE Autopilot and GDC Constraints

If the PR touches DaemonSet volumes, hostPaths, capabilities, containers, or securityContext fields, also consult [gke-constraints-review-guide.md](gke-constraints-review-guide.md).
