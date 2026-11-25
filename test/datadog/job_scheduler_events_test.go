package datadog

import (
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

const (
	kueueEventsEnvVar = "DD_CLUSTER_AGENT_KUEUE_EVENTS_ENABLED"
	kaiEventsEnvVar   = "DD_CLUSTER_AGENT_KAI_SCHEDULER_EVENTS_ENABLED"
)

func Test_jobSchedulerEvents(t *testing.T) {
	tests := []struct {
		name          string
		overrides     map[string]string
		wantKueue     bool
		wantKAI       bool
		wantNodeEvent bool // collect_workload_events of the Kueue check
	}{
		{
			name:          "disabled by default",
			wantNodeEvent: true,
		},
		{
			name: "kueue events",
			overrides: map[string]string{
				"datadog.kueue.events.enabled": "true",
			},
			wantKueue: true,
		},
		{
			name: "kai events",
			overrides: map[string]string{
				"datadog.kaiScheduler.events.enabled": "true",
			},
			wantKAI:       true,
			wantNodeEvent: true,
		},
		{
			name: "both",
			overrides: map[string]string{
				"datadog.kueue.events.enabled":        "true",
				"datadog.kaiScheduler.events.enabled": "true",
			},
			wantKueue: true,
			wantKAI:   true,
		},
		{
			name: "events flag without the scheduler support does nothing",
			overrides: map[string]string{
				"datadog.kueue.enabled":               "false",
				"datadog.kaiScheduler.enabled":        "false",
				"datadog.kueue.events.enabled":        "true",
				"datadog.kaiScheduler.events.enabled": "true",
			},
		},
		{
			name: "cluster agent below 7.87.0 keeps the node checks sending events",
			overrides: map[string]string{
				"clusterAgent.image.tag":              "7.86.0",
				"datadog.kueue.events.enabled":        "true",
				"datadog.kaiScheduler.events.enabled": "true",
			},
			wantNodeEvent: true,
		},
		{
			name: "latest tag resolves below 7.87.0",
			overrides: map[string]string{
				"clusterAgent.image.tag":              "latest",
				"datadog.kueue.events.enabled":        "true",
				"datadog.kaiScheduler.events.enabled": "true",
			},
			wantNodeEvent: true,
		},
		{
			name: "doNotCheckTag skips the version gate",
			overrides: map[string]string{
				"clusterAgent.image.tag":              "custom",
				"clusterAgent.image.doNotCheckTag":    "true",
				"datadog.kueue.events.enabled":        "true",
				"datadog.kaiScheduler.events.enabled": "true",
			},
			wantKueue: true,
			wantKAI:   true,
		},
		{
			name: "prerelease versions are accepted",
			overrides: map[string]string{
				"clusterAgent.image.tag":              "7.87.0-rc.1",
				"datadog.kueue.events.enabled":        "true",
				"datadog.kaiScheduler.events.enabled": "true",
			},
			wantKueue: true,
			wantKAI:   true,
		},
		{
			name: "leader election off keeps the node checks sending events",
			overrides: map[string]string{
				"datadog.leaderElection":              "false",
				"datadog.kueue.events.enabled":        "true",
				"datadog.kaiScheduler.events.enabled": "true",
			},
			wantNodeEvent: true,
		},
		{
			name: "kueue events with the node events already disabled",
			overrides: map[string]string{
				"datadog.kueue.collectWorkloadEvents": "false",
				"datadog.kueue.events.enabled":        "true",
			},
			wantKueue: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			overrides := map[string]string{
				"datadog.apiKeyExistingSecret":  "datadog-secret",
				"datadog.appKeyExistingSecret":  "datadog-secret",
				"datadog.clusterChecks.enabled": "true",
				"datadog.kueue.enabled":         "true",
				"datadog.kaiScheduler.enabled":  "true",
				"clusterAgent.image.tag":        "7.87.0",
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
			assert.Equal(t, tt.wantKueue, hasEnvVar(dca.Env, kueueEventsEnvVar))
			assert.Equal(t, tt.wantKAI, hasEnvVar(dca.Env, kaiEventsEnvVar))
			if tt.wantKueue || tt.wantKAI {
				for _, env := range dca.Env {
					if env.Name == "DD_LEADER_ELECTION" {
						assert.Equal(t, "true", env.Value)
					}
				}
			}

			// The events need the apiserver access the metadata collection already grants.
			var dcaRole rbacv1.ClusterRole
			require.True(t, decodeK8sResource(t, manifest, "ClusterRole", "datadog-cluster-agent", &dcaRole))
			if tt.wantKueue {
				assert.True(t, clusterRoleHasRule(dcaRole.Rules, "kueue.x-k8s.io", "workloads", "list"))
				assert.True(t, clusterRoleHasRule(dcaRole.Rules, "kueue.x-k8s.io", "workloads", "watch"))
			}
			if tt.wantKAI {
				assert.True(t, clusterRoleHasRule(dcaRole.Rules, "scheduling.run.ai", "podgroups", "list"))
				assert.True(t, clusterRoleHasRule(dcaRole.Rules, "scheduling.run.ai", "podgroups", "watch"))
				assert.True(t, clusterRoleHasRule(dcaRole.Rules, "", "events", "list"))
			}

			var confd corev1.ConfigMap
			require.True(t, decodeK8sResource(t, manifest, "ConfigMap", "datadog-cluster-agent-confd", &confd))
			if tt.overrides["datadog.kueue.enabled"] == "false" {
				assert.NotContains(t, confd.Data, "kueue.yaml")
				return
			}
			var conf struct {
				Instances []struct {
					CollectWorkloadEvents bool `yaml:"collect_workload_events"`
				} `yaml:"instances"`
			}
			require.NoError(t, yaml.Unmarshal([]byte(confd.Data["kueue.yaml"]), &conf))
			require.Len(t, conf.Instances, 1)
			wantNodeEvent := tt.wantNodeEvent
			if tt.overrides["datadog.kueue.collectWorkloadEvents"] == "false" {
				wantNodeEvent = false
			}
			assert.Equal(t, wantNodeEvent, conf.Instances[0].CollectWorkloadEvents)
		})
	}
}
