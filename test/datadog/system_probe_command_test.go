package datadog

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"

	"github.com/DataDog/helm-charts/test/common"
)

func Test_systemProbeCommand(t *testing.T) {
	tests := []struct {
		name                   string
		overrides              map[string]string
		expectSystemProbe      bool
		expectCommandContains  string
		expectCommandEquals    []string
	}{
		{
			name: "SPL with explicit discovery.enabled -- fallback is system-probe",
			overrides: map[string]string{
				"datadog.apiKeyExistingSecret":        "datadog-secret",
				"datadog.appKeyExistingSecret":        "datadog-secret",
				"datadog.discovery.enabled":           "true",
				"datadog.discovery.useSystemProbeLite": "true",
			},
			expectSystemProbe:     true,
			expectCommandContains: "|| system-probe --config=/etc/datadog-agent/system-probe.yaml",
		},
		{
			name: "SPL with enabledByDefault -- fallback is sleep infinity",
			overrides: map[string]string{
				"datadog.apiKeyExistingSecret":            "datadog-secret",
				"datadog.appKeyExistingSecret":            "datadog-secret",
				"datadog.discovery.enabledByDefault":      "true",
				"datadog.discovery.useSystemProbeLite":    "true",
			},
			expectSystemProbe:     true,
			expectCommandContains: "|| sleep infinity",
		},
		{
			name: "SPL with explicit enable plus other features -- regular system-probe",
			overrides: map[string]string{
				"datadog.apiKeyExistingSecret":           "datadog-secret",
				"datadog.appKeyExistingSecret":           "datadog-secret",
				"datadog.discovery.enabled":              "true",
				"datadog.discovery.useSystemProbeLite":   "true",
				"datadog.networkMonitoring.enabled":      "true",
			},
			expectSystemProbe:  true,
			expectCommandEquals: []string{"system-probe", "--config=/etc/datadog-agent/system-probe.yaml"},
		},
		{
			name: "no discovery -- no system-probe container",
			overrides: map[string]string{
				"datadog.apiKeyExistingSecret": "datadog-secret",
				"datadog.appKeyExistingSecret": "datadog-secret",
			},
			expectSystemProbe: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest, err := common.RenderChart(t, common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides:   tt.overrides,
			})
			require.NoError(t, err, "couldn't render template")

			var ds appsv1.DaemonSet
			common.Unmarshal(t, manifest, &ds)

			spContainer, found := getContainer(t, ds.Spec.Template.Spec.Containers, "system-probe")

			if !tt.expectSystemProbe {
				assert.False(t, found, "system-probe container should not exist")
				return
			}

			require.True(t, found, "system-probe container should exist")

			if tt.expectCommandContains != "" {
				require.Len(t, spContainer.Command, 3, "expected shell command with 3 elements")
				assert.Equal(t, "/bin/sh", spContainer.Command[0])
				assert.Equal(t, "-c", spContainer.Command[1])
				assert.True(t, strings.Contains(spContainer.Command[2], "system-probe-lite"), "command should contain system-probe-lite")
				assert.True(t, strings.Contains(spContainer.Command[2], tt.expectCommandContains),
					"command %q should contain %q", spContainer.Command[2], tt.expectCommandContains)
			}

			if tt.expectCommandEquals != nil {
				assert.Equal(t, tt.expectCommandEquals, spContainer.Command)
			}
		})
	}
}
