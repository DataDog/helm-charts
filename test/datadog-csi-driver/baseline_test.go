package datadog_csi_driver

import (
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
)

func Test_baseline_manifests(t *testing.T) {
	tests := []struct {
		name                 string
		command              common.HelmCommand
		baselineManifestPath string
		assertions           func(t *testing.T, baselineManifestPath, manifest string)
		skipTest             bool
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
			skipTest:             SkipTest,
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
			skipTest:             SkipTest,
		},
	}

	for _, tt := range tests {
		if tt.skipTest {
			continue
		}
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
	verifyBaseline(t, baselineManifestPath, manifest, appsv1.DaemonSet{}, appsv1.DaemonSet{})
}

func verifyBaseline[T any](t *testing.T, baselineManifestPath, manifest string, baseline, actual T) {
	common.Unmarshal(t, manifest, &actual)
	common.LoadFromFile(t, baselineManifestPath, &baseline)

	// Exclude "helm.sh/chart" label from comparison to avoid
	// updating baselines on every unrelated chart changes.
	ops := make(cmp.Options, 0)
	ops = append(ops, cmpopts.IgnoreMapEntries(func(k, v string) bool {
		return k == "helm.sh/chart"
	}))

	assert.True(t, cmp.Equal(baseline, actual, ops), cmp.Diff(baseline, actual))
}
