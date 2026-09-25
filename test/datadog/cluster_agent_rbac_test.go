package datadog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	rbacv1 "k8s.io/api/rbac/v1"

	"github.com/DataDog/helm-charts/test/common"
)

func Test_ClusterAgent_RBAC_AllowsReadingPodLogs(t *testing.T) {
	manifest, err := common.RenderChart(t, common.HelmCommand{
		ReleaseName: "datadog",
		ChartPath:   "../../charts/datadog",
		ShowOnly:    []string{"templates/cluster-agent-rbac.yaml"},
		Values:      []string{"../../charts/datadog/values.yaml"},
	})
	require.NoError(t, err)

	var clusterRole rbacv1.ClusterRole
	require.True(t, decodeResourceByKindAndName(manifest, "ClusterRole", "datadog-cluster-agent", &clusterRole))

	var podLogsRule *rbacv1.PolicyRule
	for i := range clusterRole.Rules {
		rule := &clusterRole.Rules[i]
		if len(rule.Resources) == 1 && rule.Resources[0] == "pods/log" {
			podLogsRule = rule
			break
		}
	}

	require.NotNil(t, podLogsRule, "Cluster Agent ClusterRole must include the pods/log subresource")
	assert.Empty(t, podLogsRule.APIGroups)
	assert.Equal(t, []string{"get"}, podLogsRule.Verbs)
}
