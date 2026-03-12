package datadog_operator

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

const (
	defaultAgentConfigURL = "https://raw.githubusercontent.com/DataDog/integrations-management/main/eks-addon/default-datadog-agent.yaml"
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
	assert.False(t, strings.HasSuffix(nameLabel, "-"), "label value should not end with a hyphen")
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

	assert.Equal(t, "datadog-operator-agent-install", job.Name)
	assert.Equal(t, "post-install,post-upgrade", job.Annotations["helm.sh/hook"])
	assert.Equal(t, "1", job.Annotations["helm.sh/hook-weight"])
	assert.Equal(t, "before-hook-creation,hook-succeeded", job.Annotations["helm.sh/hook-delete-policy"])
	assert.Equal(t, int32(5), *job.Spec.BackoffLimit)

	container := job.Spec.Template.Spec.Containers[0]
	assert.Equal(t, "bitnami/kubectl:1.31", container.Image)
}

func Test_agent_install_job_default_config_url(t *testing.T) {
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
	urlEnv := findEnvVar(env, "AGENT_CONFIG_URL")
	require.NotNil(t, urlEnv)
	assert.Equal(t, defaultAgentConfigURL, urlEnv.Value)
}

func Test_agent_install_job_custom_config_url(t *testing.T) {
	customURL := "https://example.com/my-custom-config.yaml"
	manifest, err := common.RenderChart(t, baseHelmCommand(
		map[string]string{
			"installAgents": "true",
			"apiKey":        "test-api-key",
			"agentConfigUrl": customURL,
		},
		[]string{"templates/agent-install-job.yaml"},
	))
	require.NoError(t, err)

	var job batchv1.Job
	common.Unmarshal(t, manifest, &job)

	env := job.Spec.Template.Spec.Containers[0].Env
	urlEnv := findEnvVar(env, "AGENT_CONFIG_URL")
	require.NotNil(t, urlEnv)
	assert.Equal(t, customURL, urlEnv.Value)
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
}

func Test_agent_install_job_script_appSecret_branch_when_set(t *testing.T) {
	manifest, err := common.RenderChart(t, baseHelmCommand(
		map[string]string{
			"installAgents": "true",
			"apiKey":        "test-api-key",
			"appKey":        "test-app-key",
		},
		[]string{"templates/agent-install-job.yaml"},
	))
	require.NoError(t, err)

	// When app key is set, the script should have the substitution branch
	assert.Contains(t, manifest, `sed -i "s/__DD_APP_SECRET_NAME__/$APP_SECRET_NAME/g"`)
}

func Test_agent_install_job_script_appSecret_removal_when_unset(t *testing.T) {
	manifest, err := common.RenderChart(t, baseHelmCommand(
		map[string]string{
			"installAgents": "true",
			"apiKey":        "test-api-key",
		},
		[]string{"templates/agent-install-job.yaml"},
	))
	require.NoError(t, err)

	// When app key is not set, the script should have the awk-based removal
	// that only drops appSecret blocks containing the placeholder token
	assert.Contains(t, manifest, "/appSecret:/")
	assert.Contains(t, manifest, "__DD_APP_SECRET_NAME__")
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
	assert.Equal(t, "post-install,post-upgrade", sa.Annotations["helm.sh/hook"])

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

func Test_agent_install_rbac_hook_annotations(t *testing.T) {
	manifest, err := common.RenderChart(t, baseHelmCommand(
		map[string]string{
			"installAgents": "true",
			"apiKey":        "test-api-key",
		},
		[]string{"templates/agent-install-rbac.yaml"},
	))
	require.NoError(t, err)

	docs := splitManifests(manifest)
	require.Equal(t, 3, len(docs))

	// All three resources should have hook-weight "0" (before the job at weight "1")
	for _, doc := range docs {
		assert.Contains(t, doc, `"helm.sh/hook-weight": "0"`)
		assert.Contains(t, doc, `"helm.sh/hook-delete-policy": before-hook-creation,hook-succeeded`)
	}
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

// --- appSecret awk removal logic tests ---
//
// These tests execute the actual awk program from the hook script against
// fixture inputs to verify runtime behavior.

// The awk program embedded in agent-install-job.yaml. Kept in sync manually;
// if the template changes, update this constant and the tests will catch
// regressions in the new logic.
const appSecretAwk = `
/appSecret:/ {
  match($0,/^[ \t]*/); n=RLENGTH; buf=$0 ORS; next
}
buf!="" {
  match($0,/^[ \t]*/);
  if(RLENGTH>n){ buf=buf $0 ORS; next }
  if(buf !~ /__DD_APP_SECRET_NAME__/) printf "%s",buf;
  buf=""; n=-1
}
{ print }
END { if(buf!="" && buf !~ /__DD_APP_SECRET_NAME__/) printf "%s",buf }
`

// runAwk writes input to a temp file, runs awk, and returns the output.
func runAwk(t *testing.T, input string) string {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "awk-test-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(input)
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	cmd := exec.Command("awk", appSecretAwk, tmpFile.Name())
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "awk failed: %s", string(out))
	return string(out)
}

func Test_awk_removes_placeholder_appSecret_block(t *testing.T) {
	input := `spec:
  global:
    credentials:
      apiSecret:
        secretName: my-api-secret
        keyName: api-key
      appSecret:
        secretName: __DD_APP_SECRET_NAME__
        keyName: app-key
  features:
    apm:
      enabled: true
`
	output := runAwk(t, input)
	assert.NotContains(t, output, "appSecret")
	assert.NotContains(t, output, "__DD_APP_SECRET_NAME__")
	assert.Contains(t, output, "apiSecret")
	assert.Contains(t, output, "my-api-secret")
	assert.Contains(t, output, "apm")
}

func Test_awk_preserves_real_appSecret_block(t *testing.T) {
	input := `spec:
  global:
    credentials:
      appSecret:
        secretName: my-real-secret
        keyName: app-key
`
	output := runAwk(t, input)
	assert.Contains(t, output, "appSecret")
	assert.Contains(t, output, "my-real-secret")
	assert.Contains(t, output, "keyName: app-key")
}

func Test_awk_mixed_config_removes_only_placeholder_block(t *testing.T) {
	input := `spec:
  global:
    credentials:
      appSecret:
        secretName: __DD_APP_SECRET_NAME__
        keyName: app-key
---
spec:
  global:
    credentials:
      appSecret:
        secretName: real-secret-for-other-resource
        keyName: app-key
`
	output := runAwk(t, input)
	assert.NotContains(t, output, "__DD_APP_SECRET_NAME__", "placeholder block should be removed")
	assert.Contains(t, output, "real-secret-for-other-resource", "real appSecret block should be preserved")
}

func Test_awk_no_appSecret_passes_through(t *testing.T) {
	input := `spec:
  global:
    credentials:
      apiSecret:
        secretName: my-api-secret
        keyName: api-key
  features:
    apm:
      enabled: true
`
	output := runAwk(t, input)
	assert.Equal(t, input, output, "input with no appSecret should pass through unchanged")
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
