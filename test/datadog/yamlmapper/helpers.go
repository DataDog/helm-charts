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

	"github.com/DataDog/datadog-operator/cmd/yaml-mapper/mapper"
	"github.com/DataDog/helm-charts/test/common"
	"github.com/google/go-cmp/cmp"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ### Test setup and cleanup

type TestContext struct {
	T                *testing.T
	Namespace        string
	KubectlOptions   *k8s.KubectlOptions
	TestCleanupRegistry  *TestCleanupRegistry
	cleanupSecrets   func()
	cleanupResources func()
}

// newTestContext creates a new test context with a unique namespace.
func newTestContext(t *testing.T) *TestContext {
	t.Helper()
	namespace := fmt.Sprintf("%s%s", testNamespacePrefix, strings.ToLower(random.UniqueId()))
	kubectlOptions := k8s.NewKubectlOptions("", "", namespace)
	k8s.CreateNamespace(t, kubectlOptions, namespace)

	return &TestContext{
		T:               t,
		Namespace:       namespace,
		KubectlOptions:  kubectlOptions,
		TestCleanupRegistry: &TestCleanupRegistry{},
	}
}

// SetupCleanup registers the cleanup function with t.Cleanup().
// This should be called after creating the test context.
func (tc *TestContext) SetupCleanup() {
	tc.T.Cleanup(func() {
		tc.cleanup()
	})
}

// SetupSecretsFromEnv creates secrets from environment variables if available.
func (tc *TestContext) SetupSecretsFromEnv() {
	if os.Getenv(apiKeyEnv) != "" && os.Getenv(appKeyEnv) != "" {
		tc.cleanupSecrets = common.CreateSecretFromEnv(tc.T, tc.KubectlOptions, apiKeyEnv, appKeyEnv)
	}
}

// SetupTestResources creates test-specific resources (ConfigMaps, Secrets, PriorityClasses).
func (tc *TestContext) SetupTestResources(testCase *ResourceDependentTestCase) {
	tc.cleanupResources = createTestResources(tc.T, tc.KubectlOptions, testCase)
}

// cleanup performs the full cleanup sequence for a test.
func (tc *TestContext) cleanup() {
	t := tc.T
	cleanupKubectlOptions := tc.KubectlOptions

	t.Logf("Starting cleanup for namespace %s", tc.Namespace)

	// Step 1: Delete DDA and wait for finalizers to complete
	for _, ddaFile := range tc.TestCleanupRegistry.GetFiles() {
		t.Logf("Deleting DDA from file %s", ddaFile)
		_ = k8s.RunKubectlE(t, cleanupKubectlOptions, []string{"delete", "-f", ddaFile, "--ignore-not-found", "--wait=true", "--timeout=60s"}...)
		_ = os.Remove(ddaFile)
	}

	waitForDDADeletion(t, cleanupKubectlOptions, tc.Namespace, defaultWaitTimeout)

	// Step 2: Uninstall helm charts
	tc.TestCleanupRegistry.UninstallOperator()
	tc.TestCleanupRegistry.UninstallDatadog()

	// Wait for all pods to terminate
	_ = waitForPodsTerminated(t, tc.KubectlOptions, "", defaultWaitTimeout)

	// Step 3: Clean up test resources
	t.Logf("Cleaning up test-specific resources in namespace %s", tc.Namespace)
	if tc.cleanupResources != nil {
		tc.cleanupResources()
	}
	if tc.cleanupSecrets != nil {
		tc.cleanupSecrets()
	}

	// Step 4: Delete namespace
	t.Logf("Deleting namespace %s", tc.Namespace)
	k8s.DeleteNamespace(t, cleanupKubectlOptions, tc.Namespace)
	waitForNamespaceDeletion(t, tc.Namespace, defaultWaitTimeout)

	t.Logf("Cleanup complete for namespace %s", tc.Namespace)
}

// ### Dependency Resource Creation Helpers

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

// createTestResources creates all required k8s resourcesfor a test case.
// Returns a cleanup function that removes all created resources.
func createTestResources(t *testing.T, kubectlOptions *k8s.KubectlOptions, tc *ResourceDependentTestCase) func() {
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

// ### Validation Helpers

func validateExpectedPodsAgainstValues(t *testing.T, valuesPath string, expected ExpectedComponentPods) {
	data, err := os.ReadFile(valuesPath)
	require.NoError(t, err, "Failed to read values file: %s", valuesPath)

	var parsed struct {
		Agents struct {
			Enabled *bool `yaml:"enabled"`
		} `yaml:"agents"`
		ClusterAgent struct {
			Enabled  *bool `yaml:"enabled"`
			Replicas *int  `yaml:"replicas"`
		} `yaml:"clusterAgent"`
		ClusterChecksRunner struct {
			Enabled  *bool `yaml:"enabled"`
			Replicas *int  `yaml:"replicas"`
		} `yaml:"clusterChecksRunner"`
	}
	err = yaml.Unmarshal(data, &parsed)
	require.NoError(t, err, "Failed to parse values file: %s", valuesPath)

	if parsed.Agents.Enabled != nil && !*parsed.Agents.Enabled {
		t.Fatalf("Values file disables agents but tests require agent pods: %s", valuesPath)
	}
	if parsed.ClusterAgent.Enabled != nil && !*parsed.ClusterAgent.Enabled {
		t.Fatalf("Values file disables clusterAgent but tests require clusterAgent pods: %s", valuesPath)
	}

	validateExpectedPodCount(t, "clusterAgent", parsed.ClusterAgent.Enabled, parsed.ClusterAgent.Replicas, expected.ClusterAgent, valuesPath)
	validateExpectedPodCount(t, "clusterChecksRunner", parsed.ClusterChecksRunner.Enabled, parsed.ClusterChecksRunner.Replicas, expected.ClusterChecksRunner, valuesPath)
}

func validateExpectedPodCount(t *testing.T, name string, enabled *bool, replicas *int, expected int, valuesPath string) {
	if enabled != nil {
		if *enabled && expected == 0 {
			t.Fatalf("Values file enables %s but expected count is 0: %s", name, valuesPath)
		}
		if !*enabled && expected > 0 {
			t.Fatalf("Values file disables %s but expected count is %d: %s", name, expected, valuesPath)
		}
	}
	if replicas != nil && expected != *replicas {
		t.Fatalf("Values file sets %s.replicas=%d but expected count is %d: %s", name, *replicas, expected, valuesPath)
	}
}

// assertBaseValuesCoverage verifies that all base values files are covered by a test case.
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

// ### Operator Installation

// installOperator installs the datadog-operator chart and waits for it to be ready
func installOperator(t *testing.T, kubectlOptions *k8s.KubectlOptions, namespace string, cleanup *TestCleanupRegistry) error {
	operatorInstallCmd := common.HelmCommand{
		ReleaseName: releaseDatadogOperator,
		ChartPath:   operatorChartPath,
		Overrides: map[string]string{
			"installCRDs":        "false", // CRDs managed externally (CI or locally)
			"watchNamespaces[0]": namespace,
		},
	}

	cleanUpOperator := common.InstallChart(t, kubectlOptions, operatorInstallCmd)
	cleanup.SetOperator(cleanUpOperator)

	operatorDeployments := k8s.ListDeployments(t, kubectlOptions, metav1.ListOptions{
		LabelSelector: "app.kubernetes.io/name=datadog-operator",
	})
	require.Len(t, operatorDeployments, 1, "Expected 1 deployment, got %d", len(operatorDeployments))

	err := k8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, operatorDeployments[0].Name, operatorDeployRetries, operatorDeploySleep)
	if err != nil {
		t.Logf("Failed to wait for operator deployment: %v", err)
		return err
	}
	return nil
}

// applyDDAAndWaitForAgents applies the DDA manifest and waits for operator-managed agents to be ready.
func applyDDAAndWaitForAgents(t *testing.T, kubectlOptions *k8s.KubectlOptions, ddaFilePath string) error {
	err := k8s.RunKubectlE(t, kubectlOptions, []string{"apply", "-f", ddaFilePath}...)
	if err != nil {
		t.Logf("Failed to apply DDA: %v", err)
		return err
	}

	expectedPods := expectedDsCount(t, kubectlOptions)
	err = k8s.WaitUntilNumPodsCreatedE(t, kubectlOptions, metav1.ListOptions{
		LabelSelector: operatorAgentLabelSelector,
		FieldSelector: "status.phase=Running",
	}, expectedPods, operatorDeployRetries, operatorDeploySleep)
	if err != nil {
		t.Logf("Failed to wait for operator agent pods: %v", err)
		return err
	}
	return nil
}

// ### Verification Helpers

// verifyAgentConf compares helm agent config against operator agent config.
// Differences indicate that the mapper produced a DDA that results in different
// agent runtime configuration than the original Helm values.
func verifyAgentConfEqual(t *testing.T, helmAgentConf string, operatorAgentConf string) {
	t.Helper()

	agentConfEqual := cmp.Equal(helmAgentConf, operatorAgentConf)
	if !agentConfEqual {
		diff := cmp.Diff(helmAgentConf, operatorAgentConf)
		t.Logf("Agent config diff detected (- helm, + operator):\n%s", diff)
		if isAgentConfStrict() {
			require.True(t, agentConfEqual,
				"Strict mode: helm vs operator agent config mismatch. "+
					"The mapper produced a DDA that results in different runtime configuration. "+
					"Review the diff above to identify the discrepancy.")
		}
	}
}

// assertExpectedPodHealth asserts that expected pods are running and healthy.
func assertExpectedPodHealth(t *testing.T, kubectlOptions *k8s.KubectlOptions, expected ExpectedComponentPods, selectors PodSelectors) {
	agentCount := expectedDsCount(t, kubectlOptions)
	waitForPodsHealthy(t, kubectlOptions, selectors.Agent, agentCount)

	if expected.ClusterAgent > 0 {
		waitForPodsHealthy(t, kubectlOptions, selectors.ClusterAgent, expected.ClusterAgent)
	} else {
		expectNoPods(t, kubectlOptions, selectors.ClusterAgent)
	}
	if expected.ClusterChecksRunner > 0 {
		waitForPodsHealthy(t, kubectlOptions, selectors.ClusterChecksRunner, expected.ClusterChecksRunner)
	} else {
		expectNoPods(t, kubectlOptions, selectors.ClusterChecksRunner)
	}
}

// waitForPodsHealthy verifies that expected pods are running and healthy.
func waitForPodsHealthy(t *testing.T, kubectlOptions *k8s.KubectlOptions, labelSelector string, expectedPodCount int) {
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

		for _, containerStatus := range pod.Status.ContainerStatuses {
			require.Zero(t, containerStatus.RestartCount,
				"Container %q in pod %q has %d restarts. "+
					"The container is unhealthy. Check container logs.",
				containerStatus.Name, pod.Name, containerStatus.RestartCount)
		}
	}
}

// expectNoPods asserts that no pods exist for the selector.
func expectNoPods(t *testing.T, kubectlOptions *k8s.KubectlOptions, labelSelector string) {
	t.Helper()
	pods := k8s.ListPods(t, kubectlOptions, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	require.Empty(t, pods,
		"Expected no pods for selector %q in namespace %q, but found %d",
		labelSelector, kubectlOptions.Namespace, len(pods))
}

func verifyContainers(t *testing.T, kubectlOptions *k8s.KubectlOptions, expected ExpectedComponentContainers, selectors PodSelectors) {
	assertPodContainers(t, kubectlOptions, selectors.Agent, expected.Agent)
	assertPodContainers(t, kubectlOptions, selectors.ClusterAgent, expected.ClusterAgent)
	assertPodContainers(t, kubectlOptions, selectors.ClusterChecksRunner, expected.ClusterChecksRunner)
}

// assertPodContainers verifies that all pods matching the selector contain the expected containers.
func assertPodContainers(t *testing.T, kubectlOptions *k8s.KubectlOptions, labelSelector string, expected []string) {
	if len(expected) == 0 {
		return
	}

	pods := k8s.ListPods(t, kubectlOptions, metav1.ListOptions{
		LabelSelector: labelSelector,
		FieldSelector: "status.phase=Running",
	})
	require.NotEmpty(t, pods, "No pods found for selector %s while checking containers", labelSelector)

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

// containerNames gets the container names for error messages
func containerNames(containers []corev1.Container) []string {
	names := make([]string, len(containers))
	for i, c := range containers {
		names[i] = c.Name
	}
	return names
}

// ### Mapper Helpers

// getMappingPath returns the absolute path to the mapping file.
func getMappingPath() string {
	absPath, err := filepath.Abs(mappingFileName)
	if err != nil {
		return mappingFileName
	}
	return absPath
}

// runMapper executes the yaml-mapper and returns the output DDA file path.
func runMapper(t *testing.T, valuesPath string, namespace string, cleanup *TestCleanupRegistry) (string, error) {
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

	// capture mapper logs to detect warnings and silent errors
	logCapture, restoreLog := captureMapperLogs()
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

	// check for mapper warnings and fail if strict mode is enabled
	reportMapperWarnings(t, logCapture)
	reportMapperErrors(t, logCapture)
	if err == nil && logCapture.ErrorCount() > 0 {
		t.Fatalf("Mapper logged %d error(s) without returning an error", logCapture.ErrorCount())
	}

	return destFile.Name(), err
}

// runMapperExpectError runs the mapper and expects it to return an error.
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
		Namespace:   "test-ns",
		PrintOutput: false,
	}
	newMapper := mapper.NewMapper(mapperConfig)
	return newMapper.Run()
}


