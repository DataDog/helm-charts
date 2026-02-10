// Based on script from: https://github.com/DataDog/k8s-datadog-agent-ops/blob/main/.github/workflows/automatically-bump.yaml

module.exports = async ({github, context, core, exec}) => {
  const fs = require('fs');

  const pr = context.payload.pull_request;
  if (!pr) {
    core.setFailed("No pull request found in context payload.");
    return;
  }

  // Extract commonly used repo identifiers
  const { owner, repo } = context.repo;

  const { data: labels } = await github.rest.issues.listLabelsOnIssue({
    owner,
    repo,
    issue_number: pr.number
  });
  const labelNames = labels.map(l => l.name);

  // Gather all file changes and individual commit messages.
  const fileChanges = [];
  const commitMessages = [];

  // Helper to parse a semver string (e.g., "1.2.3" or "1.2.3-dev.4") into an object
  // Supports: major.minor.patch[-prerelease]
  // Pre-release can contain alphanumeric characters and dots
  // Note: Build metadata (e.g., +build.123) is not supported
  function parseVersion(versionStr) {
    // Match semver pattern with optional pre-release (e.g., -dev.4)
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

  // Helper to produce a semver string
  function makeVersion({ major, minor, patch, prerelease }) {
    let version = `${major}.${minor}.${patch}`;
    if (prerelease) {
      version += `-${prerelease}`;
    }
    return version;
  }

  // Get the list of charts that need a version bump (or changelog update)
  const chartsToBump = [];
  for (const label of labelNames) {
    const match = label.match(/^(?<chartName>[^/]+)\/(?<versionType>minor-version|patch-version|no-version-bump)$/);
    if (match) {
      chartsToBump.push({
        chartName: match.groups.chartName,
        bumpType: match.groups.versionType
      });
    }
  }
  
  if (chartsToBump.length === 0) {
    core.info("No charts to bump found in labels.");
    return;
  }

  // Compare the base and head branches to find their merge base
  const comparison = await github.rest.repos.compareCommits({
    owner,
    repo,
    base: pr.base.ref,
    head: pr.head.ref
  });

  // Use the merge_base_commit SHA
  const mergeBaseSHA = comparison.data.merge_base_commit.sha;

  for (const info of chartsToBump) {
    const { chartName, bumpType } = info;
    core.info(`Examining '${chartName}' for a ${bumpType} update…`);

    // Get the base Chart.yaml (from the PR base branch)
    let baseChartFile;
    try {
      baseChartFile = await github.rest.repos.getContent({
        owner,
        repo,
        path: `charts/${chartName}/Chart.yaml`,
        ref: mergeBaseSHA
      });
    } catch (error) {
      core.setFailed(`Could not get base Chart.yaml for ${chartName}: ${error.message}`);
      return;
    }
    const baseContent = Buffer.from(baseChartFile.data.content, baseChartFile.data.encoding).toString();
    const baseVersionMatch = baseContent.match(/^version:\s+(\S+)/m);
    if (!baseVersionMatch) {
      core.setFailed(`No 'version:' found in base branch Chart.yaml for ${chartName}. Skipping…`);
      return;
    }
    const baseVersion = baseVersionMatch[1].trim();
    core.info(`Base version for '${chartName}' is '${baseVersion}'.`);
    
    // Validate base version format
    let baseParsed;
    try {
      baseParsed = parseVersion(baseVersion);
    } catch (error) {
      core.setFailed(`Invalid base version format '${baseVersion}' for ${chartName}: ${error.message}`);
      return;
    }

    // Read the PR Chart.yaml on the PR head branch.
    let prChartFile;
    try {
      prChartFile = await github.rest.repos.getContent({
        owner,
        repo,
        path: `charts/${chartName}/Chart.yaml`,
        ref: pr.head.ref
      });
    } catch (error) {
      core.setFailed(`Could not get PR Chart.yaml for ${chartName}: ${error.message}`);
      return;
    }
    const prContent = Buffer.from(prChartFile.data.content, prChartFile.data.encoding).toString();
    const prVersionMatch = prContent.match(/^version:\s+(\S+)/m);
    if (!prVersionMatch) {
      core.setFailed(`No 'version:' found in PR Chart.yaml for ${chartName}. Skipping…`);
      return;
    }
    const prVersion = prVersionMatch[1].trim();
    core.info(`PR version for '${chartName}' is '${prVersion}'.`);
    
    // Validate PR version format
    try {
      parseVersion(prVersion);
    } catch (error) {
      core.setFailed(`Invalid PR version format '${prVersion}' for ${chartName}: ${error.message}`);
      return;
    }

    // Calculate the desired version based on bump type.
    let desiredParsed = { ...baseParsed };
    
    // Skip version bump calculation for no-version-bump label
    if (bumpType === 'no-version-bump') {
      core.info(`No version bump requested for '${chartName}' (no-version-bump label).`);
      // desiredParsed stays the same as baseParsed
    }
    // If the base version is a pre-release, handle based on bump type
    else if (baseParsed.prerelease) {
      core.info(`Base version '${baseVersion}' is a pre-release version.`);
      
      if (bumpType === 'minor-version') {
        // For minor-version, drop the pre-release suffix to promote to full release
        core.info(`Promoting pre-release to full release version (dropping pre-release suffix).`);
        desiredParsed.prerelease = null;
      } else if (bumpType === 'patch-version') {
        // For patch-version, bump the pre-release number
        // Expected format: <letters>.<number> (e.g., dev.2, alpha.1, beta.5)
        const prereleaseMatch = baseParsed.prerelease.match(/^([a-zA-Z]+)\.(\d+)$/);
        if (prereleaseMatch) {
          const prereleasePrefix = prereleaseMatch[1];
          const prereleaseNumber = parseInt(prereleaseMatch[2], 10);
          desiredParsed.prerelease = `${prereleasePrefix}.${prereleaseNumber + 1}`;
          core.info(`Bumping pre-release number from ${baseParsed.prerelease} to ${desiredParsed.prerelease}`);
        } else {
          core.setFailed(`Pre-release format '${baseParsed.prerelease}' for ${chartName} is not supported for patch bumping. Expected format: '<letters>.<number>' (e.g., 'dev.2', 'alpha.1'). Unsupported formats include: 'rc1', 'dev-2', 'beta.1.2'.`);
          return;
        }
      }
    } else {
      // Standard semver bump for non-pre-release versions
      if (bumpType === 'patch-version') {
        desiredParsed.patch += 1;
      } else if (bumpType === 'minor-version') {
        desiredParsed.minor += 1;
        desiredParsed.patch = 0;
      }
    }
    const desiredVersion = makeVersion(desiredParsed);
    core.info(`Desired version for '${chartName}' is '${desiredVersion}'.`);

    // If the Chart.yaml version is not what we expect, update it.
    if (prVersion !== desiredVersion) {
      core.info(`For '${chartName}', base was '${baseVersion}' but PR had '${prVersion}'. Changing version to '${desiredVersion}'.`);
      const newChartContent = prContent.replace(
        /^version:\s+\S+/m,
        `version: ${desiredVersion}`
      );

      // Replace file content locally so that helm-docs.sh script can properly update READMEs
      fs.writeFileSync(`charts/${chartName}/Chart.yaml`, newChartContent, 'utf8');

      fileChanges.push({
        path: `charts/${chartName}/Chart.yaml`,
        content: newChartContent
      });
      commitMessages.push(`bump version for ${chartName} to ${desiredVersion} (${bumpType})`);
    } else {
      core.info(`'${chartName}' version is already correct ('${prVersion}').`);
    }

    // Unless the bump type is no-version-bump, prepare CHANGELOG update.
    if (bumpType === 'no-version-bump') {
      core.info(`Skipping CHANGELOG update for '${chartName}' (no-version-bump).`);
      continue;
    }

    // Get base and head CHANGELOG.md files.
    let baseChangelog, prChangelog;
    try {
      baseChangelog = await github.rest.repos.getContent({
        owner,
        repo,
        path: `charts/${chartName}/CHANGELOG.md`,
        ref: mergeBaseSHA
      });
    } catch (error) {
      core.setFailed(`Could not get base CHANGELOG.md for ${chartName}: ${error.message}`);
      return;
    }
    try {
      prChangelog = await github.rest.repos.getContent({
        owner,
        repo,
        path: `charts/${chartName}/CHANGELOG.md`,
        ref: pr.head.ref
      });
    } catch (error) {
      core.setFailed(`Could not get head CHANGELOG.md for ${chartName}: ${error.message}`);
      return;
    }
  
    // Get the changelog content and check if it has already been modified in this branch.
    let newChangelogContent;
    const changelogContent = Buffer.from(prChangelog.data.content, prChangelog.data.encoding).toString();
    const lines = changelogContent.split('\n');
    const versionHeaderIdx = lines.findIndex(line => line.trim().startsWith('##'));
    
    if (prChangelog.data.sha !== baseChangelog.data.sha) {
      core.info(`CHANGELOG.md for '${chartName}' has already been modified in this branch. Updating the latest changelog entry version.`);
      // Update the version header to the desired version.
      if (versionHeaderIdx !== -1) {
        lines[versionHeaderIdx] = `## ${desiredVersion}`;
        newChangelogContent = lines.join('\n');
        
        // Check for diff between newChangelogContent and HEAD prChangelog content
        if (newChangelogContent !== changelogContent) {                  
          fileChanges.push({
            path: `charts/${chartName}/CHANGELOG.md`,
            content: newChangelogContent
          });
          commitMessages.push(`update changelog version for ${chartName} to ${desiredVersion}`);
          // Replace file content locally so that helm-docs.sh script can properly update READMEs
          fs.writeFileSync(`charts/${chartName}/CHANGELOG.md`, newChangelogContent, 'utf8');
        } else {
          core.info(`CHANGELOG.md for '${chartName}' is already updated with latest entry version.`)
        }
      }
      continue;
    }
    
    // If the changelog has not been modified, add a new entry.
    const prLink = `https://github.com/${owner}/${repo}/pull/${pr.number}`;
    const newEntry = `## ${desiredVersion}\n\n* ${pr.title} ([#${pr.number}](${prLink})).\n\n`;
    
    if (versionHeaderIdx !== -1) {
      const headerSection = lines.slice(0, versionHeaderIdx).join('\n').trimEnd();
      const remaining = lines.slice(versionHeaderIdx).join('\n');
      newChangelogContent = `${headerSection}\n\n${newEntry}${remaining}`;
    } else {
      newChangelogContent = newEntry + changelogContent;
    }

    // Also replace file content locally so that helm-docs.sh script can properly update READMEs
    fs.writeFileSync(`charts/${chartName}/CHANGELOG.md`, newChangelogContent, 'utf8');

    fileChanges.push({
      path: `charts/${chartName}/CHANGELOG.md`,
      content: newChangelogContent
    });
    commitMessages.push(`update changelog for ${chartName} with version ${desiredVersion}`);
    core.info(`CHANGELOG.md for '${chartName}' updated with a new entry for version ${desiredVersion}.`);
  }

  core.info("Done checking all labeled charts.");

  // Update README.mds using the .github/helm-docs.sh script
  try {
    await exec.exec('/bin/bash', ['.github/helm-docs.sh']);
  } catch (exitCode) {
    // Do nothing
    // Exit code 1 from the helm-docs.sh script when there is a diff is OK
  }

  // Get head README.md file
  for (const chart of chartsToBump) {
    let prReadme;
    try {
      prReadme = await github.rest.repos.getContent({
        owner,
        repo,
        path: `charts/${chart.chartName}/README.md`,
        ref: pr.head.ref
      });
    } catch (error) {
      core.setFailed(`Could not get head README.md for ${chart.chartName}: ${error.message}`);
      return;
    }
  
    // Compare local README.md file with HEAD to check if it has been modified by helm-docs.sh
    const localReadmeContent = fs.readFileSync(`charts/${chart.chartName}/README.md`, { encoding: 'utf-8' });
    if (localReadmeContent !== Buffer.from(prReadme.data.content, prReadme.data.encoding).toString()) {
      fileChanges.push({
        path: `charts/${chart.chartName}/README.md`,
        content: localReadmeContent
      });
      commitMessages.push(`update readme for ${chart.chartName}`);
    }
  }

  core.info("Done updating README.md files.");

  // If no file changes were collected, nothing to commit.
  if (fileChanges.length === 0) {
    core.info("No file changes to commit.");
    return;
  }

  // Get the current commit of the PR head branch.
  core.info("Getting current commit of the PR head branch…");
  let branchData;
  try {
    const response = await github.rest.repos.getBranch({
      owner,
      repo,
      branch: pr.head.ref
    });
    branchData = response.data;
  } catch (error) {
    core.setFailed(`Could not get branch data: ${error.message}`);
      return;
  }
  const baseCommitSha = branchData.commit.sha;
  const baseTreeSha = branchData.commit.commit.tree.sha;
  core.info(`Base commit for branch '${pr.head.ref}' is ${baseCommitSha}`);

  // Prepare tree entries from each file change.
  // (Mode "100644" means a normal non‐executable file.)
  const treeItems = fileChanges.map(change => ({
    path: change.path,
    mode: "100644",
    type: "blob",
    content: change.content
  }));

  // Create a new tree with these modifications.
  core.info("Creating new tree with modified files…");
  let newTreeResponse;
  try {
    const response = await github.rest.git.createTree({
      owner,
      repo,
      tree: treeItems,
      base_tree: baseTreeSha
    });
    newTreeResponse = response;
  } catch (error) {
    core.setFailed(`Could not create new tree: ${error.message}`);
    return;
  }
  const newTreeSha = newTreeResponse.data.sha;
  core.info(`Created new tree ${newTreeSha}`);

  // Create a combined commit message.
  const commitMessage = `chore: update charts\n\n` +
                        commitMessages.map(msg => `- ${msg}`).join('\n');

  // Create a new commit object.
  core.info("Creating new commit object…");
  let newCommitResponse;
  try {            
    const response = await github.rest.git.createCommit({
      owner,
      repo,
      message: commitMessage,
      tree: newTreeSha,
      parents: [baseCommitSha]
    });
    newCommitResponse = response;
  } catch (error) {
    core.setFailed(`Could not create new commit: ${error.message}`);
    return;
  }
  const newCommitSha = newCommitResponse.data.sha;
  core.info(`Created new commit ${newCommitSha}`);

  // Update the head branch reference to point to the new commit.
  core.info(`Updating branch reference to point to new commit ${newCommitSha}…`);
  try {
    await github.rest.git.updateRef({
      owner,
      repo,
      ref: `heads/${pr.head.ref}`,
      sha: newCommitSha
    });
    core.info(`Branch '${pr.head.ref}' has been updated with a combined commit.`);
  } catch (error) {
    core.setFailed(`Could not update branch reference: ${error.message}`);
    return;
  }
};

