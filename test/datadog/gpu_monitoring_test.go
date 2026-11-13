package datadog

import (
	"strconv"
	"strings"
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
		name               string
		overrides          map[string]string
		expectAMD          bool
		expectRuntimeClass string
	}{
		{
			name: "amd.enabled with the runtime class unset sets the AMD env vars and no runtime class",
			overrides: map[string]string{
				"datadog.gpuMonitoring.enabled":     "true",
				"datadog.gpuMonitoring.amd.enabled": "true",
				"agents.image.tag":                  "7.86.0",
			},
			expectAMD:          true,
			expectRuntimeClass: "",
		},
		{
			name: "amd.enabled with the nvidia runtime class sets the AMD env vars and keeps it",
			overrides: map[string]string{
				"datadog.gpuMonitoring.enabled":          "true",
				"datadog.gpuMonitoring.amd.enabled":      "true",
				"datadog.gpuMonitoring.runtimeClassName": "nvidia",
				"agents.image.tag":                       "7.86.0",
			},
			expectAMD:          true,
			expectRuntimeClass: "nvidia",
		},
		{
			name: "amd.enabled with an empty runtime class sets the AMD env vars and no runtime class",
			overrides: map[string]string{
				"datadog.gpuMonitoring.enabled":          "true",
				"datadog.gpuMonitoring.amd.enabled":      "true",
				"datadog.gpuMonitoring.runtimeClassName": "",
				"agents.image.tag":                       "7.86.0",
			},
			expectAMD:          true,
			expectRuntimeClass: "",
		},
		{
			name: "amd.enabled with the amd runtime class sets the AMD env vars and keeps it",
			overrides: map[string]string{
				"datadog.gpuMonitoring.enabled":          "true",
				"datadog.gpuMonitoring.amd.enabled":      "true",
				"datadog.gpuMonitoring.runtimeClassName": "amd",
				"agents.image.tag":                       "7.86.0",
			},
			expectAMD:          true,
			expectRuntimeClass: "amd",
		},
		{
			name: "amd runtime class without amd.enabled skips the AMD env vars",
			overrides: map[string]string{
				"datadog.gpuMonitoring.enabled":          "true",
				"datadog.gpuMonitoring.runtimeClassName": "amd",
				"agents.image.tag":                       "7.86.0",
			},
			expectRuntimeClass: "amd",
		},
		{
			name: "amd.enabled false with the amd runtime class skips the AMD env vars",
			overrides: map[string]string{
				"datadog.gpuMonitoring.enabled":          "true",
				"datadog.gpuMonitoring.amd.enabled":      "false",
				"datadog.gpuMonitoring.runtimeClassName": "amd",
				"agents.image.tag":                       "7.86.0",
			},
			expectRuntimeClass: "amd",
		},
		{
			name: "default values on Agent 7.86.0 skip the AMD env vars",
			overrides: map[string]string{
				"datadog.gpuMonitoring.enabled": "true",
				"agents.image.tag":              "7.86.0",
			},
			expectRuntimeClass: "nvidia",
		},
		{
			name: "empty runtime class without amd.enabled sets no runtime class",
			overrides: map[string]string{
				"datadog.gpuMonitoring.enabled":          "true",
				"datadog.gpuMonitoring.runtimeClassName": "",
				"agents.image.tag":                       "7.86.0",
			},
			expectRuntimeClass: "",
		},
		{
			// The RuntimeClass default follows amd.enabled, not the version gate.
			name: "chart default Agent version skips the AMD env vars and sets no runtime class",
			overrides: map[string]string{
				"datadog.gpuMonitoring.enabled":     "true",
				"datadog.gpuMonitoring.amd.enabled": "true",
			},
			expectRuntimeClass: "",
		},
		{
			name: "GPU monitoring disabled renders nothing",
			overrides: map[string]string{
				"datadog.gpuMonitoring.enabled":     "false",
				"datadog.gpuMonitoring.amd.enabled": "true",
				"agents.image.tag":                  "7.86.0",
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
			assertAMDEnv(t, agentContainer.Env, tt.expectAMD)

			_, hasNvidiaDevices := getEnvValue(agentContainer.Env, "NVIDIA_VISIBLE_DEVICES")
			gpuEnabled := tt.overrides["datadog.gpuMonitoring.enabled"] == "true"
			assert.Equal(t, gpuEnabled, hasNvidiaDevices, "unexpected NVIDIA_VISIBLE_DEVICES presence")
		})
	}
}

func Test_gpuMonitoringAMDVersionGate(t *testing.T) {
	tests := []struct {
		name          string
		imageTag      string
		doNotCheckTag bool
		expectAMD     bool
	}{
		{name: "older stable release", imageTag: "7.85.0"},
		{name: "floating tag resolves to the chart default", imageTag: "latest"},
		{name: "release candidate", imageTag: "7.86.0-rc.3", expectAMD: true},
		{name: "first supported stable release", imageTag: "7.86.0", expectAMD: true},
		{name: "full image variant", imageTag: "7.86.0-full", expectAMD: true},
		{name: "unchecked custom tag", imageTag: "custom-amd-build", doNotCheckTag: true, expectAMD: true},
		{name: "unchecked older tag", imageTag: "7.85.0", doNotCheckTag: true, expectAMD: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			daemonset, err := renderGpuMonitoringDaemonset(t, map[string]string{
				"datadog.gpuMonitoring.enabled":     "true",
				"datadog.gpuMonitoring.amd.enabled": "true",
				"agents.image.tag":                  tt.imageTag,
				"agents.image.doNotCheckTag":        strconv.FormatBool(tt.doNotCheckTag),
			})
			require.NoError(t, err, "below the version gate the AMD env vars are skipped, not an error")
			assert.Nil(t, daemonset.Spec.Template.Spec.RuntimeClassName, "the RuntimeClass default follows amd.enabled, not the version gate")

			agentContainer, hasAgent := getContainer(t, daemonset.Spec.Template.Spec.Containers, "agent")
			require.True(t, hasAgent, "expected the core agent container to be rendered")
			assertAMDEnv(t, agentContainer.Env, tt.expectAMD)
		})
	}
}

func Test_gpuMonitoringAMDUserEnvWins(t *testing.T) {
	// The user's values are strings, as in a values file: --set would turn "false" into a bool.
	userEnvList := `[{"name":"DD_GPU_AMD_ENABLED","value":"false"},{"name":"AMD_VISIBLE_DEVICES","value":"1"}]`
	userEnvDict := `{"DD_GPU_AMD_ENABLED":"false","AMD_VISIBLE_DEVICES":"1"}`
	for _, key := range []string{"datadog.env", "datadog.envDict", "agents.containers.agent.env", "agents.containers.agent.envDict"} {
		t.Run(key, func(t *testing.T) {
			value := userEnvList
			if strings.HasSuffix(key, "envDict") {
				value = userEnvDict
			}
			daemonset, err := renderGpuMonitoringDaemonsetWithJSON(t, map[string]string{
				"datadog.gpuMonitoring.enabled":     "true",
				"datadog.gpuMonitoring.amd.enabled": "true",
				"agents.image.tag":                  "7.86.0",
			}, map[string]string{key: value})
			require.NoError(t, err, "couldn't render chart")

			agentContainer, hasAgent := getContainer(t, daemonset.Spec.Template.Spec.Containers, "agent")
			require.True(t, hasAgent, "expected the core agent container to be rendered")
			assert.Equal(t, []string{"false"}, envValues(agentContainer.Env, "DD_GPU_AMD_ENABLED"))
			assert.Equal(t, []string{"1"}, envValues(agentContainer.Env, "AMD_VISIBLE_DEVICES"))
		})
	}
}

func Test_gpuMonitoringAMDUnsupportedProviders(t *testing.T) {
	for _, provider := range []string{"providers.gke.autopilot", "providers.gke.gdc"} {
		t.Run(provider, func(t *testing.T) {
			daemonset, err := renderGpuMonitoringDaemonset(t, map[string]string{
				"datadog.gpuMonitoring.enabled":     "true",
				"datadog.gpuMonitoring.amd.enabled": "true",
				"agents.image.tag":                  "7.86.0",
				provider:                            "true",
			})
			require.NoError(t, err, "GKE Autopilot and GDC skip the AMD env vars, not an error")
			assert.Equal(t, "nvidia", ptrStringValue(daemonset.Spec.Template.Spec.RuntimeClassName), "amd.enabled is ignored, so the nvidia default is kept")

			agentContainer, hasAgent := getContainer(t, daemonset.Spec.Template.Spec.Containers, "agent")
			require.True(t, hasAgent, "expected the core agent container to be rendered")
			assertAMDEnv(t, agentContainer.Env, false)
		})
	}
}

func assertAMDEnv(t *testing.T, envs []corev1.EnvVar, expectAMD bool) {
	t.Helper()
	if expectAMD {
		assert.Equal(t, []string{"true"}, envValues(envs, "DD_GPU_AMD_ENABLED"))
		assert.Equal(t, []string{"all"}, envValues(envs, "AMD_VISIBLE_DEVICES"))
		return
	}
	assert.Empty(t, envValues(envs, "DD_GPU_AMD_ENABLED"), "unexpected DD_GPU_AMD_ENABLED")
	assert.Empty(t, envValues(envs, "AMD_VISIBLE_DEVICES"), "unexpected AMD_VISIBLE_DEVICES")
}

// envValues returns the value of every env entry named name, to catch duplicates.
func envValues(envs []corev1.EnvVar, name string) []string {
	var values []string
	for _, env := range envs {
		if env.Name == name {
			values = append(values, env.Value)
		}
	}
	return values
}

// renderGpuMonitoringDaemonset decodes the agent DaemonSet with the Kubernetes field names, unlike extractAgentDaemonset.
func renderGpuMonitoringDaemonset(t *testing.T, overrides map[string]string) (appsv1.DaemonSet, error) {
	t.Helper()
	return renderGpuMonitoringDaemonsetWithJSON(t, overrides, nil)
}

func renderGpuMonitoringDaemonsetWithJSON(t *testing.T, overrides, jsonOverrides map[string]string) (appsv1.DaemonSet, error) {
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
		ReleaseName:   "datadog",
		ChartPath:     "../../charts/datadog",
		ShowOnly:      []string{"templates/daemonset.yaml"},
		Values:        []string{"../../charts/datadog/values.yaml"},
		Overrides:     merged,
		OverridesJson: jsonOverrides,
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
