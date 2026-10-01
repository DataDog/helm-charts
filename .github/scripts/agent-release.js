const fs = require('node:fs');
const path = require('node:path');
const { execFileSync } = require('node:child_process');
const { createHash } = require('node:crypto');
const { parseVersion, computeBumpedVersion, extractVersionFromChart } = require('./chart-version-utils');
const renovate = require('../../renovate.json');

const branch = 'renovate/datadog-agent';
const sourcePaths = ['charts/datadog/values.yaml', 'charts/datadog/templates/_helpers.tpl'];
const chartPath = 'charts/datadog/Chart.yaml';
const changelogPath = 'charts/datadog/CHANGELOG.md';
const readmePath = 'charts/datadog/README.md';
const chartPaths = [...sourcePaths, chartPath, changelogPath, readmePath];
const patterns = renovate.customManagers.find(manager => manager.depNameTemplate === 'DataDog/datadog-agent').matchStrings;

function isAgentReleasePR(pr, repository) {
  return pr.state === 'open' && pr.user.login === 'renovate[bot]' &&
    pr.head.repo?.full_name === repository && pr.base.repo.full_name === repository &&
    pr.base.ref === 'main' && pr.head.ref === branch;
}

function isAllowedPath(file) {
  return chartPaths.includes(file) ||
    (file.startsWith('test/datadog/baseline/manifests/') && file.endsWith('.yaml') &&
      path.posix.normalize(file) === file);
}

function versionsIn(content) {
  return patterns.flatMap(pattern => [...content.matchAll(new RegExp(pattern, 'g'))]
    .map(match => match.groups.currentValue));
}

function agentVersion(contents) {
  const versions = contents.flatMap(content => {
    const found = versionsIn(content);
    if (found.length !== 3) throw new Error('Expected exactly three Agent version references per source file');
    return found;
  });
  if (new Set(versions).size !== 1) throw new Error('Agent image tags and version fallbacks must agree');
  const version = versions[0];
  if (!/^7\.(0|[1-9]\d*)\.(0|[1-9]\d*)$/.test(version)) {
    throw new Error('Expected a stable Agent 7.x.y version');
  }
  return version;
}

function replaceVersions(content, version) {
  for (const pattern of patterns) {
    content = content.replace(new RegExp(pattern, 'g'), (...args) => {
      const match = args[0];
      const oldVersion = args.at(-1).currentValue;
      const offset = match.lastIndexOf(oldVersion);
      return match.slice(0, offset) + version + match.slice(offset + oldVersion.length);
    });
  }
  return content;
}

function validateVersionUpdate(baseContents, headContents) {
  const previous = agentVersion(baseContents);
  const version = agentVersion(headContents);
  const oldParts = previous.split('.').map(Number);
  const newParts = version.split('.').map(Number);
  const firstDifference = newParts.findIndex((part, i) => part !== oldParts[i]);
  if (firstDifference === -1 || newParts[firstDifference] < oldParts[firstDifference]) {
    throw new Error(`Expected an upgrade from ${previous}, got ${version}`);
  }
  baseContents.forEach((content, i) => {
    if (replaceVersions(content, version) !== headContents[i]) {
      throw new Error(`Unexpected changes in ${sourcePaths[i]}; only Agent version substitutions are allowed`);
    }
  });
  return version;
}

async function readContent(github, repo, file, ref) {
  const { data } = await github.rest.repos.getContent({ ...repo, path: file, ref });
  if (data.type !== 'file' || data.encoding !== 'base64') throw new Error(`Expected a regular file: ${file}`);
  return Buffer.from(data.content, 'base64').toString('utf8');
}

async function changedFiles(github, repo, number) {
  const files = await github.paginate(github.rest.pulls.listFiles, { ...repo, pull_number: number, per_page: 100 });
  for (const file of files) {
    if (file.status !== 'modified' || !isAllowedPath(file.filename)) {
      throw new Error(`Unexpected PR change: ${file.status} ${file.filename}`);
    }
  }
  return files.map(file => file.filename);
}

function releaseMetadata(chart, changelog, version, repository, number) {
  const chartVersion = computeBumpedVersion(parseVersion(extractVersionFromChart(chart)), 'patch-version');
  const heading = changelog.search(/^## /m);
  if (heading === -1) throw new Error('Cannot find the current chart changelog entry');
  const entry = `## ${chartVersion}\n\n* Bump default Agent, Cluster Agent, and Cluster Checks Runner to ${version} ([#${number}](https://github.com/${repository}/pull/${number})).\n\n`;
  return {
    chart: chart.replace(/^version:\s+\S+/m, `version: ${chartVersion}`),
    changelog: changelog.slice(0, heading) + entry + changelog.slice(heading),
  };
}

async function prepare({ github, context, core, baseSha, root = '.' }) {
  const repository = `${context.repo.owner}/${context.repo.repo}`;
  const { data: pr } = await github.rest.pulls.get({ ...context.repo, pull_number: context.payload.pull_request.number });
  if (!isAgentReleasePR(pr, repository)) throw new Error('Not a same-repository Renovate Agent release PR');
  const { data: comparison } = await github.rest.repos.compareCommits({ ...context.repo, base: baseSha, head: pr.head.sha });
  if (pr.base.sha !== baseSha || comparison.merge_base_commit.sha !== baseSha) {
    throw new Error('The Renovate branch is behind main. Request a Renovate rebase before regenerating files.');
  }
  await changedFiles(github, context.repo, pr.number);
  const baseContents = sourcePaths.map(file => fs.readFileSync(path.join(root, file), 'utf8'));
  const headContents = await Promise.all(sourcePaths.map(file => readContent(github, context.repo, file, pr.head.sha)));
  const version = validateVersionUpdate(baseContents, headContents);
  const { data: release } = await github.rest.repos.getReleaseByTag({ owner: 'DataDog', repo: 'datadog-agent', tag: version });
  if (release.draft || release.prerelease || release.tag_name !== version) {
    throw new Error(`Agent ${version} is not a published stable GitHub release`);
  }

  // Build only from the trusted base checkout; no PR templates, scripts, or dependencies are executed.
  sourcePaths.forEach((file, i) => fs.writeFileSync(path.join(root, file), replaceVersions(baseContents[i], version)));
  const metadata = releaseMetadata(
    fs.readFileSync(path.join(root, chartPath), 'utf8'),
    fs.readFileSync(path.join(root, changelogPath), 'utf8'),
    version, repository, pr.number,
  );
  fs.writeFileSync(path.join(root, chartPath), metadata.chart);
  fs.writeFileSync(path.join(root, changelogPath), metadata.changelog);
  core.setOutput('head-sha', pr.head.sha);
  core.setOutput('version', version);
}

function blobSHA(contents) {
  return createHash('sha1').update(`blob ${Buffer.byteLength(contents)}\0`).update(contents).digest('hex');
}

async function publish({ github, context, core, headSha, baseSha, root = '.', git = execFileSync }) {
  const repository = `${context.repo.owner}/${context.repo.repo}`;
  const { data: pr } = await github.rest.pulls.get({ ...context.repo, pull_number: context.payload.pull_request.number });
  if (!isAgentReleasePR(pr, repository) || pr.head.sha !== headSha || pr.base.sha !== baseSha) {
    throw new Error('PR changed during generation; rerun against the current head and base');
  }
  const prPaths = await changedFiles(github, context.repo, pr.number);
  const localPaths = [
    ['diff', '--name-only', '-z'],
    ['ls-files', '--others', '--exclude-standard', '-z'],
  ].flatMap(args => git('git', args, { cwd: root, encoding: 'utf8' }).split('\0').filter(Boolean));
  const { data: commit } = await github.rest.git.getCommit({ ...context.repo, commit_sha: headSha });
  const { data: tree } = await github.rest.git.getTree({ ...context.repo, tree_sha: commit.tree.sha, recursive: '1' });
  if (tree.truncated) throw new Error('Cannot validate a truncated PR tree');
  const entries = new Map(tree.tree.map(entry => [entry.path, entry]));
  const additions = [];
  for (const file of new Set([...prPaths, ...localPaths])) {
    if (!isAllowedPath(file) || entries.get(file)?.mode !== '100644') {
      throw new Error(`Refusing to publish unexpected file: ${file}`);
    }
    const contents = fs.readFileSync(path.join(root, file));
    if (blobSHA(contents) !== entries.get(file).sha) {
      additions.push({ path: file, contents: contents.toString('base64') });
    }
  }
  if (!additions.length) {
    core.info('Agent release generated files are already up to date');
    return;
  }

  // The expected head prevents overwriting a concurrent Renovate rebase or maintainer edit.
  await github.graphql(`mutation($input: CreateCommitOnBranchInput!) {
    createCommitOnBranch(input: $input) { commit { oid } }
  }`, {
    input: {
      branch: { repositoryNameWithOwner: repository, branchName: branch },
      expectedHeadOid: headSha,
      message: { headline: 'chore(datadog): regenerate Agent release files' },
      fileChanges: { additions },
    },
  });
  core.info(`Updated ${additions.length} generated files`);
}

module.exports = {
  branch, sourcePaths, chartPath, changelogPath, readmePath,
  isAgentReleasePR, isAllowedPath, agentVersion, replaceVersions,
  validateVersionUpdate, releaseMetadata, blobSHA, prepare, publish,
};
