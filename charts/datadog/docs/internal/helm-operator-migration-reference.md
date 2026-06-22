# Helm-to-Operator Migration: Reference for AI and Maintainers

This document captures Helm v3, Kubernetes, and migration-flow constraints that affect the **Helm (datadog chart) → standalone Datadog Operator** migration. Use it when editing migration logic, NOTES.txt, or operator subchart behavior.

---

## 1. Migration Flow

1. **Enable migration** on the datadog chart (with `operator.datadogCRDs.keepCrds=true`) and run the migration job so the `DatadogAgent` CR is created.
2. **Install the standalone operator chart** with a release name whose deployment name does **not** collide with the subchart's (see §4 for the collision rule). Use `datadogCRDs.crds.datadogAgents=true` and `--take-ownership`. Duplicate operator pods are expected until step 3.
3. **Uninstall the datadog chart.** Do **not** run further `helm upgrade` on the datadog chart after migration (e.g. to disable the operator subchart); that can trigger immutable-field errors when the chart tries to recreate resources.

---

## 2. Migration Options and Chart Implementation

### 2.1 `datadog.operator.migration.preview` and `datadog.operator.migration.enabled`

Both options run the migration job (see §2.2). They differ in whether the `DatadogAgent` manifest is **applied** to the cluster.

| Option | Purpose | Behavior |
|--------|---------|----------|
| **`preview`** | Dry-run / validation | Runs the **dda-mapper** container only. Maps Helm values → DatadogAgent manifest but does **not** apply the CR. The mapped manifest can be viewed in the `dda-mapper` container logs (`kubectl logs job/<release>-dda-migration-job -c dda-mapper`). Does **not** require `operator.datadogCRDs.keepCrds`. Typical flow: enable preview → review logs → enable migration. |
| **`enabled`** | Full migration | Runs **dda-mapper** and **dda-migrator**. When mapping succeeds and the DatadogAgent CRD is present, the migrator applies the manifest (with `agent.datadoghq.com/helm-migration: "true"` annotation). Requires **`operator.datadogCRDs.keepCrds: true`** (validation fails otherwise). Also grants RBAC (get, patch, create on `datadogagents`) to the agents ServiceAccount (`rbac.yaml`). |

**Prerequisites for the migration job** (both modes): `migration-supported` must be true — i.e. `datadog.operator.enabled`, DatadogAgent CRD v2alpha1 present, and operator image tag ≥ 1.22.0 (or `operator.image.doNotCheckTag`). Both modes require **`datadog.operator.migration.userValues`** (via `--set-file`); otherwise the template fails with an error.

### 2.2 Migration Job and Dependent Resources

The migration is implemented by a Kubernetes Job (`migration-job.yaml`) and two ConfigMaps.

#### Job

- **Condition:** Renders when `migration-supported` is true AND (`migration.enabled` OR `migration.preview`).
- **Helm hook:** `post-install,post-upgrade`; `before-hook-deletion`.
- **Name:** `{{ template "datadog.fullname" . }}-dda-migration-job`.
- **Containers:**
  - **dda-mapper** (always): Uses the operator image. Runs `/yaml-mapper map` with:
    - `--sourcePath=/tmp/values.yaml` (user-provided Helm values)
    - `--mappingPath=/tmp/mapping_datadog_helm_to_datadogagent_crd.yaml`
    - `--destPath=/tmp/<release>.yaml`
    - `--ddaName`, `--namespace`
  - Writes `SUCCEEDED` or `FAILED` to `/tmp/mapper-status`.
  - In **preview** mode, or when **enabled** but DatadogAgent CRD is not ready: prints completion message only; no apply.
  - **dda-migrator** (only when `migration.enabled` AND `datadogagents-crd-ready`): Uses `bitnami/kubectl`. Waits for mapper status, injects `agent.datadoghq.com/helm-migration: "true"` into the manifest metadata, then runs `kubectl apply -f` on the DatadogAgent manifest.

#### Dependent ConfigMaps

| ConfigMap | Source | When created | Contents |
|-----------|--------|--------------|----------|
| `{release}-values-config` | `migration-values-configmap.yaml` | `migration.enabled` OR `migration.preview` (and `userValues` set) | `values.yaml` key = `datadog.operator.migration.userValues` (Helm values string). Annotated with `checksum/migration-config` for change detection. |
| `{release}-migration-mapper-config` | `migration-mapper-configmap.yaml` | `migration.enabled` OR `migration.preview` | `mapping_datadog_helm_to_datadogagent_crd.yaml` key = contents of `files/mapping_datadog_helm_to_datadogagent_crd.yaml`. |

#### Mapping file

`files/mapping_datadog_helm_to_datadogagent_crd.yaml` defines the Helm → DatadogAgent CR spec mapping. Keys are dotted Helm chart paths (e.g. `agents.image.name`); values are DatadogAgent spec paths (e.g. `spec.override.nodeAgent.image.name`). Empty string means no mapping. The `yaml-mapper` binary (packaged in the operator image) reads this file and the user values to produce the DatadogAgent manifest.

---

## 3. Helm v3 Nuances

### 3.1 Release names are unique

A Helm release is identified by `(release name, namespace)`. You **cannot** install the standalone operator chart with the same release name as the datadog chart while the datadog chart is still installed.

### 3.2 Subcharts inherit the parent release context

When the datadog chart includes datadog-operator as a subchart, both share the **same** release. In the subchart templates, **`.Release.Name`** is the parent release name (e.g. `dd`, `datadog`), not `datadog-operator`.

### 3.3 Alias overrides `.Chart.Name`

In `charts/datadog/requirements.yaml`, the operator dependency uses **`alias: operator`**. When rendered as a subchart, **`.Chart.Name` = `"operator"`** (not `"datadog-operator"`), which affects resource names and labels:
- `app.kubernetes.io/name` = **`operator`**
- `app.kubernetes.io/instance` = **parent release name**

### 3.4 `--take-ownership` (Helm 3.17+)

- Lets the current release adopt existing resources (e.g. CRDs) that would otherwise be created by the chart. Helm relabels them so the previous release no longer owns them.
- After a successful take-over, uninstalling the datadog chart will **not** delete adopted resources.
- **Limitation:** If `--take-ownership` is omitted or fails, uninstalling the datadog chart could still delete CRDs. Hence we require `keepCrds: true` as a safety net (see §5).

---

## 4. Deployment Name Collision

### 4.1 Why collisions cause errors

Kubernetes Deployment `spec.selector.matchLabels` are **immutable** after creation. If the standalone operator chart creates a Deployment with the **same name** as the subchart's but with **different selector labels**, Kubernetes rejects the update. The subchart and standalone chart produce different label values because `.Chart.Name` differs (`"operator"` vs `"datadog-operator"`) and `.Release.Name` differs.

### 4.2 Selector labels (cannot be overridden)

The operator Deployment selector uses:
- `app.kubernetes.io/name`: from `include "datadog-operator.name" .` (chart name or `nameOverride`; subchart with alias → `"operator"`)
- `app.kubernetes.io/instance`: `.Release.Name`

**`fullnameOverride` only affects resource names** (Deployment name, Service name), **not** selector labels. It cannot fix a label mismatch.

### 4.3 Fullname resolution logic

Both the subchart and standalone chart use the same `fullname` helper template:

1. If `fullnameOverride` is set → name = `fullnameOverride` (truncated to 63 chars).
2. Else compute `$name` = `nameOverride` or chart name (subchart: `"operator"`, standalone: `"datadog-operator"`).
   - If release name **contains** `$name` (substring match) → name = release name.
   - Else → name = `release-$name`.

### 4.4 When collisions occur

**Subchart Deployment name S** = `Release.Name-operator` (since chart name = alias `"operator"`).

**Standalone Deployment name T**: if the standalone release name contains `"datadog-operator"` → T = release name; else T = `release-datadog-operator`.

A collision (S = T) via reusing the subchart deployment name as the standalone release **only** occurs when S itself contains the substring `"datadog-operator"`. Since S = `"{R}-operator"`, this happens precisely when R ends with `"datadog"`.

| Datadog release | Subchart deployment (S) | S contains `"datadog-operator"`? | Standalone release = S → Collision? |
|-----------------|-------------------------|----------------------------------|--------------------------------------|
| `datadog` | `datadog-operator` | Yes | **YES** (T = `datadog-operator`) |
| `my-datadog` | `my-datadog-operator` | Yes | **YES** (T = `my-datadog-operator`) |
| `dd` | `dd-operator` | No | No (T = `dd-operator-datadog-operator`) |
| `datadog-dd` | `datadog-dd-operator` | No | No (T = `datadog-dd-operator-datadog-operator`) |
| `monitoring` | `monitoring-operator` | No | No (T = `monitoring-operator-datadog-operator`) |

The NOTES.txt conditionally warns about the forbidden release name only when `contains "datadog-operator" (include "operator-subchart-deployment-name" .)`. See `test/datadog/operator_migration_helpers_test.go` for Go tests.

### 4.5 Avoiding collisions with overrides (via `fullnameOverride` / `nameOverride` in the `datadog-operator.fullname` helper)

#### `fullnameOverride` (recommended)

Directly sets the standalone Deployment name, bypassing the `contains` substring logic entirely.

**Example** (parent release `"datadog"`, S = `"datadog-operator"`):
```
helm install datadog-operator datadog/datadog-operator --set fullnameOverride=datadog-operator-standalone
```
Deployment name = `"datadog-operator-standalone"` — no collision.

#### `nameOverride` (NOT recommended)

Replaces `"datadog-operator"` in the `contains` check. Because `contains` does **substring** matching, short override values easily match within the release name, collapsing the deployment name back to just the release name:

| `nameOverride` | `contains` check | Deployment name | Collision? |
|-----------------|-------------------|-----------------|------------|
| `"op"` | `contains "op" "datadog-operator"` = true | `"datadog-operator"` | YES |
| `"operator"` | `contains "operator" "datadog-operator"` = true | `"datadog-operator"` | YES |
| `"standalone"` | `contains "standalone" "datadog-operator"` = false | `"datadog-operator-standalone"` | No, but fragile |

---

## 5. CRDs and keepCrds

### 5.1 CRDs must not be deleted during migration

**DatadogAgent**, **DatadogAgentInternal**, and other operator-managed CRDs **must not** be removed at any point. If deleted, the cluster loses the schema and existing CRs become invalid.

### 5.2 `operator.datadogCRDs.keepCrds` (safety net)

- Annotates CRDs with `helm.sh/resource-policy: keep` so Helm skips them on uninstall.
- Required when migration is enabled. Even with `--take-ownership` transferring CRD ownership to the standalone release, `keepCrds` protects against the case where `--take-ownership` is omitted or fails.
- **Do not relax** the validation that migration requires `keepCrds`.

---

## 6. Files and Helpers to Keep Consistent

| File | Key details |
|------|-------------|
| `charts/datadog/templates/NOTES.txt` | Migration sections: (1) always warn that standalone deployment name must not match `operator-subchart-deployment-name`, (2) conditionally warn about the forbidden release name only when `contains "datadog-operator" (include "operator-subchart-deployment-name" .)`, (3) instruct user to uninstall the datadog chart. Never advise `helm upgrade` after migration. |
| `charts/datadog/templates/_helpers.tpl` | `operator-subchart-deployment-name`: exact subchart Deployment name (alias `"operator"` → `Release.Name-operator`). `operator-forbidden-standalone-release-name`: same value; warning only shown when it contains `"datadog-operator"`. `operator-standalone-install-command`: uses release name `"operator"` (deployment `operator-datadog-operator`). `migration-supported`: operator enabled + DatadogAgent CRD v2alpha1 + operator image ≥ 1.22.0. `datadogagents-crd-ready`: `Capabilities.APIVersions.Has "datadoghq.com/v2alpha1/DatadogAgent"`. |
| `charts/datadog/templates/migration-job.yaml` | Job runs when `migration-supported` AND (migration.enabled OR migration.preview). Requires userValues. dda-mapper uses operator image; dda-migrator (bitnami/kubectl) only when migration.enabled AND datadogagents-crd-ready. |
| `charts/datadog/templates/migration-values-configmap.yaml` | ConfigMap `{release}-values-config` with user Helm values. Created when migration.enabled OR migration.preview; requires userValues. |
| `charts/datadog/templates/migration-mapper-configmap.yaml` | ConfigMap `{release}-migration-mapper-config` with mapping rules. Created when migration.enabled OR migration.preview. |
| `charts/datadog/files/mapping_datadog_helm_to_datadogagent_crd.yaml` | Helm chart key → DatadogAgent spec path mapping for yaml-mapper. |
| `charts/datadog-operator/templates/deployment.yaml` | Selector and pod labels use `include "datadog-operator.name" .` and `.Release.Name`. Changing these breaks existing installs — requires a migration path (e.g. optional override). |
| `charts/datadog/requirements.yaml` | Operator dependency uses `alias: operator` → subchart `.Chart.Name` = `"operator"`. |
