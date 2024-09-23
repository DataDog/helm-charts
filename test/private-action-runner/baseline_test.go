package private_action_runner

import (
	appsv1 "k8s.io/api/apps/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
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
			name: "Private Action Runner default",
			command: common.HelmCommand{
				ReleaseName: "private-action-runner",
				ChartPath:   "../../charts/private-action-runner",
				Values:      []string{"../../charts/private-action-runner/values.yaml"},
				Overrides:   map[string]string{},
			},
			baselineManifestPath: "./baseline/Private_Action_Runner_default.yaml",
			assertions:           verifyPrivateActionRunner,
		},
		{
			name: "Enable kubernetes actions",
			command: common.HelmCommand{
				ReleaseName: "private-action-runner",
				ChartPath:   "../../charts/private-action-runner",
				Values:      []string{"../../charts/private-action-runner/values.yaml"},
				Overrides: map[string]string{
					"runners[0].kubernetesActions.controllerRevisions": "{get,list,create,update,patch,delete,deleteMultiple}",
					"runners[0].kubernetesActions.customObjects":       "{deleteMultiple}",
					"runners[0].kubernetesActions.deployments":         "{restart}",
					"runners[0].kubernetesActions.endpoints":           "{patch}",
					"runners[0].kubernetesPermissions[0].apiGroups":    "{example.com}",
					"runners[0].kubernetesPermissions[0].resources":    "{tests}",
					"runners[0].kubernetesPermissions[0].verbs":        "{list,get,create,patch,update,delete}",
				},
			},
			baselineManifestPath: "./baseline/Kubernetes_Actions.yaml",
			assertions:           verifyPrivateActionRunner,
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

func verifyPrivateActionRunner(t *testing.T, baselineManifestPath, manifest string) {
	verifyBaseline(t, baselineManifestPath, manifest, appsv1.Deployment{}, appsv1.Deployment{})
	verifyBaseline(t, baselineManifestPath, manifest, rbacv1.ClusterRole{}, rbacv1.ClusterRole{})
	verifyBaseline(t, baselineManifestPath, manifest, rbacv1.RoleBinding{}, rbacv1.RoleBinding{})
	verifyBaseline(t, baselineManifestPath, manifest, corev1.Secret{}, corev1.Secret{})
	verifyBaseline(t, baselineManifestPath, manifest, corev1.Service{}, corev1.Service{})
	verifyBaseline(t, baselineManifestPath, manifest, corev1.ServiceAccount{}, corev1.ServiceAccount{})
}

func verifyBaseline[T any](t *testing.T, baselineManifestPath, manifest string, baseline, actual T) {
	common.Unmarshal(t, manifest, &actual)
	common.LoadFromFile(t, baselineManifestPath, &baseline)

	assert.True(t, cmp.Equal(baseline, actual, cmp.Options{}), cmp.Diff(baseline, actual))
}
