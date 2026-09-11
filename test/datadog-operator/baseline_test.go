package datadog_operator

import (
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/DataDog/helm-charts/test/utils"
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
		},
	}

	for _, tt := range tests {
		if SkipTest {
			t.Skip()
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
	// The image tag and the "app.kubernetes.io/version" label (sourced from
	// Chart.AppVersion) both change with every Operator release and aren't
	// part of the chart structure this baseline is meant to protect, so
	// they're stripped before comparing.
	utils.VerifyBaseline(t, baselineManifestPath, manifest, appsv1.Deployment{}, appsv1.Deployment{}, stripReleaseVersion)
}

func stripReleaseVersion(d *appsv1.Deployment) {
	for i, c := range d.Spec.Template.Spec.Containers {
		d.Spec.Template.Spec.Containers[i].Image = utils.ImageRepository(c.Image)
	}
	delete(d.Labels, "app.kubernetes.io/version")
}

func verifyDatadogAgent(t *testing.T, baselineManifestPath, manifest string) {
	utils.VerifyBaseline(t, baselineManifestPath, manifest, v1.CustomResourceDefinition{}, v1.CustomResourceDefinition{})
}
