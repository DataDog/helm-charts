package datadog

import (
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
)

func Test_nodeAgentPodName(t *testing.T) {
	for _, targetSystem := range []string{"linux", "windows"} {
		t.Run(targetSystem, func(t *testing.T) {
			manifest, err := common.RenderChart(t, common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret": "datadog-secret",
					"targetSystem":                 targetSystem,
				},
			})
			require.NoError(t, err)

			var daemonset appsv1.DaemonSet
			common.Unmarshal(t, manifest, &daemonset)

			var agentContainerFound bool
			for _, container := range daemonset.Spec.Template.Spec.Containers {
				if container.Name != "agent" {
					continue
				}
				agentContainerFound = true
				env, found := findEnvVar(container.Env, "DD_POD_NAME")
				require.True(t, found)
				require.NotNil(t, env.ValueFrom)
				require.NotNil(t, env.ValueFrom.FieldRef)
				assert.Equal(t, "metadata.name", env.ValueFrom.FieldRef.FieldPath)
			}
			require.True(t, agentContainerFound)
		})
	}
}
