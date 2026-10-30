package datadog

import (
	"strings"
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

func TestKSMDraRBAC(t *testing.T) {
	tests := []struct {
		name      string
		overrides map[string]string
		enabled   bool
	}{
		{
			name:      "cluster agent 7.85.0",
			overrides: map[string]string{"clusterAgent.image.tag": "7.85.0"},
			enabled:   true,
		},
		{
			name:      "cluster agent 7.85.0 release candidate",
			overrides: map[string]string{"clusterAgent.image.tag": "7.85.0-rc.1"},
			enabled:   true,
		},
		{
			name:      "cluster agent 7.84.3",
			overrides: map[string]string{"clusterAgent.image.tag": "7.84.3"},
		},
		{
			name:      "latest cluster agent resolves below 7.85.0",
			overrides: map[string]string{"clusterAgent.image.tag": "latest"},
		},
		{
			name: "cluster agent doNotCheckTag",
			overrides: map[string]string{
				"clusterAgent.image.tag":           "7.84.3",
				"clusterAgent.image.doNotCheckTag": "true",
			},
			enabled: true,
		},
		{
			name: "cluster checks runner 7.85.0",
			overrides: map[string]string{
				"datadog.kubeStateMetricsCore.useClusterCheckRunners": "true",
				"clusterChecksRunner.enabled":                         "true",
				"clusterAgent.image.tag":                              "7.84.3",
				"clusterChecksRunner.image.tag":                       "7.85.0",
			},
			enabled: true,
		},
		{
			name: "cluster checks runner 7.84.3",
			overrides: map[string]string{
				"datadog.kubeStateMetricsCore.useClusterCheckRunners": "true",
				"clusterChecksRunner.enabled":                         "true",
				"clusterAgent.image.tag":                              "7.85.0",
				"clusterChecksRunner.image.tag":                       "7.84.3",
			},
		},
		{
			name: "cluster checks runner doNotCheckTag",
			overrides: map[string]string{
				"datadog.kubeStateMetricsCore.useClusterCheckRunners": "true",
				"clusterChecksRunner.enabled":                         "true",
				"clusterChecksRunner.image.tag":                       "7.84.3",
				"clusterChecksRunner.image.doNotCheckTag":             "true",
			},
			enabled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run("full mode", func(t *testing.T) {
				roles := renderKSMClusterRoles(t, draTestOverrides(tt.overrides))
				require.Contains(t, roles, "datadog-ksm-core")
				var want []string
				if tt.enabled {
					want = []string{"resourceclaims", "resourceslices", "devicetaintrules"}
				}
				assert.ElementsMatch(t, want, draResources(t, roles["datadog-ksm-core"]))
			})

			t.Run("namespaced mode", func(t *testing.T) {
				overrides := draTestOverrides(tt.overrides)
				overrides["datadog.kubeStateMetricsCore.namespaces[0]"] = "team-a"
				roles := renderKSMClusterRoles(t, overrides)
				require.Contains(t, roles, "datadog-ksm-core-namespaced")
				require.Contains(t, roles, "datadog-ksm-core")
				var wantNamespaced, wantCluster []string
				if tt.enabled {
					wantNamespaced = []string{"resourceclaims"}
					wantCluster = []string{"resourceslices", "devicetaintrules"}
				}
				assert.ElementsMatch(t, wantNamespaced, draResources(t, roles["datadog-ksm-core-namespaced"]))
				assert.ElementsMatch(t, wantCluster, draResources(t, roles["datadog-ksm-core"]))
			})
		})
	}

	t.Run("ksm core disabled", func(t *testing.T) {
		_, err := common.RenderChart(t, common.HelmCommand{
			ReleaseName: "datadog",
			ChartPath:   "../../charts/datadog",
			ShowOnly:    []string{"templates/kube-state-metrics-core-rbac.yaml"},
			Values:      []string{"../../charts/datadog/values.yaml"},
			Overrides: draTestOverrides(map[string]string{
				"datadog.kubeStateMetricsCore.enabled": "false",
				"clusterAgent.image.tag":               "7.85.0",
			}),
		})
		// helm template errors when --show-only matches no manifest.
		require.ErrorContains(t, err, "could not find template")
	})
}

func TestGPUMonitoringCDIMount(t *testing.T) {
	for _, enabled := range []bool{true, false} {
		overrides := map[string]string{
			"datadog.apiKeyExistingSecret": "datadog-secret",
			"datadog.appKeyExistingSecret": "datadog-secret",
		}
		if enabled {
			overrides["datadog.gpuMonitoring.enabled"] = "true"
		}
		manifest, err := common.RenderChart(t, common.HelmCommand{
			ReleaseName: "datadog",
			ChartPath:   "../../charts/datadog",
			ShowOnly:    []string{"templates/daemonset.yaml"},
			Values:      []string{"../../charts/datadog/values.yaml"},
			Overrides:   overrides,
		})
		require.NoError(t, err)

		var ds appsv1.DaemonSet
		common.Unmarshal(t, manifest, &ds)

		var volume *corev1.Volume
		for i := range ds.Spec.Template.Spec.Volumes {
			if ds.Spec.Template.Spec.Volumes[i].Name == "host-var-run-cdi" {
				volume = &ds.Spec.Template.Spec.Volumes[i]
			}
		}
		var mount *corev1.VolumeMount
		for _, c := range ds.Spec.Template.Spec.Containers {
			if c.Name != "agent" {
				continue
			}
			seen := map[string]bool{}
			for i, m := range c.VolumeMounts {
				assert.False(t, seen[m.MountPath], "duplicate mountPath %s", m.MountPath)
				seen[m.MountPath] = true
				if m.Name == "host-var-run-cdi" {
					mount = &c.VolumeMounts[i]
				}
			}
		}

		if !enabled {
			assert.Nil(t, volume)
			assert.Nil(t, mount)
			continue
		}
		require.NotNil(t, volume)
		require.NotNil(t, volume.HostPath)
		assert.Equal(t, "/var/run/cdi", volume.HostPath.Path)
		require.NotNil(t, volume.HostPath.Type)
		assert.Equal(t, corev1.HostPathDirectoryOrCreate, *volume.HostPath.Type)
		require.NotNil(t, mount)
		assert.Equal(t, "/host/var/run/cdi", mount.MountPath)
		assert.True(t, mount.ReadOnly)
	}
}

func draTestOverrides(overrides map[string]string) map[string]string {
	merged := map[string]string{
		"datadog.apiKeyExistingSecret":         "datadog-secret",
		"datadog.appKeyExistingSecret":         "datadog-secret",
		"datadog.kubeStateMetricsCore.enabled": "true",
	}
	for k, v := range overrides {
		merged[k] = v
	}
	return merged
}

func renderKSMClusterRoles(t *testing.T, overrides map[string]string) map[string]rbacv1.ClusterRole {
	manifest, err := common.RenderChart(t, common.HelmCommand{
		ReleaseName: "datadog",
		ChartPath:   "../../charts/datadog",
		ShowOnly:    []string{"templates/kube-state-metrics-core-rbac.yaml"},
		Values:      []string{"../../charts/datadog/values.yaml"},
		Overrides:   overrides,
	})
	require.NoError(t, err)

	roles := map[string]rbacv1.ClusterRole{}
	for _, doc := range strings.Split(manifest, "---") {
		if !strings.Contains(doc, "kind: ClusterRole\n") {
			continue
		}
		var role rbacv1.ClusterRole
		common.Unmarshal(t, doc, &role)
		if role.Kind == "ClusterRole" {
			roles[role.Name] = role
		}
	}
	return roles
}

func draResources(t *testing.T, role rbacv1.ClusterRole) []string {
	var resources []string
	for _, rule := range role.Rules {
		for _, group := range rule.APIGroups {
			if group != "resource.k8s.io" {
				continue
			}
			assert.ElementsMatch(t, []string{"list", "watch"}, rule.Verbs)
			resources = append(resources, rule.Resources...)
		}
	}
	return resources
}
