// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package yamlmapper

import (
	"sync"
	"time"
)

// =============================================================================
// Paths, env vars, and defaults
// =============================================================================

const (
	ddaOutputDir      = "baseline/dda"
	operatorChartPath = "../../../charts/datadog-operator"
	datadogChartPath  = "../../../charts/datadog"

	apiKeyEnv = "API_KEY"
	appKeyEnv = "APP_KEY"

	containerAgent       = "agent"
	containerTraceAgent  = "trace-agent"
	containerSystemProbe = "system-probe"

	// Helm release names
	releaseDatadog         = "datadog"
	releaseDatadogOperator = "datadog-operator"

	defaultPodTimeout      = 15 * time.Second
	defaultPodRetries      = 10
	defaultWaitTimeout     = 2 * time.Minute
	defaultHelmTimeout     = 3 * time.Minute
	defaultContainerdDelay = 5 * time.Second

	operatorDeployRetries = 10
	operatorDeploySleep   = 15 * time.Second

	testNamespacePrefix = "datadog-agent-"
)

// Directory paths for values files
const (
	valuesDir     = "baseline/values"
	mappingFileName   = "mapping_datadog_helm_to_datadogagent_crd.yaml"
)

// =============================================================================
// Test case types
// =============================================================================

// ExpectedComponentPods defines explicit expected pod counts for each component.
// DaemonSet Agent pod count is always calculated dynamically based on cluster node count.
type ExpectedComponentPods struct {
	ClusterAgent        int
	ClusterChecksRunner int
}

// ExpectedComponentContainers defines required container names per component.
type ExpectedComponentContainers struct {
	Agent               []string
	ClusterAgent        []string
	ClusterChecksRunner []string
}

// BaseTestCase defines a simple test case without dependencies
type BaseTestCase struct {
	Name               string
	ValuesFile         string
	ExpectedPods       ExpectedComponentPods
	ExpectedComponentContainers ExpectedComponentContainers
	SkipReason         string
}

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

// PriorityClassDef defines a PriorityClass to be created before the test
type PriorityClassDef struct {
	Name        string
	Value       int32
	Description string
}

// ResourceDependentTestCase defines a test case that requires pre-created resources.
//   - Name: test name (usually the filename)
//   - ValuesFile: path to the values file relative to the yamlmapper directory
//   - ExpectedPods: explicit pod counts for each component
//   - ExpectedComponentContainers: required container names per component
//   - ConfigMaps: ConfigMaps to create before the test
//   - Secrets: Secrets to create before the test
//   - PriorityClasses: PriorityClasses to create before the test
//   - SkipReason: if set, the test will be skipped with this reason
type ResourceDependentTestCase struct {
	Name               string
	ValuesFile         string
	ExpectedPods       ExpectedComponentPods
	ExpectedComponentContainers ExpectedComponentContainers
	ConfigMaps         []ConfigMapDef
	Secrets            []SecretDef
	PriorityClasses    []PriorityClassDef
	SkipReason         string
}

// NegativeTestCase defines a test case that expects the mapper to fail
type NegativeTestCase struct {
	Name           string
	ValuesFile     string
	ExpectedErrMsg string
	Description    string
}

// PodSelectors groups label selectors for agent components.
type PodSelectors struct {
	Agent               string
	ClusterAgent        string
	ClusterChecksRunner string
}

// =============================================================================
// Cleanup registry
// =============================================================================

// TestCleanupRegistry stores cleanup hooks for test runs.
// - datadog: datadog helm chart uninstall function
// - operator: operator chart uninstall function
type TestCleanupRegistry struct {
	mu       sync.Mutex
	datadog  func()
	operator func()
}

func (c *TestCleanupRegistry) SetDatadog(cleanup func()) {
	c.mu.Lock()
	c.datadog = cleanup
	c.mu.Unlock()
}

func (c *TestCleanupRegistry) UninstallDatadog() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.datadog != nil {
		c.datadog()
		c.datadog = nil
	}
}

func (c *TestCleanupRegistry) SetOperator(cleanup func()) {
	c.mu.Lock()
	c.operator = cleanup
	c.mu.Unlock()
}

func (c *TestCleanupRegistry) UninstallOperator() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.operator != nil {
		c.operator()
		c.operator = nil
	}
}

// =============================================================================
// Defaults
// =============================================================================

func defaultExpectedPods() ExpectedComponentPods {
	return ExpectedComponentPods{
		ClusterAgent:        1,
		ClusterChecksRunner: 0,
	}
}

// Default expected containers when left empty means "no container assertions for this component".
func defaultExpectedComponentContainers() ExpectedComponentContainers {
	return ExpectedComponentContainers{}
}

// helmPodSelectors returns label selectors for Helm-managed pods
func helmPodSelectors() PodSelectors {
	return PodSelectors{
		Agent:               helmAgentLabelSelector,
		ClusterAgent:        helmClusterAgentLabelSelector,
		ClusterChecksRunner: helmCCRLabelSelector,
	}
}

// operatorPodSelectors returns label selectors for Operator-managed pods
func operatorPodSelectors() PodSelectors {
	return PodSelectors{
		Agent:               operatorAgentLabelSelector,
		ClusterAgent:        operatorClusterAgentLabelSelector,
		ClusterChecksRunner: operatorCCRLabelSelector,
	}
}

