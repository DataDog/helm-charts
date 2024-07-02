package operator_eks_addon

import (
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
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
			name: "Addon wrapper full render",
			command: common.HelmCommand{
				ReleaseName: "datadog-operator",
				ChartPath:   "../../charts/operator-eks-addon",
				ShowOnly:    []string{},
				Values:      []string{"../../charts/operator-eks-addon/values.yaml"},
				Overrides:   map[string]string{
					// "datadog-operator.image.tag": "1.7.0",
				},
			},
			baselineManifestPath: "./baseline/chart_render_default.yaml",
			assertions:           verify,
			skipTest:             false,
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

func verify(t *testing.T, baselineManifestPath, manifest string) {
	// do nothing
}
