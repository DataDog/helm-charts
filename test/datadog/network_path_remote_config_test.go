package datadog

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"

	"github.com/DataDog/helm-charts/test/common"
)

const ddNetworkPathRemoteConfigEnabled = "DD_NETWORK_PATH_REMOTE_CONFIG_ENABLED"

func networkPathRemoteConfigOverrides(imageTag string, doNotCheckTag bool, enabled bool) map[string]string {
	return map[string]string{
		"agents.image.tag":                                  imageTag,
		"agents.image.doNotCheckTag":                        strconv.FormatBool(doNotCheckTag),
		"datadog.apiKeyExistingSecret":                      "datadog-secret",
		"datadog.appKeyExistingSecret":                      "datadog-secret",
		"datadog.networkMonitoring.enabled":                 "true",
		"datadog.networkPath.connectionsMonitoring.enabled": "true",
		"datadog.traceroute.enabled":                        "true",
		"datadog.networkPath.remoteConfig.enabled":          strconv.FormatBool(enabled),
	}
}

// Verifies that enabling Network Path Remote Configuration requires Agent 7.83.2 or newer
// unless image tag validation is disabled.
func Test_networkPathRemoteConfigAgentVersion(t *testing.T) {
	const versionError = "datadog.networkPath.remoteConfig.enabled requires a stable Datadog Agent 7.83.2 or newer"

	for _, test := range []struct {
		name          string
		imageTag      string
		doNotCheckTag bool
		expectError   bool
	}{
		{name: "last unsupported release", imageTag: "7.83.1", expectError: true},
		{name: "unverified release candidate", imageTag: "7.83.2-rc.1", expectError: true},
		{name: "verified compatible prerelease", imageTag: "7.83.2-rc.1", doNotCheckTag: true},
		{name: "first supported stable release", imageTag: "7.83.2"},
		{name: "next minor release", imageTag: "7.84.0"},
		{name: "compatible custom image", imageTag: "local", doNotCheckTag: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			manifest, err := common.RenderChart(t, common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides:   networkPathRemoteConfigOverrides(test.imageTag, test.doNotCheckTag, true),
			})

			if test.expectError {
				require.ErrorContains(t, err, versionError)
				assert.Empty(t, manifest)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, manifest)
		})
	}
}

// Verifies that the setting reaches every container that reads network_path.remote_config.enabled:
// the core Agent (scheduled tests) and the process-agent and system-probe (dynamic test filters).
func Test_networkPathRemoteConfig(t *testing.T) {
	t.Run("enabled sets the env var on each runtime container", func(t *testing.T) {
		manifest, err := common.RenderChart(t, common.HelmCommand{
			ReleaseName: "datadog",
			ChartPath:   "../../charts/datadog",
			ShowOnly:    []string{"templates/daemonset.yaml"},
			Values:      []string{"../../charts/datadog/values.yaml"},
			Overrides:   networkPathRemoteConfigOverrides("7.83.2", false, true),
		})
		require.NoError(t, err)

		var daemonSet appsv1.DaemonSet
		common.Unmarshal(t, manifest, &daemonSet)

		for _, containerName := range []string{"agent", "process-agent", "system-probe"} {
			container, found := getContainer(t, daemonSet.Spec.Template.Spec.Containers, containerName)
			require.True(t, found, "expected %s container", containerName)
			env := getNetworkPathEnv(t, container, ddNetworkPathRemoteConfigEnabled)
			assert.Equal(t, "true", env.Value, "container %s", containerName)
		}
	})

	t.Run("disabled by default sets no env var", func(t *testing.T) {
		manifest, err := common.RenderChart(t, common.HelmCommand{
			ReleaseName: "datadog",
			ChartPath:   "../../charts/datadog",
			ShowOnly:    []string{"templates/daemonset.yaml"},
			Values:      []string{"../../charts/datadog/values.yaml"},
			Overrides:   networkPathRemoteConfigOverrides("7.83.2", false, false),
		})
		require.NoError(t, err)

		var daemonSet appsv1.DaemonSet
		common.Unmarshal(t, manifest, &daemonSet)

		for _, container := range daemonSet.Spec.Template.Spec.Containers {
			for _, env := range container.Env {
				assert.NotEqual(t, ddNetworkPathRemoteConfigEnabled, env.Name, "container %s", container.Name)
			}
		}
	})
}
