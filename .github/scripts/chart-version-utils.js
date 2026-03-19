// Shared semver helpers for bump-chart-version.js and validate-chart-version.js.

// Supported pre-release format: <letters>.<number> (e.g. dev.2, alpha.1).
const PRERELEASE_FORMAT = /^([a-zA-Z]+)\.(\d+)$/;

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

// Bump the pre-release number in a prerelease string (e.g. "dev.4" → "dev.5").
// Returns the bumped string, or null if the format is not <letters>.<number>.
function bumpPrereleaseNumber(prerelease) {
  const m = prerelease.match(PRERELEASE_FORMAT);
  if (!m) return null;
  return `${m[1]}.${parseInt(m[2], 10) + 1}`;
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
      const bumped = bumpPrereleaseNumber(baseParsed.prerelease);
      if (!bumped) {
        throw new Error(
          `Pre-release format '${baseParsed.prerelease}' is not supported for patch bumping. ` +
          `Expected format: '<letters>.<number>' (e.g. 'dev.2', 'alpha.1'). ` +
          `Unsupported formats include: 'rc1', 'dev-2', 'beta.1.2'.`
        );
      }
      desired.prerelease = bumped;
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

// Return the next pre-release version string (for use in error hints), or null
// if the format is not the expected <letters>.<number> pattern.
function computeNextPrerelease(base) {
  if (!base.prerelease) return null;
  const bumped = bumpPrereleaseNumber(base.prerelease);
  if (!bumped) return null;
  return makeVersion({ ...base, prerelease: bumped });
}

// Check whether a version bump is "sequential" (no skipped versions).
//
// Valid transitions from a STABLE base (no prerelease):
//   X.Y.Z   → X.Y.(Z+1)          patch bump (stable)
//   X.Y.Z   → X.(Y+1).0          minor bump (stable)
//   X.Y.Z   → X.Y.(Z+1)-pre.N    patch bump into a pre-release cycle
//   X.Y.Z   → X.(Y+1).0-pre.N    minor bump into a pre-release cycle
//
// Valid transitions from a PRE-RELEASE base:
//   X.Y.Z-pre.N → X.Y.Z              finalize the pre-release (drop suffix)
//   X.Y.Z-pre.N → X.Y.Z-pre.(N+1)   bump pre-release number by exactly 1 (same prefix)
//
// Everything else is considered non-sequential.
function isSequentialBump(base, pr) {
  if (!base.prerelease) {
    // Stable base: the non-prerelease part of PR must be a single patch or minor step.
    const isPatch = pr.major === base.major && pr.minor === base.minor && pr.patch === base.patch + 1;
    const isMinor = pr.major === base.major && pr.minor === base.minor + 1 && pr.patch === 0;
    return isPatch || isMinor;
  }

  // Pre-release base.
  if (!pr.prerelease) {
    // Only valid: finalize by dropping the suffix (X.Y.Z-pre.N → X.Y.Z).
    return pr.major === base.major && pr.minor === base.minor && pr.patch === base.patch;
  }

  // Both pre-release: version number must be identical, same prefix, number +1 only.
  if (pr.major !== base.major || pr.minor !== base.minor || pr.patch !== base.patch) {
    return false;
  }
  const bm = base.prerelease.match(PRERELEASE_FORMAT);
  const pm = pr.prerelease.match(PRERELEASE_FORMAT);
  if (!bm || !pm || bm[1] !== pm[1]) return false;
  return parseInt(pm[2], 10) === parseInt(bm[2], 10) + 1;
}

// Decode a file blob returned by the GitHub contents API.
function decodeFileContent(fileData) {
  return Buffer.from(fileData.content, fileData.encoding).toString();
}

module.exports = { parseVersion, makeVersion, computeBumpedVersion, isSequentialBump, computeNextPrerelease, decodeFileContent };
