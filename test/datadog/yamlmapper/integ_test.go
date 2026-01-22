// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package yamlmapper

import (
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/stretchr/testify/require"
)

// TestMain runs before all tests and handles pre-test cleanup of stale resources
func TestMain(m *testing.M) {
	flag.Parse()

	// Clean up any stale resources from previous interrupted test runs
	cleanupStaleResources()

	// Run tests
	os.Exit(m.Run())
}

// =============================================================================
// Main Test Functions
// =============================================================================

// TestBaseValues runs tests for simple values files that don't require any pre-created resources.
// These files are located in values/base/ directory.
func TestBaseValues(t *testing.T) {
	// Prerequisites
	validateEnv(t)

	assertBaseValuesCoverage(t)

	for _, tc := range baseTestCases {
		tc := tc // capture range variable
		t.Run(tc.Name, func(t *testing.T) {
			if tc.SkipReason != "" {
				t.Skipf("Skipping %s: %s", tc.Name, tc.SkipReason)
			}
			runWorkloadTest(t, tc.ValuesFile, tc.ExpectedPods, tc.ExpectedContainers, nil)
		})
	}
}

// TestValuesWithDependencies runs tests for values files that require pre-created resources.
// Resources (ConfigMaps, Secrets) are created before the helm install and cleaned up after.
func TestValuesWithDependencies(t *testing.T) {
	// Prerequisites
	validateEnv(t)

	for _, tc := range testCasesWithDependencies {
		tc := tc // capture range variable
		t.Run(tc.Name, func(t *testing.T) {
			if tc.SkipReason != "" {
				t.Skipf("Skipping %s: %s", tc.Name, tc.SkipReason)
			}
			runWorkloadTest(t, tc.ValuesFile, tc.ExpectedPods, tc.ExpectedContainers, &tc)
		})
	}
}

// =============================================================================
// Test Implementation
// =============================================================================

func runWorkloadTest(t *testing.T, valuesPath string, expectedPods ExpectedPodCounts, expectedContainers ExpectedContainers, tc *TestCaseWithDependencies) {
	// Create test context with unique namespace
	ctx := NewTestContext(t)
	ctx.SetupCleanup()
	ctx.SetupSecretsFromEnv()
	ctx.SetupTestResources(tc)

	// Validate expected pods match values file
	validateExpectedPodsAgainstValues(t, valuesPath, expectedPods)

	// Install Datadog chart WITHOUT operator (operator installed separately later)
	agentInstallCmd := common.HelmCommand{
		ReleaseName: releaseDatadog,
		ChartPath:   datadogChartPath,
		Values:      []string{valuesPath},
		Overrides: map[string]string{
			"datadog.operator.enabled": "false",
		},
	}
	if testing.Verbose() {
		agentInstallCmd.Logger = logger.Discard
	}

	cleanUpDatadog := common.InstallChart(t, ctx.KubectlOptions, agentInstallCmd)
	ctx.CleanupRegistry.SetDatadog(cleanUpDatadog)

	// Verify helm-managed pods health before applying DDA
	verifyPodsHealth(t, ctx.KubectlOptions, expectedPods, HelmPodSelectors())
	verifyContainers(t, ctx.KubectlOptions, expectedContainers, HelmPodSelectors())

	// Run workload verification
	verifyWorkload(t, ctx.KubectlOptions, valuesPath, ctx.Namespace, expectedPods, expectedContainers, ctx.CleanupRegistry)
}

func verifyWorkload(t *testing.T, kubectlOptions *k8s.KubectlOptions, valuesPath string, namespace string, expectedPods ExpectedPodCounts, expectedContainers ExpectedContainers, cleanup *CleanupRegistry) {
	// Run mapper against values.yaml
	ddaFilePath, err := runMapper(t, valuesPath, namespace, cleanup)
	require.NoError(t, err, fmt.Sprintf("Mapper returned error: %s", err))

	// Get agent conf from helm-installed agent
	helmAgentConf := getHelmAgentConf(t, kubectlOptions)
	require.NotEmpty(t, helmAgentConf, "Failed to get agent conf from helm-installed agent")

	// Uninstall datadog chart and wait for all pods to be fully terminated
	// This prevents containerd state corruption from rapid pod creation/deletion
	cleanup.UninstallDatadog()
	err = waitForPodsTerminated(t, kubectlOptions, "app.kubernetes.io/managed-by=Helm", defaultHelmTimeout)
	if err != nil {
		t.Logf("Warning: %v", err)
	}

	// Small delay to let containerd stabilize after pod termination
	interTestDelay(t, defaultContainerdDelay)

	err = installOperator(t, kubectlOptions, namespace, cleanup)
	require.NoError(t, err, "Failed to install operator")

	// Apply DDA and wait for operator-managed agents
	err = applyDDAAndWaitForAgents(t, kubectlOptions, ddaFilePath)
	require.NoError(t, err, "Failed to apply DDA and wait for operator-managed agents")

	// Verify agent conf matches
	verifyAgentConf(t, kubectlOptions, helmAgentConf)

	// Verify operator-managed pods health
	verifyPodsHealth(t, kubectlOptions, expectedPods, OperatorPodSelectors())
	verifyContainers(t, kubectlOptions, expectedContainers, OperatorPodSelectors())
}
