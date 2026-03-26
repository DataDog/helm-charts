// Validates that chart version, CHANGELOG, and README badge are consistent whenever
// chart-relevant files are modified on a PR. Mirrors the path triggers from ci.yaml
// to determine which charts need validation.
//
// Designed to run AFTER bump-chart-version.js (via `needs: [bump-chart-version]`).
// All GitHub API calls for PR content use `pr.head.ref` (branch name, not SHA) so
// that any commit pushed by the bump job is visible to this validation.

// These patterns mirror the `paths:` trigger in ci.yaml exactly.
const CI_PATH_PATTERNS = [
  /^charts\/([^/]+)\/Chart\.[^/]+$/,        // Chart.yaml, Chart.lock
  /^charts\/([^/]+)\/requirements\.[^/]+$/, // requirements.yaml, requirements.lock
  /^charts\/([^/]+)\/values\.[^/]+$/,        // values.yaml, values.schema.json
  /^charts\/([^/]+)\/templates\/.+$/,        // templates/**
];

module.exports = async ({github, context, core}) => {
  const { parseVersion, makeVersion, computeBumpedVersion, isSequentialBump, computeNextPrerelease, decodeFileContent } = require('./chart-version-utils');

  const pr = context.payload.pull_request;
  if (!pr) {
    core.setFailed("No pull request found in context payload.");
    return;
  }

  const { owner, repo } = context.repo;

  // Fetch labels and changed files in parallel — both are needed before we can proceed.
  const [{ data: labelData }, files] = await Promise.all([
    github.rest.issues.listLabelsOnIssue({ owner, repo, issue_number: pr.number }),
    github.paginate(github.rest.pulls.listFiles, { owner, repo, pull_number: pr.number })
  ]);
  // Use a Set for O(1) label lookups in the per-chart loop.
  const labelNames = new Set(labelData.map(l => l.name));

  // Determine which charts have ci.yaml-relevant file changes.
  const changedCharts = new Set();
  for (const file of files) {
    for (const pattern of CI_PATH_PATTERNS) {
      const match = file.filename.match(pattern);
      if (match) {
        changedCharts.add(match[1]);
        break;
      }
    }
  }

  if (changedCharts.size === 0) {
    core.info("No chart-relevant files changed. Skipping validation.");
    return;
  }

  core.info(`Charts with relevant changes: ${[...changedCharts].join(', ')}`);

  // Get the merge base SHA for fetching "before" versions.
  const comparison = await github.rest.repos.compareCommits({
    owner,
    repo,
    base: pr.base.ref,
    head: pr.head.ref
  });
  const mergeBaseSHA = comparison.data.merge_base_commit.sha;

  const errors = [];

  for (const chartName of changedCharts) {
    core.info(`\nValidating '${chartName}'...`);

    // Charts with no-version-bump label are intentionally skipping a version bump
    // (e.g. docs-only changes). The bump job handles reverting any CHANGELOG drift.
    if (labelNames.has(`${chartName}/no-version-bump`)) {
      core.info(`Skipping '${chartName}': no-version-bump label present.`);
      continue;
    }

    // --- Fetch base Chart.yaml ---
    let baseVersion;
    try {
      const file = await github.rest.repos.getContent({
        owner,
        repo,
        path: `charts/${chartName}/Chart.yaml`,
        ref: mergeBaseSHA
      });
      const content = decodeFileContent(file.data);
      const m = content.match(/^version:\s+(\S+)/m);
      if (!m) {
        core.warning(`No 'version:' found in base Chart.yaml for '${chartName}'. Skipping.`);
        continue;
      }
      baseVersion = m[1].trim();
    } catch (e) {
      if (e.status === 404) {
        // New chart with no prior Chart.yaml — nothing to compare against.
        core.info(`'${chartName}' appears to be a new chart (no base Chart.yaml). Skipping version check.`);
        continue;
      }
      errors.push(`'${chartName}': failed to fetch base Chart.yaml: ${e.message}`);
      continue;
    }

    // --- Fetch PR Chart.yaml ---
    // Uses pr.head.ref (branch name) so we always see the latest commit on the branch,
    // including any fixup commit pushed by the bump job.
    let prVersion;
    try {
      const file = await github.rest.repos.getContent({
        owner,
        repo,
        path: `charts/${chartName}/Chart.yaml`,
        ref: pr.head.ref
      });
      const content = decodeFileContent(file.data);
      const m = content.match(/^version:\s+(\S+)/m);
      if (!m) {
        errors.push(`'${chartName}': no 'version:' found in PR Chart.yaml.`);
        continue;
      }
      prVersion = m[1].trim();
    } catch (e) {
      errors.push(`'${chartName}': failed to fetch PR Chart.yaml: ${e.message}`);
      continue;
    }

    core.info(`  Base version: ${baseVersion}`);
    core.info(`  PR version:   ${prVersion}`);

    // --- Check A: version must be bumped ---
    if (prVersion === baseVersion) {
      errors.push(
        `'${chartName}': version was not bumped (still at ${baseVersion}). ` +
        `When chart-relevant files are modified, Chart.yaml version, CHANGELOG.md, and README.md ` +
        `must all be updated. Add a '${chartName}/patch-version' or '${chartName}/minor-version' ` +
        `label to have the version bumped automatically, or update these files manually.`
      );
      continue; // No point checking changelog/readme if version isn't bumped.
    }

    let baseParsed, prParsed;
    try {
      baseParsed = parseVersion(baseVersion);
      prParsed = parseVersion(prVersion);
    } catch (e) {
      errors.push(`'${chartName}': ${e.message}`);
      continue;
    }

    const hasPatchLabel = labelNames.has(`${chartName}/patch-version`);
    const hasMinorLabel = labelNames.has(`${chartName}/minor-version`);

    // --- Check B: version correctness ---
    if (hasPatchLabel || hasMinorLabel) {
      // With a version label, the bump job calculates the expected version.
      // Validate against the same logic here as a safety net.
      const bumpType = hasMinorLabel ? 'minor-version' : 'patch-version';
      let expected;
      try {
        expected = computeBumpedVersion(baseParsed, bumpType);
      } catch (e) {
        errors.push(`'${chartName}': cannot compute expected version: ${e.message}`);
        // Fall through — still check changelog and readme against prVersion.
      }
      if (expected !== undefined && prVersion !== expected) {
        errors.push(
          `'${chartName}': version ${prVersion} does not match expected ${expected} ` +
          `for ${bumpType} bump from ${baseVersion}.`
        );
      }
    } else {
      // No version label — check that the bump is exactly one sequential semver step.
      if (!isSequentialBump(baseParsed, prParsed)) {
        const patchExpected = makeVersion({ ...baseParsed, patch: baseParsed.patch + 1, prerelease: null });
        const minorExpected = makeVersion({ major: baseParsed.major, minor: baseParsed.minor + 1, patch: 0, prerelease: null });
        let hint;
        if (baseParsed.prerelease) {
          const finalRelease = makeVersion({ ...baseParsed, prerelease: null });
          const nextPrerelease = computeNextPrerelease(baseParsed);
          hint = nextPrerelease
            ? `Expected to finalize the pre-release (${finalRelease}) or bump the pre-release number (${nextPrerelease}).`
            : `Expected to finalize the pre-release (${finalRelease}).`;
        } else {
          hint = `Expected a patch bump (${patchExpected}) or a minor bump (${minorExpected}).`;
        }
        errors.push(
          `'${chartName}': version ${prVersion} is not a sequential semver bump from ${baseVersion}. ${hint}`
        );
      }
    }

    // --- Check C: CHANGELOG entry ---
    try {
      const file = await github.rest.repos.getContent({
        owner,
        repo,
        path: `charts/${chartName}/CHANGELOG.md`,
        ref: pr.head.ref
      });
      const content = decodeFileContent(file.data);
      if (!content.split('\n').some(line => line.trim() === `## ${prVersion}`)) {
        errors.push(`'${chartName}': CHANGELOG.md does not contain an entry for version ${prVersion}.`);
      }
    } catch (e) {
      errors.push(`'${chartName}': failed to fetch CHANGELOG.md: ${e.message}`);
    }

    // --- Check D: README version badge ---
    // helm-docs generates a badge in the form: ![Version: X.Y.Z](https://img.shields.io/badge/...)
    // We check for the markdown alt text which is unambiguous even for pre-release versions.
    try {
      const file = await github.rest.repos.getContent({
        owner,
        repo,
        path: `charts/${chartName}/README.md`,
        ref: pr.head.ref
      });
      const content = decodeFileContent(file.data);
      if (!content.includes(`![Version: ${prVersion}]`)) {
        errors.push(
          `'${chartName}': README.md version badge does not reflect version ${prVersion}. ` +
          `Run helm-docs to regenerate the README, or update the badge manually.`
        );
      }
    } catch (e) {
      errors.push(`'${chartName}': failed to fetch README.md: ${e.message}`);
    }
  }

  // Report all errors at once so the PR author sees every issue in one pass.
  if (errors.length > 0) {
    for (const err of errors) {
      core.error(err);
    }
    core.setFailed(`Chart version validation failed with ${errors.length} error(s). See errors above.`);
  } else {
    core.info("All chart version validations passed.");
  }
};
