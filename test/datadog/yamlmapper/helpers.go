// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package yamlmapper

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DataDog/datadog-operator/cmd/yaml-mapper/mapper"
	"github.com/DataDog/helm-charts/test/common"
	"github.com/google/go-cmp/cmp"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// =============================================================================
// Path Helpers
// =============================================================================

// getMappingPath returns the absolute path to the mapping file.
// This ensures the mapper uses our custom mapping file instead of the embedded default.
func getMappingPath() string {
	absPath, err := filepath.Abs(mappingFileName)
	if err != nil {
		return mappingFileName
	}
	return absPath
}

// =============================================================================
// Test Context
// =============================================================================

// TestContext holds all the context needed for a single test run.
// This helps consolidate test state and simplify cleanup.
type TestContext struct {
	T                *testing.T
	Namespace        string
	KubectlOptions   *k8s.KubectlOptions
	CleanupRegistry  *CleanupRegistry
	cleanupSecrets   func()
	cleanupResources func()
}

// NewTestContext creates a new test context with a unique namespace.
func NewTestContext(t *testing.T) *TestContext {
	t.Helper()
	namespace := fmt.Sprintf("%s%s", testNamespacePrefix, strings.ToLower(random.UniqueId()))
	kubectlOptions := k8s.NewKubectlOptions("", "", namespace)
	k8s.CreateNamespace(t, kubectlOptions, namespace)

	return &TestContext{
		T:               t,
		Namespace:       namespace,
		KubectlOptions:  kubectlOptions,
		CleanupRegistry: &CleanupRegistry{},
	}
}

// SetupCleanup registers the cleanup function with t.Cleanup().
// This should be called after creating the test context.
func (tc *TestContext) SetupCleanup() {
	tc.T.Cleanup(func() {
		tc.cleanup()
	})
}

// cleanup performs the full cleanup sequence for a test.
func (tc *TestContext) cleanup() {
	t := tc.T
	cleanupKubectlOptions := quietKubectlOptions(tc.KubectlOptions)

	logVerbosef(t, "Starting cleanup for namespace %s", tc.Namespace)

	// Step 1: Delete DDA and wait for finalizers to complete
	for _, ddaFile := range tc.CleanupRegistry.GetFiles() {
		logVerbosef(t, "Deleting DDA from file %s", ddaFile)
		_ = k8s.RunKubectlE(t, cleanupKubectlOptions, []string{"delete", "-f", ddaFile, "--ignore-not-found", "--wait=true", "--timeout=60s"}...)
		_ = os.Remove(ddaFile)
	}

	waitForDDADeletion(t, cleanupKubectlOptions, tc.Namespace, defaultWaitTimeout)

	// Step 2: Uninstall helm charts
	tc.CleanupRegistry.UninstallOperator()
	tc.CleanupRegistry.UninstallDatadog()

	// Wait for all pods to terminate
	_ = waitForPodsTerminated(t, tc.KubectlOptions, "", defaultWaitTimeout)

	// Step 3: Clean up test resources
	logVerbosef(t, "Cleaning up test-specific resources in namespace %s", tc.Namespace)
	if tc.cleanupResources != nil {
		tc.cleanupResources()
	}
	if tc.cleanupSecrets != nil {
		tc.cleanupSecrets()
	}

	// Step 4: Delete namespace
	logVerbosef(t, "Deleting namespace %s", tc.Namespace)
	k8s.DeleteNamespace(t, cleanupKubectlOptions, tc.Namespace)
	waitForNamespaceDeletion(t, tc.Namespace, defaultWaitTimeout)

	logVerbosef(t, "Cleanup complete for namespace %s", tc.Namespace)
}

// SetupSecretsFromEnv creates secrets from environment variables if available.
func (tc *TestContext) SetupSecretsFromEnv() {
	if os.Getenv(apiKeyEnv) != "" && os.Getenv(appKeyEnv) != "" {
		tc.cleanupSecrets = common.CreateSecretFromEnv(tc.T, tc.KubectlOptions, apiKeyEnv, appKeyEnv)
	}
}

// SetupTestResources creates test-specific resources (ConfigMaps, Secrets, PriorityClasses).
func (tc *TestContext) SetupTestResources(testCase *TestCaseWithDependencies) {
	tc.cleanupResources = createTestResources(tc.T, tc.KubectlOptions, testCase)
}

// =============================================================================
// Resource Creation Helpers
// =============================================================================

func createPriorityClasses(t *testing.T, kubectlOptions *k8s.KubectlOptions, classes []PriorityClassDef) func() {
	if len(classes) == 0 {
		return func() {}
	}

	clientset, err := k8s.GetKubernetesClientFromOptionsE(t, kubectlOptions)
	require.NoError(t, err, "Failed to get Kubernetes client")

	created := make([]string, 0, len(classes))
	for _, class := range classes {
		pc := &schedulingv1.PriorityClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: class.Name,
			},
			Value:       class.Value,
			Description: class.Description,
		}

		_, err := clientset.SchedulingV1().PriorityClasses().Create(context.Background(), pc, metav1.CreateOptions{})
		if apierrors.IsAlreadyExists(err) {
			continue
		}
		require.NoError(t, err, "Failed to create priority class %s", class.Name)
		created = append(created, class.Name)
	}

	return func() {
		for i := len(created) - 1; i >= 0; i-- {
			name := created[i]
			err := clientset.SchedulingV1().PriorityClasses().Delete(context.Background(), name, metav1.DeleteOptions{})
			if err != nil && !apierrors.IsNotFound(err) {
				t.Logf("Failed to delete priority class %s: %v", name, err)
			}
		}
	}
}

// createTestResources creates all required ConfigMaps and Secrets for a test case.
// Returns a cleanup function that removes all created resources.
func createTestResources(t *testing.T, kubectlOptions *k8s.KubectlOptions, tc *TestCaseWithDependencies) func() {
	if tc == nil {
		return func() {}
	}

	var cleanupFuncs []func()

	if len(tc.PriorityClasses) > 0 {
		cleanup := createPriorityClasses(t, kubectlOptions, tc.PriorityClasses)
		cleanupFuncs = append(cleanupFuncs, cleanup)
	}

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
// Operator Installation
// =============================================================================

// installOperator installs the datadog-operator chart and waits for it to be ready
func installOperator(t *testing.T, kubectlOptions *k8s.KubectlOptions, namespace string, cleanup *CleanupRegistry) error {
	quietOptions := quietKubectlOptions(kubectlOptions)
	operatorInstallCmd := common.HelmCommand{
		ReleaseName: releaseDatadogOperator,
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

// =============================================================================
// Verification Helpers
// =============================================================================

// verifyAgentConf compares helm agent config against operator agent config.
// Differences indicate that the mapper produced a DDA that results in different
// runtime configuration than the original Helm values.
func verifyAgentConf(t *testing.T, kubectlOptions *k8s.KubectlOptions, helmAgentConf string) {
	t.Helper()

	operatorAgentConf := getOperatorAgentConf(t, kubectlOptions)
	if operatorAgentConf == "" {
		if isAgentConfStrict() {
			require.NotEmpty(t, operatorAgentConf,
				"Strict mode: expected operator agent config for comparison in namespace %q. "+
					"This usually means the operator-managed agent pod isn't running.",
				kubectlOptions.Namespace)
		}
		t.Log("Warning: Could not retrieve operator agent config for comparison, skipping config verification")
		return
	}

	agentConfEqual := cmp.Equal(helmAgentConf, operatorAgentConf)
	if !agentConfEqual {
		diff := cmp.Diff(helmAgentConf, operatorAgentConf)
		t.Logf("Agent config diff detected (- helm, + operator):\n%s", diff)
		if isAgentConfStrict() {
			require.True(t, agentConfEqual,
				"Strict mode: helm vs operator agent config mismatch. "+
					"This indicates the mapper produced a DDA that results in different runtime configuration. "+
					"Review the diff above to identify the discrepancy.")
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

// checkPodsHealth verifies that expected pods are running, healthy, and have no restarts.
// It provides detailed error messages to help diagnose failures.
func checkPodsHealth(t *testing.T, kubectlOptions *k8s.KubectlOptions, labelSelector string, expectedPodCount int) {
	t.Helper()

	listOpts := metav1.ListOptions{
		LabelSelector: labelSelector,
		FieldSelector: "status.phase=Running",
	}

	err := k8s.WaitUntilNumPodsCreatedE(t, kubectlOptions, listOpts, expectedPodCount, defaultPodRetries, defaultPodTimeout)
	require.NoError(t, err,
		"Timed out waiting for %d pods with selector %q in namespace %q. "+
			"Check if the pods are scheduled and images are available.",
		expectedPodCount, labelSelector, kubectlOptions.Namespace)

	podList := k8s.ListPods(t, kubectlOptions, listOpts)
	require.Len(t, podList, expectedPodCount,
		"Pod count mismatch for selector %q in namespace %q: expected %d, got %d",
		labelSelector, kubectlOptions.Namespace, expectedPodCount, len(podList))

	for _, pod := range podList {
		err = k8s.WaitUntilPodAvailableE(t, kubectlOptions, pod.Name, defaultPodRetries, defaultPodTimeout)
		require.NoError(t, err,
			"Pod %q never became available in namespace %q. Current phase: %s, Conditions: %v",
			pod.Name, kubectlOptions.Namespace, pod.Status.Phase, pod.Status.Conditions)

		// Check no restarts - indicates stability issues
		for _, containerStatus := range pod.Status.ContainerStatuses {
			require.Zero(t, containerStatus.RestartCount,
				"Container %q in pod %q has %d restart(s). "+
					"This indicates stability issues. Check container logs with: "+
					"kubectl logs %s -c %s -n %s",
				containerStatus.Name, pod.Name, containerStatus.RestartCount,
				pod.Name, containerStatus.Name, kubectlOptions.Namespace)
		}
	}
}

func verifyContainers(t *testing.T, kubectlOptions *k8s.KubectlOptions, expected ExpectedContainers, selectors PodSelectors) {
	checkContainers(t, kubectlOptions, selectors.Agent, expected.Agent)
	checkContainers(t, kubectlOptions, selectors.ClusterAgent, expected.ClusterAgent)
	checkContainers(t, kubectlOptions, selectors.ClusterChecksRunner, expected.ClusterChecksRunner)
}

// checkContainers verifies that ALL pods matching the selector contain the expected containers.
// This ensures consistency across all replicas/nodes.
func checkContainers(t *testing.T, kubectlOptions *k8s.KubectlOptions, labelSelector string, expected []string) {
	if len(expected) == 0 {
		return
	}

	pods := k8s.ListPods(t, kubectlOptions, metav1.ListOptions{
		LabelSelector: labelSelector,
		FieldSelector: "status.phase=Running",
	})
	require.NotEmpty(t, pods, "No pods found for selector %s while checking containers", labelSelector)

	// Check ALL pods, not just the first one
	for _, pod := range pods {
		actual := make(map[string]struct{}, len(pod.Spec.Containers))
		for _, container := range pod.Spec.Containers {
			actual[container.Name] = struct{}{}
		}

		for _, name := range expected {
			require.Contains(t, actual, name,
				"Expected container %s not found in pod %s (selector %s). Available containers: %v",
				name, pod.Name, labelSelector, containerNames(pod.Spec.Containers))
		}
	}
}

// containerNames extracts container names for error messages
func containerNames(containers []corev1.Container) []string {
	names := make([]string, len(containers))
	for i, c := range containers {
		names[i] = c.Name
	}
	return names
}

// =============================================================================
// Validation Helpers
// =============================================================================

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

// =============================================================================
// Mapper Helpers
// =============================================================================

// runMapper executes the yaml-mapper and returns the output DDA file path.
// If cleanup is nil, the caller is responsible for cleaning up the output file.
func runMapper(t *testing.T, valuesPath string, namespace string, cleanup *CleanupRegistry) (string, error) {
	t.Helper()
	destFile, err := os.CreateTemp("", ddaDestPath)
	require.NoError(t, err, "Failed to create temp file for DDA output")
	if cleanup != nil {
		cleanup.AddDDA(destFile.Name())
	}

	mappingFilePath := getMappingPath()
	t.Logf("Using mapping file: %s", mappingFilePath)

	absValuesPath, err := filepath.Abs(valuesPath)
	require.NoError(t, err, "Failed to get absolute path for values file: %s", valuesPath)

	// Install log capture to detect mapper warnings
	logCapture, restoreLog := InstallLogCapture()
	defer restoreLog()

	mapperConfig := mapper.MapConfig{
		MappingPath: mappingFilePath,
		SourcePath:  absValuesPath,
		DestPath:    destFile.Name(),
		Namespace:   namespace,
		PrintOutput: false,
	}
	newMapper := mapper.NewMapper(mapperConfig)
	err = newMapper.Run()

	// Check for mapper warnings (logs them and optionally fails if strict mode enabled)
	CheckMapperWarnings(t, logCapture)

	return destFile.Name(), err
}

// runMapperExpectError runs the mapper and expects it to return an error.
// Returns the error for further inspection.
func runMapperExpectError(t *testing.T, valuesPath string) error {
	t.Helper()
	destFile, err := os.CreateTemp("", ddaDestPath)
	require.NoError(t, err, "Failed to create temp file for DDA output")
	defer os.Remove(destFile.Name())

	mappingFilePath := getMappingPath()
	absValuesPath, err := filepath.Abs(valuesPath)
	require.NoError(t, err, "Failed to get absolute path for values file: %s", valuesPath)

	mapperConfig := mapper.MapConfig{
		MappingPath: mappingFilePath,
		SourcePath:  absValuesPath,
		DestPath:    destFile.Name(),
		Namespace:   "test-namespace",
		PrintOutput: false,
	}
	newMapper := mapper.NewMapper(mapperConfig)
	return newMapper.Run()
}

