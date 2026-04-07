package datadog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"

	"github.com/DataDog/helm-charts/test/common"
)

func systemProbeConfigmapCmd(overrides map[string]string) common.HelmCommand {
	merged := map[string]string{
		"datadog.apiKeyExistingSecret": "datadog-secret",
		"datadog.appKeyExistingSecret": "datadog-secret",
	}
	for k, v := range overrides {
		merged[k] = v
	}
	return common.HelmCommand{
		ReleaseName: "datadog",
		ChartPath:   "../../charts/datadog",
		ShowOnly:    []string{"templates/system-probe-configmap.yaml"},
		Overrides:   merged,
	}
}

func Test_systemProbeConfigmap_discovery(t *testing.T) {
	tests := []struct {
		name                   string
		overrides              map[string]string
		expectDiscoveryEnabled *bool
	}{
		{
			name: "discovery.enabled=false with other SP feature -- discovery block rendered as false",
			overrides: map[string]string{
				"datadog.discovery.enabled":         "false",
				"datadog.networkMonitoring.enabled": "true",
			},
			expectDiscoveryEnabled: boolPtr(false),
		},
		{
			name: "discovery.enabled=true -- discovery block rendered as true",
			overrides: map[string]string{
				"datadog.discovery.enabled": "true",
			},
			expectDiscoveryEnabled: boolPtr(true),
		},
		{
			name: "enabledByDefault=true -- discovery block rendered as true",
			overrides: map[string]string{
				"datadog.discovery.enabledByDefault": "true",
			},
			expectDiscoveryEnabled: boolPtr(true),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest, err := common.RenderChart(t, systemProbeConfigmapCmd(tt.overrides))
			require.NoError(t, err, "couldn't render template")

			var cm v1.ConfigMap
			common.Unmarshal(t, manifest, &cm)

			spYaml, ok := cm.Data["system-probe.yaml"]
			require.True(t, ok, "system-probe.yaml key not found in configmap")

			var spConfig map[string]interface{}
			require.NoError(t, yaml.Unmarshal([]byte(spYaml), &spConfig))

			if tt.expectDiscoveryEnabled != nil {
				discoveryRaw, ok := spConfig["discovery"]
				require.True(t, ok, "discovery section not found in system-probe.yaml")
				discovery, ok := discoveryRaw.(map[string]interface{})
				require.True(t, ok, "discovery section is not a map")
				assert.Equal(t, *tt.expectDiscoveryEnabled, discovery["enabled"], "unexpected discovery.enabled value")
			}
		})
	}
}

func boolPtr(b bool) *bool { return &b }

func Test_systemProbeContainer_cgroupsMount(t *testing.T) {
	tests := []struct {
		name              string
		overrides         map[string]string
		expectCgroupMount bool
	}{
		{
			name: "enabledByDefault=true -- cgroups mount present",
			overrides: map[string]string{
				"datadog.discovery.enabledByDefault": "true",
			},
			expectCgroupMount: true,
		},
		{
			name: "discovery.enabled=true -- cgroups mount present",
			overrides: map[string]string{
				"datadog.discovery.enabled": "true",
			},
			expectCgroupMount: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest, err := common.RenderChart(t, common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Overrides: func() map[string]string {
					m := map[string]string{
						"datadog.apiKeyExistingSecret": "datadog-secret",
						"datadog.appKeyExistingSecret": "datadog-secret",
					}
					for k, v := range tt.overrides {
						m[k] = v
					}
					return m
				}(),
			})
			require.NoError(t, err, "couldn't render template")

			var ds appsv1.DaemonSet
			common.Unmarshal(t, manifest, &ds)

			spContainer, found := getContainer(t, ds.Spec.Template.Spec.Containers, "system-probe")
			require.True(t, found, "system-probe container should exist")

			hasCgroups := false
			for _, vm := range spContainer.VolumeMounts {
				if vm.Name == "cgroups" {
					hasCgroups = true
					break
				}
			}
			assert.Equal(t, tt.expectCgroupMount, hasCgroups, "unexpected cgroups mount presence")
		})
	}
}
