package datadog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	"github.com/DataDog/helm-charts/test/common"
)

const (
	DDKubeActionsEnabled             = "DD_KUBEACTIONS_ENABLED"
	DDKubeActionsRemoteConfigEnabled = "DD_KUBEACTIONS_REMOTE_CONFIG_ENABLED"
	DDKubeActionsPAREnabled          = "DD_KUBEACTIONS_PRIVATE_ACTION_RUNNER_ENABLED"

	// First Cluster Agent version that understands the per-channel settings.
	channelTogglesDCATag = "7.86.0"
)

func renderKubernetesActions(t *testing.T, overrides map[string]string) (string, map[string]string, error) {
	manifest, err := common.RenderChart(t, common.HelmCommand{
		ReleaseName: "datadog",
		ChartPath:   "../../charts/datadog",
		ShowOnly:    []string{"templates/cluster-agent-deployment.yaml", "templates/cluster-agent-rbac.yaml"},
		Values:      []string{"../../charts/datadog/values.yaml"},
		Overrides:   overrides,
	})
	if err != nil {
		return manifest, nil, err
	}

	var deployment appsv1.Deployment
	common.Unmarshal(t, manifest, &deployment)
	envVars := map[string]string{}
	for _, envVar := range deployment.Spec.Template.Spec.Containers[0].Env {
		envVars[envVar.Name] = envVar.Value
	}
	return manifest, envVars, nil
}

func Test_KubernetesActions_Communication_DefaultsToRC(t *testing.T) {
	manifest, envVars, err := renderKubernetesActions(t, map[string]string{
		"datadog.kubernetesActions.enabled": "true",
	})
	require.NoError(t, err)

	assert.Equal(t, "true", envVars[DDKubeActionsEnabled])
	// The default Cluster Agent predates the per-channel settings: emit neither, so it
	// does not log them as unknown environment variables.
	assert.NotContains(t, envVars, DDKubeActionsRemoteConfigEnabled)
	assert.NotContains(t, envVars, DDKubeActionsPAREnabled)
	assert.NotContains(t, envVars, DDPAREnabled, "PAR should not be enabled by default")
	assert.NotContains(t, manifest, "datadog-private-action-runner", "PAR Role should not be created")
}

func Test_KubernetesActions_Communication_PARAndRC(t *testing.T) {
	manifest, envVars, err := renderKubernetesActions(t, map[string]string{
		"datadog.kubernetesActions.enabled":                                   "true",
		"datadog.kubernetesActions.communication.privateActionRunner.enabled": "true",
	})
	require.NoError(t, err)

	assert.Equal(t, "true", envVars[DDKubeActionsEnabled])
	assert.NotContains(t, envVars, DDKubeActionsRemoteConfigEnabled)
	assert.NotContains(t, envVars, DDKubeActionsPAREnabled)
	assert.Equal(t, "true", envVars[DDPAREnabled])
	assert.Equal(t, "true", envVars[DDPARSelfEnroll])
	assert.Equal(t, "datadog-private-action-runner-identity", envVars[DDPARIdentitySecret])
	assert.Equal(t, "true", envVars["DD_REMOTE_CONFIGURATION_ENABLED"])
	assert.Contains(t, manifest, "datadog-private-action-runner-identity", "PAR identity secret RBAC should be created")
}

func Test_KubernetesActions_Communication_PAROnly(t *testing.T) {
	_, envVars, err := renderKubernetesActions(t, map[string]string{
		"datadog.kubernetesActions.enabled":                                   "true",
		"datadog.kubernetesActions.communication.privateActionRunner.enabled": "true",
		"datadog.kubernetesActions.communication.remoteConfiguration.enabled": "false",
		"clusterAgent.image.tag":                                              channelTogglesDCATag,
	})
	require.NoError(t, err)

	assert.Equal(t, "true", envVars[DDKubeActionsEnabled])
	assert.Equal(t, "false", envVars[DDKubeActionsRemoteConfigEnabled])
	assert.NotContains(t, envVars, DDKubeActionsPAREnabled)
	assert.Equal(t, "true", envVars[DDPAREnabled])
	// The PAR still needs Remote Configuration for its task-signing keys.
	assert.Equal(t, "true", envVars["DD_REMOTE_CONFIGURATION_ENABLED"])
}

func Test_KubernetesActions_Communication_RCOnly(t *testing.T) {
	_, envVars, err := renderKubernetesActions(t, map[string]string{
		"datadog.kubernetesActions.enabled": "true",
		"clusterAgent.image.tag":            channelTogglesDCATag,
	})
	require.NoError(t, err)

	assert.Equal(t, "true", envVars[DDKubeActionsEnabled])
	assert.NotContains(t, envVars, DDKubeActionsRemoteConfigEnabled)
	// privateActionRunner.enabled defaults to false: keep Kubernetes actions out of a PAR enabled for other workflows.
	assert.Equal(t, "false", envVars[DDKubeActionsPAREnabled])
	assert.NotContains(t, envVars, DDPAREnabled)
}

func Test_KubernetesActions_Communication_PARDisabledWithPAREnabledForOtherWorkflows(t *testing.T) {
	_, envVars, err := renderKubernetesActions(t, map[string]string{
		"datadog.kubernetesActions.enabled":        "true",
		"clusterAgent.privateActionRunner.enabled": "true",
		"clusterAgent.image.tag":                   channelTogglesDCATag,
	})
	require.NoError(t, err)

	assert.Equal(t, "true", envVars[DDPAREnabled])
	assert.Equal(t, "false", envVars[DDKubeActionsPAREnabled])
}

func Test_KubernetesActions_Communication_RCDisabledRequiresClusterAgent786(t *testing.T) {
	_, _, err := renderKubernetesActions(t, map[string]string{
		"datadog.kubernetesActions.enabled":                                   "true",
		"datadog.kubernetesActions.communication.privateActionRunner.enabled": "true",
		"datadog.kubernetesActions.communication.remoteConfiguration.enabled": "false",
		"clusterAgent.image.tag":                                              "7.85.0",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "communication.remoteConfiguration.enabled=false requires Cluster Agent 7.86.0 or greater")
}

func Test_KubernetesActions_Communication_ChannelTogglesWithUncheckedTag(t *testing.T) {
	_, envVars, err := renderKubernetesActions(t, map[string]string{
		"datadog.kubernetesActions.enabled":                                   "true",
		"datadog.kubernetesActions.communication.privateActionRunner.enabled": "true",
		"datadog.kubernetesActions.communication.remoteConfiguration.enabled": "false",
		"clusterAgent.image.tag":                                              "my-custom-build",
		"clusterAgent.image.doNotCheckTag":                                    "true",
	})
	require.NoError(t, err)

	assert.Equal(t, "false", envVars[DDKubeActionsRemoteConfigEnabled])
	assert.NotContains(t, envVars, DDKubeActionsPAREnabled)
}

func Test_KubernetesActions_Communication_NoneEnabled(t *testing.T) {
	_, _, err := renderKubernetesActions(t, map[string]string{
		"datadog.kubernetesActions.enabled":                                   "true",
		"datadog.kubernetesActions.communication.privateActionRunner.enabled": "false",
		"datadog.kubernetesActions.communication.remoteConfiguration.enabled": "false",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one of datadog.kubernetesActions.communication.remoteConfiguration.enabled or datadog.kubernetesActions.communication.privateActionRunner.enabled must be true")
}

func Test_KubernetesActions_Communication_PARIgnoredWhenFeatureDisabled(t *testing.T) {
	_, envVars, err := renderKubernetesActions(t, map[string]string{
		"datadog.kubernetesActions.enabled":                                   "false",
		"datadog.kubernetesActions.communication.privateActionRunner.enabled": "true",
	})
	require.NoError(t, err)

	assert.NotContains(t, envVars, DDKubeActionsEnabled)
	assert.NotContains(t, envVars, DDPAREnabled)
}

func Test_KubernetesActions_Communication_PARRequiresLeaderElection(t *testing.T) {
	_, _, err := renderKubernetesActions(t, map[string]string{
		"datadog.kubernetesActions.enabled":                                   "true",
		"datadog.kubernetesActions.communication.privateActionRunner.enabled": "true",
		"datadog.leaderElection":                                              "false",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "selfEnroll requires leader election")
}

func Test_KubernetesActions_RBAC_CoversAllActions(t *testing.T) {
	manifest, err := common.RenderChart(t, common.HelmCommand{
		ReleaseName: "datadog",
		ChartPath:   "../../charts/datadog",
		ShowOnly:    []string{"templates/cluster-agent-kubernetes-actions-rbac.yaml"},
		Values:      []string{"../../charts/datadog/values.yaml"},
		Overrides: map[string]string{
			"datadog.kubernetesActions.enabled": "true",
		},
	})
	require.NoError(t, err)

	var clusterRole rbacv1.ClusterRole
	common.Unmarshal(t, manifest, &clusterRole)

	verbsByResource := map[string][]string{}
	for _, rule := range clusterRole.Rules {
		for _, res := range rule.Resources {
			verbsByResource[res] = append(verbsByResource[res], rule.Verbs...)
		}
	}
	// One entry per Kubernetes API call made by the kubeactions executors exposed in the
	// Datadog UI (comp/kubeactions/kubeactions/impl in datadog-agent), shared by the RC and PAR paths.
	assert.Subset(t, verbsByResource["pods"], []string{"get", "delete"})
	assert.Subset(t, verbsByResource["deployments"], []string{"get", "patch", "update"})
	// The patch DaemonSet / StatefulSet actions are not exposed in the Datadog UI:
	// do not grant cluster-wide patch on those workloads.
	assert.NotContains(t, verbsByResource, "daemonsets")
	assert.NotContains(t, verbsByResource, "statefulsets")
	assert.Subset(t, verbsByResource["replicasets"], []string{"get", "list"})
}
