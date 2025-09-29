// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package yamlmapper

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"gopkg.in/yaml.v3"
)

const (
	mappingPath   = "../../mapping_datadog_helm_to_datadogagent_crd_v2.yaml"
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
		assertions []func(t *testing.T, kubectlOptions *k8s.KubectlOptions, values string, namespace string)
	}{
		{
			name: "Minimal mapping",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../../../charts/datadog",
				Values:      []string{"./values/default-values.yaml"},
			},
			valuesPath: "./values/default-values.yaml",
			assertions: []func(t *testing.T, kubectlOptions *k8s.KubectlOptions, values string, namespace string){verifyAgentConf},
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

			cleanUpOperator := common.InstallChart(t, kubectlOptions, common.HelmCommand{
				ReleaseName: "datadog-operator",
				ChartPath:   "../../../../charts/datadog-operator",
			})
			defer cleanUpOperator()

			for _, assertion := range tt.assertions {
				assertion(t, kubectlOptions, tt.valuesPath, namespaceName)
			}

		})
	}
}

func verifyAgentConf(t *testing.T, kubectlOptions *k8s.KubectlOptions, valuesPath string, namespace string) {
	// Run mapper against values.yaml
	//os.Args = []string{
	//	"yaml-mapper",
	//	"-sourceFile=" + valuesPath,
	//	fmt.Sprintf("-mappingFile=%s", mappingPath),
	//	fmt.Sprintf("-destFile=%s", ddaDestPath),
	//	"-printOutput=true",
	//}

	destFile, err := os.CreateTemp(".", ddaDestPath)
	require.NoError(t, err)
	defer os.Remove(destFile.Name())

	MapYaml(mappingPath, valuesPath, destFile.Name(), "", namespace, false, false)

	outputBytes, err := os.ReadFile(destFile.Name())
	require.NoError(t, err)

	var ddaResult map[string]interface{}
	err = yaml.Unmarshal(outputBytes, &ddaResult)
	require.NoError(t, err)

	// Get agent conf from helm install
	helmAgentPods, err := k8s.ListPodsE(t, kubectlOptions, metav1.ListOptions{LabelSelector: "app.kubernetes.io/component=agent,app.kubernetes.io/managed-by=Helm"})
	require.NoError(t, err)
	assert.NotEmpty(t, helmAgentPods)
	helmAgentConf, err := k8s.RunKubectlAndGetOutputE(t, kubectlOptions, []string{"exec", helmAgentPods[0].Name, "--", "agent", "config"}...)
	require.NoError(t, err)
	helmAgentConf = normalizeAgentConf(helmAgentConf)

	// Apply DDA from mapper

	err = k8s.RunKubectlE(t, kubectlOptions, []string{"apply", "-f", destFile.Name()}...)
	require.NoError(t, err)
	defer k8s.RunKubectl(t, kubectlOptions, []string{"delete", "-f", destFile.Name()}...)

	time.Sleep(120 * time.Second)

	// Get agent conf from operator install
	operatorAgentPods, err := k8s.ListPodsE(t, kubectlOptions, metav1.ListOptions{})

	require.NoError(t, err)
	assert.NotEmpty(t, operatorAgentPods)
	operatorAgentConf, err := k8s.RunKubectlAndGetOutputE(t, kubectlOptions, []string{"exec", operatorAgentPods[0].Name, "--", "agent", "config"}...)
	require.NoError(t, err)
	operatorAgentConf = normalizeAgentConf(operatorAgentConf)

	// Check agent conf diff

	assert.Equal(t, helmAgentConf, operatorAgentConf)
	assert.EqualValues(t, helmAgentConf, operatorAgentConf)
}

// filterLogLines removes log lines that start with timestamps in the format "2006-01-02 15:04:05 UTC"
func normalizeAgentConf(input string) string {
	if input == "" {
		return input
	}

	var result strings.Builder
	lines := strings.Split(input, "\n")

	for _, line := range lines {
		// Skip lines that start with a timestamp
		if isTimestampLine(line) {
			continue
		}
		if result.Len() > 0 {
			result.WriteByte('\n')
		}
		result.WriteString(line)
	}

	return result.String()
}

// isTimestampLine checks if a line starts with a timestamp in the format "2006-01-02 15:04:05 UTC"
func isTimestampLine(line string) bool {
	if len(line) < 20 { // Minimum length for "2006-01-02 15:04:05"
		return false
	}

	// Check the prefix format: "2006-01-02 15:04:05 UTC"
	if len(line) >= 20 &&
		line[4] == '-' &&
		line[7] == '-' &&
		line[10] == ' ' &&
		line[13] == ':' &&
		line[16] == ':' {
		// Check if it's followed by " UTC"
		if len(line) > 20 && strings.HasPrefix(line[19:], " UTC") {
			return true
		}
	}

	return false
}
