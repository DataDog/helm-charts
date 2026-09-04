package datadog_operator

import (
	"strings"
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/DataDog/helm-charts/test/utils"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
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
			name: "Operator Deployment default",
			command: common.HelmCommand{
				ReleaseName: "datadog-operator",
				ChartPath:   "../../charts/datadog-operator",
				ShowOnly:    []string{"templates/deployment.yaml"},
				Values:      []string{"../../charts/datadog-operator/values.yaml"},
				Overrides:   map[string]string{},
			},
			baselineManifestPath: "./baseline/Operator_Deployment_default.yaml",
			assertions:           verifyOperatorDeployment,
			skipTest:             SkipTest,
		},
		{
			name: "DatadogAgent CRD default",
			command: common.HelmCommand{
				ReleaseName: "datadog-operator",
				ChartPath:   "../../charts/datadog-operator",
				// datadogCRDs is an alias defined in the chart dependency
				ShowOnly:  []string{"charts/datadogCRDs/templates/datadoghq.com_datadogagents_v1.yaml"},
				Values:    []string{"../../charts/datadog-operator/values.yaml"},
				Overrides: map[string]string{},
			},
			baselineManifestPath: "./baseline/DatadogAgent_CRD_default.yaml",
			assertions:           verifyDatadogAgent,
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

func verifyOperatorDeployment(t *testing.T, baselineManifestPath, manifest string) {
	var actual appsv1.Deployment
	common.Unmarshal(t, manifest, &actual)
	var baseline appsv1.Deployment
	common.LoadFromFile(t, baselineManifestPath, &baseline)

	// The image tag changes with every Operator release and is not part of the
	// chart structure this baseline is meant to protect, so it's stripped
	// before comparing (see CONTP-2001).
	stripImageTag(&actual)
	stripImageTag(&baseline)

	ops := cmp.Options{
		cmpopts.IgnoreMapEntries(func(k, v string) bool {
			return k == "helm.sh/chart"
		}),
	}
	assert.True(t, cmp.Equal(baseline, actual, ops), cmp.Diff(baseline, actual))
}

func stripImageTag(d *appsv1.Deployment) {
	for i, c := range d.Spec.Template.Spec.Containers {
		if repo, _, ok := strings.Cut(c.Image, ":"); ok {
			d.Spec.Template.Spec.Containers[i].Image = repo
		}
	}
}

func verifyDatadogAgent(t *testing.T, baselineManifestPath, manifest string) {
	utils.VerifyBaseline(t, baselineManifestPath, manifest, v1.CustomResourceDefinition{}, v1.CustomResourceDefinition{})
}
