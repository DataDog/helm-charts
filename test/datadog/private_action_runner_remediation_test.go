package datadog

import (
	"strings"
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/yaml"
)

const ddPARKubernetesRemediationsEnabled = "DD_PRIVATE_ACTION_RUNNER_KUBERNETES_REMEDIATIONS_ENABLED"

func Test_PrivateActionRunner_KubernetesRemediations(t *testing.T) {
	for _, tt := range []struct {
		name         string
		overrides    map[string]string
		wantPAR      bool
		wantEnv      string
		wantActions  bool
		wantRBAC     bool
		wantHelm     bool
		wantChannels map[string]string
	}{
		{name: "disabled by default"},
		{name: "PAR alone", overrides: map[string]string{"clusterAgent.privateActionRunner.enabled": "true"}, wantPAR: true, wantEnv: "false"},
		{name: "remediation requires PAR", overrides: map[string]string{"clusterAgent.privateActionRunner.k8sRemediationEnabled": "true"}},
		{name: "node PAR alone is insufficient", overrides: map[string]string{"datadog.privateActionRunner.enabled": "true", "datadog.privateActionRunner.selfEnroll": "true", "clusterAgent.privateActionRunner.k8sRemediationEnabled": "true"}},
		{name: "enabled", overrides: map[string]string{"clusterAgent.privateActionRunner.enabled": "true", "clusterAgent.privateActionRunner.k8sRemediationEnabled": "true"}, wantPAR: true, wantEnv: "true", wantActions: true, wantRBAC: true},
		{name: "explicit false", overrides: map[string]string{"clusterAgent.privateActionRunner.enabled": "true", "clusterAgent.privateActionRunner.k8sRemediationEnabled": "false"}, wantPAR: true, wantEnv: "false"},
		{name: "explicit Kubernetes Actions opt out", overrides: map[string]string{"clusterAgent.privateActionRunner.enabled": "true", "clusterAgent.privateActionRunner.k8sRemediationEnabled": "true", "datadog.kubernetesActions.enabled": "false"}, wantPAR: true, wantEnv: "true", wantRBAC: true},
		{name: "Kubernetes Actions enables PAR", overrides: map[string]string{"datadog.kubernetesActions.enabled": "true", "clusterAgent.privateActionRunner.k8sRemediationEnabled": "true"}, wantPAR: true, wantEnv: "true", wantActions: true, wantRBAC: true},
		{name: "RC only respects PAR channel opt out", overrides: map[string]string{"clusterAgent.privateActionRunner.enabled": "true", "clusterAgent.privateActionRunner.k8sRemediationEnabled": "true", "datadog.kubernetesActions.communication.privateActionRunner.enabled": "false"}, wantPAR: true, wantEnv: "true", wantActions: true, wantRBAC: true, wantChannels: map[string]string{DDKubeActionsPAREnabled: "false"}},
		{name: "PAR only respects RC channel opt out", overrides: map[string]string{"clusterAgent.privateActionRunner.enabled": "true", "clusterAgent.privateActionRunner.k8sRemediationEnabled": "true", "datadog.kubernetesActions.communication.remoteConfiguration.enabled": "false"}, wantPAR: true, wantEnv: "true", wantActions: true, wantRBAC: true, wantChannels: map[string]string{DDKubeActionsRemoteConfigEnabled: "false"}},
		{name: "Helm Actions uses resolved enablement", overrides: map[string]string{"clusterAgent.privateActionRunner.enabled": "true", "clusterAgent.privateActionRunner.k8sRemediationEnabled": "true", "datadog.kubernetesActions.helmActions.enabled": "true"}, wantPAR: true, wantEnv: "true", wantActions: true, wantRBAC: true, wantHelm: true},
		{name: "Kubernetes Actions version gate retained", overrides: map[string]string{"clusterAgent.privateActionRunner.enabled": "true", "clusterAgent.privateActionRunner.k8sRemediationEnabled": "true", "clusterAgent.image.tag": "7.83.0"}, wantPAR: true, wantEnv: "true", wantRBAC: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			overrides := map[string]string{"clusterAgent.image.tag": channelTogglesDCATag}
			for key, value := range tt.overrides {
				overrides[key] = value
			}
			manifest, err := common.RenderChart(t, common.HelmCommand{
				ReleaseName: "datadog", ChartPath: "../../charts/datadog",
				Values: []string{"../../charts/datadog/values.yaml"}, Overrides: overrides,
			})
			require.NoError(t, err)

			var deployment appsv1.Deployment
			var remediationRole *rbacv1.ClusterRole
			var remediationBinding *rbacv1.ClusterRoleBinding
			var actionsRole, helmRole bool
			for _, doc := range strings.Split(manifest, "\n---") {
				var meta struct {
					Kind     string
					Metadata struct{ Name string }
				}
				require.NoError(t, yaml.Unmarshal([]byte(doc), &meta))
				switch {
				case meta.Kind == "Deployment" && meta.Metadata.Name == "datadog-cluster-agent":
					require.NoError(t, yaml.Unmarshal([]byte(doc), &deployment))
				case meta.Kind == "ClusterRole" && meta.Metadata.Name == "datadog-private-action-runner":
					remediationRole = &rbacv1.ClusterRole{}
					require.NoError(t, yaml.Unmarshal([]byte(doc), remediationRole))
				case meta.Kind == "ClusterRoleBinding" && meta.Metadata.Name == "datadog-private-action-runner":
					remediationBinding = &rbacv1.ClusterRoleBinding{}
					require.NoError(t, yaml.Unmarshal([]byte(doc), remediationBinding))
				case meta.Kind == "ClusterRole" && meta.Metadata.Name == "datadog-kubernetes-actions":
					actionsRole = true
				case meta.Kind == "ClusterRole" && meta.Metadata.Name == "datadog-helm-action-caller":
					helmRole = true
				case meta.Kind == "DaemonSet":
					assert.NotContains(t, doc, ddPARKubernetesRemediationsEnabled, "remediations must stay on the Cluster Agent")
				}
			}
			require.NotEmpty(t, deployment.Spec.Template.Spec.Containers)
			env := map[string]string{}
			for _, item := range deployment.Spec.Template.Spec.Containers[0].Env {
				env[item.Name] = item.Value
			}
			assert.Equal(t, tt.wantPAR, env[DDPAREnabled] == "true")
			assert.Equal(t, tt.wantEnv, env[ddPARKubernetesRemediationsEnabled])
			assert.Equal(t, tt.wantActions, env[DDKubeActionsEnabled] == "true")
			for _, name := range []string{DDKubeActionsRemoteConfigEnabled, DDKubeActionsPAREnabled} {
				assert.Equal(t, tt.wantChannels[name], env[name], name)
			}
			assert.Equal(t, tt.wantActions, actionsRole)
			assert.Equal(t, tt.wantHelm, helmRole)
			require.Equal(t, tt.wantRBAC, remediationRole != nil)
			require.Equal(t, tt.wantRBAC, remediationBinding != nil)
			if tt.wantRBAC {
				assert.Equal(t, remediationRole.Name, remediationBinding.RoleRef.Name)
				assert.Equal(t, deployment.Spec.Template.Spec.ServiceAccountName, remediationBinding.Subjects[0].Name)
				permissions := map[string][]string{}
				for _, rule := range remediationRole.Rules {
					for _, group := range rule.APIGroups {
						for _, resource := range rule.Resources {
							key := group + "/" + resource
							permissions[key] = append(permissions[key], rule.Verbs...)
						}
					}
				}
				// Exact permissions for the built-in remediation bundle, including rollback's update.
				assert.Equal(t, map[string][]string{
					"apps/deployments":                      {"get", "list", "watch", "patch", "update"},
					"apps/daemonsets":                       {"get", "list", "watch"},
					"apps/statefulsets":                     {"get", "list", "watch", "patch"},
					"apps/replicasets":                      {"get", "list", "watch"},
					"/pods":                                 {"get", "list", "watch", "patch", "delete"},
					"/events":                               {"get", "list", "watch", "create", "patch"},
					"/configmaps":                           {"get", "list", "watch", "create", "update", "patch"},
					"autoscaling/horizontalpodautoscalers":  {"patch"},
					"datadoghq.com/watermarkpodautoscalers": {"get", "patch"},
					"batch/jobs":                            {"delete"},
					"batch/cronjobs":                        {"patch"},
				}, permissions)
			}
		})
	}
}

func Test_PrivateActionRunner_KubernetesRemediations_Validation(t *testing.T) {
	for _, tt := range []struct {
		name      string
		overrides map[string]string
		wantError string
	}{
		{name: "no communication channels", overrides: map[string]string{"datadog.kubernetesActions.communication.remoteConfiguration.enabled": "false", "datadog.kubernetesActions.communication.privateActionRunner.enabled": "false"}, wantError: "at least one of"},
		{name: "self enrollment needs leader election", overrides: map[string]string{"datadog.leaderElection": "false"}, wantError: "selfEnroll requires leader election"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			overrides := map[string]string{"clusterAgent.privateActionRunner.enabled": "true", "clusterAgent.privateActionRunner.k8sRemediationEnabled": "true", "clusterAgent.image.tag": channelTogglesDCATag}
			for key, value := range tt.overrides {
				overrides[key] = value
			}
			_, _, err := renderKubernetesActions(t, overrides)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantError)
		})
	}
}
