package common

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	yaml2 "k8s.io/apimachinery/pkg/util/yaml"
)

type HelmCommand struct {
	ReleaseName   string
	Namespace     string
	ChartPath     string
	ShowOnly      []string          // helm template `-s, --show-only` flag
	Values        []string          // helm template `-f, --values` flag
	Overrides     map[string]string // helm template `--set` flag
	OverridesJson map[string]string // helm template `--set-json` flag
	Logger        *logger.Logger    // logger to use for helm output. Set to logger.Discard by default.
	ExtraArgs     []string
}

func Unmarshal[T any](t *testing.T, manifest string, destObj *T) {
	helm.UnmarshalK8SYaml(t, manifest, destObj)
}

func RenderChart(t *testing.T, cmd HelmCommand) (string, error) {
	chartPath, err := filepath.Abs(cmd.ChartPath)
	require.NoError(t, err, "can't resolve absolute path", "path", cmd.ChartPath)
	require.NoError(t, err)

	namespace := "datadog-agent"
	if cmd.Namespace != "" {
		namespace = cmd.Namespace
	}

	kubectlOptions := k8s.NewKubectlOptions("", "", namespace)

	options := &helm.Options{
		KubectlOptions: kubectlOptions,
		SetValues:      cmd.Overrides,
		SetJsonValues:  cmd.OverridesJson,
		ValuesFiles:    cmd.Values,
	}

	if cmd.Logger == nil {
		options.Logger = logger.Discard
	}

	output, err := helm.RenderTemplateE(t, options, chartPath, cmd.ReleaseName, cmd.ShowOnly, cmd.ExtraArgs...)

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
	kubectlOptions.Logger = logger.Discard
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

func ReadFile(t *testing.T, filepath string) string {
	fileContent, err := os.ReadFile(filepath)
	require.NoError(t, err, "can't load manifest from file", "path", filepath)
	return string(fileContent)
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

func GetVolumeNames(ds appsv1.DaemonSet) []string {
	volumeNames := []string{}
	for _, volume := range ds.Spec.Template.Spec.Volumes {
		volumeNames = append(volumeNames, volume.Name)
	}
	return volumeNames
}

func Contains(str string, list []string) bool {
	for _, s := range list {
		if s == str {
			return true
		}
	}
	return false
}

// FilterYamlKeysMultiManifest Takes multi-document YAML and filter out keys from each document.
func FilterYamlKeysMultiManifest(manifest string, filterKeys map[string]interface{}) (string, error) {
	reader := strings.NewReader(manifest)
	decoder := yaml2.NewYAMLOrJSONDecoder(reader, 4096)
	builder := strings.Builder{}
	for {
		var obj map[string]interface{}
		// We read the next YAML document from the input stream until we reach EOF.
		// This is needed if Helm rendering contains multiple resource manifests.
		err := decoder.Decode(&obj)
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("couldn't decode manifest for filtering dynamic keys: %s", err)
		}

		if obj["kind"] == "CustomResourceDefinition" {
			continue
		}

		filterKeysRecursive(&obj, filterKeys)

		var buf bytes.Buffer
		enc := yaml.NewEncoder(&buf)
		enc.SetIndent(2) // Adjust indentation (default is 4)
		err = enc.Encode(obj)
		if err != nil {
			return "", fmt.Errorf("couldn't encode manifest after filtering: %s", err)
		}

		err = enc.Close()
		if err != nil {
			return "", fmt.Errorf("couldn't close encoder: %s", err)
		}

		output := buf.String()
		_, err = builder.WriteString(output)
		if err != nil {
			return "", fmt.Errorf("couldn't write manifest string in builder: %s", err)
		}
		builder.WriteString("---\n")
	}
	return builder.String(), nil
}

func filterKeysRecursive(yamlMap *map[string]interface{}, keys map[string]interface{}) {
	for yamlKey := range *yamlMap {
		if _, found := keys[yamlKey]; found {
			// fmt.Println("deleting key", yamlKey)
			delete(*yamlMap, yamlKey)
		} else if nested, ok := (*yamlMap)[yamlKey].(map[string]interface{}); ok {
			filterKeysRecursive(&nested, keys)
		}
	}
}

func CurrentContext(t *testing.T) string {
	val, err := k8s.RunKubectlAndGetOutputE(t, k8s.NewKubectlOptions("", "", ""), "config", "current-context")
	require.Nil(t, err)
	return val
}

func GetFullValues(t *testing.T, cmd HelmCommand, namespace string) string {
	tempFile, err := os.CreateTemp("", "helm-values-*.yaml")
	require.NoError(t, err, "failed to create temporary file")
	defer tempFile.Close()

	tempFilePath := tempFile.Name()

	// Default namespace is datadog-agent
	if namespace == "" {
		namespace = "datadog-agent"
	}
	kubectlOptions := k8s.NewKubectlOptions("", "", namespace)
	helmOptions := &helm.Options{
		KubectlOptions: kubectlOptions,
	}

	output, err := helm.RunHelmCommandAndGetOutputE(t, helmOptions, "get", "values", cmd.ReleaseName, "--all")
	require.NoError(t, err, "failed to get helm values")

	err = os.WriteFile(tempFilePath, []byte(output), 0644)
	require.NoError(t, err, "failed to write values to temporary file")

	return tempFilePath
}
