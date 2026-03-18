// Shared semver helpers for bump-chart-version.js and validate-chart-version.js.

// Parse a semver string (e.g. "1.2.3" or "1.2.3-dev.4") into its components.
// Supports: major.minor.patch[-prerelease]
// Pre-release can contain alphanumeric characters and dots.
// Note: build metadata (e.g. +build.123) is not supported.
function parseVersion(versionStr) {
  const match = versionStr.match(/^(\d+)\.(\d+)\.(\d+)(?:-([a-zA-Z0-9.]+))?$/);
  if (!match) {
    throw new Error(`Invalid version format: ${versionStr}. Expected format: major.minor.patch[-prerelease]`);
  }
  return {
    major: parseInt(match[1], 10),
    minor: parseInt(match[2], 10),
    patch: parseInt(match[3], 10),
    prerelease: match[4] || null
  };
}

// Produce a semver string from its components.
function makeVersion({ major, minor, patch, prerelease }) {
  let version = `${major}.${minor}.${patch}`;
  if (prerelease) {
    version += `-${prerelease}`;
  }
  return version;
}

// Compute the bumped version for a given bump type, mirroring the logic in
// bump-chart-version.js. Throws if the pre-release format is unsupported.
//
// bumpType: 'patch-version' | 'minor-version' | 'no-version-bump'
function computeBumpedVersion(baseParsed, bumpType) {
  const desired = { ...baseParsed };

  if (bumpType === 'no-version-bump') {
    // No change.
  } else if (baseParsed.prerelease) {
    if (bumpType === 'minor-version') {
      // Promote pre-release to full release by dropping the suffix.
      desired.prerelease = null;
    } else if (bumpType === 'patch-version') {
      // Bump the pre-release number. Expected format: <letters>.<number> (e.g. dev.2, alpha.1).
      const m = baseParsed.prerelease.match(/^([a-zA-Z]+)\.(\d+)$/);
      if (!m) {
        throw new Error(
          `Pre-release format '${baseParsed.prerelease}' is not supported for patch bumping. ` +
          `Expected format: '<letters>.<number>' (e.g. 'dev.2', 'alpha.1'). ` +
          `Unsupported formats include: 'rc1', 'dev-2', 'beta.1.2'.`
        );
      }
      desired.prerelease = `${m[1]}.${parseInt(m[2], 10) + 1}`;
    }
  } else {
    if (bumpType === 'patch-version') {
      desired.patch += 1;
    } else if (bumpType === 'minor-version') {
      desired.minor += 1;
      desired.patch = 0;
    }
  }

  return makeVersion(desired);
}

module.exports = { parseVersion, makeVersion, computeBumpedVersion };
