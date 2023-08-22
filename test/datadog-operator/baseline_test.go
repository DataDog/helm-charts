package datadog_operator

import (
	"testing"

	"github.com/DataDog/helm-charts/test/common"
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
			name: "Operator Deployment with cert manager enabled",
			command: common.HelmCommand{
				ReleaseName: "datadog-operator",
				ChartPath:   "../../charts/datadog-operator",
				ShowOnly:    []string{"templates/deployment.yaml"},
				Values:      []string{"../../charts/datadog-operator/values.yaml"},
				Overrides: map[string]string{
					"datadogCRDs.migration.datadogAgents.useCertManager":            "true",
					"datadogCRDs.migration.datadogAgents.conversionWebhook.enabled": "true",
				},
			},
			baselineManifestPath: "./baseline/Operator_Deployment_with_certManager.yaml",
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
		{
			name: "DatadogAgent CRD with cert manager enabled",
			command: common.HelmCommand{
				ReleaseName: "datadog-operator",
				ChartPath:   "../../charts/datadog-operator",
				// datadogCRDs is an alias defined in the chart dependency
				ShowOnly: []string{"charts/datadogCRDs/templates/datadoghq.com_datadogagents_v1.yaml"},
				Values:   []string{"../../charts/datadog-operator/values.yaml"},
				Overrides: map[string]string{
					"datadogCRDs.migration.datadogAgents.useCertManager":            "true",
					"datadogCRDs.migration.datadogAgents.conversionWebhook.enabled": "true",
				},
			},
			baselineManifestPath: "./baseline/DatadogAgent_CRD_with_certManager.yaml",
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
	verifyBaseline(t, baselineManifestPath, manifest, appsv1.Deployment{}, appsv1.Deployment{})
}

func verifyDatadogAgent(t *testing.T, baselineManifestPath, manifest string) {
	verifyBaseline(t, baselineManifestPath, manifest, v1.CustomResourceDefinition{}, v1.CustomResourceDefinition{})
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
