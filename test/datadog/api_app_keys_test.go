package datadog

import (
	"maps"
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	DDApiKey = "DD_API_KEY"
	DDAppKey = "DD_APP_KEY"
)

// Test_clusterAgentKeys verifies that DD_API_KEY and DD_APP_KEY are correctly injected into the
// cluster-agent container. It checks that both inline values and existing secret references produce
// the expected secretKeyRef name and key, and that DD_APP_KEY is absent when no app key is configured.
func Test_clusterAgentKeys(t *testing.T) {
	baseCmd := common.HelmCommand{
		ReleaseName: "datadog",
		ChartPath:   "../../charts/datadog",
		ShowOnly:    []string{"templates/cluster-agent-deployment.yaml"},
		Values:      []string{"../../charts/datadog/values.yaml"},
	}

	tests := []struct {
		name      string
		overrides map[string]string
		assert    func(t *testing.T, container corev1.Container)
	}{
		{
			name: "DD_APP_KEY present when datadog.appKey is set",
			overrides: map[string]string{
				"datadog.apiKey": "test-api-key",
				"datadog.appKey": "test-app-key",
			},
			assert: func(t *testing.T, c corev1.Container) {
				env, ok := findEnvVar(c.Env, DDAppKey)
				require.True(t, ok, "expected DD_APP_KEY to be present in cluster-agent container")
				require.NotNil(t, env.ValueFrom.SecretKeyRef)
				assert.Equal(t, "datadog-appkey", env.ValueFrom.SecretKeyRef.Name)
				assert.Equal(t, "app-key", env.ValueFrom.SecretKeyRef.Key)
			},
		},
		{
			name: "DD_APP_KEY present when datadog.appKeyExistingSecret is set",
			overrides: map[string]string{
				"datadog.apiKey":               "test-api-key",
				"datadog.appKeyExistingSecret": "my-app-secret",
			},
			assert: func(t *testing.T, c corev1.Container) {
				env, ok := findEnvVar(c.Env, DDAppKey)
				require.True(t, ok, "expected DD_APP_KEY to be present in cluster-agent container")
				require.NotNil(t, env.ValueFrom.SecretKeyRef)
				assert.Equal(t, "my-app-secret", env.ValueFrom.SecretKeyRef.Name)
				assert.Equal(t, "app-key", env.ValueFrom.SecretKeyRef.Key)
			},
		},
		{
			name: "DD_APP_KEY absent when neither datadog.appKey nor datadog.appKeyExistingSecret is set",
			overrides: map[string]string{
				"datadog.apiKey": "test-api-key",
			},
			assert: func(t *testing.T, c corev1.Container) {
				_, ok := findEnvVar(c.Env, DDAppKey)
				assert.False(t, ok, "expected DD_APP_KEY to be absent when no appKey is configured")
			},
		},
		{
			name: "DD_API_KEY present with default secret name when datadog.apiKey is set",
			overrides: map[string]string{
				"datadog.apiKey": "test-api-key",
			},
			assert: func(t *testing.T, c corev1.Container) {
				env, ok := findEnvVar(c.Env, DDApiKey)
				require.True(t, ok, "expected DD_API_KEY to be present in cluster-agent container")
				require.NotNil(t, env.ValueFrom.SecretKeyRef)
				assert.Equal(t, "datadog", env.ValueFrom.SecretKeyRef.Name)
				assert.Equal(t, "api-key", env.ValueFrom.SecretKeyRef.Key)
			},
		},
		{
			name: "DD_API_KEY uses existing secret when datadog.apiKeyExistingSecret is set",
			overrides: map[string]string{
				"datadog.apiKeyExistingSecret": "custom-api-secret",
			},
			assert: func(t *testing.T, c corev1.Container) {
				env, ok := findEnvVar(c.Env, DDApiKey)
				require.True(t, ok, "expected DD_API_KEY to be present in cluster-agent container")
				require.NotNil(t, env.ValueFrom.SecretKeyRef)
				assert.Equal(t, "custom-api-secret", env.ValueFrom.SecretKeyRef.Name)
				assert.Equal(t, "api-key", env.ValueFrom.SecretKeyRef.Key)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := baseCmd
			cmd.Overrides = tt.overrides
			manifest, err := common.RenderChart(t, cmd)
			require.NoError(t, err, "failed to render cluster-agent-deployment.yaml")

			var deployment appsv1.Deployment
			common.Unmarshal(t, manifest, &deployment)
			require.NotEmpty(t, deployment.Spec.Template.Spec.Containers, "expected at least one container in cluster-agent deployment")

			tt.assert(t, deployment.Spec.Template.Spec.Containers[0])
		})
	}
}

// Test_daemonsetApiKey verifies that the agent daemonset only renders when an API key is configured,
// and that DD_API_KEY in the agent container references the correct secret name and key field
// for both inline values and existing secret references.
func Test_daemonsetApiKey(t *testing.T) {
	baseCmd := common.HelmCommand{
		ReleaseName: "datadog",
		ChartPath:   "../../charts/datadog",
		ShowOnly:    []string{"templates/daemonset.yaml"},
		Values:      []string{"../../charts/datadog/values.yaml"},
	}

	t.Run("daemonset renders and DD_API_KEY uses default secret when datadog.apiKey is set", func(t *testing.T) {
		cmd := baseCmd
		cmd.Overrides = map[string]string{
			"datadog.apiKey": "test-api-key",
		}
		manifest, err := common.RenderChart(t, cmd)
		require.NoError(t, err, "expected daemonset to render when datadog.apiKey is set")
		require.NotEmpty(t, manifest, "expected non-empty manifest when datadog.apiKey is set")

		var ds appsv1.DaemonSet
		common.Unmarshal(t, manifest, &ds)
		require.NotEmpty(t, ds.Spec.Template.Spec.Containers)

		agentContainer := ds.Spec.Template.Spec.Containers[0]
		env, ok := findEnvVar(agentContainer.Env, DDApiKey)
		require.True(t, ok, "expected DD_API_KEY in agent container")
		require.NotNil(t, env.ValueFrom.SecretKeyRef)
		assert.Equal(t, "datadog", env.ValueFrom.SecretKeyRef.Name)
		assert.Equal(t, "api-key", env.ValueFrom.SecretKeyRef.Key)
	})

	t.Run("daemonset renders and DD_API_KEY uses existing secret when datadog.apiKeyExistingSecret is set", func(t *testing.T) {
		cmd := baseCmd
		cmd.Overrides = map[string]string{
			"datadog.apiKeyExistingSecret": "custom-api-secret",
		}
		manifest, err := common.RenderChart(t, cmd)
		require.NoError(t, err, "expected daemonset to render when datadog.apiKeyExistingSecret is set")
		require.NotEmpty(t, manifest, "expected non-empty manifest when datadog.apiKeyExistingSecret is set")

		var ds appsv1.DaemonSet
		common.Unmarshal(t, manifest, &ds)
		require.NotEmpty(t, ds.Spec.Template.Spec.Containers)

		agentContainer := ds.Spec.Template.Spec.Containers[0]
		env, ok := findEnvVar(agentContainer.Env, DDApiKey)
		require.True(t, ok, "expected DD_API_KEY in agent container")
		require.NotNil(t, env.ValueFrom.SecretKeyRef)
		assert.Equal(t, "custom-api-secret", env.ValueFrom.SecretKeyRef.Name)
		assert.Equal(t, "api-key", env.ValueFrom.SecretKeyRef.Key)
	})

	t.Run("daemonset does not render when neither datadog.apiKey nor datadog.apiKeyExistingSecret is set", func(t *testing.T) {
		cmd := baseCmd
		// No apiKey overrides — matches the default values.yaml where both are empty.
		_, err := common.RenderChart(t, cmd)
		assert.Error(t, err, "expected render to fail when no apiKey is configured")
	})
}

// Test_clusterChecksRunnerApiKey verifies that DD_API_KEY in the cluster checks runner container
// references the correct secret name and key field for both inline values and existing secret references.
func Test_clusterChecksRunnerApiKey(t *testing.T) {
	baseCmd := common.HelmCommand{
		ReleaseName: "datadog",
		ChartPath:   "../../charts/datadog",
		ShowOnly:    []string{"templates/agent-clusterchecks-deployment.yaml"},
		Values:      []string{"../../charts/datadog/values.yaml"},
		Overrides: map[string]string{
			"clusterChecksRunner.enabled": "true",
			"clusterChecks.enabled":       "true",
		},
	}

	tests := []struct {
		name           string
		extraOverrides map[string]string
		expectedSecret string
	}{
		{
			name: "DD_API_KEY uses default secret when datadog.apiKey is set",
			extraOverrides: map[string]string{
				"datadog.apiKey": "test-api-key",
			},
			expectedSecret: "datadog",
		},
		{
			name: "DD_API_KEY uses existing secret when datadog.apiKeyExistingSecret is set",
			extraOverrides: map[string]string{
				"datadog.apiKeyExistingSecret": "custom-api-secret",
			},
			expectedSecret: "custom-api-secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := baseCmd
			overrides := make(map[string]string)
			maps.Copy(overrides, baseCmd.Overrides)
			maps.Copy(overrides, tt.extraOverrides)
			cmd.Overrides = overrides

			manifest, err := common.RenderChart(t, cmd)
			require.NoError(t, err, "failed to render agent-clusterchecks-deployment.yaml")

			var deployment appsv1.Deployment
			common.Unmarshal(t, manifest, &deployment)
			require.NotEmpty(t, deployment.Spec.Template.Spec.Containers, "expected at least one container in CCR deployment")

			env, ok := findEnvVar(deployment.Spec.Template.Spec.Containers[0].Env, DDApiKey)
			require.True(t, ok, "expected DD_API_KEY to be present in CCR container")
			require.NotNil(t, env.ValueFrom.SecretKeyRef)
			assert.Equal(t, tt.expectedSecret, env.ValueFrom.SecretKeyRef.Name)
			assert.Equal(t, "api-key", env.ValueFrom.SecretKeyRef.Key)
		})
	}
}
