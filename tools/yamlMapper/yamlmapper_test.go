package yamlMapper_test

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/DataDog/helm-charts/tools/yamlMapper"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"gopkg.in/yaml.v3"
)

const (
	mappingPath   = "tools/yamlMapper/mapping_datadog_helm_to_datadogagent_crd_v2.yaml"
	ddaDestPath   = "tempDDADest.yaml"
	apiKeyEnv     = "API_KEY"
	appKeyEnv     = "APP_KEY"
	k8sVersionEnv = "K8S_VERSION"
)

func Test(t *testing.T) {
	// Prerequisites
	context := common.CurrentContext(t)
	t.Log("Checking current context:", context)
	if strings.Contains(strings.ToLower(context), "staging") ||
		strings.Contains(strings.ToLower(context), "prod") {
		t.Fatal("Make sure context is pointing to local cluster")
	}

	require.NotEmpty(t, os.Getenv(apiKeyEnv), "API key can't be empty")
	require.NotEmpty(t, os.Getenv(appKeyEnv), "APP key can't be empty")

	tests := []struct {
		name       string
		command    common.HelmCommand
		valuesPath string
		assertions []func(t *testing.T, kubectlOptions *k8s.KubectlOptions, values string)
	}{
		{
			name: "Minimal mapping",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				Values:      []string{"./values/default-values.yaml"},
			},
			valuesPath: "./values/default-values.yaml",
			assertions: []func(t *testing.T, kubectlOptions *k8s.KubectlOptions, values string){verifyAgentConf},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			namespaceName := fmt.Sprintf("datadog-agent-%s", strings.ToLower(random.UniqueId()))
			kubectlOptions := k8s.NewKubectlOptions("", "", namespaceName)
			k8s.CreateNamespace(t, kubectlOptions, namespaceName)
			defer k8s.DeleteNamespace(t, kubectlOptions, namespaceName)

			cleanupSecrets := common.CreateSecretFromEnv(t, kubectlOptions, apiKeyEnv, appKeyEnv)
			defer cleanupSecrets()

			//	Helm install
			cleanUpDatadog := common.InstallChart(t, kubectlOptions, tt.command)
			defer cleanUpDatadog()
			time.Sleep(120 * time.Second)
			for _, assertion := range tt.assertions {
				assertion(t, kubectlOptions, tt.valuesPath)
			}

			cleanUpOperator := common.InstallChart(t, kubectlOptions, common.HelmCommand{
				ReleaseName: "datadog-operator",
				ChartPath:   "../../charts/datadog-operator",
			})
			defer cleanUpOperator()
		})
	}
}

func verifyAgentConf(t *testing.T, kubectlOptions *k8s.KubectlOptions, valuesPath string) {
	// Run mapper against values.yaml
	os.Args = []string{
		"yaml-mapper",
		"-sourceFile=" + valuesPath,
		fmt.Sprintf("-mappingFile=%s", mappingPath),
		fmt.Sprintf("-destFile=%s", ddaDestPath),
		"-printOutput=false",
	}

	destFile, err := os.CreateTemp(".", "dest-*.yaml")
	require.NoError(t, err)
	defer os.Remove(destFile.Name())

	yamlMapper.YamlMapper()

	outputBytes, err := os.ReadFile(destFile.Name())
	require.NoError(t, err)

	var ddaResult map[string]interface{}
	err = yaml.Unmarshal(outputBytes, &ddaResult)
	require.NoError(t, err)

	// Get agent conf from helm install

	helmAgentPods, err := k8s.ListPodsE(t, kubectlOptions, metav1.ListOptions{LabelSelector: "app.kubernetes.io/component=agent,app.kubernetes.io/managed-by=Helm"})
	require.NoError(t, err)
	assert.NotEmpty(t, helmAgentPods)
	helmAgentConf, err := k8s.RunKubectlAndGetOutputE(t, kubectlOptions, []string{"exec", "-it", helmAgentPods[1].Name, "--", "agent", "config"}...)
	require.NoError(t, err)

	// Apply DDA from mapper

	err = k8s.RunKubectlE(t, kubectlOptions, []string{"apply", "-f", destFile.Name()}...)
	require.NoError(t, err)

	// Get agent conf from operator install
	operatorAgentPods, err := k8s.ListPodsE(t, kubectlOptions, metav1.ListOptions{LabelSelector: "app.kubernetes.io/component=agent,app.kubernetes.io/managed-by=datadog-operator"})
	require.NoError(t, err)
	assert.NotEmpty(t, operatorAgentPods)
	operatorAgentConf, err := k8s.RunKubectlAndGetOutputE(t, kubectlOptions, []string{"exec", "-it", operatorAgentPods[1].Name, "--", "agent", "config"}...)
	require.NoError(t, err)

	// Check agent conf diff

	assert.Equal(t, helmAgentConf, operatorAgentConf)
}
