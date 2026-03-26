# AI / Agent Guide: Datadog Helm Charts

Project-level context for AI coding assistants working on this repository.

## Specialized References

- **datadog chart — PR review guide** — When reviewing PRs to `charts/datadog`, see [charts/datadog/docs/internal/agent-review-guide.md](charts/datadog/docs/internal/agent-review-guide.md) for upgrade compatibility rules, resource naming, changelog requirements, and CI test notes. For PRs touching DaemonSet volumes, hostPaths, capabilities, or containers, also see [charts/datadog/docs/internal/gke-constraints-review-guide.md](charts/datadog/docs/internal/gke-constraints-review-guide.md).
- **Helm-to-Operator migration tooling** — When editing migration templates (`charts/datadog/templates/migration*.yaml`), mapping file (`charts/datadog/files/mapping_datadog_helm_to_datadogagent_crd.yaml`), NOTES.txt migration sections, or _helpers.tpl migration helpers, see [charts/datadog/docs/internal/helm-operator-migration-reference.md](charts/datadog/docs/internal/helm-operator-migration-reference.md) for constraints, implementation details, and file references.
