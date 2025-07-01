package datadog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"

	"github.com/DataDog/helm-charts/test/common"
)

const (
	DDAgentIpcPort                  = "DD_AGENT_IPC_PORT"
	DDAgentIpcConfigRefreshInterval = "DD_AGENT_IPC_CONFIG_REFRESH_INTERVAL"
)

type ExpectedIpcEnv struct {
	ipcPort                  string
	ipcConfigRefreshInterval string
}

func Test_otelAgentConfigs(t *testing.T) {
	tests := []struct {
		name           string
		command        common.HelmCommand
		assertions     func(t *testing.T, manifest string, expectedIpcEnv ExpectedIpcEnv)
		expectedIpcEnv ExpectedIpcEnv
	}{
		{
			name: "no ipc provided",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret":  "datadog-secret",
					"datadog.appKeyExistingSecret":  "datadog-secret",
					"datadog.otelCollector.enabled": "true",
				},
			},
			expectedIpcEnv: ExpectedIpcEnv{
				ipcPort:                  "5009",
				ipcConfigRefreshInterval: "60",
			},
			assertions: verifyOtelAgentEnvVars,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest, err := common.RenderChart(t, tt.command)
			assert.Nil(t, err, "couldn't render template")
			tt.assertions(t, manifest, tt.expectedIpcEnv)
		})
	}
}

func verifyOtelAgentEnvVars(t *testing.T, manifest string, expectedIpcEnv ExpectedIpcEnv) {
	var deployment appsv1.DaemonSet
	common.Unmarshal(t, manifest, &deployment)
	// otel agent
	otelAgentContainer, ok := getContainer(t, deployment.Spec.Template.Spec.Containers, "otel-agent")
	assert.True(t, ok)
	coreEnvs := getEnvVarMap(otelAgentContainer.Env)
	assert.Equal(t, expectedIpcEnv.ipcPort, coreEnvs[DDAgentIpcPort])
	assert.Equal(t, expectedIpcEnv.ipcConfigRefreshInterval, coreEnvs[DDAgentIpcConfigRefreshInterval])

	// core agent
	coreAgentContainer, ok := getContainer(t, deployment.Spec.Template.Spec.Containers, "agent")
	assert.True(t, ok)
	coreEnvs = getEnvVarMap(coreAgentContainer.Env)
	assert.Equal(t, expectedIpcEnv.ipcPort, coreEnvs[DDAgentIpcPort])
	assert.Equal(t, expectedIpcEnv.ipcConfigRefreshInterval, coreEnvs[DDAgentIpcConfigRefreshInterval])
}

func Test_ddotCollectorImage(t *testing.T) {
	tests := []struct {
		name         string
		command      common.HelmCommand
		expectError  bool
		errorMessage string
		assertion    func(t *testing.T, manifest string)
	}{
		{
			name: "useStandaloneImage true with agent version 7.67.0",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret":             "datadog-secret",
					"datadog.appKeyExistingSecret":             "datadog-secret",
					"datadog.otelCollector.enabled":            "true",
					"datadog.otelCollector.useStandaloneImage": "true",
					"agents.image.tag":                         "7.67.0",
				},
			},
			expectError: false,
			assertion: func(t *testing.T, manifest string) {
				verifyOtelImage(t, manifest, "gcr.io/datadoghq/ddot-collector:7.67.0")
			},
		},
		{
			name: "useStandaloneImage true with agent version 7.68.0",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret":             "datadog-secret",
					"datadog.appKeyExistingSecret":             "datadog-secret",
					"datadog.otelCollector.enabled":            "true",
					"datadog.otelCollector.useStandaloneImage": "true",
					"agents.image.tag":                         "7.68.0",
				},
			},
			expectError: false,
			assertion: func(t *testing.T, manifest string) {
				verifyOtelImage(t, manifest, "gcr.io/datadoghq/ddot-collector:7.68.0")
			},
		},
		{
			name: "useStandaloneImage true with agent version 7.66.0 should fail",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret":             "datadog-secret",
					"datadog.appKeyExistingSecret":             "datadog-secret",
					"datadog.otelCollector.enabled":            "true",
					"datadog.otelCollector.useStandaloneImage": "true",
					"agents.image.tag":                         "7.66.0",
				},
			},
			expectError:  true,
			errorMessage: "datadog.otelCollector.useStandaloneImage is only supported for agent versions 7.67.0+",
		},
		{
			name: "useStandaloneImage false with tagSuffix full",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret":             "datadog-secret",
					"datadog.appKeyExistingSecret":             "datadog-secret",
					"datadog.otelCollector.enabled":            "true",
					"datadog.otelCollector.useStandaloneImage": "false",
					"agents.image.tagSuffix":                   "full",
					"agents.image.tag":                         "7.66.0",
				},
			},
			expectError: false,
			assertion: func(t *testing.T, manifest string) {
				verifyOtelImage(t, manifest, "gcr.io/datadoghq/agent:7.66.0-full")
			},
		},
		{
			name: "useStandaloneImage false without tagSuffix full should fail",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret":             "datadog-secret",
					"datadog.appKeyExistingSecret":             "datadog-secret",
					"datadog.otelCollector.enabled":            "true",
					"datadog.otelCollector.useStandaloneImage": "false",
					"agents.image.tag":                         "7.67.0",
				},
			},
			expectError:  true,
			errorMessage: "When datadog.otelCollector.useStandaloneImage is false, agents.image.tagSuffix must be set to 'full'",
		},
		{
			name: "useStandaloneImage false with incorrect tagSuffix should fail",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret":             "datadog-secret",
					"datadog.appKeyExistingSecret":             "datadog-secret",
					"datadog.otelCollector.enabled":            "true",
					"datadog.otelCollector.useStandaloneImage": "false",
					"agents.image.tagSuffix":                   "jmx",
					"agents.image.tag":                         "7.66.0",
				},
			},
			expectError:  true,
			errorMessage: "When datadog.otelCollector.useStandaloneImage is false, agents.image.tagSuffix must be set to 'full'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest, err := common.RenderChart(t, tt.command)

			if tt.expectError {
				assert.Error(t, err, "expected an error but got none")
				if err != nil {
					assert.Contains(t, err.Error(), tt.errorMessage, "error message should contain expected text")
				}
			} else {
				assert.NoError(t, err, "expected no error but got: %v", err)
				if err == nil && tt.assertion != nil {
					tt.assertion(t, manifest)
				}
			}
		})
	}
}

func verifyOtelImage(t *testing.T, manifest string, expectedImage string) {
	var deployment appsv1.DaemonSet
	common.Unmarshal(t, manifest, &deployment)

	otelAgentContainer, ok := getContainer(t, deployment.Spec.Template.Spec.Containers, "otel-agent")
	assert.True(t, ok, "should find otel-agent container")

	assert.Equal(t, expectedImage, otelAgentContainer.Image, "should use exact expected otel image")
}
