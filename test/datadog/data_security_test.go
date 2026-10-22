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

// Data Security is set only when both flags are true. The shared library check follows its own flag.
func Test_dataSecurityFlags(t *testing.T) {
	tests := []struct {
		name               string
		dataSecurity       string
		sharedLibraryCheck string
		wantDataSecurity   bool
		wantSharedLibrary  bool
	}{
		{name: "both disabled", dataSecurity: "false", sharedLibraryCheck: "false"},
		{name: "data security only", dataSecurity: "true", sharedLibraryCheck: "false"},
		{name: "shared library check only", dataSecurity: "false", sharedLibraryCheck: "true", wantSharedLibrary: true},
		{name: "both enabled", dataSecurity: "true", sharedLibraryCheck: "true", wantDataSecurity: true, wantSharedLibrary: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			daemonSet := renderDataSecurityDaemonSet(t, map[string]string{
				"datadog.dataSecurity.enabled":       tt.dataSecurity,
				"datadog.sharedLibraryCheck.enabled": tt.sharedLibraryCheck,
			})

			agent, found := getContainer(t, daemonSet.Spec.Template.Spec.Containers, "agent")
			require.True(t, found, "expected agent container")
			env := getEnvVarMap(agent.Env)

			if tt.wantDataSecurity {
				assert.Equal(t, "true", env[ddDataSecurityEnabled])
			} else {
				assert.NotContains(t, env, ddDataSecurityEnabled)
			}
			if tt.wantSharedLibrary {
				assert.Equal(t, "true", env[ddSharedLibraryCheckEnabled])
			} else {
				assert.NotContains(t, env, ddSharedLibraryCheckEnabled)
			}
		})
	}
}
