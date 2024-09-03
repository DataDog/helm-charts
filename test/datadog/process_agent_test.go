package datadog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/DataDog/helm-charts/test/common"
)

const (
	DDProcessCollectionEnabled     = "DD_PROCESS_CONFIG_PROCESS_COLLECTION_ENABLED"
	DDContainerCollectionEnabled   = "DD_PROCESS_CONFIG_CONTAINER_COLLECTION_ENABLED"
	DDProcessDiscoveryEnabled      = "DD_PROCESS_AGENT_DISCOVERY_ENABLED"
	DDStripProcessArgs             = "DD_STRIP_PROCESS_ARGS"
	DDProcessRunInCoreAgentEnabled = "DD_PROCESS_CONFIG_RUN_IN_CORE_AGENT_ENABLED"
	DDSystemProbeEnabled           = "DD_SYSTEM_PROBE_ENABLED"
	DDNetworkMonitoringEnabled     = "DD_SYSTEM_PROBE_NETWORK_ENABLED"
	DDOrchestratorEnabled          = "DD_ORCHESTRATOR_EXPLORER_ENABLED"
)

func Test_processAgentConfigs(t *testing.T) {
	tests := []struct {
		name       string
		command    common.HelmCommand
		assertions func(t *testing.T, manifest string)
	}{
		{
			name: "default",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret": "datadog-secret",
					"datadog.appKeyExistingSecret": "datadog-secret",
				},
			},
			assertions: verifyDaemonsetMinimal,
		},
		{
			name: "default windows",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret": "datadog-secret",
					"datadog.appKeyExistingSecret": "datadog-secret",
					"targetSystem":                 "windows",
				},
			},
			assertions: verifyDaemonsetMinimalWindows,
		},
		{
			name: "all checks off",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret":                           "datadog-secret",
					"datadog.appKeyExistingSecret":                           "datadog-secret",
					"datadog.processAgent.processCollection":                 "false",
					"datadog.processAgent.containerCollection":               "false",
					"datadog.processAgent.processDiscovery":                  "false",
					"datadog.apm.instrumentation.language_detection.enabled": "false",
				},
			},
			assertions: verifyChecksOff,
		},
		{
			name: "only network monitoring enabled",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret":                           "datadog-secret",
					"datadog.appKeyExistingSecret":                           "datadog-secret",
					"datadog.processAgent.processCollection":                 "false",
					"datadog.processAgent.containerCollection":               "false",
					"datadog.processAgent.processDiscovery":                  "false",
					"datadog.apm.instrumentation.language_detection.enabled": "false",
					"datadog.networkMonitoring.enabled":                      "true",
				},
			},
			assertions: verifyOnlyNetworkMonitoringEnabled,
		},
		{
			name: "enable process checks in core agent -- linux with default version",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret":           "datadog-secret",
					"datadog.appKeyExistingSecret":           "datadog-secret",
					"datadog.processAgent.runInCoreAgent":    "true",
					"datadog.processAgent.processCollection": "true",
				},
			},
			assertions: verifyLinuxRunInCoreAgent,
		},
		{
			name: "enable process checks in core agent -- linux with latest version",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret":           "datadog-secret",
					"datadog.appKeyExistingSecret":           "datadog-secret",
					"datadog.processAgent.runInCoreAgent":    "true",
					"datadog.processAgent.processCollection": "true",
					"agents.image.tag":                       "latest",
				},
			},
			assertions: verifyLinuxRunInCoreAgent,
		},
		{
			name: "enable process checks in core agent -- linux with version 7",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret":           "datadog-secret",
					"datadog.appKeyExistingSecret":           "datadog-secret",
					"datadog.processAgent.runInCoreAgent":    "true",
					"datadog.processAgent.processCollection": "true",
					"agents.image.tag":                       "7",
				},
			},
			assertions: verifyLinuxRunInCoreAgent,
		},
		{
			name: "enable process checks in core agent -- windows",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret":        "datadog-secret",
					"datadog.appKeyExistingSecret":        "datadog-secret",
					"targetSystem":                        "windows",
					"datadog.processAgent.runInCoreAgent": "true",
				},
			},
			assertions: verifyDaemonsetMinimalWindows,
		},
		{
			name: "orchestrator enabled - latest version",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret":                           "datadog-secret",
					"datadog.appKeyExistingSecret":                           "datadog-secret",
					"datadog.processAgent.processCollection":                 "false",
					"datadog.processAgent.containerCollection":               "false",
					"datadog.processAgent.processDiscovery":                  "false",
					"datadog.apm.instrumentation.language_detection.enabled": "false",
					"datadog.orchestratorExplorer.enabled":                   "true",
				},
			},
			assertions: verifyOrchestratorEnabledLatest,
		},
		{
			name: "orchestrator enabled - old version",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret":                           "datadog-secret",
					"datadog.appKeyExistingSecret":                           "datadog-secret",
					"datadog.processAgent.processCollection":                 "false",
					"datadog.processAgent.containerCollection":               "false",
					"datadog.processAgent.processDiscovery":                  "false",
					"datadog.apm.instrumentation.language_detection.enabled": "false",
					"datadog.orchestratorExplorer.enabled":                   "true",
					"agents.image.tag":                                       "7.50.0",
				},
			},
			assertions: verifyOrchestratorEnabledOld,
		},
		{
			name: "enable process checks in core agent -- old version",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret":        "datadog-secret",
					"datadog.appKeyExistingSecret":        "datadog-secret",
					"datadog.processAgent.runInCoreAgent": "true",
					"agents.image.tag":                    "7.52.0",
				},
			},
			assertions: verifyLinuxRunInCoreAgentOld,
		},
		{
			name: "enable process checks in core agent -- do not check image tag",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret":        "datadog-secret",
					"datadog.appKeyExistingSecret":        "datadog-secret",
					"datadog.processAgent.runInCoreAgent": "true",
					"agents.image.doNotCheckTag":          "true",
				},
			},
			assertions: verifyLinuxRunInCoreAgentOld,
		},
		{
			name: "enable process checks in core agent -- env var override",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml", "values/process-run-in-core-envvars.yaml" },
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret":        "datadog-secret",
					"datadog.appKeyExistingSecret":        "datadog-secret",
					"datadog.processAgent.runInCoreAgent": "false",
					"agents.image.doNotCheckTag":          "true",
					"datadog.processAgent.processCollection": "true",
				},
			},
			assertions: verifyLinuxRunInCoreAgent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest, err := common.RenderChart(t, tt.command)
			assert.Nil(t, err, "couldn't render template")
			tt.assertions(t, manifest)
		})
	}
}

func verifyDaemonsetMinimal(t *testing.T, manifest string) {
	var deployment appsv1.DaemonSet
	common.Unmarshal(t, manifest, &deployment)
	coreAgentContainer, ok := getContainer(t, deployment.Spec.Template.Spec.Containers, "agent")
	assert.True(t, ok)
	coreEnvs := getEnvVarMap(coreAgentContainer.Env)
	assertDefaultCommonProcessEnvs(t, coreEnvs)
	assert.Equal(t, "false", coreEnvs[DDProcessRunInCoreAgentEnabled])
	assert.False(t, getPasswdMount(t, coreAgentContainer.VolumeMounts))

	processAgentContainer, ok := getContainer(t, deployment.Spec.Template.Spec.Containers, "process-agent")
	assert.True(t, ok)
	processEnvs := getEnvVarMap(processAgentContainer.Env)
	assertDefaultCommonProcessEnvs(t, processEnvs)
	assert.Equal(t, "false", coreEnvs[DDProcessRunInCoreAgentEnabled])
	assert.True(t, getPasswdMount(t, processAgentContainer.VolumeMounts))
}

func verifyDaemonsetMinimalWindows(t *testing.T, manifest string) {
	var deployment appsv1.DaemonSet
	common.Unmarshal(t, manifest, &deployment)
	coreAgentContainer, ok := getContainer(t, deployment.Spec.Template.Spec.Containers, "agent")
	assert.True(t, ok)
	coreEnvs := getEnvVarMap(coreAgentContainer.Env)
	assertDefaultCommonProcessEnvs(t, coreEnvs)
	assert.Equal(t, "", coreEnvs[DDProcessRunInCoreAgentEnabled])

	processAgentContainer, ok := getContainer(t, deployment.Spec.Template.Spec.Containers, "process-agent")
	assert.True(t, ok)
	processEnvs := getEnvVarMap(processAgentContainer.Env)
	assertDefaultCommonProcessEnvs(t, processEnvs)
	assert.Equal(t, "", processEnvs[DDProcessRunInCoreAgentEnabled])
}

func verifyLinuxRunInCoreAgent(t *testing.T, manifest string) {
	var deployment appsv1.DaemonSet
	common.Unmarshal(t, manifest, &deployment)
	coreAgentContainer, ok := getContainer(t, deployment.Spec.Template.Spec.Containers, "agent")
	assert.True(t, ok)
	coreEnvs := getEnvVarMap(coreAgentContainer.Env)
	assert.Equal(t, "true", coreEnvs[DDContainerCollectionEnabled])
	assert.Equal(t, "true", coreEnvs[DDProcessCollectionEnabled])
	assert.Equal(t, "true", coreEnvs[DDProcessDiscoveryEnabled])
	assert.Equal(t, "false", coreEnvs[DDStripProcessArgs])
	assert.Equal(t, "true", coreEnvs[DDProcessRunInCoreAgentEnabled])
	assert.True(t, getPasswdMount(t, coreAgentContainer.VolumeMounts))

	_, ok = getContainer(t, deployment.Spec.Template.Spec.Containers, "process-agent")
	assert.False(t, ok)
}

func verifyChecksOff(t *testing.T, manifest string) {
	var deployment appsv1.DaemonSet
	common.Unmarshal(t, manifest, &deployment)
	coreAgentContainer, ok := getContainer(t, deployment.Spec.Template.Spec.Containers, "agent")
	assert.True(t, ok)
	coreEnvs := getEnvVarMap(coreAgentContainer.Env)
	assertFalseCommonProcessEnvs(t, coreEnvs)
	assert.Equal(t, "false", coreEnvs[DDProcessRunInCoreAgentEnabled])
	assert.False(t, getPasswdMount(t, coreAgentContainer.VolumeMounts))

	_, ok = getContainer(t, deployment.Spec.Template.Spec.Containers, "process-agent")
	assert.False(t, ok)
}

func verifyOnlyNetworkMonitoringEnabled(t *testing.T, manifest string) {
	var deployment appsv1.DaemonSet
	common.Unmarshal(t, manifest, &deployment)
	coreAgentContainer, ok := getContainer(t, deployment.Spec.Template.Spec.Containers, "agent")
	assert.True(t, ok)
	coreEnvs := getEnvVarMap(coreAgentContainer.Env)
	assertFalseCommonProcessEnvs(t, coreEnvs)
	assert.Equal(t, "false", coreEnvs[DDProcessRunInCoreAgentEnabled])
	assert.False(t, getPasswdMount(t, coreAgentContainer.VolumeMounts))

	processAgentContainer, ok := getContainer(t, deployment.Spec.Template.Spec.Containers, "process-agent")
	assert.True(t, ok)
	processEnvs := getEnvVarMap(processAgentContainer.Env)
	assertFalseCommonProcessEnvs(t, processEnvs)
	assert.Equal(t, "false", coreEnvs[DDProcessRunInCoreAgentEnabled])
	assert.Equal(t, "true", processEnvs[DDSystemProbeEnabled])
	assert.Equal(t, "true", processEnvs[DDNetworkMonitoringEnabled])
	assert.False(t, getPasswdMount(t, processAgentContainer.VolumeMounts))
}

func verifyOrchestratorEnabledLatest(t *testing.T, manifest string) {
	var deployment appsv1.DaemonSet
	common.Unmarshal(t, manifest, &deployment)
	coreAgentContainer, ok := getContainer(t, deployment.Spec.Template.Spec.Containers, "agent")
	assert.True(t, ok)
	coreEnvs := getEnvVarMap(coreAgentContainer.Env)
	assertFalseCommonProcessEnvs(t, coreEnvs)
	assert.Equal(t, "false", coreEnvs[DDProcessRunInCoreAgentEnabled])
	assert.Equal(t, "true", coreEnvs[DDOrchestratorEnabled])
	assert.False(t, getPasswdMount(t, coreAgentContainer.VolumeMounts))

	_, ok = getContainer(t, deployment.Spec.Template.Spec.Containers, "process-agent")
	assert.False(t, ok)
}

func verifyOrchestratorEnabledOld(t *testing.T, manifest string) {
	var deployment appsv1.DaemonSet
	common.Unmarshal(t, manifest, &deployment)
	coreAgentContainer, ok := getContainer(t, deployment.Spec.Template.Spec.Containers, "agent")
	assert.True(t, ok)
	coreEnvs := getEnvVarMap(coreAgentContainer.Env)
	assertFalseCommonProcessEnvs(t, coreEnvs)
	assert.Equal(t, "false", coreEnvs[DDProcessRunInCoreAgentEnabled])
	assert.Equal(t, "true", coreEnvs[DDOrchestratorEnabled])
	assert.False(t, getPasswdMount(t, coreAgentContainer.VolumeMounts))

	processAgentContainer, ok := getContainer(t, deployment.Spec.Template.Spec.Containers, "process-agent")
	assert.True(t, ok)
	processEnvs := getEnvVarMap(processAgentContainer.Env)
	assertFalseCommonProcessEnvs(t, processEnvs)
	assert.Equal(t, "false", coreEnvs[DDProcessRunInCoreAgentEnabled])
	assert.Equal(t, "true", processEnvs[DDOrchestratorEnabled])
	assert.False(t, getPasswdMount(t, processAgentContainer.VolumeMounts))
}

func verifyLinuxRunInCoreAgentOld(t *testing.T, manifest string) {
	var deployment appsv1.DaemonSet
	common.Unmarshal(t, manifest, &deployment)
	coreAgentContainer, ok := getContainer(t, deployment.Spec.Template.Spec.Containers, "agent")
	assert.True(t, ok)
	coreEnvs := getEnvVarMap(coreAgentContainer.Env)
	assertDefaultCommonProcessEnvs(t, coreEnvs)
	assert.Equal(t, "false", coreEnvs[DDProcessRunInCoreAgentEnabled])
	assert.False(t, getPasswdMount(t, coreAgentContainer.VolumeMounts))

	processAgentContainer, ok := getContainer(t, deployment.Spec.Template.Spec.Containers, "process-agent")
	assert.True(t, ok)
	processEnvs := getEnvVarMap(processAgentContainer.Env)
	assertDefaultCommonProcessEnvs(t, processEnvs)
	assert.Equal(t, "false", coreEnvs[DDProcessRunInCoreAgentEnabled])
	assert.True(t, getPasswdMount(t, processAgentContainer.VolumeMounts))
}

func getContainer(t *testing.T, containers []corev1.Container, name string) (corev1.Container, bool) {
	for _, container := range containers {
		if container.Name == name {
			return container, true
		}
	}
	return corev1.Container{}, false
}

func assertDefaultCommonProcessEnvs(t *testing.T, envs map[string]string) {
	assert.Equal(t, "true", envs[DDContainerCollectionEnabled])
	assert.Equal(t, "false", envs[DDProcessCollectionEnabled])
	assert.Equal(t, "true", envs[DDProcessDiscoveryEnabled])
	assert.Equal(t, "false", envs[DDStripProcessArgs])
}

func assertFalseCommonProcessEnvs(t *testing.T, envs map[string]string) {
	assert.Equal(t, "false", envs[DDContainerCollectionEnabled])
	assert.Equal(t, "false", envs[DDProcessCollectionEnabled])
	assert.Equal(t, "false", envs[DDProcessDiscoveryEnabled])
	assert.Equal(t, "false", envs[DDStripProcessArgs])
}

func getPasswdMount(t *testing.T, volumeMounts []corev1.VolumeMount) bool {
	for _, vm := range volumeMounts {
		if vm.Name == "passwd" {
			return true
		}
	}
	return false
}

func getEnvVarMap(envVars []corev1.EnvVar) map[string]string {
	envVarMap := map[string]string{}
	for _, envVar := range envVars {
		envVarMap[envVar.Name] = envVar.Value
	}
	return envVarMap
}
