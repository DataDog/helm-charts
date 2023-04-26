package common

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/require"
)

type HelmCommand struct {
	ReleaseName string
	ChartPath   string
	ShowOnly    []string          // helm template `-s, --show-only` flag
	Values      string            // helm template `-f, --values` flag
	Overrides   map[string]string // helm template `--set` flag
}

func Unmarshal[T any](t *testing.T, manifest string, destObj *T) {
	helm.UnmarshalK8SYaml(t, manifest, destObj)
}

func RenderChart(t *testing.T, cmd HelmCommand) (string, error) {
	chartPath, err := filepath.Abs(cmd.ChartPath)
	require.NoError(t, err, "can't resolve absolute path", "path", cmd.ChartPath)
	require.NoError(t, err)

	options := &helm.Options{
		SetValues: cmd.Overrides,
	}

	output, err := helm.RenderTemplateE(t, options, chartPath, cmd.ReleaseName, cmd.ShowOnly,
		"-f", cmd.Values,
		"-n", "datadog-agent",
		// "--debug",
	)

	return output, err
}

func LoadFromFile[T any](t *testing.T, filepath string, destObj *T) string {
	fileContent, err := os.ReadFile(filepath)
	require.NoError(t, err, "can't load manifest from file", "path", filepath)

	content := string(fileContent)
	helm.UnmarshalK8SYaml(t, content, destObj)
	return content
}

func WriteToFile(t *testing.T, filepath, content string) {
	err := os.WriteFile(filepath, []byte(content), 0644)
	require.NoError(t, err, "can't update manifest", "path", filepath)
}
