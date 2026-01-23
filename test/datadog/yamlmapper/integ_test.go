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
	"github.com/stretchr/testify/require"
)

// TestMain runs before tests and handles cleanup of stale resources from previous interrupted test runs.
func TestMain(m *testing.M) {
	flag.Parse()

	cleanupStaleResources()

	os.Exit(m.Run())
}

// ### Main Test Functions

// TestBaseValues runs tests for simple values files that don't require any pre-created resources.
func TestBaseValues(t *testing.T) {
	// validate environment prerequisites
	validateEnv(t)
	// validate base values coverage
	assertBaseValuesCoverage(t)

	for _, tc := range baseTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			if tc.SkipReason != "" {
				t.Skipf("Skipping %s: %s", tc.Name, tc.SkipReason)
			}
			runValuesToDDAMappingTest(t, tc.ValuesFile, tc.ExpectedPods, tc.ExpectedComponentContainers, nil)
		})
	}
}

// TestValuesWithDependencies runs tests for values files that depend on extra k8s resources (ConfigMaps, Secrets, etc.).
// These resources are created before the helm install and cleaned up after the test.
func TestValuesWithDependencies(t *testing.T) {
	// validate environment prerequisites
	validateEnv(t)

	for _, tc := range testCasesWithDependencies {
		t.Run(tc.Name, func(t *testing.T) {
			if tc.SkipReason != "" {
				t.Skipf("Skipping %s: %s", tc.Name, tc.SkipReason)
			}
			runValuesToDDAMappingTest(t, tc.ValuesFile, tc.ExpectedPods, tc.ExpectedComponentContainers, &tc)
		})
	}
}

func runValuesToDDAMappingTest(t *testing.T, valuesPath string, expectedPods ExpectedComponentPods, expectedContainers ExpectedComponentContainers, tc *ResourceDependentTestCase) {
	// test setup
	ctx := newTestContext(t)
	ctx.SetupCleanup()
	ctx.SetupSecretsFromEnv()
	ctx.SetupTestResources(tc)

	// validate expected pods match test values file
	validateExpectedPodsAgainstValues(t, valuesPath, expectedPods)

	// install Helm chart and verify valid deployment
	helmAgentConf := installAndVerifyHelmAgent(t, ctx, valuesPath, expectedPods, expectedContainers)

	// run mapper against values.yaml to generate DDA
	ddaFilePath, err := runMapper(t, valuesPath)
	require.NoError(t, err, fmt.Sprintf("Mapper returned error: %s", err))

	// log the mapped DDA file contents
	ddaFileContents, err := os.ReadFile(ddaFilePath)
	if err != nil {
		t.Fatalf("Failed to read DDA file: %v", err)
	}
	t.Logf("Mapped DDA:\n%s", string(ddaFileContents))

	// uninstall Datadog chart
	ctx.TestCleanupRegistry.UninstallDatadog()
	err = waitForPodsTerminated(t, ctx.KubectlOptions, "app.kubernetes.io/managed-by=Helm", defaultHelmTimeout)
	if err != nil {
		t.Logf("Warning: %v", err)
	}
	
	// small delay to let containerd stabilize after pod termination
	interTestDelay(t, defaultContainerdDelay)

	// install Operator, deploy DDA, and verify that Operator-managed deployment matches Helm deployment
	installAndVerifyOperatorAgent(t, ctx, ctx.KubectlOptions, ddaFilePath, helmAgentConf, expectedPods, expectedContainers)
}

// installAndVerifyHelmAgent installs the Helm chart, verifies pods are healthy, and captures the agent config.
func installAndVerifyHelmAgent(t *testing.T, ctx *TestContext, valuesPath string, expectedPods ExpectedComponentPods, expectedContainers ExpectedComponentContainers) string {
	agentInstallCmd := common.HelmCommand{
		ReleaseName: releaseDatadog,
		ChartPath:   datadogChartPath,
		Values:      []string{valuesPath},
		Overrides: map[string]string{
			"datadog.operator.enabled": "false",
		},
	}

	cleanUpDatadog := common.InstallChart(t, ctx.KubectlOptions, agentInstallCmd)
	ctx.TestCleanupRegistry.SetDatadog(cleanUpDatadog)

	// verify Helm installation to ensure values.yaml is valid
	assertExpectedPodHealth(t, ctx.KubectlOptions, expectedPods, helmPodSelectors())
	verifyContainers(t, ctx.KubectlOptions, expectedContainers, helmPodSelectors())

	// capture agent config for later comparison
	helmAgentConf := getHelmAgentConf(t, ctx.KubectlOptions)
	require.NotEmpty(t, helmAgentConf, "Failed to get agent conf from helm-installed agent")

	return helmAgentConf
}

// installAndVerifyOperatorAgent installs the Operator, applies the DDA, and verifies that the Operator-managed deployment matches the original Helm deployment.
func installAndVerifyOperatorAgent(t *testing.T, ctx *TestContext, kubectlOptions *k8s.KubectlOptions, ddaFilePath string, helmAgentConf string, expectedPods ExpectedComponentPods, expectedContainers ExpectedComponentContainers) {
	err := installOperator(t, ctx.KubectlOptions, ctx.Namespace, ctx.TestCleanupRegistry)
	require.NoError(t, err, "Failed to install operator")
	
	err = applyDDAAndWaitForAgents(t, ctx.KubectlOptions, ddaFilePath)
	require.NoError(t, err, "Failed to apply DDA and wait for operator-managed agents")

	operatorAgentConf := getOperatorAgentConf(t, kubectlOptions)
	require.NotEmpty(t, operatorAgentConf, "Failed to get agent conf from operator-installed agent")

	verifyAgentConfEqual(t, helmAgentConf, operatorAgentConf)
	assertExpectedPodHealth(t, kubectlOptions, expectedPods, operatorPodSelectors())
	verifyContainers(t, kubectlOptions, expectedContainers, operatorPodSelectors())
}
