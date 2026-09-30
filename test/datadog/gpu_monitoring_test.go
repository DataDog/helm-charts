package datadog

import (
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

func Test_gpuMonitoringEnableEbpfProbes(t *testing.T) {
	tests := []struct {
		name             string
		overrides        map[string]string
		expectKey        bool
		expectEbpfProbes bool
	}{
		{
			name: "privileged mode defaults to disabled eBPF probes",
			overrides: map[string]string{
				"datadog.gpuMonitoring.enabled":        "true",
				"datadog.gpuMonitoring.privilegedMode": "true",
			},
			expectKey:        true,
			expectEbpfProbes: false,
		},
		{
			name: "privileged mode with an explicit opt-in enables eBPF probes",
			overrides: map[string]string{
				"datadog.gpuMonitoring.enabled":          "true",
				"datadog.gpuMonitoring.privilegedMode":   "true",
				"datadog.gpuMonitoring.enableEbpfProbes": "true",
			},
			expectKey:        true,
			expectEbpfProbes: true,
		},
		{
			name: "non-privileged mode does not render the setting",
			overrides: map[string]string{
				"datadog.gpuMonitoring.enabled":          "true",
				"datadog.gpuMonitoring.enableEbpfProbes": "true",
			},
			expectKey: false,
		},
		{
			name: "privileged mode without GPU monitoring does not render the setting",
			overrides: map[string]string{
				"datadog.gpuMonitoring.enabled":          "false",
				"datadog.gpuMonitoring.privilegedMode":   "true",
				"datadog.gpuMonitoring.enableEbpfProbes": "true",
				// system-probe is pulled in by another feature here, so the gpu_monitoring block still renders.
				"datadog.networkMonitoring.enabled": "true",
			},
			expectKey: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := renderGpuMonitoringManifest(t, tt.overrides)

			daemonset := extractAgentDaemonset(t, manifest)
			agentContainer, hasAgent := getContainer(t, daemonset.Spec.Template.Spec.Containers, "agent")
			require.True(t, hasAgent, "expected the core agent container to be rendered")

			envValue, hasEnv := getEnvValue(agentContainer.Env, "DD_GPU_MONITORING_ENABLE_EBPF_PROBES")
			assert.Equal(t, tt.expectKey, hasEnv, "unexpected DD_GPU_MONITORING_ENABLE_EBPF_PROBES presence")

			systemProbeConfig, hasSystemProbeConfig := extractSystemProbeConfig(t, manifest)
			require.True(t, hasSystemProbeConfig, "expected system-probe config to render")
			gpuConfig, found := nestedMap(systemProbeConfig, "gpu_monitoring")
			require.True(t, found, "expected gpu_monitoring block in system-probe config")
			configValue, hasConfigKey := gpuConfig["enable_ebpf_probes"]
			assert.Equal(t, tt.expectKey, hasConfigKey, "unexpected gpu_monitoring.enable_ebpf_probes presence")

			if !tt.expectKey {
				return
			}

			expected := "false"
			if tt.expectEbpfProbes {
				expected = "true"
			}
			assert.Equal(t, expected, envValue, "unexpected DD_GPU_MONITORING_ENABLE_EBPF_PROBES value")
			assert.Equal(t, tt.expectEbpfProbes, configValue, "unexpected gpu_monitoring.enable_ebpf_probes value")
		})
	}
}

func renderGpuMonitoringManifest(t *testing.T, overrides map[string]string) string {
	t.Helper()

	merged := map[string]string{
		"datadog.apiKeyExistingSecret": "datadog-secret",
		"datadog.appKeyExistingSecret": "datadog-secret",
	}
	for key, value := range overrides {
		merged[key] = value
	}

	manifest, err := common.RenderChart(t, common.HelmCommand{
		ReleaseName: "datadog",
		ChartPath:   "../../charts/datadog",
		Values:      []string{"../../charts/datadog/values.yaml"},
		Overrides:   merged,
	})
	require.NoError(t, err, "couldn't render chart")
	return manifest
}

func getEnvValue(envs []corev1.EnvVar, name string) (string, bool) {
	for _, env := range envs {
		if env.Name == name {
			return env.Value, true
		}
	}
	return "", false
}
