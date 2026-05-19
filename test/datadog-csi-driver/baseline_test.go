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
		{
			name: "CSI Driver with container resources set",
			command: common.HelmCommand{
				ReleaseName: "datadog-csi-driver",
				ChartPath:   "../../charts/datadog-csi-driver",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values: []string{
					"../../charts/datadog-csi-driver/values.yaml",
					"./manifests/added_resources.yaml",
				},
				Overrides: map[string]string{},
			},
			baselineManifestPath: "./baseline/CSI_Driver_resources.yaml",
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

func findCSIDriverEnvVar(env []corev1.EnvVar, name string) (corev1.EnvVar, bool) {
	for _, e := range env {
		if e.Name == name {
			return e, true
		}
	}
	return corev1.EnvVar{}, false
}

func Test_csi_driver_registryAllowList_envVar_only_when_explicitly_configured(t *testing.T) {
	tests := []struct {
		name        string
		overrides   map[string]string
		wantPresent bool
		wantValue   string
	}{
		{
			name:        "default values - env var is not set",
			overrides:   map[string]string{},
			wantPresent: false,
		},
		{
			name: "explicit allow list - env var is set",
			overrides: map[string]string{
				"global.apmRegistryAllowList[0]": "public.ecr.aws/datadog",
				"global.apmRegistryAllowList[1]": "gcr.io/datadoghq",
			},
			wantPresent: true,
			wantValue:   "public.ecr.aws/datadog,gcr.io/datadoghq",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest, err := common.RenderChart(t, common.HelmCommand{
				ReleaseName: "datadog-csi-driver",
				ChartPath:   "../../charts/datadog-csi-driver",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog-csi-driver/values.yaml"},
				Overrides:   tt.overrides,
			})
			require.NoError(t, err, "failed to render chart")

			var daemonSet appsv1.DaemonSet
			common.Unmarshal(t, manifest, &daemonSet)
			require.NotEmpty(t, daemonSet.Spec.Template.Spec.Containers, "expected at least one container in csi driver daemonset")

			csiContainer := daemonSet.Spec.Template.Spec.Containers[0]
			envVar, found := findCSIDriverEnvVar(csiContainer.Env, "DD_REGISTRY_ALLOW_LIST")

			if !tt.wantPresent {
				require.False(t, found, "did not expect DD_REGISTRY_ALLOW_LIST to be present")
				return
			}

			require.True(t, found, "expected DD_REGISTRY_ALLOW_LIST to be present")
			require.Equal(t, tt.wantValue, envVar.Value)
		})
	}
}

