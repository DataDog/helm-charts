// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package yamlmapper

import (
	"time"
)

// =============================================================================
// Constants
// =============================================================================

const (
	// File paths and names
	ddaDestPath       = "tempDDADest.yaml"
	operatorChartPath = "../../../charts/datadog-operator"
	datadogChartPath  = "../../../charts/datadog"

	// Environment variables
	apiKeyEnv = "API_KEY"
	appKeyEnv = "APP_KEY"

	// Container names
	containerAgent       = "agent"
	containerTraceAgent  = "trace-agent"
	containerSystemProbe = "system-probe"

	// Helm release names
	releaseDatadog         = "datadog"
	releaseDatadogOperator = "datadog-operator"

	// Timeouts
	defaultPodTimeout      = 15 * time.Second
	defaultPodRetries      = 10
	defaultWaitTimeout     = 2 * time.Minute
	defaultHelmTimeout     = 3 * time.Minute
	defaultContainerdDelay = 5 * time.Second
)

// =============================================================================
// Path Configuration
// =============================================================================

const (
	// Directory paths for values files
	baseValuesDir     = "values/base"
	negativeValuesDir = "values/negative"
	mappingFileName   = "mapping_datadog_helm_to_datadogagent_crd.yaml"
)

// =============================================================================
// Type Definitions
// =============================================================================

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

// BaseTestCase defines a simple test case without dependencies
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

// PriorityClassDef defines a PriorityClass to be created before the test
type PriorityClassDef struct {
	Name        string
	Value       int32
	Description string
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
	// PriorityClasses to create before the test
	PriorityClasses []PriorityClassDef
	// SkipReason if set, the test will be skipped with this reason
	SkipReason string
}

// PodSelectors contains label selectors for different pod types
type PodSelectors struct {
	Agent               string
	ClusterAgent        string
	ClusterChecksRunner string
}

// NegativeTestCase defines a test case that expects the mapper to fail
type NegativeTestCase struct {
	Name           string
	ValuesFile     string
	ExpectedErrMsg string // Substring that should appear in the error message
	Description    string // Human-readable description of what's being tested
}

// valuesExpectations is used to parse values files for validation
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

// =============================================================================
// Default Values
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

// HelmPodSelectors returns label selectors for Helm-managed pods
func HelmPodSelectors() PodSelectors {
	return PodSelectors{
		Agent:               helmAgentLabelSelector,
		ClusterAgent:        helmClusterAgentLabelSelector,
		ClusterChecksRunner: helmCCRLabelSelector,
	}
}

// OperatorPodSelectors returns label selectors for Operator-managed pods
func OperatorPodSelectors() PodSelectors {
	return PodSelectors{
		Agent:               operatorAgentLabelSelector,
		ClusterAgent:        operatorClusterAgentLabelSelector,
		ClusterChecksRunner: operatorCCRLabelSelector,
	}
}

