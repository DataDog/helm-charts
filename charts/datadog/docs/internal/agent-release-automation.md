# Agent release update PRs

Mend-hosted Renovate tracks published stable `7.x.y` releases of
[`DataDog/datadog-agent`](https://github.com/DataDog/datadog-agent/releases),
including both patch updates (for example, `7.84.0` to `7.84.1`) and minor updates
(for example, `7.84.1` to `7.85.0`). It opens a single `renovate/datadog-agent` PR for the three default image tags
and their three floating-tag version fallbacks. Go module dependencies remain
owned by ADMS; the existing GitHub Actions manager remains enabled.

The repository-wide `minimumReleaseAge` is one day for both Agent updates and
GitHub Actions updates. Agent patch and minor releases become eligible 24 hours
after publication; PR creation then depends on Renovate's next run and available
PR capacity, not the exact 24-hour mark. Prereleases, Agent 6, Agent 8, and image
build variants are excluded. Approval and merging are
manual, and the PR receives the `datadog/patch-version` label.

## Generated files

`.github/workflows/agent-release.yaml` handles only open, same-repository PRs
created by `renovate[bot]` on the exact `renovate/datadog-agent` branch. It:

1. Requires the branch to include its current `main` base.
2. Verifies that the three image tags and three template fallbacks all move to
   the same newer stable GitHub release, with no other source changes.
3. Reconstructs those changes in a trusted base checkout. PR scripts, templates,
   dependency declarations, and workflow definitions are never executed.
4. Checks availability of the default Agent and Cluster Agent image tags in GCR
   and public ECR. This is not an exhaustive check of every variant, architecture,
   or regional mirror.
5. Calculates a patch chart version using the shared chart-version utility and
   writes a changelog entry. `appVersion` remains unchanged.
6. Regenerates test baselines using locked Helm dependencies, reruns the Datadog
   unit tests, lints the chart, and regenerates its README with helm-docs.
7. Commits only the allowed chart and baseline files through the GitHub API.
   The commit is signed by GitHub and uses an expected head SHA so it cannot
   overwrite a concurrent rebase or edit. Identical generated output is a no-op.

The generic chart-version **bump** job skips this specific bot PR so it cannot
race the generation workflow or regenerate the README from old defaults. The
existing chart-version **validation** job still runs. Regular PRs retain their
existing chart-version behavior.

The workflow uses the narrowly scoped
`.github/chainguard/self.agent-release.create-commit.sts.yaml` policy. Its
`dd-octo-sts` installation token allows the generated commit to trigger the normal
PR checks. The bot's author email is listed in Renovate's `gitIgnoredAuthors` so
generated commits do not stop Renovate from maintaining the branch. A Renovate
rebase can discard these commits; the next synchronization regenerates them.

## Reviewing an update

- Review the upstream release notes and rendered baseline changes. An Agent
  version bump may activate version-gated chart behavior; do not assume all
  baseline changes are image substitutions.
- Confirm the generation workflow and all normal chart checks pass on the final
  commit, including GKE-related checks where applicable.
- Keep feature minimum-version guards and compatibility-test pins unchanged
  unless a separate, intentional change is necessary. The custom manager does
  not globally replace the old version string.
- Approve and merge normally. There is no automatic approval or merge step.

## Rollout and troubleshooting

The workflow and STS policy must be merged to `main` before they can process a
Renovate PR. A local configuration validation does not verify the live Mend
scheduler, token exchange, or write permissions. Check the first generated PR
end to end after rollout.

Renovate's default concurrent-PR limit is ten. An existing backlog can keep an
eligible Agent update queued. Check the [Dependency Dashboard](https://github.com/DataDog/helm-charts/issues/2544)
for rate-limited updates; clear capacity or agree on an explicit PR-limit policy
before relying on automatic creation after the one-day release delay.

If the Renovate PR is behind `main`, request a Renovate rebase; the package rule
also configures automatic rebasing when behind the base branch. Do not merge a
PR with failed generation checks.

If a release or image is unavailable, or generation fails transiently, rerun the
failed workflow once the problem is resolved. If generation detects unexpected
source edits or new version-reference layouts, update the extraction rules and
tests in a normal PR instead of widening the privileged workflow's allowlist.

When taking over a PR manually, prefer a separate branch rather than editing
this bot-owned branch. To pause automatic Agent updates, set `enabled: false` on
the Agent-specific package rule in `renovate.json`; GitHub Actions and ADMS updates
are independent.

## Validation

```sh
make unit-test-ci-scripts
npx --yes --package renovate@44.129.0 renovate-config-validator --strict renovate.json
go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12 \
  .github/workflows/agent-release.yaml \
  .github/workflows/chart-version.yaml \
  .github/workflows/test-ci-scripts.yaml
```

Unit tests use the real chart and Renovate regex patterns, mocked GitHub APIs,
and temporary files. They do not open PRs, publish commits, or invoke Renovate
against GitHub. Generation can also be smoke-tested in a disposable checkout by
applying a stable release through the version helper, regenerating baselines,
and running the normal Datadog chart tests.
