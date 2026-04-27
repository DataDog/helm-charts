// Based on script from: https://github.com/DataDog/k8s-datadog-agent-ops/blob/main/.github/workflows/automatically-bump.yaml

module.exports = async ({github, context, core, exec}) => {
  const fs = require('fs');
  const { parseVersion, computeBumpedVersion, decodeFileContent, extractVersionFromChart } = require('./chart-version-utils');

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
    const baseContent = decodeFileContent(baseChartFile.data);
    const baseVersion = extractVersionFromChart(baseContent);
    if (!baseVersion) {
      core.setFailed(`No 'version:' found in base branch Chart.yaml for ${chartName}. Skipping…`);
      return;
    }
    core.info(`Base version for '${chartName}' is '${baseVersion}'.`);

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
    const prContent = decodeFileContent(prChartFile.data);
    const prVersion = extractVersionFromChart(prContent);
    if (!prVersion) {
      core.setFailed(`No 'version:' found in PR Chart.yaml for ${chartName}. Skipping…`);
      return;
    }
    core.info(`PR version for '${chartName}' is '${prVersion}'.`);

    // Calculate the desired version based on bump type.
    let desiredVersion;
    try {
      desiredVersion = computeBumpedVersion(baseParsed, bumpType);
    } catch (error) {
      core.setFailed(`Could not compute bumped version for ${chartName}: ${error.message}`);
      return;
    }
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

    if (bumpType === 'no-version-bump') {
      if (prChangelog.data.sha !== baseChangelog.data.sha) {
        core.info(`Reverting CHANGELOG.md for '${chartName}' to merge-base version (no-version-bump).`);
        const baseChangelogContent = decodeFileContent(baseChangelog.data);
        fileChanges.push({
          path: `charts/${chartName}/CHANGELOG.md`,
          content: baseChangelogContent
        });
        commitMessages.push(`revert changelog for ${chartName} (no-version-bump)`);
        fs.writeFileSync(`charts/${chartName}/CHANGELOG.md`, baseChangelogContent, 'utf8');
      } else {
        core.info(`CHANGELOG.md for '${chartName}' matches merge-base, nothing to do here (no-version-bump).`);
      }
      continue;
    }
  
    // Get the changelog content and check if it has already been modified in this branch.
    let newChangelogContent;
    const changelogContent = decodeFileContent(prChangelog.data);
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
    if (localReadmeContent !== decodeFileContent(prReadme.data)) {
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
    newTreeResponse = await github.rest.git.createTree({
      owner,
      repo,
      tree: treeItems,
      base_tree: baseTreeSha
    });
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

    newCommitResponse = await github.rest.git.createCommit({
      owner,
      repo,
      message: commitMessage,
      tree: newTreeSha,
      parents: [baseCommitSha]
    });
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
  }
};

