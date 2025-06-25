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
