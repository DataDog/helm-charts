# AI / Agent Guide: Datadog Helm Charts

Project-level context for AI coding assistants working on this repository.

## Specialized References

- **datadog chart — PR review guide** — When reviewing PRs to `charts/datadog`, see [charts/datadog/docs/internal/agent-review-guide.md](charts/datadog/docs/internal/agent-review-guide.md) for upgrade compatibility rules, resource naming, changelog requirements, and CI test notes. Also always see [charts/datadog/docs/internal/gke-constraints-review-guide.md](charts/datadog/docs/internal/gke-constraints-review-guide.md) and actively check that any changes to DaemonSet containers, commands, args, volumes, capabilities, or securityContext are compliant with the WorkloadAllowlist constraints documented there — any mismatch or unexempted default restriction can silently break installs on GKE Autopilot and GKE GDC.
- **Helm-to-Operator migration tooling** — When editing migration templates (`charts/datadog/templates/migration*.yaml`), mapping file (`charts/datadog/files/mapping_datadog_helm_to_datadogagent_crd.yaml`), NOTES.txt migration sections, or _helpers.tpl migration helpers, see [charts/datadog/docs/internal/helm-operator-migration-reference.md](charts/datadog/docs/internal/helm-operator-migration-reference.md) for constraints, implementation details, and file references.
- **New `charts/datadog` values and the migration mapping** — Every new `charts/datadog` value must be added to `charts/datadog/files/mapping_datadog_helm_to_datadogagent_crd.yaml`, or explicitly left unmapped there, so the Helm-to-Operator preview and migration handle it.
- **GKE Autopilot / GDC beyond the datadog DaemonSet** — The WorkloadAllowlist constraints also apply to `charts/datadog-csi-driver`, and to hostPaths, capabilities, or containers added for other providers (e.g. Talos): gate them so they never render when `providers.gke.autopilot` or `providers.gke.gdc` is set.

## Repository Map

- `charts/` — `cloudprem`, `datadog`, `datadog-crds`, `datadog-csi-driver`, `datadog-operator`, `extended-daemon-set`, `observability-pipelines-worker`, `private-action-runner`, `synthetics-private-location`.
- `test/` — Go unit, integration, and e2e tests; golden manifests live in `test/<chart>/baseline/manifests/`.
- **Generated — never edit by hand:** chart `README.md` files (`.github/helm-docs.sh`), baseline manifests (`make update-test-baselines-<chart>`), and the `AUTO-GENERATED RULES` section of `charts/datadog-operator/templates/clusterrole.yaml` (CI fails on hand edits; manual RBAC goes in its `MANUAL RULES` section).
- **Bot-owned:** `datadog-operator` / `datadog-crds` release PRs and Agent version bumps are opened by release automation and Renovate. Don't bump these by hand.

## Every Chart Change (all charts, not only datadog)

1. Bump `charts/<chart>/Chart.yaml` (patch for fixes, minor for features) and apply the matching `<chart>/patch-version`, `<chart>/minor-version`, or `<chart>/no-version-bump` label. An image or appVersion bump still needs a chart version bump.
2. Add an entry to `charts/<chart>/CHANGELOG.md`.
3. Regenerate docs with `.github/helm-docs.sh` after bumping.
4. Run `make unit-test-<chart>` and `make update-test-baselines-<chart>` where the chart has them (datadog via `-datadog` / `-datadog-agent`, operator, datadog-csi-driver, private-action-runner), and review the baseline diffs you commit.
5. After bumping a subchart dependency (e.g. `datadog-operator` in `charts/datadog/requirements.yaml`), regenerate baselines and check that RBAC still covers versions users can still pin.
6. Open PRs from a branch in `DataDog/helm-charts`, not a fork, so CI jobs that need write permissions (labels, version checks, OIDC) can run.

## Template Conventions

- **Survive `helm upgrade --reuse-values`.** Reused values from older releases lack newly added maps, so read new nested values defensively: `dig "feature" "enabled" false .Values` or `(.Values.feature | default dict).enabled`, never a bare `.Values.feature.enabled`.
- **Gate a feature and its RBAC with the same helper.** Don't restate the condition in the workload and RBAC templates; they drift (e.g. a component enabled without permission to list what it needs).
- **Enforce feature dependencies.** If a feature needs another one (e.g. Remote Configuration), enable it or `fail` with a clear message; never render a configuration that cannot work.
- **Don't collide with user workarounds.** Before adding a default env var, volume, or mount, check whether users can already provide it through `extraEnv`, `envFrom`, or `volumes`/`extraVolumes`: avoid duplicate volume names (Kubernetes rejects the workload) and `env` defaults that silently override `envFrom` values.
- **No magic strings.** Repeated names, ports, and paths come from named templates or values.
- **Validate input precisely.** Use `regexMatch` rather than `contains` when checking user-supplied values.
- **Keep files focused.** Put a feature's helpers in their own `_<feature>.tpl`; put embedded scripts and config in `files/` and load them with `.Files.Get` instead of inlining them in templates.
- **Production-ready containers.** New containers get readiness and liveness probes, resource defaults, and a securityContext.
- **Document every new value** with helm-docs comments in `values.yaml`, and in `values.schema.json` where the chart has one (`datadog`, `private-action-runner`).
- **CRDs before CRs.** Don't render a custom resource whose CRD the same install may not have registered yet.

## Tests

- Every fix gets a unit test or baseline case that **fails without the fix**; a baseline that only reflects incidental changes doesn't count.
- Cover feature combinations, including the case where other features are off: their default permissions, volumes, or env vars can hide a gap.

## Reviewing PRs

- Review only the changes the PR introduces, and give a concrete failure scenario (values, upgrade path, resulting error) for each finding.
- **Commit signing:** don't run `git verify-commit` in a review checkout. The commit there may be one the review environment created, not the PR's, which has produced false "unsigned commit" findings. GitHub enforces signing; if you must check, read `verification.verified` for the PR's own commits from the GitHub API.
