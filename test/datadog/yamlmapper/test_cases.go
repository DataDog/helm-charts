// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package yamlmapper

// =============================================================================
// Base Test Cases
// =============================================================================

// baseTestCases defines all base values test cases with explicit pod counts
// Note: Many individual test files have been consolidated to reduce test run time
var baseTestCases = []BaseTestCase{
	// Consolidated test files (combine multiple related feature tests)
	{Name: "global-settings-consolidated-values.yaml", ValuesFile: baseValuesDir + "/global-settings-consolidated-values.yaml", ExpectedPods: defaultExpectedPods(), ExpectedContainers: defaultExpectedContainers()},
	{Name: "admission-controller-consolidated-values.yaml", ValuesFile: baseValuesDir + "/admission-controller-consolidated-values.yaml", ExpectedPods: defaultExpectedPods(), ExpectedContainers: defaultExpectedContainers()},
	{Name: "cluster-agent-features-consolidated-values.yaml", ValuesFile: baseValuesDir + "/cluster-agent-features-consolidated-values.yaml", ExpectedPods: defaultExpectedPods(), ExpectedContainers: defaultExpectedContainers()},
	{Name: "apm-logs-process-consolidated-values.yaml", ValuesFile: baseValuesDir + "/apm-logs-process-consolidated-values.yaml", ExpectedPods: defaultExpectedPods(), ExpectedContainers: ExpectedContainers{
		Agent: []string{containerTraceAgent},
	}},

	// Existing combined files
	{Name: "combined-apm-logs-values.yaml", ValuesFile: baseValuesDir + "/combined-apm-logs-values.yaml", ExpectedPods: defaultExpectedPods(), ExpectedContainers: defaultExpectedContainers()},
	{Name: "combined-full-observability-values.yaml", ValuesFile: baseValuesDir + "/combined-full-observability-values.yaml", ExpectedPods: defaultExpectedPods(), ExpectedContainers: defaultExpectedContainers()},

	// Baseline and minimal tests
	{Name: "default-minimal.yaml", ValuesFile: baseValuesDir + "/default-minimal.yaml", ExpectedPods: defaultExpectedPods(), ExpectedContainers: defaultExpectedContainers()},

	// Features that need separate tests (special pod counts or mutually exclusive configurations)
	{Name: "feature-cluster-checks-values.yaml", ValuesFile: baseValuesDir + "/feature-cluster-checks-values.yaml", ExpectedPods: ExpectedPodCounts{
		AgentDaemonset:      true,
		ClusterAgent:        1,
		ClusterChecksRunner: 2,
	}, ExpectedContainers: ExpectedContainers{
		ClusterChecksRunner: []string{containerAgent},
	}},
	{Name: "feature-dogstatsd-hostport-values.yaml", ValuesFile: baseValuesDir + "/feature-dogstatsd-hostport-values.yaml", ExpectedPods: defaultExpectedPods(), ExpectedContainers: defaultExpectedContainers()},
	{Name: "feature-dogstatsd-values.yaml", ValuesFile: baseValuesDir + "/feature-dogstatsd-values.yaml", ExpectedPods: defaultExpectedPods(), ExpectedContainers: defaultExpectedContainers()},

	// Skipped tests (require kernel features not available in kind)
	{Name: "feature-npm-values.yaml", ValuesFile: baseValuesDir + "/feature-npm-values.yaml", ExpectedPods: defaultExpectedPods(),
		ExpectedContainers: defaultExpectedContainers(),
		SkipReason:         "NPM requires kernel features not available in kind"},
	{Name: "feature-system-probe-checks-values.yaml", ValuesFile: baseValuesDir + "/feature-system-probe-checks-values.yaml", ExpectedPods: defaultExpectedPods(),
		ExpectedContainers: ExpectedContainers{
			Agent: []string{containerSystemProbe},
		},
		SkipReason: "System probe requires kernel features not available in kind"},
}

// =============================================================================
// Test Cases with Dependencies
// =============================================================================

// testCasesWithDependencies defines all test cases that require pre-created resources
var testCasesWithDependencies = []TestCaseWithDependencies{
	{
		Name:               "global-credentials-existing-secret-values.yaml",
		ValuesFile:         "values/global-credentials-existing-secret-values.yaml",
		ExpectedPods:       defaultExpectedPods(),
		ExpectedContainers: defaultExpectedContainers(),
		Secrets: []SecretDef{
			{Name: "my-datadog-api-secret", Data: map[string]string{"api-key": "00000000000000000000000000000000"}},
			{Name: "my-datadog-app-secret", Data: map[string]string{"app-key": "0000000000000000000000000000000000000000"}},
			{Name: "my-cluster-agent-token-secret", Data: map[string]string{"token": "test-cluster-agent-token-value"}},
		},
	},
	{
		Name:               "override-cluster-agent-values.yaml",
		ValuesFile:         "values/override-cluster-agent-values.yaml",
		ExpectedPods:       defaultExpectedPods(),
		ExpectedContainers: defaultExpectedContainers(),
		ConfigMaps: []ConfigMapDef{
			{Name: "cluster-agent-config", Data: map[string]string{"DD_LOG_LEVEL": "debug"}},
		},
		PriorityClasses: []PriorityClassDef{
			{Name: "cluster-agent-critical", Value: 1000000000, Description: "Ensures Cluster Agent pods run with elevated priority"},
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
			ClusterChecksRunner: []string{containerAgent},
		},
		ConfigMaps: []ConfigMapDef{
			{Name: "ccr-config", Data: map[string]string{"DD_LOG_LEVEL": "debug"}},
		},
		PriorityClasses: []PriorityClassDef{
			{Name: "ccr-critical", Value: 1000000000, Description: "Ensures Cluster Checks Runner pods run with elevated priority"},
		},
	},
	{
		Name:               "override-node-agent-values.yaml",
		ValuesFile:         "values/override-node-agent-values.yaml",
		ExpectedPods:       defaultExpectedPods(),
		ExpectedContainers: defaultExpectedContainers(),
		PriorityClasses: []PriorityClassDef{
			{Name: "datadog-agent-critical", Value: 1000000000, Description: "Ensures Datadog Agent pods run with elevated priority"},
		},
	},
	{
		Name:               "global-envfrom-values.yaml",
		ValuesFile:         "values/global-envfrom-values.yaml",
		ExpectedPods:       defaultExpectedPods(),
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
// Negative Test Cases
// =============================================================================

// negativeTestCases defines test cases where the mapper should return an error
var negativeTestCases = []NegativeTestCase{
	{
		Name:           "unsupported-helm-key",
		ValuesFile:     negativeValuesDir + "/unsupported-key-values.yaml",
		ExpectedErrMsg: "error",
		Description:    "Mapper should error when values file contains unmapped/unsupported Helm keys",
	},
	{
		Name:           "multiple-unsupported-keys",
		ValuesFile:     negativeValuesDir + "/multiple-unsupported-keys-values.yaml",
		ExpectedErrMsg: "error",
		Description:    "Mapper should error when values file contains multiple unmapped keys",
	},
}

