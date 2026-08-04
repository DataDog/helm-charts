package datadog

import (
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

const instrumentationCRDControllerEnabledEnvVar = "DD_INSTRUMENTATION_CRD_CONTROLLER_ENABLED"

func Test_instrumentationCRDControllerVersionGate(t *testing.T) {
	tests := []struct {
		name      string
		overrides map[string]string
		enabled   bool
	}{
		{
			name: "disabled with supported versions",
			overrides: map[string]string{
				"datadog.instrumentationCrd.enabled": "false",
				"agents.image.tag":                   "7.82.0",
				"clusterAgent.image.tag":             "7.82.0",
			},
			enabled: false,
		},
		{
			name: "enabled at minimum versions",
			overrides: map[string]string{
				"datadog.instrumentationCrd.enabled": "true",
				"agents.image.tag":                   "7.82.0",
				"clusterAgent.image.tag":             "7.82.0",
			},
			enabled: true,
		},
		{
			name: "enabled with prerelease versions",
			overrides: map[string]string{
				"datadog.instrumentationCrd.enabled": "true",
				"agents.image.tag":                   "7.82.0-rc.1",
				"clusterAgent.image.tag":             "7.82.0-rc.1",
			},
			enabled: true,
		},
		{
			name: "disabled below minimum node agent version",
			overrides: map[string]string{
				"datadog.instrumentationCrd.enabled": "true",
				"agents.image.tag":                   "7.81.9",
				"clusterAgent.image.tag":             "7.82.0",
			},
			enabled: false,
		},
		{
			name: "disabled below minimum cluster agent version",
			overrides: map[string]string{
				"datadog.instrumentationCrd.enabled": "true",
				"agents.image.tag":                   "7.82.0",
				"clusterAgent.image.tag":             "7.81.9",
			},
			enabled: false,
		},
		{
			name: "floating node agent tag follows get-agent-version policy",
			overrides: map[string]string{
				"datadog.instrumentationCrd.enabled": "true",
				"agents.image.tag":                   "latest-jmx",
				"clusterAgent.image.tag":             "7.82.0",
			},
			enabled: false,
		},
		{
			name: "floating cluster agent tag follows get-cluster-agent-version policy",
			overrides: map[string]string{
				"datadog.instrumentationCrd.enabled": "true",
				"agents.image.tag":                   "7.82.0",
				"clusterAgent.image.tag":             "latest",
			},
			enabled: false,
		},
		{
			name: "enabled when both tag checks are skipped",
			overrides: map[string]string{
				"datadog.instrumentationCrd.enabled": "true",
				"agents.image.tag":                   "custom-agent-tag",
				"agents.image.doNotCheckTag":         "true",
				"clusterAgent.image.tag":             "custom-cluster-agent-tag",
				"clusterAgent.image.doNotCheckTag":   "true",
			},
			enabled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest, err := common.RenderChart(t, common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides:   instrumentationCRDOverrides(tt.overrides),
			})
			require.NoError(t, err, "couldn't render chart")

			assertInstrumentationCRDControllerState(t, manifest, tt.enabled)
		})
	}
}

func instrumentationCRDOverrides(overrides map[string]string) map[string]string {
	merged := map[string]string{
		"datadog.apiKeyExistingSecret": "datadog-secret",
		"datadog.appKeyExistingSecret": "datadog-secret",
	}
	for key, value := range overrides {
		merged[key] = value
	}
	return merged
}

func assertInstrumentationCRDControllerState(t *testing.T, manifest string, enabled bool) {
	t.Helper()

	var daemonset appsv1.DaemonSet
	require.True(t, decodeResourceByKindAndName(manifest, "DaemonSet", "datadog", &daemonset))
	agentContainer, found := getContainer(t, daemonset.Spec.Template.Spec.Containers, "agent")
	require.True(t, found)
	assert.Equal(t, enabled, hasEnvVar(agentContainer.Env, instrumentationCRDControllerEnabledEnvVar))

	var deployment appsv1.Deployment
	require.True(t, decodeResourceByKindAndName(manifest, "Deployment", "datadog-cluster-agent", &deployment))
	clusterAgentContainer, found := getContainer(t, deployment.Spec.Template.Spec.Containers, "cluster-agent")
	require.True(t, found)
	assert.Equal(t, enabled, hasEnvVar(clusterAgentContainer.Env, instrumentationCRDControllerEnabledEnvVar))

	var clusterRole rbacv1.ClusterRole
	require.True(t, decodeResourceByKindAndName(manifest, "ClusterRole", "datadog-cluster-agent", &clusterRole))
	assert.Equal(t, enabled, clusterRoleHasResource(clusterRole.Rules, "datadoginstrumentations"))
	assert.Equal(t, enabled, clusterRoleHasResource(clusterRole.Rules, "datadoginstrumentations/status"))
}

func hasEnvVar(envVars []corev1.EnvVar, name string) bool {
	for _, envVar := range envVars {
		if envVar.Name == name {
			return true
		}
	}
	return false
}

func clusterRoleHasResource(rules []rbacv1.PolicyRule, resource string) bool {
	for _, rule := range rules {
		for _, candidate := range rule.Resources {
			if candidate == resource {
				return true
			}
		}
	}
	return false
}
