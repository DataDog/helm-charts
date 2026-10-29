package datadog

import (
	"encoding/json"
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

const kueueEnabledEnvVar = "DD_CLUSTER_AGENT_KUEUE_ENABLED"

func Test_kueue(t *testing.T) {
	tests := []struct {
		name              string
		overrides         map[string]string
		wantMetadata      bool
		wantCheck         bool
		wantNodeAgentRBAC bool
		wantService       string
		wantNamespace     string
		wantEvents        bool
	}{
		{
			name: "disabled by default",
		},
		{
			name: "enabled with cluster checks",
			overrides: map[string]string{
				"datadog.kueue.enabled": "true",
			},
			wantMetadata:      true,
			wantCheck:         true,
			wantNodeAgentRBAC: true,
			wantService:       "kueue-controller-manager-metrics-service",
			wantNamespace:     "kueue-system",
			wantEvents:        true,
		},
		{
			name: "metrics service override and workload events disabled",
			overrides: map[string]string{
				"datadog.kueue.enabled":                  "true",
				"datadog.kueue.collectWorkloadEvents":    "false",
				"datadog.kueue.metricsService.name":      "my-kueue-metrics",
				"datadog.kueue.metricsService.namespace": "my-kueue",
			},
			wantMetadata:  true,
			wantCheck:     true,
			wantService:   "my-kueue-metrics",
			wantNamespace: "my-kueue",
		},
		{
			name: "cluster checks disabled only enables metadata collection",
			overrides: map[string]string{
				"datadog.kueue.enabled":         "true",
				"datadog.clusterChecks.enabled": "false",
			},
			wantMetadata: true,
		},
		{
			name: "node agent below minimum version only enables metadata collection",
			overrides: map[string]string{
				"datadog.kueue.enabled": "true",
				"agents.image.tag":      "7.81.0",
			},
			wantMetadata: true,
		},
		{
			name: "cluster agent below minimum version disables the feature",
			overrides: map[string]string{
				"datadog.kueue.enabled":  "true",
				"clusterAgent.image.tag": "7.81.0",
			},
		},
		{
			name: "prerelease versions are accepted",
			overrides: map[string]string{
				"datadog.kueue.enabled":  "true",
				"agents.image.tag":       "7.82.0-rc.1",
				"clusterAgent.image.tag": "7.82.0-rc.1",
			},
			wantMetadata:      true,
			wantCheck:         true,
			wantNodeAgentRBAC: true,
			wantService:       "kueue-controller-manager-metrics-service",
			wantNamespace:     "kueue-system",
			wantEvents:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			overrides := map[string]string{
				"datadog.apiKeyExistingSecret":  "datadog-secret",
				"datadog.appKeyExistingSecret":  "datadog-secret",
				"datadog.clusterChecks.enabled": "true",
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
			assert.Equal(t, tt.wantMetadata, hasEnvVar(dca.Env, kueueEnabledEnvVar))
			assert.Equal(t, tt.wantCheck, confdVolumeHasPath(deployment.Spec.Template.Spec.Volumes, "kueue.d/kueue.yaml"))

			var dcaRole rbacv1.ClusterRole
			require.True(t, decodeK8sResource(t, manifest, "ClusterRole", "datadog-cluster-agent", &dcaRole))
			assert.Equal(t, tt.wantMetadata, clusterRoleHasRule(dcaRole.Rules, "kueue.x-k8s.io", "clusterqueues", "list"))
			assert.Equal(t, tt.wantMetadata, clusterRoleHasRule(dcaRole.Rules, "kueue.x-k8s.io", "workloads", "watch"))

			var agentRole rbacv1.ClusterRole
			require.True(t, decodeK8sResource(t, manifest, "ClusterRole", "datadog", &agentRole))
			assert.Equal(t, tt.wantNodeAgentRBAC, clusterRoleHasRule(agentRole.Rules, "kueue.x-k8s.io", "workloads", "list"))

			var confd corev1.ConfigMap
			require.True(t, decodeK8sResource(t, manifest, "ConfigMap", "datadog-cluster-agent-confd", &confd))
			checkConfig, found := confd.Data["kueue.yaml"]
			require.Equal(t, tt.wantCheck, found)
			if !tt.wantCheck {
				return
			}

			var conf struct {
				AdvancedADIdentifiers []struct {
					KubeEndpoints struct {
						Name      string `yaml:"name"`
						Namespace string `yaml:"namespace"`
					} `yaml:"kube_endpoints"`
				} `yaml:"advanced_ad_identifiers"`
				ClusterCheck bool `yaml:"cluster_check"`
				Instances    []struct {
					OpenMetricsEndpoint   string `yaml:"openmetrics_endpoint"`
					CollectWorkloadEvents bool   `yaml:"collect_workload_events"`
				} `yaml:"instances"`
			}
			require.NoError(t, yaml.Unmarshal([]byte(checkConfig), &conf))
			require.Len(t, conf.AdvancedADIdentifiers, 1)
			assert.Equal(t, tt.wantService, conf.AdvancedADIdentifiers[0].KubeEndpoints.Name)
			assert.Equal(t, tt.wantNamespace, conf.AdvancedADIdentifiers[0].KubeEndpoints.Namespace)
			assert.True(t, conf.ClusterCheck)
			require.Len(t, conf.Instances, 1)
			assert.Equal(t, "https://%%host%%:%%port%%/metrics", conf.Instances[0].OpenMetricsEndpoint)
			assert.Equal(t, tt.wantEvents, conf.Instances[0].CollectWorkloadEvents)
		})
	}
}

func confdVolumeHasPath(volumes []corev1.Volume, path string) bool {
	for _, volume := range volumes {
		if volume.Name != "confd" || volume.ConfigMap == nil {
			continue
		}
		for _, item := range volume.ConfigMap.Items {
			if item.Path == path {
				return true
			}
		}
	}
	return false
}

func clusterRoleHasRule(rules []rbacv1.PolicyRule, apiGroup, resource, verb string) bool {
	for _, rule := range rules {
		if common.Contains(apiGroup, rule.APIGroups) && common.Contains(resource, rule.Resources) && common.Contains(verb, rule.Verbs) {
			return true
		}
	}
	return false
}

// decodeK8sResource honors JSON field tags, unlike decodeResourceByKindAndName,
// which drops camelCase fields such as `apiGroups` or `configMap`.
func decodeK8sResource(t *testing.T, manifest, kind, name string, dest interface{}) bool {
	var raw map[string]interface{}
	if !decodeResourceByKindAndName(manifest, kind, name, &raw) {
		return false
	}
	data, err := json.Marshal(raw)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(data, dest))
	return true
}
