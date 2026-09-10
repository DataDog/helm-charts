package datadog_operator

import (
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/DataDog/helm-charts/test/utils"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// Test_baseline_deployment and Test_baseline_crd used to be rows of one shared
// table-driven test, but they have different baseline-regeneration semantics:
// the Deployment baseline is normalized against release-version noise (see
// stripReleaseVersion) so it never needs regenerating on a release, while the
// CRD baseline has no such normalization and must be regenerated whenever
// datadog-crds brings in a real schema change. They're split into separate
// top-level tests so release automation can regenerate only the CRD baseline
// via `go test -run '^Test_baseline_crd$'` (see Makefile's
// update-test-baselines-operator-crd) without a Deployment template regression
// silently getting absorbed into that same automated commit (see CONTP-2001).

func Test_baseline_deployment(t *testing.T) {
	if SkipTest {
		t.Skip()
	}
	command := common.HelmCommand{
		ReleaseName: "datadog-operator",
		ChartPath:   "../../charts/datadog-operator",
		ShowOnly:    []string{"templates/deployment.yaml"},
		Values:      []string{"../../charts/datadog-operator/values.yaml"},
		Overrides:   map[string]string{},
	}
	baselineManifestPath := "./baseline/Operator_Deployment_default.yaml"

	manifest, err := common.RenderChart(t, command)
	assert.Nil(t, err, "couldn't render template")
	t.Log("update baselines", common.UpdateBaselines)
	if common.UpdateBaselines {
		common.WriteToFile(t, baselineManifestPath, manifest)
	}

	verifyOperatorDeployment(t, baselineManifestPath, manifest)
}

func Test_baseline_crd(t *testing.T) {
	if SkipTest {
		t.Skip()
	}
	command := common.HelmCommand{
		ReleaseName: "datadog-operator",
		ChartPath:   "../../charts/datadog-operator",
		// datadogCRDs is an alias defined in the chart dependency
		ShowOnly:  []string{"charts/datadogCRDs/templates/datadoghq.com_datadogagents_v1.yaml"},
		Values:    []string{"../../charts/datadog-operator/values.yaml"},
		Overrides: map[string]string{},
	}
	baselineManifestPath := "./baseline/DatadogAgent_CRD_default.yaml"

	manifest, err := common.RenderChart(t, command)
	assert.Nil(t, err, "couldn't render template")
	t.Log("update baselines", common.UpdateBaselines)
	if common.UpdateBaselines {
		common.WriteToFile(t, baselineManifestPath, manifest)
	}

	verifyDatadogAgent(t, baselineManifestPath, manifest)
}

func verifyOperatorDeployment(t *testing.T, baselineManifestPath, manifest string) {
	// The image tag and the "app.kubernetes.io/version" label (sourced from
	// Chart.AppVersion) both change with every Operator release and aren't
	// part of the chart structure this baseline is meant to protect, so
	// they're stripped before comparing (see CONTP-2001).
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
