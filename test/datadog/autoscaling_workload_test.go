package datadog

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupHelmRepos(t *testing.T) {
	cmd := exec.Command("helm", "repo", "add", "datadog", "https://helm.datadoghq.com", "--force-update")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("helm repo add output (may already exist):\n%s", string(output))
	}

	cmd = exec.Command("helm", "repo", "add", "prometheus-community", "https://prometheus-community.github.io/helm-charts", "--force-update")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Logf("helm repo add output (may already exist):\n%s", string(output))
	}

	cmd = exec.Command("helm", "repo", "update")
	cmd.Run()
}

func TestNoConditionPathWarning(t *testing.T) {
	setupHelmRepos(t)

	chartPath, err := filepath.Abs("../../charts/datadog")
	require.NoError(t, err)

	cmd := exec.Command("helm", "dependency", "build", chartPath)
	cmd.Dir = chartPath
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	assert.NotContains(t, outputStr, "Condition path", "should not emit 'Condition path' warning")
	assert.NotContains(t, outputStr, "returned non-bool value", "should not emit 'non-bool value' warning")

	if err != nil {
		t.Logf("helm dependency build output:\n%s", outputStr)
	}
	require.NoError(t, err, "helm dependency build should succeed")
}

func TestAutoscalingWorkloadEnabledDefaultNoWarning(t *testing.T) {
	setupHelmRepos(t)

	chartPath, err := filepath.Abs("../../charts/datadog")
	require.NoError(t, err)

	cmd := exec.Command("helm", "template", "datadog", chartPath)
	cmd.Dir = chartPath
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	assert.NotContains(t, outputStr, "Condition path", "should not emit 'Condition path' warning")
	assert.NotContains(t, outputStr, "returned non-bool value", "should not emit 'non-bool value' warning")

	if err != nil {
		t.Logf("helm template output:\n%s", outputStr)
	}
	require.NoError(t, err, "helm template should succeed")
}

func TestDatadogCrdsConditionWithMetricsProvider(t *testing.T) {
	chartPath, err := filepath.Abs("../../charts/datadog")
	require.NoError(t, err)

	t.Run("CRDs included when metricsProvider enabled", func(t *testing.T) {
		cmd := exec.Command("helm", "template", "datadog", chartPath,
			"--set", "clusterAgent.metricsProvider.useDatadogMetrics=true",
			"--set", "datadog.apiKeyExistingSecret=test",
			"--set", "datadog.appKeyExistingSecret=test")
		cmd.Dir = chartPath
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Logf("helm template output:\n%s", string(output))
		}
		require.NoError(t, err)

		outputStr := string(output)
		assert.Contains(t, outputStr, "DatadogMetrics", "CRDs should be included when clusterAgent.metricsProvider.useDatadogMetrics is true")
	})

	t.Run("no warning when using default values", func(t *testing.T) {
		cmd := exec.Command("helm", "template", "datadog", chartPath,
			"--set", "datadog.apiKeyExistingSecret=test",
			"--set", "datadog.appKeyExistingSecret=test")
		cmd.Dir = chartPath
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Logf("helm template output:\n%s", string(output))
		}
		require.NoError(t, err)

		outputStr := string(output)
		assert.NotContains(t, outputStr, "Condition path", "should not emit 'Condition path' warning with default values")
		assert.NotContains(t, outputStr, "returned non-bool value", "should not emit 'non-bool value' warning with default values")
	})
}

func TestRequirementsYamlCondition(t *testing.T) {
	requirementsPath, err := filepath.Abs("../../charts/datadog/requirements.yaml")
	require.NoError(t, err)

	content, err := os.ReadFile(requirementsPath)
	require.NoError(t, err)

	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")

	inDatadogCrdsSection := false
	for _, line := range lines {
		if strings.Contains(line, "name: datadog-crds") {
			inDatadogCrdsSection = true
		}
		if inDatadogCrdsSection && strings.Contains(line, "condition:") {
			assert.Contains(t, line, "datadog.autoscaling.workload.enabled",
				"requirements.yaml should include datadog.autoscaling.workload.enabled in condition")
			assert.Contains(t, line, "clusterAgent.metricsProvider.useDatadogMetrics",
				"requirements.yaml should include clusterAgent.metricsProvider.useDatadogMetrics in condition")
			return
		}
		if inDatadogCrdsSection && strings.HasPrefix(strings.TrimSpace(line), "name:") && !strings.Contains(line, "datadog-crds") {
			t.Fatal("datadog-crds condition not found")
		}
	}
	t.Fatal("datadog-crds dependency not found in requirements.yaml")
}

func TestAutoscalingWorkloadEnabledNotInValuesYaml(t *testing.T) {
	valuesPath, err := filepath.Abs("../../charts/datadog/values.yaml")
	require.NoError(t, err)

	content, err := os.ReadFile(valuesPath)
	require.NoError(t, err)

	contentStr := string(content)
	assert.NotContains(t, contentStr, "autoscaling:\n    workload:\n      enabled:",
		"datadog.autoscaling.workload.enabled should not be documented in values.yaml to avoid Helm warning")
	assert.NotContains(t, contentStr, "autoscaling:\n    workload:\n      #",
		"autoscaling.workload.enabled section should be removed from values.yaml")
}

func TestAutoscalingWorkloadEnabledCanBeSetViaCli(t *testing.T) {
	setupHelmRepos(t)

	chartPath, err := filepath.Abs("../../charts/datadog")
	require.NoError(t, err)

	cmd := exec.Command("helm", "template", "datadog", chartPath,
		"--set", "datadog.autoscaling.workload.enabled=true",
		"--set", "datadog.apiKeyExistingSecret=test",
		"--set", "datadog.appKeyExistingSecret=test")
	cmd.Dir = chartPath
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Logf("helm template output:\n%s", string(output))
	}
	require.NoError(t, err)

	outputStr := string(output)
	assert.NotContains(t, outputStr, "Condition path", "should not emit 'Condition path' warning when explicitly set")
	assert.NotContains(t, outputStr, "returned non-bool value", "should not emit 'non-bool value' warning when explicitly set")
}
