const { test } = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const renovate = require('../../renovate.json');
const {
  branch, sourcePaths, chartPath, changelogPath, readmePath,
  isAgentReleasePR, isAllowedPath, agentVersion, replaceVersions,
  validateVersionUpdate, blobSHA, prepare, publish,
} = require('./agent-release');

const repoRoot = path.resolve(__dirname, '../..');
const baseContents = sourcePaths.map(file => fs.readFileSync(path.join(repoRoot, file), 'utf8'));
const currentVersion = agentVersion(baseContents);
const nextVersion = `7.${Number(currentVersion.split('.')[1]) + 1}.0`;
const headContents = baseContents.map(content => replaceVersions(content, nextVersion));
const baselinePath = 'test/datadog/baseline/manifests/example.yaml';
const rule = renovate.packageRules.find(rule => rule.matchPackageNames?.includes('DataDog/datadog-agent'));

function pullRequest() {
  return {
    number: 42, state: 'open', user: { login: 'renovate[bot]' },
    head: { ref: branch, sha: 'head', repo: { full_name: 'DataDog/helm-charts' } },
    base: { ref: 'main', sha: 'base', repo: { full_name: 'DataDog/helm-charts' } },
  };
}

function fixture(t) {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), 'agent-release-test-'));
  t.after(() => fs.rmSync(root, { recursive: true, force: true }));
  const contents = {
    ...Object.fromEntries(sourcePaths.map((file, i) => [file, baseContents[i]])),
    [chartPath]: 'apiVersion: v1\nversion: 3.10.9\nappVersion: "7"\n',
    [changelogPath]: '# Changelog\n\n## 3.10.9\n\n* Previous change.\n',
    [readmePath]: '# Datadog\n',
    [baselinePath]: 'kind: Pod\n',
  };
  function checkoutBase() {
    for (const [file, content] of Object.entries(contents)) {
      fs.mkdirSync(path.dirname(path.join(root, file)), { recursive: true });
      fs.writeFileSync(path.join(root, file), content);
    }
  }
  checkoutBase();
  const state = {
    pr: pullRequest(), mergeBase: 'base',
    files: sourcePaths.map(filename => ({ filename, status: 'modified' })),
    head: { ...contents, ...Object.fromEntries(sourcePaths.map((file, i) => [file, headContents[i]])) },
    release: { tag_name: nextVersion, prerelease: false, draft: false },
    outputs: {}, mutations: [], reads: [], modes: {},
  };
  const github = {
    rest: {
      pulls: { get: async () => ({ data: state.pr }), listFiles: 'listFiles' },
      repos: {
        compareCommits: async () => ({ data: { merge_base_commit: { sha: state.mergeBase } } }),
        getContent: async ({ path, ref }) => {
          assert.equal(ref, state.pr.head.sha);
          state.reads.push({ path, ref });
          return { data: { type: 'file', encoding: 'base64', content: Buffer.from(state.head[path]).toString('base64') } };
        },
        getReleaseByTag: async args => {
          assert.deepEqual(args, { owner: 'DataDog', repo: 'datadog-agent', tag: nextVersion });
          if (state.releaseError) throw state.releaseError;
          return { data: state.release };
        },
      },
      git: {
        getCommit: async ({ commit_sha }) => {
          assert.equal(commit_sha, state.pr.head.sha);
          return { data: { tree: { sha: 'tree' } } };
        },
        getTree: async () => ({ data: {
          truncated: false,
          tree: Object.entries(state.head).map(([path, contents]) => ({ path, mode: state.modes[path] ?? '100644', sha: blobSHA(contents) })),
        } }),
      },
    },
    paginate: async () => state.files,
    graphql: async (query, variables) => {
      assert.match(query, /createCommitOnBranch/);
      state.mutations.push(variables.input);
    },
  };
  return {
    root, state, github, checkoutBase, baseSha: 'base', headSha: 'head',
    core: { setOutput: (key, value) => { state.outputs[key] = value; }, info: () => {} },
    context: { repo: { owner: 'DataDog', repo: 'helm-charts' }, payload: { pull_request: { number: 42 } } },
    git: (_command, args) => args[0] === 'diff' ? Object.keys(contents).join('\0') + '\0' : '',
  };
}

test('Renovate tracks exactly the three image tags and three fallbacks in the real chart', () => {
  assert.equal(validateVersionUpdate(baseContents, headContents), nextVersion);
  assert.equal(agentVersion(headContents), nextVersion);
  for (const [i, content] of baseContents.entries()) {
    assert.equal(replaceVersions(headContents[i], currentVersion), content);
    const oldLines = content.split('\n');
    const changedLines = headContents[i].split('\n').filter((line, n) => line !== oldLines[n]);
    assert.equal(changedLines.length, 3);
    assert.ok(changedLines.every(line => /tag: |\$version = /.test(line)));
  }
  assert.ok(headContents[1].includes('$version = "6.55.1"'));
  assert.deepEqual(
    headContents[1].split('\n').filter(line => /semverCompare/.test(line)),
    baseContents[1].split('\n').filter(line => /semverCompare/.test(line)),
  );
});

test('Renovate keeps GitHub Actions enabled, uses stable releases, and requires manual merging', () => {
  assert.deepEqual(renovate.enabledManagers, ['github-actions', 'custom.regex']);
  assert.equal(rule.automerge, false);
  assert.equal(rule.ignoreUnstable, true);
  assert.equal(rule.groupSlug, 'datadog-agent');
  assert.equal(`renovate/${rule.branchTopic}`, branch);
  assert.equal(rule.rebaseWhen, 'behind-base-branch');
  assert.equal(renovate.minimumReleaseAge, '7 days');
  assert.equal(rule.minimumReleaseAge, '1 day');
  const actionsRule = renovate.packageRules.find(rule => rule.matchManagers?.includes('github-actions'));
  assert.equal(actionsRule.minimumReleaseAge ?? renovate.minimumReleaseAge, '7 days');
  const allowed = new RegExp(rule.allowedVersions.slice(1, -1));
  for (const version of ['7.84.0', '7.84.1', '7.100.0']) assert.ok(allowed.test(version));
  for (const version of ['6.55.1', '8.0.0', '7.85.0-rc.1', '7.84.0-jmx', 'latest']) assert.ok(!allowed.test(version));
});

test('patch releases update all six references without requiring a minor version change', () => {
  const base = baseContents.map(content => replaceVersions(content, '7.84.0'));
  const head = base.map(content => replaceVersions(content, '7.84.1'));
  assert.equal(validateVersionUpdate(base, head), '7.84.1');
  assert.equal(agentVersion(head), '7.84.1');
});

test('version comparison uses numeric semver ordering', () => {
  const base = baseContents.map(content => replaceVersions(content, '7.99.9'));
  const head = base.map(content => replaceVersions(content, '7.100.0'));
  assert.equal(validateVersionUpdate(base, head), '7.100.0');
});

test('version updates reject no-op, downgrade, mismatched, missing, and prerelease references', () => {
  assert.throws(() => validateVersionUpdate(baseContents, baseContents), /Expected an upgrade/);
  assert.throws(() => validateVersionUpdate(headContents, baseContents), /Expected an upgrade/);
  assert.throws(() => validateVersionUpdate(baseContents, [headContents[0], baseContents[1]]), /must agree/);
  assert.throws(() => agentVersion([baseContents[0], '']), /exactly three/);
  const rc = baseContents.map(content => replaceVersions(content, `${nextVersion}-rc.1`));
  assert.throws(() => validateVersionUpdate(baseContents, rc), /exactly three/);
});

test('version updates reject injected YAML, template code, and changes to feature guards', () => {
  assert.throws(() => validateVersionUpdate(baseContents, [headContents[0] + '\nmalicious: true\n', headContents[1]]), /Unexpected changes/);
  assert.throws(() => validateVersionUpdate(baseContents, [headContents[0], headContents[1] + '\n{{ fail "injected" }}']), /Unexpected changes/);
  assert.throws(() => validateVersionUpdate(baseContents, [headContents[0], headContents[1].replace('semverCompare', 'injected')]), /Unexpected changes/);
});

test('only an open same-repository Renovate Agent PR is eligible', () => {
  assert.ok(isAgentReleasePR(pullRequest(), 'DataDog/helm-charts'));
  for (const mutate of [
    pr => { pr.user.login = 'someone'; },
    pr => { pr.state = 'closed'; },
    pr => { pr.head.repo.full_name = 'someone/helm-charts'; },
    pr => { pr.head.repo = null; },
    pr => { pr.base.ref = 'release'; },
    pr => { pr.head.ref = 'renovate/datadog-agent-malicious'; },
  ]) {
    const pr = pullRequest();
    mutate(pr);
    assert.ok(!isAgentReleasePR(pr, 'DataDog/helm-charts'));
  }
});

test('allowed paths exclude code, dependency locks, and other charts', () => {
  for (const file of [...sourcePaths, chartPath, changelogPath, readmePath, baselinePath]) assert.ok(isAllowedPath(file));
  for (const file of ['test/go.mod', '.github/workflows/agent-release.yaml', 'charts/datadog/requirements.lock', 'charts/datadog-operator/values.yaml', 'test/datadog/baseline/manifests/payload.sh']) assert.ok(!isAllowedPath(file));
});

test('prepare validates the release, snapshots the head, and builds from trusted base data', async t => {
  const f = fixture(t);
  f.state.head[chartPath] = 'untrusted dependencies and chart metadata';
  f.state.head[changelogPath] = 'untrusted changelog';
  f.state.files.push(...[chartPath, changelogPath].map(filename => ({ filename, status: 'modified' })));
  await prepare(f);
  assert.deepEqual(f.state.outputs, { 'head-sha': 'head', version: nextVersion });
  assert.ok(f.state.reads.every(read => read.ref === 'head'));
  assert.equal(fs.readFileSync(path.join(f.root, sourcePaths[0]), 'utf8'), headContents[0]);
  assert.equal(fs.readFileSync(path.join(f.root, chartPath), 'utf8'), 'apiVersion: v1\nversion: 3.10.10\nappVersion: "7"\n');
  const changelog = fs.readFileSync(path.join(f.root, changelogPath), 'utf8');
  assert.ok(changelog.startsWith('# Changelog\n\n## 3.10.10\n'));
  assert.ok(changelog.endsWith('## 3.10.9\n\n* Previous change.\n'));
  assert.ok(changelog.includes(nextVersion));
  assert.ok(changelog.includes('https://github.com/DataDog/helm-charts/pull/42'));
  assert.equal(f.state.mutations.length, 0);
});

for (const [name, mutate, message] of [
  ['stale base', f => { f.state.pr.base.sha = 'new-base'; }, /behind main/],
  ['branch behind main', f => { f.state.mergeBase = 'old-base'; }, /behind main/],
  ['unexpected file', f => { f.state.files.push({ filename: 'test/go.mod', status: 'modified' }); }, /Unexpected PR change/],
  ['deleted file', f => { f.state.files[0].status = 'removed'; }, /Unexpected PR change/],
  ['prerelease', f => { f.state.release.prerelease = true; }, /published stable/],
  ['unavailable release', f => { f.state.releaseError = Object.assign(new Error('Not Found'), { status: 404 }); }, /Not Found/],
  ['template injection', f => { f.state.head[sourcePaths[1]] += '\n{{ fail "injected" }}'; }, /Unexpected changes/],
]) {
  test(`prepare rejects ${name} without modifying the checkout`, async t => {
    const f = fixture(t);
    mutate(f);
    await assert.rejects(prepare(f), message);
    assert.equal(fs.readFileSync(path.join(f.root, sourcePaths[0]), 'utf8'), baseContents[0]);
    assert.equal(f.state.mutations.length, 0);
  });
}

test('publish creates a single optimistic-concurrency commit and is idempotent', async t => {
  const f = fixture(t);
  function generate() {
    fs.writeFileSync(path.join(f.root, readmePath), '# Regenerated docs\n');
    fs.writeFileSync(path.join(f.root, baselinePath), 'kind: Pod\n# Regenerated baseline\n');
  }
  await prepare(f);
  generate();
  await publish(f);
  assert.equal(f.state.mutations.length, 1);
  const mutation = f.state.mutations[0];
  assert.equal(mutation.expectedHeadOid, 'head');
  assert.equal(mutation.branch.branchName, branch);
  assert.deepEqual(mutation.fileChanges.additions.map(file => file.path).sort(), [chartPath, changelogPath, readmePath, baselinePath].sort());
  for (const file of mutation.fileChanges.additions) {
    f.state.head[file.path] = Buffer.from(file.contents, 'base64').toString();
    f.state.files.push({ filename: file.path, status: 'modified' });
  }
  f.state.pr.head.sha = 'generated-head';
  f.checkoutBase();
  await prepare(f);
  f.headSha = f.state.outputs['head-sha'];
  assert.equal(f.headSha, 'generated-head');
  generate();
  await publish(f);
  assert.equal(f.state.mutations.length, 1);
});

for (const [name, mutate, message] of [
  ['head moved', f => { f.state.pr.head.sha = 'new-head'; }, /PR changed/],
  ['base moved', f => { f.state.pr.base.sha = 'new-base'; }, /PR changed/],
  ['closed PR', f => { f.state.pr.state = 'closed'; }, /PR changed/],
  ['symlink', f => { f.state.modes[readmePath] = '120000'; }, /unexpected file/],
  ['unexpected local output', f => { f.git = () => 'test/go.mod\0'; }, /unexpected file/],
  ['untracked output', f => { f.git = (_command, args) => args[0] === 'ls-files' ? 'unexpected.txt\0' : ''; }, /unexpected file/],
]) {
  test(`publish refuses ${name}`, async t => {
    const f = fixture(t);
    await prepare(f);
    mutate(f);
    await assert.rejects(publish(f), message);
    assert.equal(f.state.mutations.length, 0);
  });
}

test('publish restores stale generated content even when it matches the local base', async t => {
  const f = fixture(t);
  await prepare(f);
  f.state.files.push({ filename: baselinePath, status: 'modified' });
  f.state.head[baselinePath] = 'stale baseline\n';
  f.git = () => '';
  await publish(f);
  const file = f.state.mutations[0].fileChanges.additions.find(file => file.path === baselinePath);
  assert.equal(Buffer.from(file.contents, 'base64').toString(), 'kind: Pod\n');
});
