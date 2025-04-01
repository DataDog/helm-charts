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
				ReleaseName: "default-test",
				ChartPath:   "../../charts/private-action-runner",
				Values:      []string{"../../charts/private-action-runner/values.yaml"},
				Overrides:   map[string]string{},
			},
			snapshotName: "default",
			assertions:   verifyPrivateActionRunner,
		}, {
			name: "Private Action Runner example file",
			command: common.HelmCommand{
				ReleaseName: "example-test",
				ChartPath:   "../../charts/private-action-runner",
				Values:      []string{"../../charts/private-action-runner/examples/values.yaml"},
				Overrides:   map[string]string{},
			},
			snapshotName: "example",
			assertions:   verifyPrivateActionRunner,
		},
		{
			name: "Enable kubernetes actions",
			command: common.HelmCommand{
				ReleaseName: "kubernetes-test",
				ChartPath:   "../../charts/private-action-runner",
				Values:      []string{"../../charts/private-action-runner/values.yaml"},
				OverridesJson: map[string]string{
					"runner.roleType": `"ClusterRole"`,
					"runner.kubernetesActions.controllerRevisions": `["get","list","create","update","patch","delete","deleteMultiple"]`,
					"runner.kubernetesActions.customObjects":       `["deleteMultiple"]`,
					"runner.kubernetesActions.deployments":         `["restart"]`,
					"runner.kubernetesActions.endpoints":           `["patch"]`,
					"runner.kubernetesPermissions[0].apiGroups":    `["example.com"]`,
					"runner.kubernetesPermissions[0].resources":    `["tests"]`,
					"runner.kubernetesPermissions[0].verbs":        `["list","get","create","patch","update","delete"]`,
				},
			},
			snapshotName: "enable-kubernetes-actions",
			assertions:   verifyPrivateActionRunner,
		},
		{
			name: "Specify certain config overrides",
			command: common.HelmCommand{
				ReleaseName: "override-test",
				ChartPath:   "../../charts/private-action-runner",
				Values:      []string{"../../charts/private-action-runner/values.yaml"},
				OverridesJson: map[string]string{
					"fullnameOverride": `"custom-full-name"`,
					"runner.env":       `[ {"name": "FOO", "value": "foo"}, {"name": "BAR", "value": "bar"} ]`,
				},
			},
			snapshotName: "config-overrides",
			assertions:   verifyPrivateActionRunner,
		},
		{
			name: "Specify secrets externally",
			command: common.HelmCommand{
				ReleaseName: "secrets-test",
				ChartPath:   "../../charts/private-action-runner",
				Values:      []string{"../../charts/private-action-runner/values.yaml"},
				OverridesJson: map[string]string{
					"runner.runnerIdentitySecret": `"the-name-of-the-secret"`,
					"runner.config.urn":           ``,
					"runner.config.privateKey":    ``,
					"runner.credentialSecrets":    `[{"secretName": "first-secret"}, {"secretName": "second-secret", "directoryName": "second-secret-directory"}]`,
				},
			},
			snapshotName: "external-secrets",
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
