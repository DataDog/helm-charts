package datadog

import (
	"strings"
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKSMPodCollectionOnNodeVersionGate(t *testing.T) {
	tests := []struct {
		name      string
		overrides map[string]string
		enabled   bool
	}{
		{
			name: "supported at minimum versions",
			overrides: map[string]string{
				"agents.image.tag":       "7.82.0",
				"clusterAgent.image.tag": "7.82.0",
			},
			enabled: true,
		},
		{
			name: "unsupported below minimum node agent version",
			overrides: map[string]string{
				"agents.image.tag":       "7.81.9",
				"clusterAgent.image.tag": "7.82.0",
			},
		},
		{
			name: "unsupported below minimum cluster agent version",
			overrides: map[string]string{
				"agents.image.tag":       "7.82.0",
				"clusterAgent.image.tag": "7.81.9",
			},
		},
		{
			name: "unused old cluster checks runner does not affect support",
			overrides: map[string]string{
				"agents.image.tag":              "7.82.0",
				"clusterAgent.image.tag":        "7.82.0",
				"clusterChecksRunner.image.tag": "7.81.9",
			},
			enabled: true,
		},
		{
			name: "used cluster checks runner at minimum version is supported",
			overrides: map[string]string{
				"agents.image.tag":       "7.82.0",
				"clusterAgent.image.tag": "7.82.0",
				"datadog.kubeStateMetricsCore.useClusterCheckRunners": "true",
				"clusterChecksRunner.image.tag":                       "7.82.0",
			},
			enabled: true,
		},
		{
			name: "used cluster checks runner prerelease is supported",
			overrides: map[string]string{
				"agents.image.tag":       "7.82.0",
				"clusterAgent.image.tag": "7.82.0",
				"datadog.kubeStateMetricsCore.useClusterCheckRunners": "true",
				"clusterChecksRunner.image.tag":                       "7.82.0-rc.1",
			},
			enabled: true,
		},
		{
			name: "used old cluster checks runner is unsupported",
			overrides: map[string]string{
				"agents.image.tag":       "7.82.0",
				"clusterAgent.image.tag": "7.82.0",
				"datadog.kubeStateMetricsCore.useClusterCheckRunners": "true",
				"clusterChecksRunner.image.tag":                       "7.81.9",
			},
		},
		{
			name: "floating cluster checks runner latest follows cluster agent version policy",
			overrides: map[string]string{
				"agents.image.tag":       "7.82.0",
				"clusterAgent.image.tag": "7.82.0",
				"datadog.kubeStateMetricsCore.useClusterCheckRunners": "true",
				"clusterChecksRunner.image.tag":                       "latest",
			},
			enabled: true,
		},
		{
			name: "custom cluster checks runner tag can skip version check",
			overrides: map[string]string{
				"agents.image.tag":       "7.82.0",
				"clusterAgent.image.tag": "7.82.0",
				"datadog.kubeStateMetricsCore.useClusterCheckRunners": "true",
				"clusterChecksRunner.image.tag":                       "custom",
				"clusterChecksRunner.image.doNotCheckTag":             "true",
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
				Overrides:   ksmPodCollectionOverrides(tt.overrides),
			})
			require.NoError(t, err, "couldn't render chart")

			hasNodeConfig := strings.Contains(manifest, "pod_collection_mode: node_kubelet")
			assert.Equal(t, tt.enabled, hasNodeConfig, "unexpected KSM node pod collection state")
		})
	}
}

func ksmPodCollectionOverrides(overrides map[string]string) map[string]string {
	merged := map[string]string{
		"datadog.apiKeyExistingSecret":                   "datadog-secret",
		"datadog.appKeyExistingSecret":                   "datadog-secret",
		"datadog.kubeStateMetricsCore.enabled":           "true",
		"datadog.kubeStateMetricsCore.podCollectionMode": "node_kubelet",
	}
	for key, value := range overrides {
		merged[key] = value
	}
	return merged
}
