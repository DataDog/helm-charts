package datadog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"

	"github.com/DataDog/helm-charts/test/common"
)

const (
	ddDataSecurityEnabled       = "DD_DATA_SECURITY_ENABLED"
	ddSharedLibraryCheckEnabled = "DD_SHARED_LIBRARY_CHECK_ENABLED"
)

func renderDataSecurityDaemonSet(t *testing.T, overrides map[string]string) appsv1.DaemonSet {
	t.Helper()

	base := map[string]string{
		"datadog.apiKeyExistingSecret": "datadog-secret",
		"datadog.appKeyExistingSecret": "datadog-secret",
	}
	for k, v := range overrides {
		base[k] = v
	}

	manifest, err := common.RenderChart(t, common.HelmCommand{
		ReleaseName: "datadog",
		ChartPath:   "../../charts/datadog",
		ShowOnly:    []string{"templates/daemonset.yaml"},
		Values:      []string{"../../charts/datadog/values.yaml"},
		Overrides:   base,
	})
	require.NoError(t, err)

	var daemonSet appsv1.DaemonSet
	common.Unmarshal(t, manifest, &daemonSet)
	return daemonSet
}

// Verifies that Data Security is disabled by default: the core agent does not get the env var.
func Test_dataSecurityDisabledByDefault(t *testing.T) {
	daemonSet := renderDataSecurityDaemonSet(t, nil)

	agent, found := getContainer(t, daemonSet.Spec.Template.Spec.Containers, "agent")
	require.True(t, found, "expected agent container")
	env := getEnvVarMap(agent.Env)
	assert.NotContains(t, env, ddDataSecurityEnabled)
	assert.NotContains(t, env, ddSharedLibraryCheckEnabled)
}

// Verifies that an explicit `false` behaves like the default.
func Test_dataSecurityExplicitlyDisabled(t *testing.T) {
	daemonSet := renderDataSecurityDaemonSet(t, map[string]string{
		"datadog.dataSecurity.enabled": "false",
	})

	for _, container := range daemonSet.Spec.Template.Spec.Containers {
		assert.NotContains(t, getEnvVarMap(container.Env), ddDataSecurityEnabled, "container %s", container.Name)
	}
}

// Verifies that enabling Data Security sets the env var on the core agent container only.
func Test_dataSecurityEnabled(t *testing.T) {
	daemonSet := renderDataSecurityDaemonSet(t, map[string]string{
		"datadog.dataSecurity.enabled": "true",
		// system-probe and security-agent containers are also rendered, to prove they are untouched.
		"datadog.networkMonitoring.enabled":     "true",
		"datadog.securityAgent.runtime.enabled": "true",
	})

	agent, found := getContainer(t, daemonSet.Spec.Template.Spec.Containers, "agent")
	require.True(t, found, "expected agent container")
	env := getEnvVarMap(agent.Env)
	assert.Equal(t, "true", env[ddDataSecurityEnabled])
	assert.NotContains(t, env, ddSharedLibraryCheckEnabled)

	for _, container := range daemonSet.Spec.Template.Spec.Containers {
		if container.Name == "agent" {
			continue
		}
		assert.NotContains(t, getEnvVarMap(container.Env), ddDataSecurityEnabled, "container %s", container.Name)
	}
}

// Verifies that shared library checks are disabled by default.
func Test_sharedLibraryCheckDisabledByDefault(t *testing.T) {
	daemonSet := renderDataSecurityDaemonSet(t, nil)

	for _, container := range daemonSet.Spec.Template.Spec.Containers {
		assert.NotContains(t, getEnvVarMap(container.Env), ddSharedLibraryCheckEnabled, "container %s", container.Name)
	}
}

// Verifies that enabling shared library checks sets the env var on the core agent container only.
func Test_sharedLibraryCheckEnabled(t *testing.T) {
	daemonSet := renderDataSecurityDaemonSet(t, map[string]string{
		"datadog.sharedLibraryCheck.enabled":    "true",
		"datadog.networkMonitoring.enabled":     "true",
		"datadog.securityAgent.runtime.enabled": "true",
	})

	agent, found := getContainer(t, daemonSet.Spec.Template.Spec.Containers, "agent")
	require.True(t, found, "expected agent container")
	assert.Equal(t, "true", getEnvVarMap(agent.Env)[ddSharedLibraryCheckEnabled])

	for _, container := range daemonSet.Spec.Template.Spec.Containers {
		if container.Name == "agent" {
			continue
		}
		assert.NotContains(t, getEnvVarMap(container.Env), ddSharedLibraryCheckEnabled, "container %s", container.Name)
	}
}
