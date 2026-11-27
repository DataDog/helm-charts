package datadog

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/DataDog/helm-charts/test/common"
)

// otelCollectorConfigWithReceiver returns an OTel Collector config that declares
// the given receiver and uses it in a logs pipeline.
func otelCollectorConfigWithReceiver(name string) string {
	return fmt.Sprintf(`receivers:
  %s:
    include: [/var/log/pods/*/*/*.log]
exporters:
  debug:
service:
  pipelines:
    logs:
      receivers: [%s]
      exporters: [debug]
`, name, name)
}

func Test_otelAgentFileLogReceiverMountsLogs(t *testing.T) {
	logMounts := []string{"logpodpath", "logscontainerspath"}

	tests := []struct {
		receiver       string
		logsEnabled    string
		expectedMounts bool
	}{
		{receiver: "file_log", logsEnabled: "true", expectedMounts: true},
		{receiver: "file_log/custom", logsEnabled: "true", expectedMounts: true},
		{receiver: "filelog", logsEnabled: "true", expectedMounts: true},
		{receiver: "filelog/custom", logsEnabled: "true", expectedMounts: true},
		{receiver: "file_log", logsEnabled: "false", expectedMounts: false},
		{receiver: "otlp", logsEnabled: "true", expectedMounts: false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s/logs_enabled_%s", tt.receiver, tt.logsEnabled), func(t *testing.T) {
			manifest, err := common.RenderChart(t, common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret":       "datadog-secret",
					"datadog.appKeyExistingSecret":       "datadog-secret",
					"datadog.otelCollector.enabled":      "true",
					"datadog.otelCollector.logs.enabled": tt.logsEnabled,
				},
				// Use --set-json so the multi-line config isn't parsed by --set.
				OverridesJson: map[string]string{
					"datadog.otelCollector.config": fmt.Sprintf("%q", otelCollectorConfigWithReceiver(tt.receiver)),
				},
			})
			require.NoError(t, err, "couldn't render template")

			var ds appsv1.DaemonSet
			common.Unmarshal(t, manifest, &ds)
			otelAgent := findContainer(t, ds.Spec.Template.Spec.Containers, "otel-agent")

			mounts := map[string]bool{}
			for _, m := range otelAgent.VolumeMounts {
				mounts[m.Name] = true
			}
			for _, name := range logMounts {
				assert.Equal(t, tt.expectedMounts, mounts[name], "unexpected presence of volume mount %q", name)
			}
		})
	}
}

func findContainer(t *testing.T, containers []corev1.Container, name string) corev1.Container {
	for _, c := range containers {
		if c.Name == name {
			return c
		}
	}
	require.FailNow(t, "container not found", "name", name)
	return corev1.Container{}
}
