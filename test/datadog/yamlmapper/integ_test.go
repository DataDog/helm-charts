// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package yamlmapper

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/google/go-cmp/cmp"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// mapperPackage is the Go package path for the yaml-mapper command
const mapperPackage = "github.com/DataDog/datadog-operator/cmd/yaml-mapper"

// TestMain runs before all tests and handles pre-test cleanup of stale resources
func TestMain(m *testing.M) {
	flag.Parse()

	// Clean up any stale resources from previous interrupted test runs
	cleanupStaleResources()

	// Run tests
	os.Exit(m.Run())
}

const (
	ddaDestPath       = "tempDDADest.yaml"
	apiKeyEnv         = "API_KEY"
	appKeyEnv         = "APP_KEY"
	operatorChartPath = "../../../charts/datadog-operator"
)

// getMappingPath returns the absolute path to the mapping file.
// This ensures the mapper uses our custom mapping file instead of the embedded default.
func getMappingPath() string {
	// Get absolute path to ensure the mapper finds the correct file
	absPath, err := filepath.Abs("mapping_datadog_helm_to_datadogagent_crd.yaml")
	if err != nil {
		// Fall back to relative path if abs fails
		return "mapping_datadog_helm_to_datadogagent_crd.yaml"
	}
	return absPath
}

// Directory paths for values files
const (
	baseValuesDir = "values/base"
)

// ExpectedPodCounts defines explicit expected pod counts for each component.
// Agent pods are daemonset-based unless AgentDaemonset is false.
type ExpectedPodCounts struct {
	AgentDaemonset      bool
	AgentPods           int
	ClusterAgent        int
	ClusterChecksRunner int
}

// ExpectedContainers defines required container names per component.
// Empty slices mean "no container assertions for this component".
type ExpectedContainers struct {
	Agent               []string
	ClusterAgent        []string
	ClusterChecksRunner []string
}

type BaseTestCase struct {
	Name               string
	ValuesFile         string
	ExpectedPods       ExpectedPodCounts
	ExpectedContainers ExpectedContainers
	SkipReason         string
}

// =============================================================================
// Resource Definition Types
// =============================================================================

// ConfigMapDef defines a ConfigMap to be created before the test
type ConfigMapDef struct {
	Name string
	Data map[string]string
}

// SecretDef defines a Secret to be created before the test
type SecretDef struct {
	Name string
	Data map[string]string
}

// TestCaseWithDependencies defines a test case that requires pre-created resources
type TestCaseWithDependencies struct {
	// Name is used for the test name (usually the filename)
	Name string
	// ValuesFile is the path to the values file relative to the yamlmapper directory
	ValuesFile string
	// ExpectedPods defines explicit pod counts for each component
	ExpectedPods ExpectedPodCounts
	// ExpectedContainers defines required container names per component
	ExpectedContainers ExpectedContainers
	// ConfigMaps to create before the test
	ConfigMaps []ConfigMapDef
	// Secrets to create before the test
	Secrets []SecretDef
	// SkipReason if set, the test will be skipped with this reason
	SkipReason string
}

// =============================================================================
// Test Cases with Dependencies
// =============================================================================

func defaultExpectedPods() ExpectedPodCounts {
	return ExpectedPodCounts{
		AgentDaemonset:      true,
		ClusterAgent:        1,
		ClusterChecksRunner: 0,
	}
}

func defaultExpectedContainers() ExpectedContainers {
	return ExpectedContainers{}
}

// baseTestCases defines all base values test cases with explicit pod counts
var baseTestCases = []BaseTestCase{
	{Name: "combined-apm-logs-values.yaml", ValuesFile: baseValuesDir + "/combined-apm-logs-values.yaml", ExpectedPods: defaultExpectedPods(), ExpectedContainers: defaultExpectedContainers()},
	{Name: "combined-full-observability-values.yaml", ValuesFile: baseValuesDir + "/combined-full-observability-values.yaml", ExpectedPods: defaultExpectedPods(), ExpectedContainers: defaultExpectedContainers()},
	{Name: "default-minimal.yaml", ValuesFile: baseValuesDir + "/default-minimal.yaml", ExpectedPods: defaultExpectedPods(), ExpectedContainers: defaultExpectedContainers()},
	{Name: "feature-admission-controller-k8s-events-values.yaml", ValuesFile: baseValuesDir + "/feature-admission-controller-k8s-events-values.yaml", ExpectedPods: defaultExpectedPods(), ExpectedContainers: defaultExpectedContainers()},
	{Name: "feature-admission-controller-sidecar-values.yaml", ValuesFile: baseValuesDir + "/feature-admission-controller-sidecar-values.yaml", ExpectedPods: defaultExpectedPods(), ExpectedContainers: defaultExpectedContainers()},
	{Name: "feature-admission-controller-values.yaml", ValuesFile: baseValuesDir + "/feature-admission-controller-values.yaml", ExpectedPods: defaultExpectedPods(), ExpectedContainers: defaultExpectedContainers()},
	{Name: "feature-apm-instrumentation-targets-values.yaml", ValuesFile: baseValuesDir + "/feature-apm-instrumentation-targets-values.yaml", ExpectedPods: defaultExpectedPods(), ExpectedContainers: defaultExpectedContainers()},
	{Name: "feature-apm-instrumentation-values.yaml", ValuesFile: baseValuesDir + "/feature-apm-instrumentation-values.yaml", ExpectedPods: defaultExpectedPods(), ExpectedContainers: defaultExpectedContainers()},
	{Name: "feature-apm-values.yaml", ValuesFile: baseValuesDir + "/feature-apm-values.yaml", ExpectedPods: defaultExpectedPods(), ExpectedContainers: ExpectedContainers{
		Agent: []string{"trace-agent"},
	}},
	{Name: "feature-cluster-checks-values.yaml", ValuesFile: baseValuesDir + "/feature-cluster-checks-values.yaml", ExpectedPods: ExpectedPodCounts{
		AgentDaemonset:      true,
		ClusterAgent:        1,
		ClusterChecksRunner: 2,
	}, ExpectedContainers: ExpectedContainers{
		ClusterChecksRunner: []string{"agent"},
	}},
	{Name: "feature-dogstatsd-hostport-values.yaml", ValuesFile: baseValuesDir + "/feature-dogstatsd-hostport-values.yaml", ExpectedPods: defaultExpectedPods(), ExpectedContainers: defaultExpectedContainers()},
	{Name: "feature-dogstatsd-values.yaml", ValuesFile: baseValuesDir + "/feature-dogstatsd-values.yaml", ExpectedPods: defaultExpectedPods(), ExpectedContainers: defaultExpectedContainers()},
	{Name: "feature-event-collection-values.yaml", ValuesFile: baseValuesDir + "/feature-event-collection-values.yaml", ExpectedPods: defaultExpectedPods(), ExpectedContainers: defaultExpectedContainers()},
	{Name: "feature-helm-check-values.yaml", ValuesFile: baseValuesDir + "/feature-helm-check-values.yaml", ExpectedPods: defaultExpectedPods(), ExpectedContainers: defaultExpectedContainers()},
	{Name: "feature-ksm-core-values.yaml", ValuesFile: baseValuesDir + "/feature-ksm-core-values.yaml", ExpectedPods: defaultExpectedPods(), ExpectedContainers: defaultExpectedContainers()},
	{Name: "feature-logs-values.yaml", ValuesFile: baseValuesDir + "/feature-logs-values.yaml", ExpectedPods: defaultExpectedPods(), ExpectedContainers: defaultExpectedContainers()},
	{Name: "feature-npm-values.yaml", ValuesFile: baseValuesDir + "/feature-npm-values.yaml", ExpectedPods: defaultExpectedPods(),
		ExpectedContainers: defaultExpectedContainers(),
		SkipReason: "NPM requires kernel features not available in kind"},
	{Name: "feature-orchestrator-explorer-values.yaml", ValuesFile: baseValuesDir + "/feature-orchestrator-explorer-values.yaml", ExpectedPods: defaultExpectedPods(), ExpectedContainers: defaultExpectedContainers()},
	{Name: "feature-process-agent-values.yaml", ValuesFile: baseValuesDir + "/feature-process-agent-values.yaml", ExpectedPods: defaultExpectedPods(), ExpectedContainers: ExpectedContainers{
		Agent: []string{"agent", "trace-agent"},
	}},
	{Name: "feature-prometheus-scrape-values.yaml", ValuesFile: baseValuesDir + "/feature-prometheus-scrape-values.yaml", ExpectedPods: defaultExpectedPods(), ExpectedContainers: defaultExpectedContainers()},
	{Name: "feature-remote-config-values.yaml", ValuesFile: baseValuesDir + "/feature-remote-config-values.yaml", ExpectedPods: defaultExpectedPods(), ExpectedContainers: defaultExpectedContainers()},
	{Name: "feature-system-probe-checks-values.yaml", ValuesFile: baseValuesDir + "/feature-system-probe-checks-values.yaml", ExpectedPods: defaultExpectedPods(),
		ExpectedContainers: ExpectedContainers{
			Agent: []string{"system-probe"},
		},
		SkipReason: "System probe requires kernel features not available in kind"},
	{Name: "global-cluster-values.yaml", ValuesFile: baseValuesDir + "/global-cluster-values.yaml", ExpectedPods: defaultExpectedPods(), ExpectedContainers: defaultExpectedContainers()},
	{Name: "global-labels-annotations-tags-values.yaml", ValuesFile: baseValuesDir + "/global-labels-annotations-tags-values.yaml", ExpectedPods: defaultExpectedPods(), ExpectedContainers: defaultExpectedContainers()},
	{Name: "global-logging-values.yaml", ValuesFile: baseValuesDir + "/global-logging-values.yaml", ExpectedPods: defaultExpectedPods(), ExpectedContainers: defaultExpectedContainers()},
	{Name: "global-metadata-values.yaml", ValuesFile: baseValuesDir + "/global-metadata-values.yaml", ExpectedPods: defaultExpectedPods(), ExpectedContainers: defaultExpectedContainers()},
	{Name: "global-network-policy-values.yaml", ValuesFile: baseValuesDir + "/global-network-policy-values.yaml", ExpectedPods: defaultExpectedPods(), ExpectedContainers: defaultExpectedContainers()},
	{Name: "global-origin-detection-values.yaml", ValuesFile: baseValuesDir + "/global-origin-detection-values.yaml", ExpectedPods: defaultExpectedPods(), ExpectedContainers: defaultExpectedContainers()},
	{Name: "global-site-endpoint-values.yaml", ValuesFile: baseValuesDir + "/global-site-endpoint-values.yaml", ExpectedPods: defaultExpectedPods(), ExpectedContainers: defaultExpectedContainers()},
	{Name: "global-tags-extended-values.yaml", ValuesFile: baseValuesDir + "/global-tags-extended-values.yaml", ExpectedPods: defaultExpectedPods(), ExpectedContainers: defaultExpectedContainers()},
	{Name: "global-tags-values.yaml", ValuesFile: baseValuesDir + "/global-tags-values.yaml", ExpectedPods: defaultExpectedPods(), ExpectedContainers: defaultExpectedContainers()},
}
	
// testCasesWithDependencies defines all test cases that require pre-created resources
var testCasesWithDependencies = []TestCaseWithDependencies{
	{
		Name:       "global-credentials-existing-secret-values.yaml",
		ValuesFile: "values/global-credentials-existing-secret-values.yaml",
		ExpectedPods: defaultExpectedPods(),
		ExpectedContainers: defaultExpectedContainers(),
		Secrets: []SecretDef{
			{Name: "my-datadog-api-secret", Data: map[string]string{"api-key": "00000000000000000000000000000000"}},
			{Name: "my-datadog-app-secret", Data: map[string]string{"app-key": "0000000000000000000000000000000000000000"}},
			{Name: "my-cluster-agent-token-secret", Data: map[string]string{"token": "test-cluster-agent-token-value"}},
		},
	},
	{
		Name:       "override-cluster-agent-values.yaml",
		ValuesFile: "values/override-cluster-agent-values.yaml",
		ExpectedPods: defaultExpectedPods(),
		ExpectedContainers: defaultExpectedContainers(),
		ConfigMaps: []ConfigMapDef{
			{Name: "cluster-agent-config", Data: map[string]string{"DD_LOG_LEVEL": "debug"}},
		},
	},
	{
		Name:       "override-cluster-checks-runner-values.yaml",
		ValuesFile: "values/override-cluster-checks-runner-values.yaml",
		ExpectedPods: ExpectedPodCounts{
			AgentDaemonset:      true,
			ClusterAgent:        1,
			ClusterChecksRunner: 2,
		},
		ExpectedContainers: ExpectedContainers{
			ClusterChecksRunner: []string{"agent"},
		},
		ConfigMaps: []ConfigMapDef{
			{Name: "ccr-config", Data: map[string]string{"DD_LOG_LEVEL": "debug"}},
		},
	},
	{
		Name:       "override-node-agent-values.yaml",
		ValuesFile: "values/override-node-agent-values.yaml",
		ExpectedPods: defaultExpectedPods(),
		ExpectedContainers: defaultExpectedContainers(),
		// No dependencies for node agent - kept here for organization with other override tests
	},
	{
		Name:       "global-misc-values.yaml",
		ValuesFile: "values/global-misc-values.yaml",
		ExpectedPods: defaultExpectedPods(),
		ExpectedContainers: defaultExpectedContainers(),
		ConfigMaps: []ConfigMapDef{
			{Name: "datadog-env-config", Data: map[string]string{"DD_LOG_LEVEL": "debug", "DD_TAGS": "env:test"}},
		},
		Secrets: []SecretDef{
			{Name: "datadog-env-secrets", Data: map[string]string{"DD_SECRET_VAR": "secret-value"}},
		},
	},
}

// =============================================================================
// Test Functions
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
// Resource Creation Helpers
// =============================================================================

// createTestResources creates all required ConfigMaps and Secrets for a test case.
// Returns a cleanup function that removes all created resources.
func createTestResources(t *testing.T, kubectlOptions *k8s.KubectlOptions, tc *TestCaseWithDependencies) func() {
	if tc == nil {
		return func() {}
	}

	var cleanupFuncs []func()

	// Create ConfigMaps
	for _, cm := range tc.ConfigMaps {
		cleanup := common.CreateConfigMap(t, kubectlOptions, cm.Name, cm.Data)
		cleanupFuncs = append(cleanupFuncs, cleanup)
	}

	// Create Secrets
	for _, secret := range tc.Secrets {
		cleanup := common.CreateSecret(t, kubectlOptions, secret.Name, secret.Data)
		cleanupFuncs = append(cleanupFuncs, cleanup)
	}

	return func() {
		for _, cleanup := range cleanupFuncs {
			cleanup()
		}
	}
}

// =============================================================================
// Test Implementation
// =============================================================================

func runWorkloadTest(t *testing.T, valuesPath string, expectedPods ExpectedPodCounts, expectedContainers ExpectedContainers, tc *TestCaseWithDependencies) {
	cleanupRegistry := &CleanupRegistry{}

	namespaceName := fmt.Sprintf("%s%s", testNamespacePrefix, strings.ToLower(random.UniqueId()))
	kubectlOptions := k8s.NewKubectlOptions("", "", namespaceName)
	k8s.CreateNamespace(t, kubectlOptions, namespaceName)

	// Track cleanup functions for resources that should be cleaned up before namespace deletion
	var cleanupSecrets func()
	var cleanupTestResources func()

	// Register cleanup early so it runs even if the test fails mid-way
	t.Cleanup(func() {
		cleanupKubectlOptions := quietKubectlOptions(kubectlOptions)

		if testing.Verbose() {
			t.Logf("Starting cleanup for namespace %s", namespaceName)
		}

		// Step 1: Delete DDA and wait for finalizers to complete
		for _, ddaFile := range cleanupRegistry.GetFiles() {
			if testing.Verbose() {
				t.Logf("Deleting DDA from file %s", ddaFile)
			}
			_ = k8s.RunKubectlE(t, cleanupKubectlOptions, []string{"delete", "-f", ddaFile, "--ignore-not-found", "--wait=true", "--timeout=60s"}...)
			_ = os.Remove(ddaFile)
		}

		// Wait for all DatadogAgent resources in this namespace to be fully deleted
		waitForDDADeletion(t, cleanupKubectlOptions, namespaceName, 2*time.Minute)

		// Step 2: Uninstall operator chart
		cleanupRegistry.UninstallOperator()

		// Clean up datadog chart (if not already uninstalled)
		cleanupRegistry.UninstallDatadog()

		// Wait for all pods to terminate
		_ = waitForPodsTerminated(t, kubectlOptions, "", 2*time.Minute)

		// Clean up test resources (ConfigMaps, Secrets) explicitly
		// This ensures resources are deleted even if namespace deletion hangs
		if testing.Verbose() {
			t.Logf("Cleaning up test-specific resources in namespace %s", namespaceName)
		}
		if cleanupTestResources != nil {
			cleanupTestResources()
		}
		if cleanupSecrets != nil {
			cleanupSecrets()
		}

		// Step 3: Delete namespace and wait for it to be fully removed
		if testing.Verbose() {
			t.Logf("Deleting namespace %s", namespaceName)
		}
		k8s.DeleteNamespace(t, cleanupKubectlOptions, namespaceName)

		// Wait for namespace to be fully deleted before next test
		waitForNamespaceDeletion(t, namespaceName, 2*time.Minute)

		if testing.Verbose() {
			t.Logf("Cleanup complete for namespace %s", namespaceName)
		}
	})

	if os.Getenv(apiKeyEnv) != "" && os.Getenv(appKeyEnv) != "" {
		cleanupSecrets = common.CreateSecretFromEnv(t, kubectlOptions, apiKeyEnv, appKeyEnv)
	}

	// Create test-specific resources (ConfigMaps, Secrets) if defined
	cleanupTestResources = createTestResources(t, kubectlOptions, tc)

	// Install Datadog chart WITHOUT operator (operator installed separately later)
	agentInstallCmd := common.HelmCommand{
		ReleaseName: "datadog",
		ChartPath:   "../../../charts/datadog",
		Values:      []string{valuesPath},
		Overrides: map[string]string{
			"datadog.operator.enabled": "false",
		},
	}
	if testing.Verbose() {
		agentInstallCmd.Logger = logger.Discard
	}
	validateExpectedPodsAgainstValues(t, valuesPath, expectedPods)

	cleanUpDatadog := common.InstallChart(t, kubectlOptions, agentInstallCmd)
	cleanupRegistry.SetDatadog(cleanUpDatadog)

	// Verify helm-managed pods health before applying DDA
	verifyPodsHealth(t, kubectlOptions, expectedPods, HelmPodSelectors())
	verifyContainers(t, kubectlOptions, expectedContainers, HelmPodSelectors())

	// Run workload verification
	verifyWorkload(t, kubectlOptions, valuesPath, namespaceName, expectedPods, expectedContainers, cleanupRegistry)
}

func verifyWorkload(t *testing.T, kubectlOptions *k8s.KubectlOptions, valuesPath string, namespace string, expectedPods ExpectedPodCounts, expectedContainers ExpectedContainers, cleanup *CleanupRegistry) {
	// Run mapper against values.yaml
	ddaFilePath := runMapper(t, valuesPath, namespace, cleanup)

	// Get agent conf from helm-installed agent
	helmAgentConf := getHelmAgentConf(t, kubectlOptions)
	require.NotEmpty(t, helmAgentConf, "Failed to get agent conf from helm-installed agent")

	// Uninstall datadog chart and wait for all pods to be fully terminated
	// This prevents containerd state corruption from rapid pod creation/deletion
	cleanup.UninstallDatadog()
	err := waitForPodsTerminated(t, kubectlOptions, "app.kubernetes.io/managed-by=Helm", 3*time.Minute)
	if err != nil {
		t.Logf("Warning: %v", err)
	}

	// Small delay to let containerd stabilize after pod termination
	interTestDelay(t, 5*time.Second)

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

// installOperator installs the datadog-operator chart and waits for it to be ready
func installOperator(t *testing.T, kubectlOptions *k8s.KubectlOptions, namespace string, cleanup *CleanupRegistry) error {
	quietOptions := quietKubectlOptions(kubectlOptions)
	operatorInstallCmd := common.HelmCommand{
		ReleaseName: "datadog-operator",
		ChartPath:   operatorChartPath,
		Overrides: map[string]string{
			"installCRDs":        "false", // CRDs managed externally (CI or developer)
			"watchNamespaces[0]": namespace,
		},
	}
	if testing.Verbose() {
		operatorInstallCmd.Logger = logger.Discard
	}
	cleanUpOperator := common.InstallChart(t, quietOptions, operatorInstallCmd)
	cleanup.SetOperator(cleanUpOperator)

	operatorDeployments := k8s.ListDeployments(t, quietOptions, metav1.ListOptions{
		LabelSelector: "app.kubernetes.io/name=datadog-operator",
	})
	require.Len(t, operatorDeployments, 1, "Expected 1 deployment, got %d", len(operatorDeployments))

	// Wait for operator to be ready
	err := k8s.WaitUntilDeploymentAvailableE(t, quietOptions, operatorDeployments[0].Name, 10, 15*time.Second)
	if err != nil {
		t.Logf("Failed to wait for operator deployment: %v", err)
		return err
	}
	return nil
}

// applyDDAAndWaitForAgents applies the DDA manifest and waits for operator-managed agents to be running
func applyDDAAndWaitForAgents(t *testing.T, kubectlOptions *k8s.KubectlOptions, ddaFilePath string) error {
	quietOptions := quietKubectlOptions(kubectlOptions)
	err := k8s.RunKubectlE(t, quietOptions, []string{"apply", "-f", ddaFilePath}...)
	if err != nil {
		t.Logf("Failed to apply DDA: %v", err)
		return err
	}

	expectedPods := expectedDsCount(t, quietOptions)
	err = k8s.WaitUntilNumPodsCreatedE(t, quietOptions, metav1.ListOptions{
		LabelSelector: "agent.datadoghq.com/component=agent,app.kubernetes.io/managed-by=datadog-operator",
		FieldSelector: "status.phase=Running",
	}, expectedPods, 10, 15*time.Second)
	if err != nil {
		t.Logf("Failed to wait for operator agent pods: %v", err)
		return err
	}
	return nil
}

// verifyAgentConf compares helm agent config against operator agent config
func verifyAgentConf(t *testing.T, kubectlOptions *k8s.KubectlOptions, helmAgentConf string) {
	operatorAgentConf := getOperatorAgentConf(t, kubectlOptions)
	if operatorAgentConf == "" {
		if isAgentConfStrict() {
			require.NotEmpty(t, operatorAgentConf, "Strict mode: expected operator agent config for comparison")
		}
		t.Log("Warning: Could not retrieve operator agent config for comparison, skipping config verification")
		return
	}

	agentConfEqual := cmp.Equal(helmAgentConf, operatorAgentConf)
	if !agentConfEqual {
		t.Logf("Agent conf diff: %s", cmp.Diff(helmAgentConf, operatorAgentConf))
		if isAgentConfStrict() {
			require.True(t, agentConfEqual, "Strict mode: helm vs operator agent config mismatch")
		}
	}
}

// verifyPodsHealth asserts that expected pods are running and healthy (no restarts)
func verifyPodsHealth(t *testing.T, kubectlOptions *k8s.KubectlOptions, expected ExpectedPodCounts, selectors PodSelectors) {
	agentCount := expected.AgentPods
	if expected.AgentDaemonset {
		agentCount = expectedDsCount(t, kubectlOptions)
	}
	checkPodsHealth(t, kubectlOptions, selectors.Agent, agentCount)

	if expected.ClusterAgent > 0 {
		checkPodsHealth(t, kubectlOptions, selectors.ClusterAgent, expected.ClusterAgent)
	}
	if expected.ClusterChecksRunner > 0 {
		checkPodsHealth(t, kubectlOptions, selectors.ClusterChecksRunner, expected.ClusterChecksRunner)
	}
}

type valuesExpectations struct {
	ClusterAgent struct {
		Enabled  *bool `yaml:"enabled"`
		Replicas *int  `yaml:"replicas"`
	} `yaml:"clusterAgent"`
	ClusterChecksRunner struct {
		Enabled  *bool `yaml:"enabled"`
		Replicas *int  `yaml:"replicas"`
	} `yaml:"clusterChecksRunner"`
}

func validateExpectedPodsAgainstValues(t *testing.T, valuesPath string, expected ExpectedPodCounts) {
	data, err := os.ReadFile(valuesPath)
	require.NoError(t, err, "Failed to read values file: %s", valuesPath)

	var parsed valuesExpectations
	err = yaml.Unmarshal(data, &parsed)
	require.NoError(t, err, "Failed to parse values file: %s", valuesPath)

	if parsed.ClusterAgent.Enabled != nil {
		if *parsed.ClusterAgent.Enabled && expected.ClusterAgent == 0 {
			t.Fatalf("Values file enables clusterAgent but expected count is 0: %s", valuesPath)
		}
		if !*parsed.ClusterAgent.Enabled && expected.ClusterAgent > 0 {
			t.Fatalf("Values file disables clusterAgent but expected count is %d: %s", expected.ClusterAgent, valuesPath)
		}
	}
	if parsed.ClusterAgent.Replicas != nil {
		if expected.ClusterAgent != *parsed.ClusterAgent.Replicas {
			t.Fatalf("Values file sets clusterAgent.replicas=%d but expected count is %d: %s",
				*parsed.ClusterAgent.Replicas, expected.ClusterAgent, valuesPath)
		}
	}

	if parsed.ClusterChecksRunner.Enabled != nil {
		if *parsed.ClusterChecksRunner.Enabled && expected.ClusterChecksRunner == 0 {
			t.Fatalf("Values file enables clusterChecksRunner but expected count is 0: %s", valuesPath)
		}
		if !*parsed.ClusterChecksRunner.Enabled && expected.ClusterChecksRunner > 0 {
			t.Fatalf("Values file disables clusterChecksRunner but expected count is %d: %s", expected.ClusterChecksRunner, valuesPath)
		}
	}
	if parsed.ClusterChecksRunner.Replicas != nil {
		if expected.ClusterChecksRunner != *parsed.ClusterChecksRunner.Replicas {
			t.Fatalf("Values file sets clusterChecksRunner.replicas=%d but expected count is %d: %s",
				*parsed.ClusterChecksRunner.Replicas, expected.ClusterChecksRunner, valuesPath)
		}
	}
}

func assertBaseValuesCoverage(t *testing.T) {
	entries, err := os.ReadDir(baseValuesDir)
	require.NoError(t, err, "Failed to read base values directory")

	filesOnDisk := map[string]struct{}{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		filesOnDisk[entry.Name()] = struct{}{}
	}

	listed := map[string]int{}
	for _, tc := range baseTestCases {
		listed[tc.Name]++
	}

	var missing []string
	for name := range filesOnDisk {
		if listed[name] == 0 {
			missing = append(missing, name)
		}
	}

	var extra []string
	for name, count := range listed {
		if count > 1 {
			t.Fatalf("Duplicate base test case entry for %s", name)
		}
		if _, ok := filesOnDisk[name]; !ok {
			extra = append(extra, name)
		}
	}

	if len(missing) > 0 || len(extra) > 0 {
		t.Fatalf("Base values coverage mismatch: missing=%v extra=%v", missing, extra)
	}
}

type PodSelectors struct {
	Agent               string
	ClusterAgent        string
	ClusterChecksRunner string
}

func HelmPodSelectors() PodSelectors {
	return PodSelectors{
		Agent:               helmAgentLabelSelector,
		ClusterAgent:        helmClusterAgentLabelSelector,
		ClusterChecksRunner: helmCCRLabelSelector,
	}
}

func OperatorPodSelectors() PodSelectors {
	return PodSelectors{
		Agent:               operatorAgentLabelSelector,
		ClusterAgent:        operatorClusterAgentLabelSelector,
		ClusterChecksRunner: operatorCCRLabelSelector,
	}
}

func checkPodsHealth(t *testing.T, kubectlOptions *k8s.KubectlOptions, labelSelector string, expectedPodCount int) {
	err := k8s.WaitUntilNumPodsCreatedE(t, kubectlOptions, metav1.ListOptions{
		LabelSelector: labelSelector,
		FieldSelector: "status.phase=Running",
	}, expectedPodCount, 10, 15*time.Second)
	require.NoError(t, err, "Failed to wait for expected %d pods to be created for selector %s", expectedPodCount, labelSelector)

	podList := k8s.ListPods(t, kubectlOptions, metav1.ListOptions{
		LabelSelector: labelSelector,
		FieldSelector: "status.phase=Running",
	})
	require.Len(t, podList, expectedPodCount, "Expected %d pods, got %d for selector %s", expectedPodCount, len(podList), labelSelector)

	for _, pod := range podList {
		err = k8s.WaitUntilPodAvailableE(t, kubectlOptions, pod.Name, 10, 15*time.Second)
		require.NoError(t, err, "Failed to wait for pod %s to be available", pod.Name)

		// Check no restarts
		for _, containerStatus := range pod.Status.ContainerStatuses {
			require.Zero(t, containerStatus.RestartCount, "Container %s in pod %s has %d restarts",
				containerStatus.Name, pod.Name, containerStatus.RestartCount)
		}
	}
}

func verifyContainers(t *testing.T, kubectlOptions *k8s.KubectlOptions, expected ExpectedContainers, selectors PodSelectors) {
	checkContainers(t, kubectlOptions, selectors.Agent, expected.Agent)
	checkContainers(t, kubectlOptions, selectors.ClusterAgent, expected.ClusterAgent)
	checkContainers(t, kubectlOptions, selectors.ClusterChecksRunner, expected.ClusterChecksRunner)
}

func checkContainers(t *testing.T, kubectlOptions *k8s.KubectlOptions, labelSelector string, expected []string) {
	if len(expected) == 0 {
		return
	}

	pods := k8s.ListPods(t, kubectlOptions, metav1.ListOptions{
		LabelSelector: labelSelector,
		FieldSelector: "status.phase=Running",
	})
	require.NotEmpty(t, pods, "No pods found for selector %s while checking containers", labelSelector)

	pod := pods[0]
	actual := map[string]struct{}{}
	for _, container := range pod.Spec.Containers {
		actual[container.Name] = struct{}{}
	}

	for _, name := range expected {
		if _, ok := actual[name]; !ok {
			t.Fatalf("Expected container %s not found in pod %s (selector %s)", name, pod.Name, labelSelector)
		}
	}
}

func runMapper(t *testing.T, valuesPath string, namespace string, cleanup *CleanupRegistry) string {
	destFile, err := os.CreateTemp("", ddaDestPath)
	require.NoError(t, err)
	cleanup.AddDDA(destFile.Name())

	// Use absolute path to ensure we use our custom mapping file
	mappingFilePath := getMappingPath()
	t.Logf("Using mapping file: %s", mappingFilePath)

	// Run mapper as external process to avoid global state pollution between tests. 
	// TODO: run the mapper package directly when bug is fixed in mapper binary
	cmd := exec.Command("go", "run", mapperPackage, "map",
		"--mappingPath", mappingFilePath,
		"--sourcePath", valuesPath,
		"--destPath", destFile.Name(),
		"--namespace", namespace,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Mapper output: %s", string(output))
	}
	require.NoError(t, err, "Mapper failed: %s", string(output))

	return destFile.Name()
}
