package private_action_runner

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"

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
		},
		{
			name: "Private Action Runner deprecated mode values",
			command: common.HelmCommand{
				ReleaseName: "deprecated-modes-test",
				ChartPath:   "../../charts/private-action-runner",
				Values:      []string{"../../charts/private-action-runner/values.yaml"},
				OverridesJson: map[string]string{
					"runner.config.modes": `["workflowAutomation", "appBuilder"]`,
				},
			},
			snapshotName: "deprecated-modes",
			assertions:   verifyPrivateActionRunner,
		},
		{
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
					"runner.kubernetesActions.deployments":         `["restart", "rollback", "scale"]`,
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
					"fullnameOverride":                 `"custom-full-name"`,
					"runner.env":                       `[ {"name": "FOO", "value": "foo"}, {"name": "BAR", "value": "bar"} ]`,
					"runner.config.allowIMDSEndpoint":  `true`,
					"runner.config.tags":               `["foo:bar", "bar:baz"]`,
					"runner.config.taskTimeoutSeconds": `25`,
					"runner.config.httpTimeoutSeconds": `10`,
					"image.pullPolicy":                 `"Always"`,
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
					"runner.env":                  `[{"name": "FOO", "value": "foo"}]`,
					"runner.credentialSecrets":    `[{"secretName": "first-secret"}, {"secretName": "second-secret", "directoryName": "second-secret-directory"}]`,
				},
			},
			snapshotName: "external-secrets",
			assertions:   verifyPrivateActionRunner,
		},
		{
			name: "Custom resource requirements",
			command: common.HelmCommand{
				ReleaseName: "resources-test",
				ChartPath:   "../../charts/private-action-runner",
				Values:      []string{"../../charts/private-action-runner/values.yaml"},
				OverridesJson: map[string]string{
					"runner.resources.limits.cpu":      `"500m"`,
					"runner.resources.limits.memory":   `"2Gi"`,
					"runner.resources.requests.cpu":    `"100m"`,
					"runner.resources.requests.memory": `"512Mi"`,
				},
			},
			snapshotName: "custom-resources",
			assertions:   verifyPrivateActionRunner,
		},
		{
			name: "Custom pod scheduling",
			command: common.HelmCommand{
				ReleaseName: "resources-test",
				ChartPath:   "../../charts/private-action-runner",
				Values:      []string{"../../charts/private-action-runner/values.yaml"},
				OverridesJson: map[string]string{
					"runner.nodeSelector": `{"kubernetes.io/os": "linux"}`,
					"runner.tolerations":  `[{"key": "taint.custom.com/key", "effect": "NoSchedule", "operator": "Exists"}]`,
					"runner.affinity":     `{"nodeAffinity": {"requiredDuringSchedulingIgnoredDuringExecution": {"nodeSelectorTerms": [{"matchExpressions": [{"key": "kubernetes.io/arch", "operator": "In", "values": ["amd64"]}]}]}}}`,
				},
			},
			snapshotName: "custom-pod-scheduling",
			assertions:   verifyPrivateActionRunner,
		},
		{
			name: "Scripts configuration",
			command: common.HelmCommand{
				ReleaseName: "scripts-test",
				ChartPath:   "../../charts/private-action-runner",
				Values:      []string{"../../charts/private-action-runner/values.yaml"},
				OverridesJson: map[string]string{
					"runner.credentialFiles": `[{"fileName": "script.yaml", "data": "echoInBash:\n  command: [\"bash\", \"/home/scriptuser/hello-from-bash.sh\"]"}]`,
					"runner.scriptFiles":     `[{"fileName": "hello-from-bash.sh", "data": "#!/bin/bash\necho \"Hello World from bash!\""}]`,
				},
			},
			snapshotName: "scripts-configuration",
			assertions:   verifyPrivateActionRunner,
		},
		{
			name: "Service annotations",
			command: common.HelmCommand{
				ReleaseName: "scripts-test",
				ChartPath:   "../../charts/private-action-runner",
				Values:      []string{"../../charts/private-action-runner/values.yaml"},
				OverridesJson: map[string]string{
					"service.annotations": `{"example.com/custom-annotation": "custom-value", "service.beta.kubernetes.io/aws-load-balancer-type": "nlb"}`,
				},
			},
			snapshotName: "service-annotations",
			assertions:   verifyPrivateActionRunner,
		},
		{
			name: "Deployment metadata annotations",
			command: common.HelmCommand{
				ReleaseName: "deployment-metadata-test",
				ChartPath:   "../../charts/private-action-runner",
				Values:      []string{"../../charts/private-action-runner/values.yaml"},
				OverridesJson: map[string]string{
					"deployment.metadata.annotations": `{"example.com/owner": "platform-team", "deployment.kubernetes.io/revision": "1"}`,
				},
			},
			snapshotName: "deployment-metadata-annotations",
			assertions:   verifyPrivateActionRunner,
		},
		{
			name: "Deployment metadata labels",
			command: common.HelmCommand{
				ReleaseName: "deployment-metadata-labels-test",
				ChartPath:   "../../charts/private-action-runner",
				Values:      []string{"../../charts/private-action-runner/values.yaml"},
				OverridesJson: map[string]string{
					"deployment.metadata.labels": `{"custom-label": "custom-value", "environment": "production"}`,
				},
			},
			snapshotName: "deployment-metadata-labels",
			assertions:   verifyPrivateActionRunner,
		},
		{
			name: "Runner pod annotations",
			command: common.HelmCommand{
				ReleaseName: "runner-pod-test",
				ChartPath:   "../../charts/private-action-runner",
				Values:      []string{"../../charts/private-action-runner/values.yaml"},
				OverridesJson: map[string]string{
					"runner.podAnnotations": `{"prometheus.io/scrape": "true", "prometheus.io/port": "9016"}`,
				},
			},
			snapshotName: "pod-annotations",
			assertions:   verifyPrivateActionRunner,
		},
		{
			name: "Deployment metadata and runner pod annotations",
			command: common.HelmCommand{
				ReleaseName: "deployment-runner-test",
				ChartPath:   "../../charts/private-action-runner",
				Values:      []string{"../../charts/private-action-runner/values.yaml"},
				OverridesJson: map[string]string{
					"deployment.metadata.annotations": `{"example.com/owner": "platform-team"}`,
					"runner.podAnnotations":           `{"prometheus.io/scrape": "true"}`,
				},
			},
			snapshotName: "deployment-runner-annotations",
			assertions:   verifyPrivateActionRunner,
		},
		{
			name: "SecurityContextConstraints enabled",
			command: common.HelmCommand{
				ReleaseName: "scc-test",
				ChartPath:   "../../charts/private-action-runner",
				Values:      []string{"../../charts/private-action-runner/values.yaml"},
				OverridesJson: map[string]string{
					"runner.podSecurity.securityContextConstraints.create": `true`,
					"runner.podSecurity.privileged":                        `false`,
					"runner.podSecurity.requiredDropCapabilities":          `["ALL"]`,
					"runner.podSecurity.seLinuxContext.type":               `"RunAsAny"`,
				},
			},
			snapshotName: "scc-enabled",
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
