package datadog

import (
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

func Test_kaiScheduler(t *testing.T) {
	tests := []struct {
		name         string
		overrides    map[string]string
		wantMetadata bool
		wantGPU      bool
	}{
		{
			name: "disabled by default",
		},
		{
			name: "enabled without GPU monitoring",
			overrides: map[string]string{
				"datadog.kaiScheduler.enabled": "true",
			},
			wantMetadata: true,
		},
		{
			name: "enabled with GPU monitoring",
			overrides: map[string]string{
				"datadog.kaiScheduler.enabled":  "true",
				"datadog.gpuMonitoring.enabled": "true",
			},
			wantMetadata: true,
			wantGPU:      true,
		},
		{
			name: "cluster agent below minimum version disables metadata collection",
			overrides: map[string]string{
				"datadog.kaiScheduler.enabled": "true",
				"clusterAgent.image.tag":       "7.85.0",
			},
		},
		{
			name: "prerelease versions are accepted",
			overrides: map[string]string{
				"datadog.kaiScheduler.enabled": "true",
				"clusterAgent.image.tag":       "7.86.0-rc.1",
			},
			wantMetadata: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			overrides := map[string]string{
				"datadog.apiKeyExistingSecret": "datadog-secret",
				"datadog.appKeyExistingSecret": "datadog-secret",
				"clusterAgent.image.tag":       "7.86.0",
			}
			for k, v := range tt.overrides {
				overrides[k] = v
			}

			manifest, err := common.RenderChart(t, common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides:   overrides,
			})
			require.NoError(t, err, "couldn't render chart")

			var deployment appsv1.Deployment
			require.True(t, decodeK8sResource(t, manifest, "Deployment", "datadog-cluster-agent", &deployment))
			dca, found := getContainer(t, deployment.Spec.Template.Spec.Containers, "cluster-agent")
			require.True(t, found)
			assert.Equal(t, tt.wantMetadata, hasEnvVar(dca.Env, "DD_CLUSTER_AGENT_KAI_SCHEDULER_ENABLED"))

			var dcaRole rbacv1.ClusterRole
			require.True(t, decodeK8sResource(t, manifest, "ClusterRole", "datadog-cluster-agent", &dcaRole))
			assert.Equal(t, tt.wantMetadata, clusterRoleHasRule(dcaRole.Rules, "scheduling.run.ai", "podgroups", "list"))
			assert.Equal(t, tt.wantMetadata, clusterRoleHasRule(dcaRole.Rules, "scheduling.run.ai", "podgroups", "watch"))

			var daemonset appsv1.DaemonSet
			require.True(t, decodeK8sResource(t, manifest, "DaemonSet", "datadog", &daemonset))
			agent, found := getContainer(t, daemonset.Spec.Template.Spec.Containers, "agent")
			require.True(t, found)
			assert.Equal(t, tt.wantGPU, hasEnvVar(agent.Env, "DD_GPU_KAI_SCHEDULER_ENABLED"))
		})
	}
}
