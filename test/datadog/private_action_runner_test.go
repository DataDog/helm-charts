package datadog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/DataDog/helm-charts/test/common"
)

const (
	DDPAREnabled          = "DD_PRIVATE_ACTION_RUNNER_ENABLED"
	DDPARSelfEnroll       = "DD_PRIVATE_ACTION_RUNNER_SELF_ENROLL"
	DDPARURN              = "DD_PRIVATE_ACTION_RUNNER_URN"
	DDPARPrivateKey       = "DD_PRIVATE_ACTION_RUNNER_PRIVATE_KEY"
	DDPARActionsAllowlist = "DD_PRIVATE_ACTION_RUNNER_ACTIONS_ALLOWLIST"
	DDPARIdentitySecret   = "DD_PRIVATE_ACTION_RUNNER_IDENTITY_SECRET_NAME"
)

func selectPAREnvVars(envVars []corev1.EnvVar) map[string]string {
	parEnvVarNames := []string{
		DDPAREnabled,
		DDPARSelfEnroll,
		DDPARURN,
		DDPARPrivateKey,
		DDPARActionsAllowlist,
		DDPARIdentitySecret,
	}

	selection := map[string]string{}

	for _, envVar := range envVars {
		for _, name := range parEnvVarNames {
			if envVar.Name == name {
				selection[name] = envVar.Value
			}
		}
	}
	return selection
}

func Test_PrivateActionRunner_Disabled(t *testing.T) {
	manifest, err := common.RenderChart(t, common.HelmCommand{
		ReleaseName: "datadog",
		ChartPath:   "../../charts/datadog",
		ShowOnly:    []string{"templates/cluster-agent-deployment.yaml", "templates/cluster-agent-rbac.yaml"},
		Values:      []string{"../../charts/datadog/values.yaml"},
		Overrides: map[string]string{
			"clusterAgent.privateActionRunner.enabled": "false",
		},
	})
	require.NoError(t, err)

	// Verify PAR env vars are not present
	var deployment appsv1.Deployment
	common.Unmarshal(t, manifest, &deployment)
	envVars := selectPAREnvVars(deployment.Spec.Template.Spec.Containers[0].Env)

	assert.Empty(t, envVars[DDPAREnabled], "PAR should not be enabled")

	// Verify PAR RBAC Role is not created
	assert.NotContains(t, manifest, "datadog-private-action-runner", "PAR Role should not be created")
}

func Test_PrivateActionRunner_Enabled_SelfEnroll(t *testing.T) {
	manifest, err := common.RenderChart(t, common.HelmCommand{
		ReleaseName: "datadog",
		ChartPath:   "../../charts/datadog",
		ShowOnly:    []string{"templates/cluster-agent-deployment.yaml", "templates/cluster-agent-rbac.yaml"},
		Values:      []string{"../../charts/datadog/values.yaml"},
		Overrides: map[string]string{
			"clusterAgent.privateActionRunner.enabled":    "true",
			"clusterAgent.privateActionRunner.selfEnroll": "true",
		},
	})
	require.NoError(t, err)

	// Verify deployment has PAR env vars
	var deployment appsv1.Deployment
	common.Unmarshal(t, manifest, &deployment)
	envVars := selectPAREnvVars(deployment.Spec.Template.Spec.Containers[0].Env)

	assert.Equal(t, "true", envVars[DDPAREnabled])
	assert.Equal(t, "true", envVars[DDPARSelfEnroll])
	assert.Empty(t, envVars[DDPARURN], "URN should not be set in self-enroll mode")
	assert.Empty(t, envVars[DDPARPrivateKey], "Private key should not be set in self-enroll mode")

	// Verify PAR RBAC is created
	assert.Contains(t, manifest, "kind: Role")
	assert.Contains(t, manifest, "datadog-private-action-runner")
	assert.Contains(t, manifest, "datadog-private-action-runner-identity")
}

func Test_PrivateActionRunner_Enabled_WithCredentials(t *testing.T) {
	manifest, err := common.RenderChart(t, common.HelmCommand{
		ReleaseName: "datadog",
		ChartPath:   "../../charts/datadog",
		ShowOnly:    []string{"templates/cluster-agent-deployment.yaml"},
		Values:      []string{"../../charts/datadog/values.yaml"},
		Overrides: map[string]string{
			"clusterAgent.privateActionRunner.enabled":    "true",
			"clusterAgent.privateActionRunner.selfEnroll": "false",
			"clusterAgent.privateActionRunner.urn":        "urn:datadog:private-action-runner:organization:123:runner:abc",
			"clusterAgent.privateActionRunner.privateKey": "test-private-key",
		},
	})
	require.NoError(t, err)

	var deployment appsv1.Deployment
	common.Unmarshal(t, manifest, &deployment)
	envVars := selectPAREnvVars(deployment.Spec.Template.Spec.Containers[0].Env)

	assert.Equal(t, "true", envVars[DDPAREnabled])
	assert.Empty(t, envVars[DDPARSelfEnroll])
	assert.Equal(t, "urn:datadog:private-action-runner:organization:123:runner:abc", envVars[DDPARURN])
	assert.Equal(t, "test-private-key", envVars[DDPARPrivateKey])
}

func Test_PrivateActionRunner_Enabled_WithExistingSecret(t *testing.T) {
	manifest, err := common.RenderChart(t, common.HelmCommand{
		ReleaseName: "datadog",
		ChartPath:   "../../charts/datadog",
		ShowOnly:    []string{"templates/cluster-agent-deployment.yaml"},
		Values:      []string{"../../charts/datadog/values.yaml"},
		Overrides: map[string]string{
			"clusterAgent.privateActionRunner.enabled":        "true",
			"clusterAgent.privateActionRunner.selfEnroll":     "false",
			"clusterAgent.privateActionRunner.identityFromExistingSecret": "my-par-secret",
		},
	})
	require.NoError(t, err)

	var deployment appsv1.Deployment
	common.Unmarshal(t, manifest, &deployment)
	container := deployment.Spec.Template.Spec.Containers[0]

	// Find URN env var and verify it uses valueFrom
	var urnEnv, privateKeyEnv *corev1.EnvVar
	for i := range container.Env {
		if container.Env[i].Name == DDPARURN {
			urnEnv = &container.Env[i]
		}
		if container.Env[i].Name == DDPARPrivateKey {
			privateKeyEnv = &container.Env[i]
		}
	}

	require.NotNil(t, urnEnv, "URN env var should exist")
	require.NotNil(t, privateKeyEnv, "Private key env var should exist")

	assert.NotNil(t, urnEnv.ValueFrom, "URN should use valueFrom")
	assert.NotNil(t, urnEnv.ValueFrom.SecretKeyRef, "URN should reference secret")
	assert.Equal(t, "my-par-secret", urnEnv.ValueFrom.SecretKeyRef.Name)
	assert.Equal(t, "urn", urnEnv.ValueFrom.SecretKeyRef.Key)

	assert.NotNil(t, privateKeyEnv.ValueFrom, "Private key should use valueFrom")
	assert.NotNil(t, privateKeyEnv.ValueFrom.SecretKeyRef, "Private key should reference secret")
	assert.Equal(t, "my-par-secret", privateKeyEnv.ValueFrom.SecretKeyRef.Name)
	assert.Equal(t, "private_key", privateKeyEnv.ValueFrom.SecretKeyRef.Key)
}

func Test_PrivateActionRunner_Enabled_WithActionsAllowlist(t *testing.T) {
	manifest, err := common.RenderChart(t, common.HelmCommand{
		ReleaseName: "datadog",
		ChartPath:   "../../charts/datadog",
		ShowOnly:    []string{"templates/cluster-agent-deployment.yaml"},
		Values:      []string{"../../charts/datadog/values.yaml"},
		OverridesJson: map[string]string{
			"clusterAgent.privateActionRunner.enabled":          `true`,
			"clusterAgent.privateActionRunner.selfEnroll":       `true`,
			"clusterAgent.privateActionRunner.actionsAllowlist": `["com.datadoghq.http.request", "com.datadoghq.traceroute"]`,
		},
	})
	require.NoError(t, err)

	var deployment appsv1.Deployment
	common.Unmarshal(t, manifest, &deployment)
	envVars := selectPAREnvVars(deployment.Spec.Template.Spec.Containers[0].Env)

	assert.Equal(t, "true", envVars[DDPAREnabled])
	assert.Contains(t, envVars[DDPARActionsAllowlist], "com.datadoghq.http.request")
	assert.Contains(t, envVars[DDPARActionsAllowlist], "com.datadoghq.traceroute")
}

func Test_PrivateActionRunner_RBAC(t *testing.T) {
	manifest, err := common.RenderChart(t, common.HelmCommand{
		ReleaseName: "datadog",
		ChartPath:   "../../charts/datadog",
		ShowOnly:    []string{"templates/cluster-agent-rbac.yaml"},
		Values:      []string{"../../charts/datadog/values.yaml"},
		Overrides: map[string]string{
			"clusterAgent.privateActionRunner.enabled":            "true",
			"clusterAgent.privateActionRunner.identitySecretName": "my-custom-secret",
		},
	})
	require.NoError(t, err)

	// Verify PAR Role exists
	assert.Contains(t, manifest, "kind: Role")
	assert.Contains(t, manifest, "name: datadog-private-action-runner")

	// Verify secret name is referenced
	assert.Contains(t, manifest, "my-custom-secret")

	// Verify permissions are present
	assert.Contains(t, manifest, "resources: [\"secrets\"]")
	assert.Contains(t, manifest, "verbs: [\"get\", \"update\", \"create\"]")

	// Verify RoleBinding is created
	assert.Contains(t, manifest, "kind: RoleBinding")
	assert.Contains(t, manifest, "roleRef:")
	assert.Contains(t, manifest, "name: datadog-private-action-runner")
	assert.Contains(t, manifest, "- kind: ServiceAccount")
	assert.Contains(t, manifest, "name: datadog-cluster-agent")
}

func Test_PrivateActionRunner_RBAC_Not_Created_When_Disabled(t *testing.T) {
	manifest, err := common.RenderChart(t, common.HelmCommand{
		ReleaseName: "datadog",
		ChartPath:   "../../charts/datadog",
		ShowOnly:    []string{"templates/cluster-agent-rbac.yaml"},
		Values:      []string{"../../charts/datadog/values.yaml"},
		Overrides: map[string]string{
			"clusterAgent.privateActionRunner.enabled": "false",
		},
	})
	require.NoError(t, err)

	// Verify PAR Role is not in the manifest
	assert.NotContains(t, manifest, "name: datadog-private-action-runner")
	// Also verify the identity secret name is not referenced
	assert.NotContains(t, manifest, "datadog-private-action-runner-identity")
}

func Test_PrivateActionRunner_Validation_SelfEnrollWithoutLeaderElection(t *testing.T) {
	_, err := common.RenderChart(t, common.HelmCommand{
		ReleaseName: "datadog",
		ChartPath:   "../../charts/datadog",
		ShowOnly:    []string{"templates/cluster-agent-deployment.yaml"},
		Values:      []string{"../../charts/datadog/values.yaml"},
		Overrides: map[string]string{
			"clusterAgent.privateActionRunner.enabled":    "true",
			"clusterAgent.privateActionRunner.selfEnroll": "true",
			"datadog.leaderElection":                      "false",
		},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "selfEnroll requires leader election to be enabled")
}

func Test_PrivateActionRunner_Validation_ManualModeWithoutCredentials(t *testing.T) {
	_, err := common.RenderChart(t, common.HelmCommand{
		ReleaseName: "datadog",
		ChartPath:   "../../charts/datadog",
		ShowOnly:    []string{"templates/cluster-agent-deployment.yaml"},
		Values:      []string{"../../charts/datadog/values.yaml"},
		Overrides: map[string]string{
			"clusterAgent.privateActionRunner.enabled":    "true",
			"clusterAgent.privateActionRunner.selfEnroll": "false",
		},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "you must provide either clusterAgent.privateActionRunner.identityFromExistingSecret or both clusterAgent.privateActionRunner.urn and clusterAgent.privateActionRunner.privateKey")
}

func Test_PrivateActionRunner_Validation_ManualModeWithOnlyURN(t *testing.T) {
	_, err := common.RenderChart(t, common.HelmCommand{
		ReleaseName: "datadog",
		ChartPath:   "../../charts/datadog",
		ShowOnly:    []string{"templates/cluster-agent-deployment.yaml"},
		Values:      []string{"../../charts/datadog/values.yaml"},
		Overrides: map[string]string{
			"clusterAgent.privateActionRunner.enabled":    "true",
			"clusterAgent.privateActionRunner.selfEnroll": "false",
			"clusterAgent.privateActionRunner.urn":        "urn:datadog:private-action-runner:organization:123:runner:abc",
		},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "you must provide either clusterAgent.privateActionRunner.identityFromExistingSecret or both clusterAgent.privateActionRunner.urn and clusterAgent.privateActionRunner.privateKey")
}

func Test_PrivateActionRunner_Validation_ManualModeWithOnlyPrivateKey(t *testing.T) {
	_, err := common.RenderChart(t, common.HelmCommand{
		ReleaseName: "datadog",
		ChartPath:   "../../charts/datadog",
		ShowOnly:    []string{"templates/cluster-agent-deployment.yaml"},
		Values:      []string{"../../charts/datadog/values.yaml"},
		Overrides: map[string]string{
			"clusterAgent.privateActionRunner.enabled":       "true",
			"clusterAgent.privateActionRunner.selfEnroll":    "false",
			"clusterAgent.privateActionRunner.privateKey": "test-key",
		},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "you must provide either clusterAgent.privateActionRunner.identityFromExistingSecret or both clusterAgent.privateActionRunner.urn and clusterAgent.privateActionRunner.privateKey")
}
