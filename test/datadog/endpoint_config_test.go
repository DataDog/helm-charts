package datadog

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
)

// endpointConfigCmd returns a HelmCommand that renders only the endpoint-config ConfigMap.
func endpointConfigCmd(releaseName string, overrides map[string]string) common.HelmCommand {
	merged := map[string]string{
		"datadog.kubelet.tlsVerify": "false",
		"datadog.operator.enabled":  "false",
	}
	for k, v := range overrides {
		merged[k] = v
	}
	return common.HelmCommand{
		ReleaseName: releaseName,
		ChartPath:   "../../charts/datadog",
		ShowOnly:    []string{"templates/datadog-endpoint-configmap.yaml"},
		Overrides:   merged,
	}
}

func Test_endpoint_config_standard(t *testing.T) {
	tests := []struct {
		name    string
		command common.HelmCommand
		verify  func(t *testing.T, cm v1.ConfigMap)
	}{
		{
			name:    "default release name",
			command: endpointConfigCmd("datadog", map[string]string{"datadog.apiKeyExistingSecret": "datadog-secret"}),
			verify: func(t *testing.T, cm v1.ConfigMap) {
				assert.Equal(t, "datadog-endpoint-config", cm.Name)
			},
		},
		{
			name:    "custom release name",
			command: endpointConfigCmd("my-release", map[string]string{"datadog.apiKeyExistingSecret": "datadog-secret"}),
			verify: func(t *testing.T, cm v1.ConfigMap) {
				assert.Equal(t, "my-release-endpoint-config", cm.Name)
			},
		},
		{
			name:    "discovery label present",
			command: endpointConfigCmd("datadog", map[string]string{"datadog.apiKeyExistingSecret": "datadog-secret"}),
			verify: func(t *testing.T, cm v1.ConfigMap) {
				assert.Equal(t, "endpoint-config", cm.Labels["datadoghq.com/component"])
			},
		},
		{
			name:    "instance label matches release name",
			command: endpointConfigCmd("my-release", map[string]string{"datadog.apiKeyExistingSecret": "datadog-secret"}),
			verify: func(t *testing.T, cm v1.ConfigMap) {
				assert.Equal(t, "my-release", cm.Labels["app.kubernetes.io/instance"])
			},
		},
		{
			name:    "api key secret name from apiKeyExistingSecret",
			command: endpointConfigCmd("datadog", map[string]string{"datadog.apiKeyExistingSecret": "my-api-secret"}),
			verify: func(t *testing.T, cm v1.ConfigMap) {
				assert.Equal(t, "my-api-secret", cm.Data["api-key-secret-name"])
			},
		},
		{
			name:    "multi-release uniqueness",
			command: endpointConfigCmd("dd1", map[string]string{"datadog.apiKeyExistingSecret": "datadog-secret"}),
			verify: func(t *testing.T, cm v1.ConfigMap) {
				assert.Equal(t, "dd1-endpoint-config", cm.Name)

				// Render a second release and verify names differ
				manifest2, err := common.RenderChart(t, endpointConfigCmd("dd2", map[string]string{"datadog.apiKeyExistingSecret": "datadog-secret"}))
				require.NoError(t, err)
				var cm2 v1.ConfigMap
				common.Unmarshal(t, manifest2, &cm2)
				assert.Equal(t, "dd2-endpoint-config", cm2.Name)
				assert.NotEqual(t, cm.Name, cm2.Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest, err := common.RenderChart(t, tt.command)
			require.NoError(t, err)
			var cm v1.ConfigMap
			common.Unmarshal(t, manifest, &cm)
			tt.verify(t, cm)
		})
	}
}

func Test_endpoint_config_aliased(t *testing.T) {
	// Resolve the absolute path to the datadog chart for the file:// dependency URL
	datadogChartPath, err := filepath.Abs("../../charts/datadog")
	require.NoError(t, err)

	// Read chart version from Chart.yaml to use in the wrapper
	chartYamlBytes, err := os.ReadFile(filepath.Join(datadogChartPath, "Chart.yaml"))
	require.NoError(t, err)
	version := extractChartVersion(t, string(chartYamlBytes))

	// Create a temporary wrapper chart with two aliased dependencies
	tmpDir := t.TempDir()
	wrapperChart := `apiVersion: v2
name: test-wrapper
version: 0.0.1
dependencies:
  - name: datadog
    version: "` + version + `"
    repository: "file://` + datadogChartPath + `"
    alias: datadog-linux
  - name: datadog
    version: "` + version + `"
    repository: "file://` + datadogChartPath + `"
    alias: datadog-windows
`
	err = os.WriteFile(filepath.Join(tmpDir, "Chart.yaml"), []byte(wrapperChart), 0644)
	require.NoError(t, err)

	// Run helm dependency update on the wrapper chart
	cmd := exec.Command("helm", "dependency", "update", tmpDir)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "helm dependency update failed: %s", string(output))

	releaseName := "myrel"
	baseOverrides := map[string]string{
		"datadog-linux.datadog.apiKeyExistingSecret":   "linux-api-secret",
		"datadog-linux.datadog.kubelet.tlsVerify":      "false",
		"datadog-linux.datadog.operator.enabled":       "false",
		"datadog-windows.datadog.apiKeyExistingSecret": "windows-api-secret",
		"datadog-windows.datadog.kubelet.tlsVerify":    "false",
		"datadog-windows.datadog.operator.enabled":     "false",
	}

	// Render the linux alias ConfigMap
	linuxManifest, err := common.RenderChart(t, common.HelmCommand{
		ReleaseName: releaseName,
		ChartPath:   tmpDir,
		ShowOnly:    []string{"charts/datadog-linux/templates/datadog-endpoint-configmap.yaml"},
		Overrides:   baseOverrides,
	})
	require.NoError(t, err, "failed to render linux alias ConfigMap")

	var linuxCM v1.ConfigMap
	common.Unmarshal(t, linuxManifest, &linuxCM)

	// Render the windows alias ConfigMap
	windowsManifest, err := common.RenderChart(t, common.HelmCommand{
		ReleaseName: releaseName,
		ChartPath:   tmpDir,
		ShowOnly:    []string{"charts/datadog-windows/templates/datadog-endpoint-configmap.yaml"},
		Overrides:   baseOverrides,
	})
	require.NoError(t, err, "failed to render windows alias ConfigMap")

	var windowsCM v1.ConfigMap
	common.Unmarshal(t, windowsManifest, &windowsCM)

	t.Run("aliased name linux", func(t *testing.T) {
		assert.Equal(t, "datadog-linux-myrel-endpoint-config", linuxCM.Name)
	})

	t.Run("aliased name windows", func(t *testing.T) {
		assert.Equal(t, "datadog-windows-myrel-endpoint-config", windowsCM.Name)
	})

	t.Run("uniqueness", func(t *testing.T) {
		assert.NotEqual(t, linuxCM.Name, windowsCM.Name)
	})

	t.Run("same instance label", func(t *testing.T) {
		assert.Equal(t, releaseName, linuxCM.Labels["app.kubernetes.io/instance"])
		assert.Equal(t, releaseName, windowsCM.Labels["app.kubernetes.io/instance"])
	})

	t.Run("discovery label", func(t *testing.T) {
		assert.Equal(t, "endpoint-config", linuxCM.Labels["datadoghq.com/component"])
		assert.Equal(t, "endpoint-config", windowsCM.Labels["datadoghq.com/component"])
	})

	t.Run("data isolation", func(t *testing.T) {
		assert.Equal(t, "linux-api-secret", linuxCM.Data["api-key-secret-name"])
		assert.Equal(t, "windows-api-secret", windowsCM.Data["api-key-secret-name"])
		assert.NotEqual(t, linuxCM.Data["api-key-secret-name"], windowsCM.Data["api-key-secret-name"])
	})
}

// extractChartVersion parses the top-level version field from Chart.yaml content.
// It skips indented lines to avoid matching dependency version fields.
func extractChartVersion(t *testing.T, content string) string {
	t.Helper()
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "version:") {
			return strings.Trim(line[len("version:"):], " \t\"'")
		}
	}
	t.Fatal("could not find version in Chart.yaml")
	return ""
}
