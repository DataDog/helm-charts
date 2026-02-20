package common

import (
	"bytes"
	"context"
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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	// --wait and --timeout for clean state transitions between chart installs
	// 3m timeout to wait for readiness probes
	extraArgs := []string{"--wait", "--timeout", "3m"}
	if len(cmd.ExtraArgs) > 0 {
		extraArgs = append(extraArgs, cmd.ExtraArgs...)
	}

	helmOptions := &helm.Options{
		KubectlOptions: kubectlOptions,
		SetValues:      cmd.Overrides,
		ValuesFiles:    cmd.Values,
		ExtraArgs:      map[string][]string{"install": extraArgs},
	}
	if cmd.Logger != nil {
		helmOptions.Logger = cmd.Logger
	}

	releaseName := cmd.ReleaseName + "-" + strings.ToLower(random.UniqueId())
	t.Log("Installing release", releaseName)

	helm.Install(t, helmOptions, helmChartPath, releaseName)

	return func() {
		t.Log("Deleting release", releaseName)
		// use --wait to ensure resources are fully cleaned up before returning
		deleteOptions := &helm.Options{
			KubectlOptions: kubectlOptions,
			ExtraArgs:      map[string][]string{"delete": {"--wait", "--timeout", "2m"}},
		}
		if cmd.Logger != nil {
			deleteOptions.Logger = cmd.Logger
		}
		helm.Delete(t, deleteOptions, releaseName, true)
	}
}

// CreateSecretFromEnv creates a Kubernetes secret from environment variables
func CreateSecretFromEnv(t *testing.T, kubectlOptions *k8s.KubectlOptions, apiKeyEnv, appKeyEnv string) (cleanupFunc func()) {
	apiKey := os.Getenv(apiKeyEnv)
	appKey := os.Getenv(appKeyEnv)

	return CreateSecret(t, kubectlOptions, "datadog-secret", map[string]string{
		"api-key": apiKey,
		"app-key": appKey,
	})
}

// CreateSecret creates a Kubernetes secret with the given name and key-value pairs
func CreateSecret(t *testing.T, kubectlOptions *k8s.KubectlOptions, secretName string, data map[string]string) (cleanupFunc func()) {
	t.Logf("Creating secret %s", secretName)

	clientset, err := k8s.GetKubernetesClientFromOptionsE(t, kubectlOptions)
	require.NoError(t, err, "Failed to get Kubernetes client")

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: kubectlOptions.Namespace,
		},
		StringData: data,
	}

	_, err = clientset.CoreV1().Secrets(kubectlOptions.Namespace).Create(context.Background(), secret, metav1.CreateOptions{})
	require.NoError(t, err, "Failed to create secret %s", secretName)

	return func() {
		t.Logf("Deleting secret %s", secretName)
		_ = clientset.CoreV1().Secrets(kubectlOptions.Namespace).Delete(context.Background(), secretName, metav1.DeleteOptions{})
	}
}

// CreateConfigMap creates a Kubernetes configmap with the given name and key-value pairs
func CreateConfigMap(t *testing.T, kubectlOptions *k8s.KubectlOptions, configMapName string, data map[string]string) (cleanupFunc func()) {
	t.Logf("Creating configmap %s", configMapName)

	clientset, err := k8s.GetKubernetesClientFromOptionsE(t, kubectlOptions)
	require.NoError(t, err, "Failed to get Kubernetes client")

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: kubectlOptions.Namespace,
		},
		Data: data,
	}

	_, err = clientset.CoreV1().ConfigMaps(kubectlOptions.Namespace).Create(context.Background(), configMap, metav1.CreateOptions{})
	require.NoError(t, err, "Failed to create configmap %s", configMapName)

	return func() {
		t.Logf("Deleting configmap %s", configMapName)
		_ = clientset.CoreV1().ConfigMaps(kubectlOptions.Namespace).Delete(context.Background(), configMapName, metav1.DeleteOptions{})
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
