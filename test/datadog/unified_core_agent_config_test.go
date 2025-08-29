package datadog

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"

	"github.com/DataDog/helm-charts/test/common"
)

// Test_UnifiedCoreAgentConfig tests that all DD-prefixed environment variables from process-agent and trace-agent
// are also set in the core agent with the same values.
// This is to ensure that the process-specific (process-agent and trace-agent) configurations are correctly propagated to the core agent for metadata payload.
func Test_UnifiedCoreAgentConfig(t *testing.T) {
	tests := []struct {
		name       string
		command    common.HelmCommand
		assertFunc func(t *testing.T, manifest string)
	}{
		{
			name: "DD-prefixed env vars in core agent",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret":      "datadog-secret",
					"datadog.appKeyExistingSecret":      "datadog-secret",
					"datadog.networkMonitoring.enabled": "true", // Required to enable process-agent
				},
			},
			assertFunc: verifyEnvVarsSync,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest, err := common.RenderChart(t, tt.command)
			assert.Nil(t, err, "couldn't render template")
			tt.assertFunc(t, manifest)
		})
	}
}

// verifyEnvVarsSync checks that all DD-prefixed environment variables from process-agent and trace-agent
// are also set in the core agent with the same values.
func verifyEnvVarsSync(t *testing.T, manifest string) {
	var daemonset appsv1.DaemonSet
	common.Unmarshal(t, manifest, &daemonset)

	// Get containers
	coreAgentContainer, coreAgentFound := getContainer(t, daemonset.Spec.Template.Spec.Containers, "agent")
	assert.True(t, coreAgentFound, "Core agent container not found")

	processAgentContainer, processAgentFound := getContainer(t, daemonset.Spec.Template.Spec.Containers, "process-agent")
	assert.True(t, processAgentFound, "Process agent container not found")

	traceAgentContainer, traceAgentFound := getContainer(t, daemonset.Spec.Template.Spec.Containers, "trace-agent")
	assert.True(t, traceAgentFound, "Trace agent container not found")

	// Get environment variable maps
	coreAgentEnvMap := getEnvVarMap(coreAgentContainer.Env)
	processAgentEnvMap := getEnvVarMap(processAgentContainer.Env)
	traceAgentEnvMap := getEnvVarMap(traceAgentContainer.Env)

	// Check that all DD-prefixed env vars from process agent are in core agent with the same value
	var missingOrDifferentVars []string

	for envName, envValue := range processAgentEnvMap {
		if strings.HasPrefix(envName, "DD_") {

			coreValue, exists := coreAgentEnvMap[envName]
			if !exists {
				missingOrDifferentVars = append(missingOrDifferentVars,
					fmt.Sprintf("%s (missing from core-agent)", envName))
			} else if coreValue != envValue {
				missingOrDifferentVars = append(missingOrDifferentVars,
					fmt.Sprintf("%s (value differs: process-agent=%q, core-agent=%q)",
						envName, envValue, coreValue))
			}
		}
	}

	// Check that all DD-prefixed env vars from trace agent are in core agent with the same value
	for envName, envValue := range traceAgentEnvMap {
		if strings.HasPrefix(envName, "DD_") {

			coreValue, exists := coreAgentEnvMap[envName]
			if !exists {
				missingOrDifferentVars = append(missingOrDifferentVars,
					fmt.Sprintf("%s (missing from core-agent)", envName))
			} else if coreValue != envValue {
				missingOrDifferentVars = append(missingOrDifferentVars,
					fmt.Sprintf("%s (value differs: trace-agent=%q, core-agent=%q)",
						envName, envValue, coreValue))
			}
		}
	}

	// Assert that all required variables are synced correctly
	assert.Empty(t, missingOrDifferentVars,
		"Found missing or different DD-prefixed environment variables in core agent: %v",
		missingOrDifferentVars)
}
