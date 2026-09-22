package datadog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DataDog/helm-charts/test/common"
)

func Test_KubernetesActions_HelmActions_Enabled(t *testing.T) {
	manifest, err := common.RenderChart(t, common.HelmCommand{
		ReleaseName: "datadog",
		ChartPath:   "../../charts/datadog",
		ShowOnly:    []string{"templates/cluster-agent-helm-actions-rbac.yaml"},
		Values:      []string{"../../charts/datadog/values.yaml"},
		Overrides: map[string]string{
			"datadog.kubernetesActions.enabled":             "true",
			"datadog.kubernetesActions.helmActions.enabled": "true",
			"clusterAgent.image.tag":                        "7.85.0",
		},
	})
	require.NoError(t, err)

	assert.Contains(t, manifest, "name: datadog-helm-action-caller")
	assert.Contains(t, manifest, "name: datadog-helm-release-storage")
	assert.Contains(t, manifest, "name: datadog-helm-to-resources")
	assert.Contains(t, manifest, "name: datadog-helm-actions")
	assert.Contains(t, manifest, "kind: ServiceAccount")
}

func Test_KubernetesActions_HelmActions_Enabled_WithoutKubernetesActionsEnabled(t *testing.T) {
	// helmActions.enabled=true alone must not grant the RBAC: kubernetesActions.enabled
	// gates the feature path (DD_KUBEACTIONS_ENABLED) that the Job's permissions rely on.
	manifest, err := common.RenderChart(t, common.HelmCommand{
		ReleaseName: "datadog",
		ChartPath:   "../../charts/datadog",
		ShowOnly:    []string{"templates/cluster-agent-helm-actions-rbac.yaml"},
		Values:      []string{"../../charts/datadog/values.yaml"},
		Overrides: map[string]string{
			"datadog.kubernetesActions.enabled":             "false",
			"datadog.kubernetesActions.helmActions.enabled": "true",
			"clusterAgent.image.tag":                        "7.85.0",
		},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "could not find template")
	assert.NotContains(t, manifest, "helm-action-caller")
}

func Test_KubernetesActions_HelmActions_Disabled_By_Default(t *testing.T) {
	manifest, err := common.RenderChart(t, common.HelmCommand{
		ReleaseName: "datadog",
		ChartPath:   "../../charts/datadog",
		ShowOnly:    []string{"templates/cluster-agent-helm-actions-rbac.yaml"},
		Values:      []string{"../../charts/datadog/values.yaml"},
	})
	require.Error(t, err)
	assert.NotContains(t, manifest, "helm-action-caller")
}

func Test_KubernetesActions_HelmActions_RequiresCompatibleClusterAgent(t *testing.T) {
	manifest, err := common.RenderChart(t, common.HelmCommand{
		ReleaseName: "datadog",
		ChartPath:   "../../charts/datadog",
		ShowOnly:    []string{"templates/cluster-agent-helm-actions-rbac.yaml"},
		Values:      []string{"../../charts/datadog/values.yaml"},
		Overrides: map[string]string{
			"datadog.kubernetesActions.enabled":             "true",
			"datadog.kubernetesActions.helmActions.enabled": "true",
			"clusterAgent.image.tag":                        "7.82.3",
		},
	})
	require.Error(t, err)
	assert.NotContains(t, manifest, "helm-action-caller")
}
