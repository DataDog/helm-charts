package datadog

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/DataDog/helm-charts/test/common"
)

const ddNetworkPathCollectorFilters = "DD_NETWORK_PATH_COLLECTOR_FILTERS"

// Verifies that non-empty filters require Agent 7.83.2 or newer unless image tag validation is disabled.
func Test_networkPathCollectorFiltersAgentVersion(t *testing.T) {
	const filtersJSON = `[{"type":"exclude","match_domain":"*.example.com"}]`
	const versionError = "datadog.networkPath.collector.filters requires a stable Datadog Agent 7.83.2 or newer"

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
			overrides := map[string]string{
				"agents.image.tag":                  test.imageTag,
				"agents.image.doNotCheckTag":        strconv.FormatBool(test.doNotCheckTag),
				"datadog.apiKeyExistingSecret":      "datadog-secret",
				"datadog.appKeyExistingSecret":      "datadog-secret",
				"datadog.networkMonitoring.enabled": "true",
			}
			manifest, err := common.RenderChart(t, common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides:   overrides,
				OverridesJson: map[string]string{
					"datadog.networkPath.collector.filters": filtersJSON,
				},
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

// Verifies that Helm rejects filters that do not match the documented schema.
func Test_networkPathCollectorFiltersSchema(t *testing.T) {
	for _, test := range []struct {
		name        string
		filtersJSON string
	}{
		{name: "object instead of list", filtersJSON: `{"type":"exclude","match_domain":"example.com"}`},
		{name: "scalar list item", filtersJSON: `["exclude"]`},
		{name: "missing type", filtersJSON: `[{"match_domain":"example.com"}]`},
		{name: "invalid type", filtersJSON: `[{"type":"drop","match_domain":"example.com"}]`},
		{name: "missing matcher", filtersJSON: `[{"type":"exclude"}]`},
		{name: "empty matcher", filtersJSON: `[{"type":"exclude","match_domain":""}]`},
		{name: "invalid domain strategy", filtersJSON: `[{"type":"exclude","match_domain":"example.com","match_domain_strategy":"glob"}]`},
		{name: "domain strategy without domain", filtersJSON: `[{"type":"exclude","match_ip":"10.0.0.1","match_domain_strategy":"regex"}]`},
		{name: "unknown field", filtersJSON: `[{"type":"exclude","match_domain":"example.com","unknown":"value"}]`},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := common.RenderChart(t, common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"agents.image.tag":             "7.83.2",
					"datadog.apiKeyExistingSecret": "datadog-secret",
					"datadog.appKeyExistingSecret": "datadog-secret",
				},
				OverridesJson: map[string]string{
					"datadog.networkPath.collector.filters": test.filtersJSON,
				},
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), "values don't meet the specifications")
			require.Contains(t, err.Error(), "/datadog/networkPath/collector/filters")
		})
	}
}

// Verifies filter rendering across Agent containers, platforms, GKE modes, and empty values.
func Test_networkPathCollectorFilters(t *testing.T) {
	filters := []map[string]string{
		{
			"match_domain": "*.example.com",
			"type":         "exclude",
		},
		{
			"match_domain":          `^api-[0-9]+\.example\.com$`,
			"match_domain_strategy": "regex",
			"type":                  "include",
		},
		{
			"match_ip": "10.0.0.0/8",
			"type":     "exclude",
		},
	}
	filtersJSON, err := json.Marshal(filters)
	require.NoError(t, err)

	t.Run("configured filters are added to each possible runtime container", func(t *testing.T) {
		manifest, err := common.RenderChart(t, common.HelmCommand{
			ReleaseName: "datadog",
			ChartPath:   "../../charts/datadog",
			ShowOnly:    []string{"templates/daemonset.yaml"},
			Values:      []string{"values/network-path-filters-values.yaml"},
		})
		require.NoError(t, err)

		var daemonSet appsv1.DaemonSet
		common.Unmarshal(t, manifest, &daemonSet)

		for _, containerName := range []string{"agent", "process-agent", "system-probe"} {
			container, found := getContainer(t, daemonSet.Spec.Template.Spec.Containers, containerName)
			require.True(t, found, "expected %s container", containerName)
			assertNetworkPathFilters(t, container, filters)
		}
	})

	t.Run("configured filters are added to Agent containers on Windows", func(t *testing.T) {
		manifest, err := common.RenderChart(t, common.HelmCommand{
			ReleaseName: "datadog",
			ChartPath:   "../../charts/datadog",
			ShowOnly:    []string{"templates/daemonset.yaml"},
			Values:      []string{"../../charts/datadog/values.yaml"},
			Overrides: map[string]string{
				"agents.image.tag":                  "7.83.2",
				"datadog.apiKeyExistingSecret":      "datadog-secret",
				"datadog.appKeyExistingSecret":      "datadog-secret",
				"datadog.networkMonitoring.enabled": "true",
				"datadog.traceroute.enabled":        "true",
				"targetSystem":                      "windows",
			},
			OverridesJson: map[string]string{
				"datadog.networkPath.collector.filters": string(filtersJSON),
			},
		})
		require.NoError(t, err)

		var daemonSet appsv1.DaemonSet
		common.Unmarshal(t, manifest, &daemonSet)

		for _, containerName := range []string{"agent", "process-agent"} {
			container, found := getContainer(t, daemonSet.Spec.Template.Spec.Containers, containerName)
			require.True(t, found, "expected %s container", containerName)
			assertNetworkPathFilters(t, container, filters)
		}
		_, found := getContainer(t, daemonSet.Spec.Template.Spec.Containers, "system-probe")
		assert.False(t, found)
	})

	t.Run("configured filters render in GKE Distributed Cloud mode", func(t *testing.T) {
		manifest, err := common.RenderChart(t, common.HelmCommand{
			ReleaseName: "datadog",
			ChartPath:   "../../charts/datadog",
			ShowOnly:    []string{"templates/daemonset.yaml"},
			Values:      []string{"../../charts/datadog/values.yaml"},
			Overrides: map[string]string{
				"agents.image.tag":             "7.83.2",
				"datadog.apiKeyExistingSecret": "datadog-secret",
				"datadog.appKeyExistingSecret": "datadog-secret",
				"providers.gke.gdc":            "true",
			},
			OverridesJson: map[string]string{
				"datadog.networkPath.collector.filters": string(filtersJSON),
			},
		})
		require.NoError(t, err)

		var daemonSet appsv1.DaemonSet
		common.Unmarshal(t, manifest, &daemonSet)
		requireContainerNames(t, daemonSet, "agent")
		assertNetworkPathFilters(t, daemonSet.Spec.Template.Spec.Containers[0], filters)
		verifyDaemonsetGDCConstraints(t, manifest)
	})

	t.Run("configured filters render in GKE Autopilot AllowlistedV2Workload mode", func(t *testing.T) {
		manifest, err := common.RenderChart(t, common.HelmCommand{
			ReleaseName: "datadog",
			ChartPath:   "../../charts/datadog",
			ShowOnly:    []string{"templates/daemonset.yaml"},
			Values:      []string{"../../charts/datadog/values.yaml"},
			Overrides: map[string]string{
				"agents.image.tag":                  "7.83.2",
				"datadog.envDict.HELM_FORCE_RENDER": "false",
				"datadog.apiKeyExistingSecret":      "datadog-secret",
				"datadog.appKeyExistingSecret":      "datadog-secret",
				"providers.gke.autopilot":           "true",
			},
			OverridesJson: map[string]string{
				"datadog.networkPath.collector.filters": string(filtersJSON),
			},
		})
		require.NoError(t, err)

		var daemonSet appsv1.DaemonSet
		common.Unmarshal(t, manifest, &daemonSet)
		requireContainerNames(t, daemonSet, "agent")
		assertNetworkPathFilters(t, daemonSet.Spec.Template.Spec.Containers[0], filters)
		verifyDaemonsetAutopilotAllowlistedV2WorkloadMinimal(t, manifest)
	})

	t.Run("configured filters render in GKE Autopilot WorkloadAllowlist mode", func(t *testing.T) {
		manifest, err := common.RenderChart(t, common.HelmCommand{
			ReleaseName: "datadog",
			ChartPath:   "../../charts/datadog",
			ShowOnly:    []string{"templates/daemonset.yaml"},
			Values:      []string{"../../charts/datadog/values.yaml"},
			Overrides: map[string]string{
				"agents.image.tag":                  "7.83.2",
				"datadog.envDict.HELM_FORCE_RENDER": "true",
				"datadog.apiKeyExistingSecret":      "datadog-secret",
				"datadog.appKeyExistingSecret":      "datadog-secret",
				"datadog.networkMonitoring.enabled": "true",
				"datadog.traceroute.enabled":        "true",
				"providers.gke.autopilot":           "true",
			},
			OverridesJson: map[string]string{
				"datadog.networkPath.collector.filters": string(filtersJSON),
			},
		})
		require.NoError(t, err)

		var daemonSet appsv1.DaemonSet
		common.Unmarshal(t, manifest, &daemonSet)
		requireContainerNames(t, daemonSet, "agent", "process-agent", "system-probe")
		for _, container := range daemonSet.Spec.Template.Spec.Containers {
			assertNetworkPathFilters(t, container, filters)
		}
		verifyAutopilotWorkloadAllowlistConstraints(t, manifest)
	})

	for _, test := range []struct {
		name          string
		filtersJSON   string
		setJSONFilter bool
	}{
		{name: "default filters"},
		{name: "explicit empty filters", filtersJSON: "[]", setJSONFilter: true},
		{name: "explicit null filters", filtersJSON: "null", setJSONFilter: true},
	} {
		t.Run(test.name+" do not render the filter environment variable", func(t *testing.T) {
			overridesJSON := map[string]string{}
			if test.setJSONFilter {
				overridesJSON["datadog.networkPath.collector.filters"] = test.filtersJSON
			}
			manifest, err := common.RenderChart(t, common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret":      "datadog-secret",
					"datadog.appKeyExistingSecret":      "datadog-secret",
					"datadog.networkMonitoring.enabled": "true",
				},
				OverridesJson: overridesJSON,
			})
			require.NoError(t, err)

			var daemonSet appsv1.DaemonSet
			common.Unmarshal(t, manifest, &daemonSet)
			for _, container := range daemonSet.Spec.Template.Spec.Containers {
				assert.NotContains(t, getEnvVarMap(container.Env), ddNetworkPathCollectorFilters, "container %s", container.Name)
			}
		})
	}
}

func assertNetworkPathFilters(t *testing.T, container corev1.Container, expected []map[string]string) {
	t.Helper()
	env := getNetworkPathFilterEnv(t, container)

	var actual []map[string]string
	require.NoError(t, json.Unmarshal([]byte(env.Value), &actual), "container %s", container.Name)
	assert.Equal(t, expected, actual, "container %s", container.Name)
}

func getNetworkPathFilterEnv(t *testing.T, container corev1.Container) corev1.EnvVar {
	t.Helper()

	var matching []corev1.EnvVar
	for _, env := range container.Env {
		if env.Name == ddNetworkPathCollectorFilters {
			matching = append(matching, env)
		}
	}
	require.Len(t, matching, 1, "container %s", container.Name)
	return matching[0]
}
