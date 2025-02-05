package private_action_runner

import (
	"github.com/gruntwork-io/terratest/modules/helm"
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
)

func Test_baseline_manifests(t *testing.T) {
	tests := []struct {
		name         string
		command      common.HelmCommand
		snapshotName string
		assertions   func(t *testing.T, manifest, snapshotName string)
	}{
		{
			name: "Private Action Runner default",
			command: common.HelmCommand{
				ReleaseName: "private-action-runner",
				ChartPath:   "../../charts/private-action-runner",
				Values:      []string{"../../charts/private-action-runner/values.yaml"},
				Overrides:   map[string]string{},
			},
			snapshotName: "default",
			assertions:   verifyPrivateActionRunner,
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
			snapshotName: "enable-kubernetes-actions",
			assertions:   verifyPrivateActionRunner,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest, err := common.RenderChart(t, tt.command)
			assert.Nil(t, err, "couldn't render template")
			t.Log("update baselines", common.UpdateBaselines)
			if common.UpdateBaselines {
				helm.UpdateSnapshot(t, &helm.Options{}, manifest, tt.snapshotName)
			}

			tt.assertions(t, manifest, tt.snapshotName)
		})
	}
}

func verifyPrivateActionRunner(t *testing.T, manifest string, snapshotName string) {
	diffCount := helm.DiffAgainstSnapshot(t, &helm.Options{}, manifest, snapshotName)
	assert.Equal(t, 0, diffCount, "manifests are different")
}
