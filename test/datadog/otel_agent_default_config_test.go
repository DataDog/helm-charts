package datadog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"

	"github.com/DataDog/helm-charts/test/common"
)

type otelCollectorConfig struct {
	Exporters map[string]any `yaml:"exporters"`
	Service   struct {
		Pipelines map[string]struct {
			Exporters []string `yaml:"exporters"`
		} `yaml:"pipelines"`
	} `yaml:"service"`
}

// Test_otelAgentDefaultGatewayConfigOTLPHTTPExporterName checks that the default
// DaemonSet OTel config used with the OTel Agent Gateway names the OTLP/HTTP exporter
// `otlp_http` on Agent 7.76.0+ (OTel Collector v0.144.0+) and keeps the deprecated
// `otlphttp` alias on older Agents.
func Test_otelAgentDefaultGatewayConfigOTLPHTTPExporterName(t *testing.T) {
	tests := []struct {
		name         string
		overrides    map[string]string
		expectedName string
	}{
		{name: "default agent image", overrides: map[string]string{}, expectedName: "otlp_http"},
		{name: "agent 7.75.0", overrides: map[string]string{"agents.image.tag": "7.75.0"}, expectedName: "otlphttp"},
		{name: "agent 7.76.0", overrides: map[string]string{"agents.image.tag": "7.76.0"}, expectedName: "otlp_http"},
		{name: "agent 7.76.0 prerelease", overrides: map[string]string{"agents.image.tag": "7.76.0-rc.1"}, expectedName: "otlp_http"},
		{name: "floating latest tag", overrides: map[string]string{"agents.image.tag": "latest"}, expectedName: "otlp_http"},
		{
			name:         "doNotCheckTag",
			overrides:    map[string]string{"agents.image.tag": "7.70.0", "agents.image.doNotCheckTag": "true"},
			expectedName: "otlp_http",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			overrides := map[string]string{
				"datadog.apiKeyExistingSecret":  "datadog-secret",
				"datadog.appKeyExistingSecret":  "datadog-secret",
				"datadog.otelCollector.enabled": "true",
				"otelAgentGateway.enabled":      "true",
			}
			for k, v := range tt.overrides {
				overrides[k] = v
			}

			manifest, err := common.RenderChart(t, common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/otel-configmap.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides:   overrides,
			})
			require.NoError(t, err, "couldn't render template")

			var cm corev1.ConfigMap
			common.Unmarshal(t, manifest, &cm)
			raw, ok := cm.Data["otel-config.yaml"]
			require.True(t, ok, "otel-config.yaml not found in ConfigMap")

			var config otelCollectorConfig
			require.NoError(t, yaml.Unmarshal([]byte(raw), &config))

			otherName := "otlphttp"
			if tt.expectedName == "otlphttp" {
				otherName = "otlp_http"
			}
			assert.Contains(t, config.Exporters, tt.expectedName)
			assert.NotContains(t, config.Exporters, otherName)

			require.NotEmpty(t, config.Service.Pipelines)
			for name, pipeline := range config.Service.Pipelines {
				assert.Contains(t, pipeline.Exporters, tt.expectedName, "pipeline %q", name)
				assert.NotContains(t, pipeline.Exporters, otherName, "pipeline %q", name)
			}
		})
	}
}
