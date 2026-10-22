package datadog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"

	"github.com/DataDog/helm-charts/test/common"
)

const (
	ddDataSecurityEnabled        = "DD_DATA_SECURITY_ENABLED"
	ddSharedLibraryCheckEnabled  = "DD_SHARED_LIBRARY_CHECK_ENABLED"
	ddRemoteConfigurationEnabled = "DD_REMOTE_CONFIGURATION_ENABLED"
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

// datadog.dataSecurity.enabled sets both Data Security and the shared library check on the core agent.
func Test_dataSecurityEnabled(t *testing.T) {
	tests := []struct {
		name    string
		enabled string
		want    bool
	}{
		{name: "disabled", enabled: "false"},
		{name: "enabled", enabled: "true", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			daemonSet := renderDataSecurityDaemonSet(t, map[string]string{
				"datadog.dataSecurity.enabled": tt.enabled,
			})

			agent, found := getContainer(t, daemonSet.Spec.Template.Spec.Containers, "agent")
			require.True(t, found, "expected agent container")
			env := getEnvVarMap(agent.Env)

			if tt.want {
				assert.Equal(t, "true", env[ddDataSecurityEnabled])
				assert.Equal(t, "true", env[ddSharedLibraryCheckEnabled])
			} else {
				assert.NotContains(t, env, ddDataSecurityEnabled)
				assert.NotContains(t, env, ddSharedLibraryCheckEnabled)
			}
		})
	}
}

// Enabling Data Security turns Remote Configuration on for the node Agent, unless the global switch is off.
func Test_dataSecurityEnablesRemoteConfiguration(t *testing.T) {
	tests := []struct {
		name      string
		overrides map[string]string
		want      string
	}{
		{
			name: "enabled despite datadog.remoteConfiguration disabled",
			overrides: map[string]string{
				"datadog.dataSecurity.enabled":        "true",
				"datadog.remoteConfiguration.enabled": "false",
			},
			want: "true",
		},
		{
			name: "global remoteConfiguration remains a kill switch",
			overrides: map[string]string{
				"datadog.dataSecurity.enabled": "true",
				"remoteConfiguration.enabled":  "false",
			},
			want: "false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			daemonSet := renderDataSecurityDaemonSet(t, tt.overrides)

			agent, found := getContainer(t, daemonSet.Spec.Template.Spec.Containers, "agent")
			require.True(t, found, "expected agent container")
			assert.Equal(t, tt.want, getEnvVarMap(agent.Env)[ddRemoteConfigurationEnabled])
		})
	}
}
