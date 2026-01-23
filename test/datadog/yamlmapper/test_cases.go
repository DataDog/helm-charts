// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package yamlmapper

// ### Base Test Cases

var baseTestCases = []BaseTestCase{
	{
		Name:               "global-settings-values.yaml",
		ValuesFile:         baseValuesDir + "/global-settings-values.yaml",
		ExpectedPods:       defaultExpectedPods(),
		ExpectedComponentContainers: defaultExpectedComponentContainers(),
	},
	{
		Name:               "admission-controller-values.yaml",
		ValuesFile:         baseValuesDir + "/admission-controller-values.yaml",
		ExpectedPods:       defaultExpectedPods(),
		ExpectedComponentContainers: defaultExpectedComponentContainers(),
	},
	{
		Name:               "cluster-agent-features-values.yaml",
		ValuesFile:         baseValuesDir + "/cluster-agent-features-values.yaml",
		ExpectedPods:       defaultExpectedPods(),
		ExpectedComponentContainers: defaultExpectedComponentContainers(),
	},
	{
		Name:       "apm-logs-process-values.yaml",
		ValuesFile: baseValuesDir + "/apm-logs-process-values.yaml",
		ExpectedPods: defaultExpectedPods(),
		ExpectedComponentContainers: ExpectedComponentContainers{
			Agent: []string{containerTraceAgent},
		},
	},
	{
		Name:       "apm-port-enabled-values.yaml",
		ValuesFile: baseValuesDir + "/apm-port-enabled-values.yaml",
		ExpectedPods: defaultExpectedPods(),
		ExpectedComponentContainers: ExpectedComponentContainers{
			Agent: []string{containerTraceAgent},
		},
	},
	{
		Name:       "apm-use-localservice-values.yaml",
		ValuesFile: baseValuesDir + "/apm-use-localservice-values.yaml",
		ExpectedPods: defaultExpectedPods(),
		ExpectedComponentContainers: ExpectedComponentContainers{
			Agent: []string{containerTraceAgent},
		},
	},
	{
		Name:               "apm-logs-values.yaml",
		ValuesFile:         baseValuesDir + "/apm-logs-values.yaml",
		ExpectedPods:       defaultExpectedPods(),
		ExpectedComponentContainers: defaultExpectedComponentContainers(),
	},
	{
		Name:               "full-observability-values.yaml",
		ValuesFile:         baseValuesDir + "/full-observability-values.yaml",
		ExpectedPods:       defaultExpectedPods(),
		ExpectedComponentContainers: defaultExpectedComponentContainers(),
	},
	{
		Name:               "default-minimal.yaml",
		ValuesFile:         baseValuesDir + "/default-minimal.yaml",
		ExpectedPods:       defaultExpectedPods(),
		ExpectedComponentContainers: defaultExpectedComponentContainers(),
	},
	{
		Name:       "cluster-checks-values.yaml",
		ValuesFile: baseValuesDir + "/cluster-checks-values.yaml",
		ExpectedPods: ExpectedComponentPods{
			ClusterAgent:        1,
			ClusterChecksRunner: 2,
		},
		ExpectedComponentContainers: ExpectedComponentContainers{
			ClusterChecksRunner: []string{containerAgent},
		},
	},
	{
		Name:               "dogstatsd-hostport-values.yaml",
		ValuesFile:         baseValuesDir + "/dogstatsd-hostport-values.yaml",
		ExpectedPods:       defaultExpectedPods(),
		ExpectedComponentContainers: defaultExpectedComponentContainers(),
	},
	{
		Name:               "dogstatsd-uds-values.yaml",
		ValuesFile:         baseValuesDir + "/dogstatsd-uds-values.yaml",
		ExpectedPods:       defaultExpectedPods(),
		ExpectedComponentContainers: defaultExpectedComponentContainers(),
	},

	// Skipped tests (require kernel features not available in kind)
	{
		Name:               "npm-values.yaml",
		ValuesFile:         baseValuesDir + "/npm-values.yaml",
		ExpectedPods:       defaultExpectedPods(),
		ExpectedComponentContainers: defaultExpectedComponentContainers(),
		SkipReason:         "NPM requires kernel features not available in kind",
	},
	{
		Name:       "system-probe-checks-values.yaml",
		ValuesFile: baseValuesDir + "/system-probe-checks-values.yaml",
		ExpectedPods: defaultExpectedPods(),
		ExpectedComponentContainers: ExpectedComponentContainers{
			Agent: []string{containerSystemProbe},
		},
		SkipReason: "System probe requires kernel features not available in kind",
	},
}

// ### Test Cases with Dependencies
var testCasesWithDependencies = []ResourceDependentTestCase{
	{
		Name:               "global-credentials-existing-secret-values.yaml",
		ValuesFile:         "values/global-credentials-existing-secret-values.yaml",
		ExpectedPods:       defaultExpectedPods(),
		ExpectedComponentContainers: defaultExpectedComponentContainers(),
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
		ExpectedComponentContainers: defaultExpectedComponentContainers(),
		ConfigMaps: []ConfigMapDef{
			{
				Name: "cluster-agent-config",
				Data: map[string]string{"DD_LOG_LEVEL": "debug"},
			},
		},
		PriorityClasses: []PriorityClassDef{
			{
				Name:        "cluster-agent-critical",
				Value:       1000000000,
				Description: "Ensures Cluster Agent pods run with elevated priority",
			},
		},
	},
	{
		Name:       "override-cluster-checks-runner-values.yaml",
		ValuesFile: "values/override-cluster-checks-runner-values.yaml",
		ExpectedPods: ExpectedComponentPods{
			ClusterAgent:        1,
			ClusterChecksRunner: 2,
		},
		ExpectedComponentContainers: ExpectedComponentContainers{
			ClusterChecksRunner: []string{containerAgent},
		},
		ConfigMaps: []ConfigMapDef{
			{
				Name: "ccr-config",
				Data: map[string]string{"DD_LOG_LEVEL": "debug"},
			},
		},
		PriorityClasses: []PriorityClassDef{
			{
				Name:        "ccr-critical",
				Value:       1000000000,
				Description: "Ensures Cluster Checks Runner pods run with elevated priority",
			},
		},
	},
	{
		Name:               "override-node-agent-values.yaml",
		ValuesFile:         "values/override-node-agent-values.yaml",
		ExpectedPods:       defaultExpectedPods(),
		ExpectedComponentContainers: defaultExpectedComponentContainers(),
		PriorityClasses: []PriorityClassDef{
			{
				Name:        "datadog-agent-critical",
				Value:       1000000000,
				Description: "Ensures Datadog Agent pods run with elevated priority",
			},
		},
	},
	{
		Name:               "global-envfrom-values.yaml",
		ValuesFile:         "values/global-envfrom-values.yaml",
		ExpectedPods:       defaultExpectedPods(),
		ExpectedComponentContainers: defaultExpectedComponentContainers(),
		ConfigMaps: []ConfigMapDef{
			{
				Name: "datadog-env-config",
				Data: map[string]string{"DD_LOG_LEVEL": "debug", "DD_TAGS": "env:test"},
			},
		},
		Secrets: []SecretDef{
			{
				Name: "datadog-env-secrets",
				Data: map[string]string{"DD_SECRET_VAR": "secret-value"},
			},
		},
	},
}

// ### Negative Test Cases

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
