package datadog

import (
	"bufio"
	"bytes"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	semver "github.com/Masterminds/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
)

func Test_serviceDiscoveryResolvedDefaulting(t *testing.T) {
	defaultAgentTag := getDefaultAgentTag(t)
	defaultDiscoveryEnabled := shouldAutoEnableDiscoveryFromTag(defaultAgentTag)

	tests := []struct {
		name                        string
		overrides                   map[string]string
		expectSystemProbe           bool
		expectDiscoveryBlock        bool
		expectDiscoveryEnabled      bool
		expectUseSystemProbeLiteKey bool
		expectUseSystemProbeLite    bool
	}{
		{
			name: "omitted discovery with agent 7.78.0 enables discovery",
			overrides: map[string]string{
				"agents.image.tag": "7.78.0",
			},
			expectSystemProbe:           true,
			expectDiscoveryBlock:        true,
			expectDiscoveryEnabled:      true,
			expectUseSystemProbeLiteKey: true,
			expectUseSystemProbeLite:    true,
		},
		{
			name: "omitted discovery with agent 7.77.9 disables discovery",
			overrides: map[string]string{
				"agents.image.tag": "7.77.9",
			},
			expectSystemProbe:    false,
			expectDiscoveryBlock: false,
		},
		{
			name: "omitted discovery with floating agent 6 disables discovery",
			overrides: map[string]string{
				"agents.image.tag": "6",
			},
			expectSystemProbe:    false,
			expectDiscoveryBlock: false,
		},
		{
			name: "omitted discovery with floating agent 6-jmx disables discovery",
			overrides: map[string]string{
				"agents.image.tag": "6-jmx",
			},
			expectSystemProbe:    false,
			expectDiscoveryBlock: false,
		},
		{
			name: "omitted discovery with floating agent 7 disables discovery",
			overrides: map[string]string{
				"agents.image.tag": "7",
			},
			expectSystemProbe:    false,
			expectDiscoveryBlock: false,
		},
		{
			name: "explicit false with agent 7.78.0 keeps discovery disabled",
			overrides: map[string]string{
				"agents.image.tag":          "7.78.0",
				"datadog.discovery.enabled": "false",
			},
			expectSystemProbe:    false,
			expectDiscoveryBlock: false,
		},
		{
			name: "explicit true with agent 7.77.9 keeps discovery enabled",
			overrides: map[string]string{
				"agents.image.tag":          "7.77.9",
				"datadog.discovery.enabled": "true",
			},
			expectSystemProbe:           true,
			expectDiscoveryBlock:        true,
			expectDiscoveryEnabled:      true,
			expectUseSystemProbeLiteKey: true,
			expectUseSystemProbeLite:    false,
		},
		{
			name: "omitted discovery with partial image override inherits default tag policy",
			overrides: map[string]string{
				"agents.image.name": "custom-agent",
			},
			expectSystemProbe:           defaultDiscoveryEnabled,
			expectDiscoveryBlock:        defaultDiscoveryEnabled,
			expectDiscoveryEnabled:      defaultDiscoveryEnabled,
			expectUseSystemProbeLiteKey: defaultDiscoveryEnabled,
			expectUseSystemProbeLite:    defaultDiscoveryEnabled,
		},
		{
			name: "omitted discovery with latest follows get-agent-version policy",
			overrides: map[string]string{
				"agents.image.tag": "latest",
			},
			expectSystemProbe:           shouldAutoEnableDiscoveryFromTag("latest"),
			expectDiscoveryBlock:        shouldAutoEnableDiscoveryFromTag("latest"),
			expectDiscoveryEnabled:      shouldAutoEnableDiscoveryFromTag("latest"),
			expectUseSystemProbeLiteKey: shouldAutoEnableDiscoveryFromTag("latest"),
			expectUseSystemProbeLite:    shouldAutoEnableDiscoveryFromTag("latest"),
		},
		{
			name: "omitted discovery with latest-jmx follows get-agent-version policy",
			overrides: map[string]string{
				"agents.image.tag": "latest-jmx",
			},
			expectSystemProbe:           shouldAutoEnableDiscoveryFromTag("latest-jmx"),
			expectDiscoveryBlock:        shouldAutoEnableDiscoveryFromTag("latest-jmx"),
			expectDiscoveryEnabled:      shouldAutoEnableDiscoveryFromTag("latest-jmx"),
			expectUseSystemProbeLiteKey: shouldAutoEnableDiscoveryFromTag("latest-jmx"),
			expectUseSystemProbeLite:    shouldAutoEnableDiscoveryFromTag("latest-jmx"),
		},
		{
			name: "omitted discovery with unparseable tag is treated as latest",
			overrides: map[string]string{
				"agents.image.tag": "nightly",
			},
			expectSystemProbe:           true,
			expectDiscoveryBlock:        true,
			expectDiscoveryEnabled:      true,
			expectUseSystemProbeLiteKey: true,
			expectUseSystemProbeLite:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := renderDiscoveryManifest(t, tt.overrides)
			daemonset := extractAgentDaemonset(t, manifest)
			systemProbeContainer, hasSystemProbe := getContainer(t, daemonset.Spec.Template.Spec.Containers, "system-probe")
			assert.Equal(t, tt.expectSystemProbe, hasSystemProbe, "unexpected system-probe container presence")
			if tt.expectSystemProbe {
				assert.NotEmpty(t, systemProbeContainer.Command, "expected system-probe container command to be rendered")
			}

			systemProbeConfig, hasSystemProbeConfig := extractSystemProbeConfig(t, manifest)
			assert.Equal(t, tt.expectDiscoveryBlock, hasSystemProbeConfig, "unexpected system-probe config presence")

			if tt.expectDiscoveryBlock {
				discoveryConfig, found := nestedMap(systemProbeConfig, "discovery")
				require.True(t, found, "expected discovery block in system-probe config")
				assert.Equal(t, tt.expectDiscoveryEnabled, discoveryConfig["enabled"], "unexpected resolved discovery enabled value")

				useSystemProbeLiteValue, found := discoveryConfig["use_system_probe_lite"]
				assert.Equal(t, tt.expectUseSystemProbeLiteKey, found, "unexpected discovery.use_system_probe_lite presence")
				if tt.expectUseSystemProbeLiteKey {
					assert.Equal(t, tt.expectUseSystemProbeLite, useSystemProbeLiteValue, "unexpected discovery.use_system_probe_lite value")
				}
			}
		})
	}
}

func Test_serviceDiscoveryExplicitFalseRendersWhenAnotherSystemProbeFeatureIsEnabled(t *testing.T) {
	manifest := renderDiscoveryManifest(t, map[string]string{
		"agents.image.tag":                  "7.78.0",
		"datadog.discovery.enabled":         "false",
		"datadog.networkMonitoring.enabled": "true",
	})
	systemProbeConfig, ok := extractSystemProbeConfig(t, manifest)
	require.True(t, ok, "expected system-probe config to render")

	discoveryConfig, found := nestedMap(systemProbeConfig, "discovery")
	require.True(t, found, "expected discovery block to be rendered when discovery is explicitly disabled")
	assert.Equal(t, false, discoveryConfig["enabled"])
	assert.Equal(t, false, discoveryConfig["use_system_probe_lite"])

	networkConfig, found := nestedMap(systemProbeConfig, "network_config")
	require.True(t, found, "expected network_config block to be rendered")
	assert.Equal(t, true, networkConfig["enabled"])
}

func renderDiscoveryManifest(t *testing.T, overrides map[string]string) string {
	t.Helper()
	manifest, err := common.RenderChart(t, common.HelmCommand{
		ReleaseName: "datadog",
		ChartPath:   "../../charts/datadog",
		Values:      []string{"../../charts/datadog/values.yaml"},
		Overrides:   mergeDiscoveryOverrides(overrides),
	})
	require.NoError(t, err, "couldn't render chart")
	return manifest
}

func extractAgentDaemonset(t *testing.T, manifest string) appsv1.DaemonSet {
	t.Helper()
	var daemonset appsv1.DaemonSet
	require.True(t, decodeResourceByKindAndName(manifest, "DaemonSet", "datadog", &daemonset), "expected agent daemonset to be rendered")
	return daemonset
}

func extractSystemProbeConfig(t *testing.T, manifest string) (map[string]interface{}, bool) {
	t.Helper()
	var configMap corev1.ConfigMap
	if !decodeResourceByKindAndName(manifest, "ConfigMap", "datadog-system-probe-config", &configMap) {
		return nil, false
	}

	var config map[string]interface{}
	require.NoError(t, yaml.Unmarshal([]byte(configMap.Data["system-probe.yaml"]), &config))
	return config, true
}

func mergeDiscoveryOverrides(overrides map[string]string) map[string]string {
	merged := map[string]string{
		"datadog.apiKeyExistingSecret": "datadog-secret",
		"datadog.appKeyExistingSecret": "datadog-secret",
	}

	for key, value := range overrides {
		merged[key] = value
	}

	tag, hasTagOverride := overrides["agents.image.tag"]
	if hasTagOverride && isFloatingDiscoveryTag(tag) {
		merged["agents.image.doNotCheckTag"] = "true"
		merged["agents.useConfigMap"] = "false"
		merged["datadog.kubeStateMetricsCore.enabled"] = "false"
	}

	return merged
}

func nestedMap(root map[string]interface{}, key string) (map[string]interface{}, bool) {
	value, found := root[key]
	if !found {
		return nil, false
	}

	nested, ok := value.(map[string]interface{})
	return nested, ok
}

func decodeResourceByKindAndName(manifest, kind, name string, dest interface{}) bool {
	reader := yamlutil.NewYAMLReader(bufio.NewReader(bytes.NewReader([]byte(manifest))))

	for {
		resourceBytes, err := reader.Read()
		if err != nil {
			break
		}

		var resource map[string]interface{}
		if err := yaml.Unmarshal(resourceBytes, &resource); err != nil {
			return false
		}
		if len(resource) == 0 {
			continue
		}

		resourceKind, _ := resource["kind"].(string)
		metadata, _ := resource["metadata"].(map[string]interface{})
		resourceName, _ := metadata["name"].(string)
		if resourceKind != kind || resourceName != name {
			continue
		}

		if err := yaml.Unmarshal(resourceBytes, dest); err != nil {
			return false
		}
		return true
	}

	return false
}

func getDefaultAgentTag(t *testing.T) string {
	t.Helper()

	valuesFile, err := os.ReadFile("../../charts/datadog/values.yaml")
	require.NoError(t, err, "couldn't read chart values")

	var values struct {
		Agents struct {
			Image struct {
				Tag string `yaml:"tag"`
			} `yaml:"image"`
		} `yaml:"agents"`
	}

	require.NoError(t, yaml.Unmarshal(valuesFile, &values), "couldn't parse chart values")
	require.NotEmpty(t, values.Agents.Image.Tag, "expected a default agent tag in values.yaml")

	return values.Agents.Image.Tag
}

func shouldAutoEnableDiscoveryFromTag(tag string) bool {
	tag = normalizeDiscoveryTag(tag)
	switch tag {
	case "6":
		tag = "6.55.1"
	case "7", "latest":
		tag = "7.67.0"
	}

	normalized := normalizeDiscoveryVersion(tag)
	if normalized == "" {
		return true
	}

	version, err := semver.NewVersion(normalized)
	if err != nil {
		return false
	}

	constraint, err := semver.NewConstraint(">= 7.78.0-0")
	if err != nil {
		return false
	}

	return constraint.Check(version)
}

func normalizeDiscoveryVersion(tag string) string {
	tag = normalizeDiscoveryTag(tag)

	threeSegmentVersion := regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+([-.+][0-9A-Za-z.-]+)?$`)
	twoSegmentVersion := regexp.MustCompile(`^([0-9]+\.[0-9]+)([-.+].*)?$`)

	if threeSegmentVersion.MatchString(tag) {
		return tag
	}

	matches := twoSegmentVersion.FindStringSubmatch(tag)
	if len(matches) == 3 {
		return matches[1] + ".0" + matches[2]
	}

	return ""
}

func normalizeDiscoveryTag(tag string) string {
	tag = strings.TrimSpace(tag)
	return strings.TrimSuffix(tag, "-jmx")
}

func isFloatingDiscoveryTag(tag string) bool {
	tag = normalizeDiscoveryTag(tag)
	return tag == "latest" || normalizeDiscoveryVersion(tag) == ""
}
