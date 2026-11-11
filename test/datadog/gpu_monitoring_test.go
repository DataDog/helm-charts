package datadog

import (
	"strconv"
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
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

func Test_gpuMonitoringAMD(t *testing.T) {
	tests := []struct {
		name                string
		overrides           map[string]string
		expectAMD           bool
		expectRuntimeClass  string
		expectNvidiaDevices bool
	}{
		{
			name: "AMD enabled sets the env vars and keeps the nvidia runtime class default",
			overrides: map[string]string{
				"datadog.gpuMonitoring.enabled":     "true",
				"datadog.gpuMonitoring.amd.enabled": "true",
				"agents.image.tag":                  "7.86.0",
			},
			expectAMD:           true,
			expectRuntimeClass:  "nvidia",
			expectNvidiaDevices: true,
		},
		{
			name: "amd runtime class is used with AMD",
			overrides: map[string]string{
				"datadog.gpuMonitoring.enabled":          "true",
				"datadog.gpuMonitoring.amd.enabled":      "true",
				"datadog.gpuMonitoring.runtimeClassName": "amd",
				"agents.image.tag":                       "7.86.0",
			},
			expectAMD:           true,
			expectRuntimeClass:  "amd",
			expectNvidiaDevices: true,
		},
		{
			name: "empty runtime class uses the default runtime class with AMD",
			overrides: map[string]string{
				"datadog.gpuMonitoring.enabled":          "true",
				"datadog.gpuMonitoring.amd.enabled":      "true",
				"datadog.gpuMonitoring.runtimeClassName": "",
				"agents.image.tag":                       "7.86.0",
			},
			expectAMD:           true,
			expectRuntimeClass:  "",
			expectNvidiaDevices: true,
		},
		{
			name: "custom runtime class is used with AMD",
			overrides: map[string]string{
				"datadog.gpuMonitoring.enabled":          "true",
				"datadog.gpuMonitoring.amd.enabled":      "true",
				"datadog.gpuMonitoring.runtimeClassName": "my-amd",
				"agents.image.tag":                       "7.86.0",
			},
			expectAMD:           true,
			expectRuntimeClass:  "my-amd",
			expectNvidiaDevices: true,
		},
		{
			name: "AMD disabled keeps the NVIDIA setup",
			overrides: map[string]string{
				"datadog.gpuMonitoring.enabled":     "true",
				"datadog.gpuMonitoring.amd.enabled": "false",
			},
			expectRuntimeClass:  "nvidia",
			expectNvidiaDevices: true,
		},
		{
			name: "AMD without GPU monitoring does nothing",
			overrides: map[string]string{
				"datadog.gpuMonitoring.enabled":     "false",
				"datadog.gpuMonitoring.amd.enabled": "true",
				"agents.image.tag":                  "7.85.0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			daemonset, err := renderGpuMonitoringDaemonset(t, tt.overrides)
			require.NoError(t, err, "couldn't render chart")
			assert.Equal(t, tt.expectRuntimeClass, ptrStringValue(daemonset.Spec.Template.Spec.RuntimeClassName), "unexpected runtimeClassName")

			agentContainer, hasAgent := getContainer(t, daemonset.Spec.Template.Spec.Containers, "agent")
			require.True(t, hasAgent, "expected the core agent container to be rendered")

			amdEnabled, hasAMDEnabled := getEnvValue(agentContainer.Env, "DD_GPU_AMD_ENABLED")
			amdDevices, hasAMDDevices := getEnvValue(agentContainer.Env, "AMD_VISIBLE_DEVICES")
			assert.Equal(t, tt.expectAMD, hasAMDEnabled, "unexpected DD_GPU_AMD_ENABLED presence")
			assert.Equal(t, tt.expectAMD, hasAMDDevices, "unexpected AMD_VISIBLE_DEVICES presence")
			if tt.expectAMD {
				assert.Equal(t, "true", amdEnabled)
				assert.Equal(t, "all", amdDevices)
			}

			_, hasNvidiaDevices := getEnvValue(agentContainer.Env, "NVIDIA_VISIBLE_DEVICES")
			assert.Equal(t, tt.expectNvidiaDevices, hasNvidiaDevices, "unexpected NVIDIA_VISIBLE_DEVICES presence")
		})
	}
}

func Test_gpuMonitoringAMDDisabledMatchesDefault(t *testing.T) {
	// A fixed token keeps the generated Cluster Agent token and its checksum stable across renders.
	base := map[string]string{
		"datadog.gpuMonitoring.enabled": "true",
		"clusterAgent.token":            "0123456789abcdef0123456789abcdef",
	}
	disabled := map[string]string{
		"datadog.gpuMonitoring.enabled":     "true",
		"datadog.gpuMonitoring.amd.enabled": "false",
		"clusterAgent.token":                "0123456789abcdef0123456789abcdef",
	}
	baseDaemonset, err := renderGpuMonitoringDaemonset(t, base)
	require.NoError(t, err, "couldn't render chart")
	disabledDaemonset, err := renderGpuMonitoringDaemonset(t, disabled)
	require.NoError(t, err, "couldn't render chart")
	assert.Equal(t, baseDaemonset, disabledDaemonset)
}

func Test_gpuMonitoringAMDVersionGate(t *testing.T) {
	const versionError = "datadog.gpuMonitoring.amd.enabled requires Datadog Agent 7.86.0 or newer"

	tests := []struct {
		name          string
		imageTag      string
		doNotCheckTag bool
		expectError   bool
	}{
		{name: "older stable release", imageTag: "7.85.0", expectError: true},
		{name: "floating tag resolves to the chart default", imageTag: "latest", expectError: true},
		{name: "release candidate", imageTag: "7.86.0-rc.3"},
		{name: "first supported stable release", imageTag: "7.86.0"},
		{name: "full image variant", imageTag: "7.86.0-full"},
		{name: "unchecked custom tag", imageTag: "custom-amd-build", doNotCheckTag: true},
		{name: "unchecked older tag", imageTag: "7.85.0", doNotCheckTag: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			daemonset, err := renderGpuMonitoringDaemonset(t, map[string]string{
				"datadog.gpuMonitoring.enabled":          "true",
				"datadog.gpuMonitoring.amd.enabled":      "true",
				"datadog.gpuMonitoring.runtimeClassName": "amd",
				"agents.image.tag":                       tt.imageTag,
				"agents.image.doNotCheckTag":             strconv.FormatBool(tt.doNotCheckTag),
			})

			if tt.expectError {
				require.ErrorContains(t, err, versionError)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, "amd", ptrStringValue(daemonset.Spec.Template.Spec.RuntimeClassName))
		})
	}
}

// renderGpuMonitoringDaemonset decodes the agent DaemonSet with the Kubernetes field names, unlike extractAgentDaemonset.
func renderGpuMonitoringDaemonset(t *testing.T, overrides map[string]string) (appsv1.DaemonSet, error) {
	t.Helper()

	merged := map[string]string{
		"datadog.apiKeyExistingSecret": "datadog-secret",
		"datadog.appKeyExistingSecret": "datadog-secret",
	}
	for key, value := range overrides {
		merged[key] = value
	}

	var daemonset appsv1.DaemonSet
	manifest, err := common.RenderChart(t, common.HelmCommand{
		ReleaseName: "datadog",
		ChartPath:   "../../charts/datadog",
		ShowOnly:    []string{"templates/daemonset.yaml"},
		Values:      []string{"../../charts/datadog/values.yaml"},
		Overrides:   merged,
	})
	if err != nil {
		return daemonset, err
	}
	common.Unmarshal(t, manifest, &daemonset)
	return daemonset, nil
}

func ptrStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
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

func Test_gpuMonitoringAMDUnsupportedProviders(t *testing.T) {
	for _, provider := range []string{"providers.gke.autopilot", "providers.gke.gdc"} {
		t.Run(provider, func(t *testing.T) {
			_, err := renderGpuMonitoringDaemonset(t, map[string]string{
				"datadog.gpuMonitoring.enabled":     "true",
				"datadog.gpuMonitoring.amd.enabled": "true",
				"agents.image.tag":                  "7.86.0",
				provider:                            "true",
			})
			require.Error(t, err)
			assert.Contains(t, err.Error(), "datadog.gpuMonitoring.amd.enabled is not supported on GKE Autopilot or GDC")
		})
	}
}
