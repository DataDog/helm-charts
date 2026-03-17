package datadog_csi_driver

import (
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/DataDog/helm-charts/test/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func Test_baseline_manifests(t *testing.T) {
	tests := []struct {
		name                 string
		command              common.HelmCommand
		baselineManifestPath string
		assertions           func(t *testing.T, baselineManifestPath, manifest string)
	}{
		{
			name: "CSI Driver DaemonSet default",
			command: common.HelmCommand{
				ReleaseName: "datadog-csi-driver",
				ChartPath:   "../../charts/datadog-csi-driver",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog-csi-driver/values.yaml"},
				Overrides:   map[string]string{},
			},
			baselineManifestPath: "./baseline/CSI_Driver_default.yaml",
			assertions:           verifyCSIDriverDaemonSet,
		},
		{
			name: "CSI Driver with annotations and security context set",
			command: common.HelmCommand{
				ReleaseName: "datadog-csi-driver",
				ChartPath:   "../../charts/datadog-csi-driver",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values: []string{
					"../../charts/datadog-csi-driver/values.yaml",
					"./manifests/added_annotation_and_securitycontext.yaml",
				},
				Overrides: map[string]string{},
			},
			baselineManifestPath: "./baseline/CSI_Driver_annotation_and_securitycontext.yaml",
			assertions:           verifyCSIDriverDaemonSet,
		},
		{
			name: "CSI Driver with nodeSelector and nodeAffinity set",
			command: common.HelmCommand{
				ReleaseName: "datadog-csi-driver",
				ChartPath:   "../../charts/datadog-csi-driver",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values: []string{
					"../../charts/datadog-csi-driver/values.yaml",
					"./manifests/added_nodeselector_and_nodeaffinity.yaml",
				},
				Overrides: map[string]string{},
			},
			baselineManifestPath: "./baseline/CSI_Driver_nodeselector_and_nodeaffinity.yaml",
			assertions:           verifyCSIDriverDaemonSet,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest, err := common.RenderChart(t, tt.command)
			assert.Nil(t, err, "couldn't render template")
			t.Log("update baselines", common.UpdateBaselines)
			if common.UpdateBaselines {
				common.WriteToFile(t, tt.baselineManifestPath, manifest)
			}

			tt.assertions(t, tt.baselineManifestPath, manifest)
		})
	}
}

func verifyCSIDriverDaemonSet(t *testing.T, baselineManifestPath, manifest string) {
	utils.VerifyBaseline(t, baselineManifestPath, manifest, appsv1.DaemonSet{}, appsv1.DaemonSet{})
}

func TestRegistryAllowList(t *testing.T) {
	tests := []struct {
		name   string
		values []string
		assert func(t *testing.T, env []corev1.EnvVar)
	}{
		{
			name: "DD_REGISTRY_ALLOW_LIST set when registryAllowList is non-empty",
			values: []string{
				"../../charts/datadog-csi-driver/values.yaml",
				"./manifests/registry_allow_list_set.yaml",
			},
			assert: func(t *testing.T, env []corev1.EnvVar) {
				e, ok := findEnvVar(env, "DD_REGISTRY_ALLOW_LIST")
				require.True(t, ok, "expected DD_REGISTRY_ALLOW_LIST to be present")
				assert.Equal(t, "public.ecr.aws/datadog,gcr.io/datadoghq", e.Value)
			},
		},
		{
			name:   "DD_REGISTRY_ALLOW_LIST absent when registryAllowList is empty",
			values: []string{"../../charts/datadog-csi-driver/values.yaml"},
			assert: func(t *testing.T, env []corev1.EnvVar) {
				_, ok := findEnvVar(env, "DD_REGISTRY_ALLOW_LIST")
				assert.False(t, ok, "expected DD_REGISTRY_ALLOW_LIST to be absent")
			},
		},
		{
			name:   "DD_REGISTRY_ALLOW_LIST absent when registryAllowList is undefined",
			values: []string{},
			assert: func(t *testing.T, env []corev1.EnvVar) {
				_, ok := findEnvVar(env, "DD_REGISTRY_ALLOW_LIST")
				assert.False(t, ok, "expected DD_REGISTRY_ALLOW_LIST to be absent")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest, err := common.RenderChart(t, common.HelmCommand{
				ReleaseName: "datadog-csi-driver",
				ChartPath:   "../../charts/datadog-csi-driver",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      tt.values,
			})
			require.Nil(t, err, "couldn't render template")

			var ds appsv1.DaemonSet
			common.Unmarshal(t, manifest, &ds)
			require.NotEmpty(t, ds.Spec.Template.Spec.Containers)

			var csiNodeDriverEnv []corev1.EnvVar
			for _, c := range ds.Spec.Template.Spec.Containers {
				if c.Name == "csi-node-driver" {
					csiNodeDriverEnv = c.Env
					break
				}
			}
			require.NotNil(t, csiNodeDriverEnv, "expected csi-node-driver container")

			tt.assert(t, csiNodeDriverEnv)
		})
	}
}

func findEnvVar(env []corev1.EnvVar, name string) (corev1.EnvVar, bool) {
	for _, e := range env {
		if e.Name == name {
			return e, true
		}
	}
	return corev1.EnvVar{}, false
}
