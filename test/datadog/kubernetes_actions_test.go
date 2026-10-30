package datadog

import (
	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestKubernetesActionsDefaults(t *testing.T) {
	for _, tc := range []struct {
		name      string
		overrides map[string]string
		enabled   bool
	}{
		{name: "default", enabled: true},
		{name: "explicit opt out", overrides: map[string]string{"datadog.kubernetesActions.enabled": "false"}},
		{name: "old image", overrides: map[string]string{"clusterAgent.image.tag": "7.78.0"}},
		{name: "remote configuration opt out", overrides: map[string]string{"remoteConfiguration.enabled": "false"}},
		{name: "GDC", overrides: map[string]string{"providers.gke.gdc": "true"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			manifest, err := common.RenderChart(t, common.HelmCommand{
				ReleaseName: "datadog", ChartPath: "../../charts/datadog",
				Values: []string{"../../charts/datadog/values.yaml"}, Overrides: tc.overrides,
			})
			require.NoError(t, err)
			assert.Equal(t, tc.enabled, strings.Contains(manifest, "name: DD_KUBEACTIONS_ENABLED"))
			if tc.enabled {
				assert.Contains(t, manifest, "name: datadog-kubernetes-actions")
			} else {
				assert.NotContains(t, manifest, "name: datadog-kubernetes-actions")
			}
			assert.NotContains(t, manifest, "name: DD_PRIVATE_ACTION_RUNNER_ENABLED")
		})
	}
}
