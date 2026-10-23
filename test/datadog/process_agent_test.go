package datadog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	DDLanguageDetectionEnabled     = "DD_LANGUAGE_DETECTION_ENABLED"
)

func Test_processAgentConfigs(t *testing.T) {
	tests := []struct {
		name       string
		command    common.HelmCommand
		assertions func(t *testing.T, manifest string)
	}{
		{
			name: "default behavior -- linux",
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
			assertions: verifyDefaultDaemonset,
		},
		{
			name: "default behavior -- windows",
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
			assertions: verifyDaemonsetWindowsProcessAgentChecks,
		},
		{
			name: "no checks -- linux",
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
			name: "enable process checks in core agent -- linux",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret":           "datadog-secret",
					"datadog.appKeyExistingSecret":           "datadog-secret",
					"datadog.processAgent.processCollection": "true",
					"agents.image.tag":                       "7.60.0",
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
					"datadog.apiKeyExistingSecret": "datadog-secret",
					"datadog.appKeyExistingSecret": "datadog-secret",
					"targetSystem":                 "windows",
					"agents.image.tag":             "7.60.0",
				},
			},
			assertions: verifyDaemonsetWindowsProcessAgentChecks,
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
					"datadog.apiKeyExistingSecret": "datadog-secret",
					"datadog.appKeyExistingSecret": "datadog-secret",
					"agents.image.tag":             "7.52.0",
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
					"datadog.apiKeyExistingSecret": "datadog-secret",
					"datadog.appKeyExistingSecret": "datadog-secret",
					"agents.image.doNotCheckTag":   "true",
				},
			},
			assertions: verifyDefaultDaemonset,
		},
		{
			name: "language detection on process agent",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret":                           "datadog-secret",
					"datadog.appKeyExistingSecret":                           "datadog-secret",
					"datadog.processAgent.processCollection":                 "true",
					"agents.image.tag":                                       "7.56",
					"datadog.apm.instrumentation.language_detection.enabled": "true",
					"datadog.apm.instrumentation.enabled":                    "true",
				},
			},
			assertions: verifyLanguageDetectionInProcessAgent,
		},
		{
			name: "language detection on core agent",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret":                           "datadog-secret",
					"datadog.appKeyExistingSecret":                           "datadog-secret",
					"datadog.processAgent.processCollection":                 "true",
					"agents.image.tag":                                       "7.60.0",
					"datadog.apm.instrumentation.language_detection.enabled": "true",
					"datadog.apm.instrumentation.enabled":                    "true",
				},
			},
			assertions: verifyLanguageDetectionInCoreAgent,
		},
		{
			name: "unprivileged agent handling",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret":      "datadog-secret",
					"datadog.appKeyExistingSecret":      "datadog-secret",
					"datadog.securityContext.runAsUser": "100",
				},
			},
			assertions: verifyUnprivilegedAgentHandling,
		},
		{
			name: "trying to disable run in core agent",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret":        "datadog-secret",
					"datadog.appKeyExistingSecret":        "datadog-secret",
					"datadog.processAgent.runInCoreAgent": "false",
				},
			},
			assertions: verifyDefaultDaemonset,
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

func TestProcessPasswdMountPaths(t *testing.T) {
	const legacy = "/etc/passwd"
	const host = "/host/etc/passwd"
	expectedCoreLegacy := map[string][]string{"agent": {legacy}}
	expectedCoreHost := map[string][]string{"agent": {host}}
	expectedNone := map[string][]string{}

	tests := []struct {
		name           string
		tag            string
		overrides      map[string]string
		expectedMounts map[string][]string
	}{
		{name: "old stable", tag: "7.85.99", expectedMounts: expectedCoreLegacy},
		{name: "first supported release", tag: "7.86.0", expectedMounts: expectedCoreHost},
		{name: "newer patch", tag: "7.86.1", expectedMounts: expectedCoreHost},
		{name: "newer minor", tag: "7.87.0", expectedMounts: expectedCoreHost},
		{name: "jmx tag", tag: "7.86.0-jmx", expectedMounts: expectedCoreHost},
		{name: "full suffix", tag: "7.86.0", overrides: map[string]string{"agents.image.tagSuffix": "full"}, expectedMounts: expectedCoreHost},
		{name: "fips tag", tag: "7.86.0-fips", expectedMounts: expectedCoreHost},
		{name: "fips full tag", tag: "7.86.0-fips-full", expectedMounts: expectedCoreHost},
		{name: "fips jmx tag", tag: "7.86.0-fips-jmx", expectedMounts: expectedCoreHost},
		{name: "version prefix uses standard parsing", tag: "v7.86.0", expectedMounts: expectedCoreHost},
		{name: "tag suffix uses configured version", tag: "7.86.0", overrides: map[string]string{"agents.image.tagSuffix": "custom"}, expectedMounts: expectedCoreHost},
		{name: "devel predates resolver", tag: "7.86.0-devel", expectedMounts: expectedCoreLegacy},
		{name: "release candidate", tag: "7.86.0-rc.1", expectedMounts: expectedCoreLegacy},
		{name: "future prerelease", tag: "7.87.0-rc.1", expectedMounts: expectedCoreLegacy},
		{name: "major alias", tag: "7", expectedMounts: expectedCoreLegacy},
		{name: "latest alias", tag: "latest", expectedMounts: expectedCoreLegacy},
		{name: "minor alias", tag: "7.86", expectedMounts: expectedCoreHost},
		{name: "custom version-looking tag", tag: "branch-7.86.0", overrides: map[string]string{"agents.image.doNotCheckTag": "true"}, expectedMounts: expectedCoreLegacy},
		{name: "custom suffix", tag: "7.86.0", overrides: map[string]string{"agents.image.tagSuffix": "custom", "agents.image.doNotCheckTag": "true"}, expectedMounts: expectedCoreLegacy},
		{name: "digest wins over tag", tag: "7.86.0", overrides: map[string]string{"agents.image.digest": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}, expectedMounts: expectedCoreLegacy},
		{name: "skip check is not capability", tag: "local-build", overrides: map[string]string{"agents.image.doNotCheckTag": "true"}, expectedMounts: expectedCoreLegacy},
		{name: "skip check still checks known tag", tag: "7.85.0", overrides: map[string]string{"agents.image.doNotCheckTag": "true"}, expectedMounts: expectedCoreLegacy},
		{name: "skip check keeps legacy for supported version", tag: "7.86.0", overrides: map[string]string{"agents.image.doNotCheckTag": "true"}, expectedMounts: expectedCoreLegacy},
		{name: "custom repository", tag: "7.86.0", overrides: map[string]string{"agents.image.repository": "registry.example:5000/mirror/agent"}, expectedMounts: expectedCoreHost},
		{name: "supported old separate process agent", tag: "7.59.0",
			expectedMounts: map[string][]string{"process-agent": {legacy}}},
		{name: "core and network process sidecar", tag: "7.86.0", overrides: map[string]string{"datadog.networkMonitoring.enabled": "true"},
			expectedMounts: map[string][]string{"agent": {host}, "process-agent": {host}}},
		{name: "autopilot legacy", tag: "7.86.0", overrides: map[string]string{"providers.gke.autopilot": "true"}, expectedMounts: expectedCoreLegacy},
		{name: "autopilot workload allowlist", tag: "7.86.0", overrides: map[string]string{"providers.gke.autopilot": "true", "datadog.envDict.HELM_FORCE_RENDER": "true"}, expectedMounts: expectedCoreLegacy},
		{name: "GDC exclusion", tag: "7.86.0", overrides: map[string]string{"providers.gke.gdc": "true"}, expectedMounts: expectedNone},
		{name: "Talos exclusion", tag: "7.86.0", overrides: map[string]string{"providers.talos.enabled": "true"}, expectedMounts: expectedNone},
		{name: "Windows exclusion", tag: "7.86.0", overrides: map[string]string{"targetSystem": "windows"}, expectedMounts: expectedNone},
		{name: "mount disabled", tag: "7.86.0", overrides: map[string]string{"datadog.disablePasswdMount": "true"}, expectedMounts: expectedNone},
		{name: "nonroot eligibility unchanged", tag: "7.86.0", overrides: map[string]string{"datadog.securityContext.runAsUser": "100"}, expectedMounts: expectedNone},
		{name: "security agent unchanged", tag: "7.86.0", overrides: map[string]string{"datadog.securityAgent.compliance.enabled": "true"},
			expectedMounts: map[string][]string{"agent": {host}, "security-agent": {legacy}}},
		{name: "system probe unchanged", tag: "7.86.0", overrides: map[string]string{"datadog.securityAgent.compliance.enabled": "true", "datadog.securityAgent.compliance.runInSystemProbe": "true"},
			expectedMounts: map[string][]string{"agent": {host}, "system-probe": {legacy}}},
		{name: "no process checks", tag: "7.86.0", overrides: map[string]string{
			"datadog.processAgent.processCollection": "false", "datadog.processAgent.processDiscovery": "false",
			"datadog.processAgent.containerCollection": "false", "datadog.apm.instrumentation.language_detection.enabled": "false",
		}, expectedMounts: expectedNone},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			overrides := map[string]string{"agents.image.tag": tt.tag, "datadog.apiKeyExistingSecret": "datadog-secret", "datadog.processAgent.processCollection": "true"}
			for key, value := range tt.overrides {
				overrides[key] = value
			}
			manifest, err := common.RenderChart(t, common.HelmCommand{
				ReleaseName: "datadog", ChartPath: "../../charts/datadog", ShowOnly: []string{"templates/daemonset.yaml"}, Overrides: overrides,
			})
			require.NoError(t, err)
			var ds appsv1.DaemonSet
			common.Unmarshal(t, manifest, &ds)
			// Keep every path so duplicate mounts fail the comparison too.
			actualMounts := map[string][]string{}
			for _, container := range ds.Spec.Template.Spec.Containers {
				for _, mount := range container.VolumeMounts {
					switch mount.Name {
					case "passwd":
						require.True(t, mount.ReadOnly)
						actualMounts[container.Name] = append(actualMounts[container.Name], mount.MountPath)
					case "group":
						assert.Equal(t, "/etc/group", mount.MountPath)
					}
				}
			}
			assert.Equal(t, tt.expectedMounts, actualMounts)
		})
	}
}

func TestProcessPasswdPreservesHostEtcConfiguration(t *testing.T) {
	for _, hostEtc := range []string{"/host/etc", "/custom/etc", ""} {
		t.Run("HOST_ETC="+hostEtc, func(t *testing.T) {
			manifest, err := common.RenderChart(t, common.HelmCommand{
				ReleaseName: "datadog", ChartPath: "../../charts/datadog", ShowOnly: []string{"templates/daemonset.yaml"},
				Overrides: map[string]string{"agents.image.tag": "7.86.0", "datadog.apiKeyExistingSecret": "datadog-secret", "datadog.envDict.HOST_ETC": hostEtc, "datadog.envDict.DD_IGNORE_HOST_ETC": "true"},
			})
			require.NoError(t, err)
			var ds appsv1.DaemonSet
			common.Unmarshal(t, manifest, &ds)
			container, found := getContainer(t, ds.Spec.Template.Spec.Containers, "agent")
			require.True(t, found)
			expectedEnv := []corev1.EnvVar{{Name: "HOST_ETC", Value: hostEtc}, {Name: "DD_IGNORE_HOST_ETC", Value: "true"}}
			var hostEtcEnv []corev1.EnvVar
			for _, env := range container.Env {
				if env.Name == "HOST_ETC" || env.Name == "DD_IGNORE_HOST_ETC" {
					hostEtcEnv = append(hostEtcEnv, env)
				}
			}
			assert.ElementsMatch(t, expectedEnv, hostEtcEnv, "preserve explicit configuration without duplicates")
		})
	}
}

func verifyDefaultDaemonset(t *testing.T, manifest string) {
	var deployment appsv1.DaemonSet
	common.Unmarshal(t, manifest, &deployment)
	coreAgentContainer, ok := getContainer(t, deployment.Spec.Template.Spec.Containers, "agent")
	assert.True(t, ok)
	coreEnvs := getEnvVarMap(coreAgentContainer.Env)
	assertDefaultCommonProcessEnvs(t, coreEnvs)
	assert.Equal(t, "true", coreEnvs[DDProcessRunInCoreAgentEnabled])
	assert.True(t, getPasswdMount(t, coreAgentContainer.VolumeMounts))

	_, ok = getContainer(t, deployment.Spec.Template.Spec.Containers, "process-agent")
	assert.False(t, ok)
}

func verifyDaemonsetWindowsProcessAgentChecks(t *testing.T, manifest string) {
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

func verifyLanguageDetectionInCoreAgent(t *testing.T, manifest string) {
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
	assert.Equal(t, "true", coreEnvs[DDLanguageDetectionEnabled])
	assert.True(t, getPasswdMount(t, coreAgentContainer.VolumeMounts))

	_, ok = getContainer(t, deployment.Spec.Template.Spec.Containers, "process-agent")
	assert.False(t, ok)
}

func verifyLanguageDetectionInProcessAgent(t *testing.T, manifest string) {
	var deployment appsv1.DaemonSet
	common.Unmarshal(t, manifest, &deployment)
	coreAgentContainer, ok := getContainer(t, deployment.Spec.Template.Spec.Containers, "agent")
	assert.True(t, ok)
	coreEnvs := getEnvVarMap(coreAgentContainer.Env)
	assert.Equal(t, "false", coreEnvs[DDProcessRunInCoreAgentEnabled])
	assert.False(t, getPasswdMount(t, coreAgentContainer.VolumeMounts))

	processAgentContainer, ok := getContainer(t, deployment.Spec.Template.Spec.Containers, "process-agent")
	assert.True(t, ok)
	processEnvs := getEnvVarMap(processAgentContainer.Env)
	assert.Equal(t, "true", processEnvs[DDContainerCollectionEnabled])
	assert.Equal(t, "true", processEnvs[DDProcessCollectionEnabled])
	assert.Equal(t, "true", processEnvs[DDProcessDiscoveryEnabled])
	assert.Equal(t, "false", processEnvs[DDStripProcessArgs])
	assert.Equal(t, "false", processEnvs[DDProcessRunInCoreAgentEnabled])
	assert.Equal(t, "true", processEnvs[DDLanguageDetectionEnabled])
	assert.True(t, getPasswdMount(t, processAgentContainer.VolumeMounts))
}

func verifyChecksOff(t *testing.T, manifest string) {
	var deployment appsv1.DaemonSet
	common.Unmarshal(t, manifest, &deployment)
	coreAgentContainer, ok := getContainer(t, deployment.Spec.Template.Spec.Containers, "agent")
	assert.True(t, ok)
	coreEnvs := getEnvVarMap(coreAgentContainer.Env)
	assertFalseCommonProcessEnvs(t, coreEnvs)
	assert.Equal(t, "true", coreEnvs[DDProcessRunInCoreAgentEnabled])
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
	assert.Equal(t, "true", coreEnvs[DDProcessRunInCoreAgentEnabled])
	assert.False(t, getPasswdMount(t, coreAgentContainer.VolumeMounts))

	processAgentContainer, ok := getContainer(t, deployment.Spec.Template.Spec.Containers, "process-agent")
	assert.True(t, ok)
	processEnvs := getEnvVarMap(processAgentContainer.Env)
	assertFalseCommonProcessEnvs(t, processEnvs)
	assert.Equal(t, "true", coreEnvs[DDProcessRunInCoreAgentEnabled])
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
	assert.Equal(t, "true", coreEnvs[DDProcessRunInCoreAgentEnabled])
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

func verifyUnprivilegedAgentHandling(t *testing.T, manifest string) {
	var deployment appsv1.DaemonSet
	common.Unmarshal(t, manifest, &deployment)
	coreAgentContainer, ok := getContainer(t, deployment.Spec.Template.Spec.Containers, "agent")
	assert.True(t, ok)
	assert.False(t, getPasswdMount(t, coreAgentContainer.VolumeMounts))
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
