package datadog_operator

import (
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
		// needsUpdate reports whether the baseline should actually be
		// rewritten when regenerating. Nil means always rewrite. The
		// Deployment baseline sets this so a routine Operator release
		// (image tag / version bump only) doesn't touch the file.
		needsUpdate func(t *testing.T, baselineManifestPath, manifest string) bool
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
			needsUpdate:          deploymentNeedsUpdate,
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
			// No needsUpdate: a real datadog-crds schema change is real
			// content, so this baseline always regenerates.
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
			if common.UpdateBaselines && (tt.needsUpdate == nil || tt.needsUpdate(t, tt.baselineManifestPath, manifest)) {
				common.WriteToFile(t, tt.baselineManifestPath, manifest)
			}

			tt.assertions(t, tt.baselineManifestPath, manifest)
		})
	}
}

// deploymentNeedsUpdate reports whether the rendered Deployment differs from
// the baseline once release-specific fields (image tag, version label) are
// stripped from both, so regenerating on a routine release is a no-op.
func deploymentNeedsUpdate(t *testing.T, baselineManifestPath, manifest string) bool {
	var actual, baseline appsv1.Deployment
	common.Unmarshal(t, manifest, &actual)
	common.LoadFromFile(t, baselineManifestPath, &baseline)
	stripReleaseVersion(&actual)
	stripReleaseVersion(&baseline)

	ops := cmpopts.IgnoreMapEntries(func(k, v string) bool {
		return k == "helm.sh/chart"
	})
	return !cmp.Equal(baseline, actual, ops)
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
