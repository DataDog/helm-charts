// Unit tests for chart-version-utils.js
// Run with: node --test .github/scripts/chart-version-utils.test.js

const { test } = require('node:test');
const assert = require('node:assert/strict');
const {
  parseVersion,
  makeVersion,
  computeBumpedVersion,
  isSequentialBump,
  computeNextPrerelease,
} = require('./chart-version-utils');

// ---------------------------------------------------------------------------
// parseVersion
// ---------------------------------------------------------------------------

test('parseVersion: stable version', () => {
  assert.deepEqual(parseVersion('1.2.3'), { major: 1, minor: 2, patch: 3, prerelease: null, vPrefix: false });
});

test('parseVersion: pre-release version', () => {
  assert.deepEqual(parseVersion('1.2.3-dev.4'), { major: 1, minor: 2, patch: 3, prerelease: 'dev.4', vPrefix: false });
});

test('parseVersion: zeros', () => {
  assert.deepEqual(parseVersion('0.0.0'), { major: 0, minor: 0, patch: 0, prerelease: null, vPrefix: false });
});

test('parseVersion: alpha pre-release', () => {
  assert.deepEqual(parseVersion('3.187.0-alpha.1'), { major: 3, minor: 187, patch: 0, prerelease: 'alpha.1', vPrefix: false });
});

test('parseVersion: v-prefixed version', () => {
  assert.deepEqual(parseVersion('v0.3.2'), { major: 0, minor: 3, patch: 2, prerelease: null, vPrefix: true });
});

test('parseVersion: YAML-quoted version', () => {
  assert.deepEqual(parseVersion('"2.14.1"'), { major: 2, minor: 14, patch: 1, prerelease: null, vPrefix: false });
});

test('parseVersion: invalid - missing patch', () => {
  assert.throws(() => parseVersion('1.2'), /Invalid version format/);
});

test('parseVersion: invalid - four parts', () => {
  assert.throws(() => parseVersion('1.2.3.4'), /Invalid version format/);
});

test('parseVersion: invalid - non-numeric', () => {
  assert.throws(() => parseVersion('abc'), /Invalid version format/);
});

test('parseVersion: invalid - empty string', () => {
  assert.throws(() => parseVersion(''), /Invalid version format/);
});

// ---------------------------------------------------------------------------
// makeVersion
// ---------------------------------------------------------------------------

test('makeVersion: stable', () => {
  assert.equal(makeVersion({ major: 1, minor: 2, patch: 3, prerelease: null }), '1.2.3');
});

test('makeVersion: pre-release', () => {
  assert.equal(makeVersion({ major: 1, minor: 2, patch: 3, prerelease: 'dev.4' }), '1.2.3-dev.4');
});

test('makeVersion: zeros', () => {
  assert.equal(makeVersion({ major: 0, minor: 0, patch: 0, prerelease: null }), '0.0.0');
});

test('makeVersion: v-prefixed', () => {
  assert.equal(makeVersion({ major: 0, minor: 3, patch: 2, prerelease: null, vPrefix: true }), 'v0.3.2');
});

// ---------------------------------------------------------------------------
// computeBumpedVersion
// ---------------------------------------------------------------------------

test('computeBumpedVersion: stable + patch-version', () => {
  assert.equal(computeBumpedVersion(parseVersion('1.2.3'), 'patch-version'), '1.2.4');
});

test('computeBumpedVersion: stable + minor-version', () => {
  assert.equal(computeBumpedVersion(parseVersion('1.2.3'), 'minor-version'), '1.3.0');
});

test('computeBumpedVersion: stable + no-version-bump', () => {
  assert.equal(computeBumpedVersion(parseVersion('1.2.3'), 'no-version-bump'), '1.2.3');
});

test('computeBumpedVersion: pre-release + patch-version bumps pre-release number', () => {
  assert.equal(computeBumpedVersion(parseVersion('1.2.3-dev.4'), 'patch-version'), '1.2.3-dev.5');
});

test('computeBumpedVersion: pre-release + patch-version with number 0', () => {
  assert.equal(computeBumpedVersion(parseVersion('1.2.3-dev.0'), 'patch-version'), '1.2.3-dev.1');
});

test('computeBumpedVersion: pre-release + minor-version promotes to full release', () => {
  assert.equal(computeBumpedVersion(parseVersion('1.2.3-dev.4'), 'minor-version'), '1.2.3');
});

test('computeBumpedVersion: pre-release + no-version-bump unchanged', () => {
  assert.equal(computeBumpedVersion(parseVersion('1.2.3-dev.4'), 'no-version-bump'), '1.2.3-dev.4');
});

test('computeBumpedVersion: v-prefixed + patch-version preserves v prefix', () => {
  assert.equal(computeBumpedVersion(parseVersion('v0.3.2'), 'patch-version'), 'v0.3.3');
});

test('computeBumpedVersion: v-prefixed + minor-version preserves v prefix', () => {
  assert.equal(computeBumpedVersion(parseVersion('v0.3.2'), 'minor-version'), 'v0.4.0');
});

test('computeBumpedVersion: pre-release patch-version throws on unsupported format rc1', () => {
  assert.throws(() => computeBumpedVersion(parseVersion('1.2.3-rc1'), 'patch-version'), /not supported for patch bumping/);
});

test('computeBumpedVersion: pre-release patch-version throws on unsupported format dev-2', () => {
  assert.throws(() => computeBumpedVersion(parseVersion('1.2.3-dev-2'), 'patch-version'), /Invalid version format/);
});

// ---------------------------------------------------------------------------
// isSequentialBump — stable base
// ---------------------------------------------------------------------------

test('isSequentialBump: stable → patch (valid)', () => {
  assert.equal(isSequentialBump(parseVersion('1.2.3'), parseVersion('1.2.4')), true);
});

test('isSequentialBump: stable → minor (valid)', () => {
  assert.equal(isSequentialBump(parseVersion('1.2.3'), parseVersion('1.3.0')), true);
});

test('isSequentialBump: stable → patch into pre-release (valid)', () => {
  assert.equal(isSequentialBump(parseVersion('1.2.3'), parseVersion('1.2.4-dev.1')), true);
});

test('isSequentialBump: stable → minor into pre-release (valid)', () => {
  assert.equal(isSequentialBump(parseVersion('1.2.3'), parseVersion('1.3.0-dev.1')), true);
});

test('isSequentialBump: stable → skipped patch (invalid)', () => {
  assert.equal(isSequentialBump(parseVersion('1.2.3'), parseVersion('1.2.5')), false);
});

test('isSequentialBump: stable → minor with non-zero patch (invalid)', () => {
  assert.equal(isSequentialBump(parseVersion('1.2.3'), parseVersion('1.3.1')), false);
});

test('isSequentialBump: stable → major (invalid)', () => {
  assert.equal(isSequentialBump(parseVersion('1.2.3'), parseVersion('2.0.0')), false);
});

test('isSequentialBump: stable → regression (invalid)', () => {
  assert.equal(isSequentialBump(parseVersion('1.2.3'), parseVersion('1.2.2')), false);
});

// ---------------------------------------------------------------------------
// isSequentialBump — pre-release base
// ---------------------------------------------------------------------------

test('isSequentialBump: pre-release → finalize (valid)', () => {
  assert.equal(isSequentialBump(parseVersion('1.2.3-dev.4'), parseVersion('1.2.3')), true);
});

test('isSequentialBump: pre-release → bump pre-release number (valid)', () => {
  assert.equal(isSequentialBump(parseVersion('1.2.3-dev.4'), parseVersion('1.2.3-dev.5')), true);
});

test('isSequentialBump: pre-release → skipping finalization to bump patch (invalid)', () => {
  assert.equal(isSequentialBump(parseVersion('1.2.3-dev.4'), parseVersion('1.2.4')), false);
});

test('isSequentialBump: pre-release → skipping finalization to bump minor (invalid)', () => {
  assert.equal(isSequentialBump(parseVersion('1.2.3-dev.4'), parseVersion('1.3.0')), false);
});

test('isSequentialBump: pre-release → skipped pre-release number (invalid)', () => {
  assert.equal(isSequentialBump(parseVersion('1.2.3-dev.4'), parseVersion('1.2.3-dev.6')), false);
});

test('isSequentialBump: pre-release → pre-release regression (invalid)', () => {
  assert.equal(isSequentialBump(parseVersion('1.2.3-dev.4'), parseVersion('1.2.3-dev.3')), false);
});

test('isSequentialBump: pre-release → different prefix (invalid)', () => {
  assert.equal(isSequentialBump(parseVersion('1.2.3-dev.4'), parseVersion('1.2.3-alpha.5')), false);
});

test('isSequentialBump: pre-release → different version base (invalid)', () => {
  assert.equal(isSequentialBump(parseVersion('1.2.3-dev.4'), parseVersion('1.2.4-dev.5')), false);
});

// ---------------------------------------------------------------------------
// computeNextPrerelease
// ---------------------------------------------------------------------------

test('computeNextPrerelease: bumps pre-release number', () => {
  assert.equal(computeNextPrerelease(parseVersion('1.2.3-dev.4')), '1.2.3-dev.5');
});

test('computeNextPrerelease: works with alpha', () => {
  assert.equal(computeNextPrerelease(parseVersion('3.187.0-alpha.1')), '3.187.0-alpha.2');
});

test('computeNextPrerelease: returns null for stable version', () => {
  assert.equal(computeNextPrerelease(parseVersion('1.2.3')), null);
});

test('computeNextPrerelease: returns null for unsupported pre-release format rc1', () => {
  // rc1 doesn't match <letters>.<number>, so parse it manually
  assert.equal(computeNextPrerelease({ major: 1, minor: 2, patch: 3, prerelease: 'rc1' }), null);
});

test('computeNextPrerelease: returns null for unsupported format beta.1.2', () => {
  assert.equal(computeNextPrerelease({ major: 1, minor: 2, patch: 3, prerelease: 'beta.1.2' }), null);
});
