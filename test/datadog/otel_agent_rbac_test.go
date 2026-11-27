package datadog

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	rbacv1 "k8s.io/api/rbac/v1"

	"github.com/DataDog/helm-charts/test/common"
)

// otelCollectorConfigWithProcessor returns an OTel Collector config that declares
// the given processor (with an optional body) and uses it in a traces pipeline.
func otelCollectorConfigWithProcessor(name, body string) string {
	return fmt.Sprintf(`receivers:
  otlp:
    protocols:
      grpc:
        endpoint: "0.0.0.0:4317"
processors:
  %s:%s
exporters:
  debug:
service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [%s]
      exporters: [debug]
`, name, body, name)
}

func Test_otelAgentK8sAttributesRBAC(t *testing.T) {
	targets := []struct {
		name      string
		template  string
		overrides map[string]string
		configKey string
	}{
		{
			name:     "daemonset otel-agent",
			template: "templates/otel-agent-rbac.yaml",
			overrides: map[string]string{
				"datadog.apiKeyExistingSecret":  "datadog-secret",
				"datadog.appKeyExistingSecret":  "datadog-secret",
				"datadog.otelCollector.enabled": "true",
			},
			configKey: "datadog.otelCollector.config",
		},
		{
			name:     "otel agent gateway",
			template: "templates/otel-agent-gateway-rbac.yaml",
			overrides: map[string]string{
				"datadog.apiKeyExistingSecret": "datadog-secret",
				"datadog.appKeyExistingSecret": "datadog-secret",
				"otelAgentGateway.enabled":     "true",
			},
			configKey: "otelAgentGateway.config",
		},
	}

	processors := []struct {
		name          string
		body          string
		expectedRules bool
	}{
		{name: "k8s_attributes", expectedRules: true},
		{name: "k8s_attributes/custom", expectedRules: true},
		{name: "k8sattributes", expectedRules: true},
		{name: "k8sattributes/custom", expectedRules: true},
		{name: "k8s_attributes", body: "\n    passthrough: true", expectedRules: false},
		{name: "k8sattributes", body: "\n    passthrough: true", expectedRules: false},
		{name: "batch", expectedRules: false},
	}

	for _, target := range targets {
		for _, processor := range processors {
			testName := fmt.Sprintf("%s/%s", target.name, processor.name)
			if processor.body != "" {
				testName += "_passthrough"
			}
			t.Run(testName, func(t *testing.T) {
				overrides := map[string]string{}
				for k, v := range target.overrides {
					overrides[k] = v
				}
				config := otelCollectorConfigWithProcessor(processor.name, processor.body)

				manifest, err := common.RenderChart(t, common.HelmCommand{
					ReleaseName: "datadog",
					ChartPath:   "../../charts/datadog",
					ShowOnly:    []string{target.template},
					Values:      []string{"../../charts/datadog/values.yaml"},
					Overrides:   overrides,
					// Use --set-json so the multi-line config isn't parsed by --set.
					OverridesJson: map[string]string{
						target.configKey: fmt.Sprintf("%q", config),
					},
				})

				if !processor.expectedRules {
					// helm template --show-only fails when the template renders nothing.
					require.Error(t, err, "expected no ClusterRole to be rendered")
					assert.Contains(t, err.Error(), "could not find template")
					return
				}

				require.NoError(t, err, "couldn't render template")
				clusterRole := findClusterRole(t, manifest)
				assertK8sAttributesRules(t, clusterRole.Rules)
			})
		}
	}
}

func findClusterRole(t *testing.T, manifest string) rbacv1.ClusterRole {
	for _, doc := range strings.Split(manifest, "\n---") {
		if strings.Contains(doc, "kind: ClusterRole\n") {
			var clusterRole rbacv1.ClusterRole
			common.Unmarshal(t, doc, &clusterRole)
			return clusterRole
		}
	}
	require.FailNow(t, "ClusterRole not found in manifest")
	return rbacv1.ClusterRole{}
}

func assertK8sAttributesRules(t *testing.T, rules []rbacv1.PolicyRule) {
	assert.Contains(t, rules, rbacv1.PolicyRule{
		APIGroups: []string{""},
		Resources: []string{"pods", "namespaces"},
		Verbs:     []string{"get", "watch", "list"},
	})
	assert.Contains(t, rules, rbacv1.PolicyRule{
		APIGroups: []string{"apps"},
		Resources: []string{"replicasets"},
		Verbs:     []string{"get", "list", "watch"},
	})
}
