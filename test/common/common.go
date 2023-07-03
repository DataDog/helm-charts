package common

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
)

type HelmCommand struct {
	ReleaseName string
	ChartPath   string
	ShowOnly    []string          // helm template `-s, --show-only` flag
	Values      []string          // helm template `-f, --values` flag
	Overrides   map[string]string // helm template `--set` flag
}

func Unmarshal[T any](t *testing.T, manifest string, destObj *T) {
	helm.UnmarshalK8SYaml(t, manifest, destObj)
}

func RenderChart(t *testing.T, cmd HelmCommand) (string, error) {
	chartPath, err := filepath.Abs(cmd.ChartPath)
	require.NoError(t, err, "can't resolve absolute path", "path", cmd.ChartPath)
	require.NoError(t, err)

	kubectlOptions := k8s.NewKubectlOptions("", "", "datadog-agent")

	options := &helm.Options{
		KubectlOptions: kubectlOptions,
		SetValues:      cmd.Overrides,
		ValuesFiles:    cmd.Values,
	}

	output, err := helm.RenderTemplateE(t, options, chartPath, cmd.ReleaseName, cmd.ShowOnly)

	return output, err
}

func InstallChart(t *testing.T, kubectlOptions *k8s.KubectlOptions, cmd HelmCommand) (cleanupFunc func()) {
	helmChartPath, err := filepath.Abs(cmd.ChartPath)
	require.NoError(t, err)

	helmOptions := &helm.Options{
		KubectlOptions: kubectlOptions,
		SetValues:      cmd.Overrides,
		ValuesFiles:    cmd.Values,
	}

	releaseName := cmd.ReleaseName + "-" + strings.ToLower(random.UniqueId())
	t.Log("Installing release", releaseName)

	helm.Install(t, helmOptions, helmChartPath, releaseName)

	return func() {
		t.Log("Deleting release", releaseName)
		helm.Delete(t, helmOptions, releaseName, true)
	}
}

func CreateSecretFromEnv(t *testing.T, kubectlOptions *k8s.KubectlOptions, apiKeyEnv, appKeyEnv string) (cleanupFunc func()) {
	apiKey := os.Getenv(apiKeyEnv)
	appKey := os.Getenv(appKeyEnv)

	// Setup Datadog Agent
	t.Log("Creating secret")
	k8s.RunKubectl(t, kubectlOptions, "create", "secret", "generic", "datadog-secret",
		"--from-literal",
		"api-key="+apiKey,
		"--from-literal",
		"app-key="+appKey)
	return func() {
		t.Log("Deleting secret")
		k8s.RunKubectl(t, kubectlOptions, "delete", "secret", "datadog-secret")
	}
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
