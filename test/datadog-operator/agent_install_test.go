package datadog_operator

import (
	"strings"
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)


func baseHelmCommand(overrides map[string]string, showOnly []string) common.HelmCommand {
	return common.HelmCommand{
		ReleaseName: "datadog-operator",
		ChartPath:   "../../charts/datadog-operator",
		ShowOnly:    showOnly,
		Values:      []string{"../../charts/datadog-operator/values.yaml"},
		Overrides:   overrides,
	}
}

// splitManifests splits a multi-document YAML string into individual documents.
func splitManifests(manifest string) []string {
	manifest = strings.TrimPrefix(manifest, "---")
	parts := strings.Split(manifest, "\n---")
	var docs []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			docs = append(docs, trimmed)
		}
	}
	return docs
}

// --- Validation tests ---

func Test_agent_install_fails_without_api_key(t *testing.T) {
	_, err := common.RenderChart(t, baseHelmCommand(
		map[string]string{
			"installAgents": "true",
		},
		nil,
	))
	require.Error(t, err, "should fail when installAgents is true but no API key source is configured")
	assert.Contains(t, err.Error(), "no API key source")
}

func Test_agent_install_fails_when_datadogAgent_disabled(t *testing.T) {
	_, err := common.RenderChart(t, baseHelmCommand(
		map[string]string{
			"installAgents":       "true",
			"apiKey":              "test-api-key",
			"datadogAgent.enabled": "false",
		},
		nil,
	))
	require.Error(t, err, "should fail when installAgents is true but datadogAgent controller is disabled")
	assert.Contains(t, err.Error(), "datadogAgent.enabled")
}

// --- installAgents=false (default) ---

func Test_agent_install_disabled_by_default(t *testing.T) {
	manifest, err := common.RenderChart(t, baseHelmCommand(
		map[string]string{},
		nil, // render all templates
	))
	require.NoError(t, err)
	assert.NotContains(t, manifest, "agent-install", "agent-install resources should not be rendered when installAgents is false")
}

// --- Job template tests ---

func Test_agent_install_name_label_within_63_chars(t *testing.T) {
	// nameOverride at 63 chars (the max from the name helper) + "-agent-install" = 77,
	// which must be truncated back to 63.
	longName := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" // 63 chars
	manifest, err := common.RenderChart(t, baseHelmCommand(
		map[string]string{
			"installAgents": "true",
			"apiKey":        "test-api-key",
			"nameOverride":  longName,
		},
		[]string{"templates/agent-install-job.yaml"},
	))
	require.NoError(t, err)

	var job batchv1.Job
	common.Unmarshal(t, manifest, &job)
	nameLabel := job.Spec.Template.Labels["app.kubernetes.io/name"]
	assert.LessOrEqual(t, len(nameLabel), 63, "app.kubernetes.io/name label must be <= 63 chars, got %d: %q", len(nameLabel), nameLabel)
	assert.True(t, strings.HasSuffix(nameLabel, "-agent-install"), "label must preserve -agent-install suffix, got %q", nameLabel)
}

func Test_agent_install_fullname_preserves_suffix(t *testing.T) {
	// A 63-char fullnameOverride should be truncated to 49 before appending
	// -agent-install, producing a 63-char name that doesn't collide with
	// the unsuffixed fullname.
	longFullname := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" // 63 chars
	manifest, err := common.RenderChart(t, baseHelmCommand(
		map[string]string{
			"installAgents":    "true",
			"apiKey":           "test-api-key",
			"fullnameOverride": longFullname,
		},
		[]string{"templates/agent-install-job.yaml"},
	))
	require.NoError(t, err)

	var job batchv1.Job
	common.Unmarshal(t, manifest, &job)
	assert.Contains(t, job.Name, "-agent-install-", "Job name must contain -agent-install- suffix before revision")
	assert.NotEqual(t, longFullname, job.Name, "Job name must not collide with unsuffixed fullname")
}

func Test_agent_install_job_rendered_with_apiKey(t *testing.T) {
	manifest, err := common.RenderChart(t, baseHelmCommand(
		map[string]string{
			"installAgents": "true",
			"apiKey":        "test-api-key",
		},
		[]string{"templates/agent-install-job.yaml"},
	))
	require.NoError(t, err)

	var job batchv1.Job
	common.Unmarshal(t, manifest, &job)

	assert.Equal(t, "datadog-operator-agent-install-1", job.Name)
	assert.Empty(t, job.Annotations["helm.sh/hook"], "should not use Helm hooks (not supported by EKS add-ons)")
	assert.Equal(t, int32(5), *job.Spec.BackoffLimit)
	assert.NotNil(t, job.Spec.TTLSecondsAfterFinished)
	assert.Equal(t, int32(300), *job.Spec.TTLSecondsAfterFinished)

	container := job.Spec.Template.Spec.Containers[0]
	assert.Equal(t, "bitnami/kubectl:1.31", container.Image)
}

func Test_agent_install_job_uses_bundled_configmap(t *testing.T) {
	manifest, err := common.RenderChart(t, baseHelmCommand(
		map[string]string{
			"installAgents": "true",
			"apiKey":        "test-api-key",
		},
		[]string{"templates/agent-install-job.yaml"},
	))
	require.NoError(t, err)

	var job batchv1.Job
	common.Unmarshal(t, manifest, &job)

	// ConfigMap volume should be mounted
	require.Equal(t, 1, len(job.Spec.Template.Spec.Volumes))
	assert.Equal(t, "agent-config", job.Spec.Template.Spec.Volumes[0].Name)
	assert.NotNil(t, job.Spec.Template.Spec.Volumes[0].ConfigMap)

	require.Equal(t, 1, len(job.Spec.Template.Spec.Containers[0].VolumeMounts))
	assert.Equal(t, "/etc/agent-config", job.Spec.Template.Spec.Containers[0].VolumeMounts[0].MountPath)
	assert.True(t, job.Spec.Template.Spec.Containers[0].VolumeMounts[0].ReadOnly)
}

func Test_agent_install_configmap_contains_default_agent_cr(t *testing.T) {
	manifest, err := common.RenderChart(t, baseHelmCommand(
		map[string]string{
			"installAgents": "true",
			"apiKey":        "test-api-key",
		},
		[]string{"templates/agent-default-config.yaml"},
	))
	require.NoError(t, err)
	assert.Contains(t, manifest, "kind: ConfigMap")
	assert.Contains(t, manifest, "agent-config.yaml")
	assert.Contains(t, manifest, "__DD_API_SECRET_NAME__")
	assert.Contains(t, manifest, "kind: DatadogAgent")
}

func Test_agent_install_job_api_secret_from_apiKey(t *testing.T) {
	manifest, err := common.RenderChart(t, baseHelmCommand(
		map[string]string{
			"installAgents": "true",
			"apiKey":        "test-api-key",
		},
		[]string{"templates/agent-install-job.yaml"},
	))
	require.NoError(t, err)

	var job batchv1.Job
	common.Unmarshal(t, manifest, &job)

	env := job.Spec.Template.Spec.Containers[0].Env
	apiEnv := findEnvVar(env, "API_SECRET_NAME")
	require.NotNil(t, apiEnv)
	assert.Equal(t, "datadog-operator-apikey", apiEnv.Value)
}

func Test_agent_install_job_api_secret_from_existing_secret(t *testing.T) {
	manifest, err := common.RenderChart(t, baseHelmCommand(
		map[string]string{
			"installAgents":       "true",
			"apiKeyExistingSecret": "my-api-secret",
		},
		[]string{"templates/agent-install-job.yaml"},
	))
	require.NoError(t, err)

	var job batchv1.Job
	common.Unmarshal(t, manifest, &job)

	env := job.Spec.Template.Spec.Containers[0].Env
	apiEnv := findEnvVar(env, "API_SECRET_NAME")
	require.NotNil(t, apiEnv)
	assert.Equal(t, "my-api-secret", apiEnv.Value)
}

func Test_agent_install_job_no_app_secret_when_appKey_unset(t *testing.T) {
	manifest, err := common.RenderChart(t, baseHelmCommand(
		map[string]string{
			"installAgents": "true",
			"apiKey":        "test-api-key",
		},
		[]string{"templates/agent-install-job.yaml"},
	))
	require.NoError(t, err)

	var job batchv1.Job
	common.Unmarshal(t, manifest, &job)

	env := job.Spec.Template.Spec.Containers[0].Env
	appEnv := findEnvVar(env, "APP_SECRET_NAME")
	assert.Nil(t, appEnv, "APP_SECRET_NAME should not be set when no app key is configured")
}

func Test_agent_install_job_app_secret_from_appKey(t *testing.T) {
	manifest, err := common.RenderChart(t, baseHelmCommand(
		map[string]string{
			"installAgents": "true",
			"apiKey":        "test-api-key",
			"appKey":        "test-app-key",
		},
		[]string{"templates/agent-install-job.yaml"},
	))
	require.NoError(t, err)

	var job batchv1.Job
	common.Unmarshal(t, manifest, &job)

	env := job.Spec.Template.Spec.Containers[0].Env
	appEnv := findEnvVar(env, "APP_SECRET_NAME")
	require.NotNil(t, appEnv, "APP_SECRET_NAME should be set when appKey is configured")
	assert.Equal(t, "datadog-operator-appkey", appEnv.Value)
}

func Test_agent_install_job_app_secret_from_existing_secret(t *testing.T) {
	manifest, err := common.RenderChart(t, baseHelmCommand(
		map[string]string{
			"installAgents":       "true",
			"apiKey":              "test-api-key",
			"appKeyExistingSecret": "my-app-secret",
		},
		[]string{"templates/agent-install-job.yaml"},
	))
	require.NoError(t, err)

	var job batchv1.Job
	common.Unmarshal(t, manifest, &job)

	env := job.Spec.Template.Spec.Containers[0].Env
	appEnv := findEnvVar(env, "APP_SECRET_NAME")
	require.NotNil(t, appEnv, "APP_SECRET_NAME should be set when appKeyExistingSecret is configured")
	assert.Equal(t, "my-app-secret", appEnv.Value)
}

func Test_agent_install_job_script_substitutes_api_key(t *testing.T) {
	manifest, err := common.RenderChart(t, baseHelmCommand(
		map[string]string{
			"installAgents": "true",
			"apiKey":        "test-api-key",
		},
		[]string{"templates/agent-install-job.yaml"},
	))
	require.NoError(t, err)

	assert.Contains(t, manifest, `sed -i "s/__DD_API_SECRET_NAME__/$API_SECRET_NAME/g"`)
	assert.Contains(t, manifest, `sed -i "s/__DD_NAMESPACE__/$NAMESPACE/g"`)
	assert.Contains(t, manifest, "kubectl auth can-i create datadogagents", "script should wait for RBAC propagation before proceeding")
}

func Test_agent_install_job_substitutes_app_secret_when_set(t *testing.T) {
	manifest, err := common.RenderChart(t, baseHelmCommand(
		map[string]string{
			"installAgents": "true",
			"apiKey":        "test-api-key",
			"appKey":        "test-app-key",
		},
		[]string{"templates/agent-install-job.yaml"},
	))
	require.NoError(t, err)

	assert.Contains(t, manifest, `sed -i "s/__DD_APP_SECRET_NAME__/$APP_SECRET_NAME/g"`)
}

func Test_agent_install_job_no_app_secret_substitution_when_unset(t *testing.T) {
	manifest, err := common.RenderChart(t, baseHelmCommand(
		map[string]string{
			"installAgents": "true",
			"apiKey":        "test-api-key",
		},
		[]string{"templates/agent-install-job.yaml"},
	))
	require.NoError(t, err)

	// No app secret substitution or awk when app key is not configured
	assert.NotContains(t, manifest, "__DD_APP_SECRET_NAME__")
	assert.NotContains(t, manifest, "awk")
}

func Test_agent_install_configmap_includes_appSecret_when_set(t *testing.T) {
	manifest, err := common.RenderChart(t, baseHelmCommand(
		map[string]string{
			"installAgents": "true",
			"apiKey":        "test-api-key",
			"appKey":        "test-app-key",
		},
		[]string{"templates/agent-default-config.yaml"},
	))
	require.NoError(t, err)
	assert.Contains(t, manifest, "appSecret:")
	assert.Contains(t, manifest, "__DD_APP_SECRET_NAME__")
}

func Test_agent_install_configmap_excludes_appSecret_when_unset(t *testing.T) {
	manifest, err := common.RenderChart(t, baseHelmCommand(
		map[string]string{
			"installAgents": "true",
			"apiKey":        "test-api-key",
		},
		[]string{"templates/agent-default-config.yaml"},
	))
	require.NoError(t, err)
	assert.NotContains(t, manifest, "appSecret:")
}

func Test_agent_install_job_inherits_nodeSelector(t *testing.T) {
	manifest, err := common.RenderChart(t, baseHelmCommand(
		map[string]string{
			"installAgents": "true",
			"apiKey":        "test-api-key",
		},
		[]string{"templates/agent-install-job.yaml"},
	))
	require.NoError(t, err)

	var job batchv1.Job
	common.Unmarshal(t, manifest, &job)

	nodeSelector := job.Spec.Template.Spec.NodeSelector
	assert.Equal(t, "linux", nodeSelector["kubernetes.io/os"])
}

func Test_agent_install_job_propagates_imagePullSecrets(t *testing.T) {
	cmd := baseHelmCommand(
		map[string]string{
			"installAgents": "true",
			"apiKey":        "test-api-key",
		},
		[]string{"templates/agent-install-job.yaml"},
	)
	cmd.OverridesJson = map[string]string{
		"imagePullSecrets": `[{"name": "my-registry-secret"}]`,
	}
	manifest, err := common.RenderChart(t, cmd)
	require.NoError(t, err)

	var job batchv1.Job
	common.Unmarshal(t, manifest, &job)

	pullSecrets := job.Spec.Template.Spec.ImagePullSecrets
	require.Equal(t, 1, len(pullSecrets))
	assert.Equal(t, "my-registry-secret", pullSecrets[0].Name)
}

func Test_agent_install_job_no_imagePullSecrets_when_unset(t *testing.T) {
	manifest, err := common.RenderChart(t, baseHelmCommand(
		map[string]string{
			"installAgents": "true",
			"apiKey":        "test-api-key",
		},
		[]string{"templates/agent-install-job.yaml"},
	))
	require.NoError(t, err)

	var job batchv1.Job
	common.Unmarshal(t, manifest, &job)
	assert.Empty(t, job.Spec.Template.Spec.ImagePullSecrets)
}

func Test_agent_install_job_uses_operator_sa_when_sa_create_false(t *testing.T) {
	manifest, err := common.RenderChart(t, baseHelmCommand(
		map[string]string{
			"installAgents":        "true",
			"apiKey":               "test-api-key",
			"serviceAccount.create": "false",
			"serviceAccount.name":   "my-preprovisioned-sa",
		},
		[]string{"templates/agent-install-job.yaml"},
	))
	require.NoError(t, err)

	var job batchv1.Job
	common.Unmarshal(t, manifest, &job)
	assert.Equal(t, "my-preprovisioned-sa", job.Spec.Template.Spec.ServiceAccountName)
}

func Test_agent_install_job_uses_operator_sa_when_rbac_create_false(t *testing.T) {
	manifest, err := common.RenderChart(t, baseHelmCommand(
		map[string]string{
			"installAgents": "true",
			"apiKey":        "test-api-key",
			"rbac.create":   "false",
		},
		[]string{"templates/agent-install-job.yaml"},
	))
	require.NoError(t, err)

	var job batchv1.Job
	common.Unmarshal(t, manifest, &job)
	// When rbac.create=false, the dedicated SA has no bindings so the Job
	// must fall back to the operator's SA.
	assert.Equal(t, "datadog-operator", job.Spec.Template.Spec.ServiceAccountName)
}

// --- Namespace resolution tests ---

func Test_agent_install_namespace_defaults_to_release(t *testing.T) {
	cmd := baseHelmCommand(
		map[string]string{
			"installAgents": "true",
			"apiKey":        "test-api-key",
		},
		[]string{"templates/agent-install-job.yaml"},
	)
	cmd.Namespace = "my-release-ns"
	manifest, err := common.RenderChart(t, cmd)
	require.NoError(t, err)

	var job batchv1.Job
	common.Unmarshal(t, manifest, &job)
	nsEnv := findEnvVar(job.Spec.Template.Spec.Containers[0].Env, "NAMESPACE")
	require.NotNil(t, nsEnv)
	assert.Equal(t, "my-release-ns", nsEnv.Value)
}

func Test_agent_install_fails_when_watchNamespacesAgent_excludes_release_ns(t *testing.T) {
	cmd := baseHelmCommand(
		map[string]string{
			"installAgents": "true",
			"apiKey":        "test-api-key",
		},
		[]string{"templates/agent-install-job.yaml"},
	)
	cmd.Namespace = "release-ns"
	cmd.OverridesJson = map[string]string{
		"watchNamespacesAgent": `["agents-ns", "other-ns"]`,
	}
	_, err := common.RenderChart(t, cmd)
	require.Error(t, err, "should fail when watchNamespacesAgent excludes the release namespace")
	assert.Contains(t, err.Error(), "watchNamespacesAgent")
	assert.Contains(t, err.Error(), "release-ns")
}

func Test_agent_install_namespace_release_ns_in_watchNamespacesAgent(t *testing.T) {
	cmd := baseHelmCommand(
		map[string]string{
			"installAgents": "true",
			"apiKey":        "test-api-key",
		},
		[]string{"templates/agent-install-job.yaml"},
	)
	cmd.Namespace = "release-ns"
	cmd.OverridesJson = map[string]string{
		"watchNamespacesAgent": `["release-ns", "other-ns"]`,
	}
	manifest, err := common.RenderChart(t, cmd)
	require.NoError(t, err)

	var job batchv1.Job
	common.Unmarshal(t, manifest, &job)
	nsEnv := findEnvVar(job.Spec.Template.Spec.Containers[0].Env, "NAMESPACE")
	require.NotNil(t, nsEnv)
	assert.Equal(t, "release-ns", nsEnv.Value, "should prefer release namespace when it is in the watch list")
}

func Test_agent_install_namespace_watchAll(t *testing.T) {
	cmd := baseHelmCommand(
		map[string]string{
			"installAgents": "true",
			"apiKey":        "test-api-key",
		},
		[]string{"templates/agent-install-job.yaml"},
	)
	cmd.Namespace = "release-ns"
	cmd.OverridesJson = map[string]string{
		"watchNamespacesAgent": `[""]`,
	}
	manifest, err := common.RenderChart(t, cmd)
	require.NoError(t, err)

	var job batchv1.Job
	common.Unmarshal(t, manifest, &job)
	nsEnv := findEnvVar(job.Spec.Template.Spec.Containers[0].Env, "NAMESPACE")
	require.NotNil(t, nsEnv)
	assert.Equal(t, "release-ns", nsEnv.Value, "should use release namespace when watching all namespaces")
}

func Test_agent_install_fails_when_watchNamespaces_excludes_release_ns(t *testing.T) {
	cmd := baseHelmCommand(
		map[string]string{
			"installAgents": "true",
			"apiKey":        "test-api-key",
		},
		[]string{"templates/agent-install-job.yaml"},
	)
	cmd.Namespace = "release-ns"
	cmd.OverridesJson = map[string]string{
		"watchNamespaces": `["monitoring"]`,
	}
	_, err := common.RenderChart(t, cmd)
	require.Error(t, err, "should fail when watchNamespaces excludes the release namespace")
	assert.Contains(t, err.Error(), "watchNamespaces")
	assert.Contains(t, err.Error(), "release-ns")
}

func Test_agent_install_namespace_watchNamespaces_includes_release_ns(t *testing.T) {
	cmd := baseHelmCommand(
		map[string]string{
			"installAgents": "true",
			"apiKey":        "test-api-key",
		},
		[]string{"templates/agent-install-job.yaml"},
	)
	cmd.Namespace = "release-ns"
	cmd.OverridesJson = map[string]string{
		"watchNamespaces": `["release-ns", "other"]`,
	}
	manifest, err := common.RenderChart(t, cmd)
	require.NoError(t, err)

	var job batchv1.Job
	common.Unmarshal(t, manifest, &job)
	nsEnv := findEnvVar(job.Spec.Template.Spec.Containers[0].Env, "NAMESPACE")
	require.NotNil(t, nsEnv)
	assert.Equal(t, "release-ns", nsEnv.Value)
}

// --- RBAC template tests ---

func Test_agent_install_rbac_uses_namespaced_role(t *testing.T) {
	manifest, err := common.RenderChart(t, baseHelmCommand(
		map[string]string{
			"installAgents": "true",
			"apiKey":        "test-api-key",
		},
		[]string{"templates/agent-install-rbac.yaml"},
	))
	require.NoError(t, err)

	// Must be Role/RoleBinding, not ClusterRole/ClusterRoleBinding
	assert.Contains(t, manifest, "kind: Role")
	assert.Contains(t, manifest, "kind: RoleBinding")
	assert.NotContains(t, manifest, "kind: ClusterRole")
	assert.NotContains(t, manifest, "kind: ClusterRoleBinding")

	docs := splitManifests(manifest)
	require.Equal(t, 3, len(docs), "RBAC template should render 3 documents: ServiceAccount, Role, RoleBinding")

	// Parse ServiceAccount
	var sa corev1.ServiceAccount
	common.Unmarshal(t, docs[0], &sa)
	assert.Equal(t, "datadog-operator-agent-install", sa.Name)

	// Parse Role
	var role rbacv1.Role
	common.Unmarshal(t, docs[1], &role)
	assert.Equal(t, "datadog-operator-agent-install", role.Name)
	require.Equal(t, 1, len(role.Rules))
	assert.Equal(t, []string{"datadoghq.com"}, role.Rules[0].APIGroups)
	assert.Equal(t, []string{"datadogagents"}, role.Rules[0].Resources)
	assert.Equal(t, []string{"get", "list", "create", "update", "patch"}, role.Rules[0].Verbs)

	// Parse RoleBinding
	var rb rbacv1.RoleBinding
	common.Unmarshal(t, docs[2], &rb)
	assert.Equal(t, "datadog-operator-agent-install", rb.Name)
	assert.Equal(t, "Role", rb.RoleRef.Kind)
	assert.Equal(t, "datadog-operator-agent-install", rb.RoleRef.Name)
	require.Equal(t, 1, len(rb.Subjects))
	assert.Equal(t, "ServiceAccount", rb.Subjects[0].Kind)
	assert.Equal(t, "datadog-operator-agent-install", rb.Subjects[0].Name)
}

func Test_agent_install_rbac_no_hook_annotations(t *testing.T) {
	manifest, err := common.RenderChart(t, baseHelmCommand(
		map[string]string{
			"installAgents": "true",
			"apiKey":        "test-api-key",
		},
		[]string{"templates/agent-install-rbac.yaml"},
	))
	require.NoError(t, err)

	// No Helm hook annotations (not supported by EKS add-ons)
	assert.NotContains(t, manifest, "helm.sh/hook")
	assert.NotContains(t, manifest, "helm.sh/hook-weight")
	assert.NotContains(t, manifest, "helm.sh/hook-delete-policy")
}

func Test_agent_install_rbac_skips_sa_when_serviceAccount_create_false(t *testing.T) {
	manifest, err := common.RenderChart(t, baseHelmCommand(
		map[string]string{
			"installAgents":        "true",
			"apiKey":               "test-api-key",
			"serviceAccount.create": "false",
		},
		[]string{"templates/agent-install-rbac.yaml"},
	))
	require.NoError(t, err)
	docs := splitManifests(manifest)
	// Only Role and RoleBinding, no standalone ServiceAccount document
	require.Equal(t, 2, len(docs))
	assert.Contains(t, docs[0], "kind: Role")
	assert.Contains(t, docs[1], "kind: RoleBinding")
}

func Test_agent_install_rbac_nothing_when_rbac_create_false(t *testing.T) {
	// When rbac.create=false, no Role/RoleBinding are created and the
	// dedicated SA is also skipped (it would have no bindings).
	_, err := common.RenderChart(t, baseHelmCommand(
		map[string]string{
			"installAgents": "true",
			"apiKey":        "test-api-key",
			"rbac.create":   "false",
		},
		[]string{"templates/agent-install-rbac.yaml"},
	))
	assert.Error(t, err, "no resources should be rendered when rbac.create=false")
}

func Test_agent_install_rbac_not_rendered_when_disabled(t *testing.T) {
	_, err := common.RenderChart(t, baseHelmCommand(
		map[string]string{},
		[]string{"templates/agent-install-rbac.yaml"},
	))
	// When installAgents is false, the template produces no output.
	// helm template --show-only returns an error for empty templates.
	assert.Error(t, err)
}

func Test_agent_install_job_not_rendered_when_disabled(t *testing.T) {
	_, err := common.RenderChart(t, baseHelmCommand(
		map[string]string{},
		[]string{"templates/agent-install-job.yaml"},
	))
	assert.Error(t, err)
}

// --- Helper ---

func findEnvVar(envs []corev1.EnvVar, name string) *corev1.EnvVar {
	for i, env := range envs {
		if env.Name == name {
			return &envs[i]
		}
	}
	return nil
}
