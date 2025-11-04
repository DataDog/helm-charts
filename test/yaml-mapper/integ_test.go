// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

//go:build integration

package yaml_mapper

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DataDog/datadog-operator/cmd/yaml-mapper/mapper"
	"github.com/DataDog/helm-charts/test/common"
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"gopkg.in/yaml.v3"
)

const (
	mappingPath   = "../../tools/yaml-mapper/mapping_datadog_helm_to_datadogagent_crd.yaml"
	ddaDestPath   = "tempDDADest.yaml"
	apiKeyEnv     = "API_KEY"
	appKeyEnv     = "APP_KEY"
	k8sVersionEnv = "K8S_VERSION"
)

func Test(t *testing.T) {
	t.Skip()
	// Prerequisites
	validateEnv(t)

	tests := []struct {
		name               string
		agentInstallCmd    common.HelmCommand
		operatorInstallCmd common.HelmCommand
		valuesPath         string
		assertions         []func(t *testing.T, kubectlOptions *k8s.KubectlOptions, values string, namespace string) (ddaName string)
	}{
		{
			name: "Minimal mapping",
			agentInstallCmd: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				Values:      []string{"./values/default-values.yaml"},
				Overrides:   map[string]string{"datadog.operator.enabled": "false"},
			},
			operatorInstallCmd: common.HelmCommand{
				ReleaseName: "operator",
				ChartPath:   "../../charts/datadog-operator",
			},
			valuesPath: "./values/default-values.yaml",
			assertions: []func(t *testing.T, kubectlOptions *k8s.KubectlOptions, values string, namespace string) (ddaName string){verifyAgentConf},
		},
		{
			name: "Agent confd configmap",
			agentInstallCmd: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				Values:      []string{"./values/confd-values.yaml"},
				Overrides:   map[string]string{"datadog.operator.enabled": "false"},
			},
			operatorInstallCmd: common.HelmCommand{
				ReleaseName: "operator",
				ChartPath:   "../../charts/datadog-operator",
			},
			valuesPath: "./values/confd-values.yaml",
			assertions: []func(t *testing.T, kubectlOptions *k8s.KubectlOptions, values string, namespace string) (ddaName string){verifyConfigData},
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

			//	Install Datadog chart
			cleanUpDatadog := common.InstallChart(t, kubectlOptions, tt.agentInstallCmd)
			defer cleanUpDatadog()

			datadogReleaseName := getHelmReleaseName(t, kubectlOptions, namespaceName, tt.agentInstallCmd.ReleaseName)

			valuesCmd := common.HelmCommand{
				ReleaseName: datadogReleaseName,
			}
			actualValuesFilePath := common.GetFullValues(t, valuesCmd, namespaceName)
			defer os.Remove(actualValuesFilePath)

			t.Log("GetFullValues created temp file:", actualValuesFilePath)

			// Install Operator chart
			cleanUpOperator := common.InstallChart(t, kubectlOptions, tt.operatorInstallCmd)
			defer cleanUpOperator()

			var ddasToCleanup []string
			for _, assertion := range tt.assertions {
				ddaName := assertion(t, kubectlOptions, actualValuesFilePath, namespaceName)
				if ddaName != "" {
					ddasToCleanup = append(ddasToCleanup, ddaName)
				}
			}
			if len(ddasToCleanup) > 0 {
				for _, ddaName := range ddasToCleanup {
					k8s.RunKubectl(t, kubectlOptions, []string{"delete", "-f", ddaName}...)
					os.Remove(ddaName)
				}
			}
		})
	}
}

func TestValues(t *testing.T) {
	validateEnv(t)

	currDir, err := os.Getwd()
	require.NoError(t, err)
	valuesDirPath, err := filepath.Abs(filepath.Join(currDir, "../../charts/datadog/ci"))
	require.NoError(t, err)
	paths, err := os.ReadDir(valuesDirPath)
	require.NoError(t, err)

	for _, maybeFile := range paths {
		var valPath string
		var valName string
		if !maybeFile.IsDir() {
			valPath = filepath.Join(valuesDirPath, maybeFile.Name())
			fileInfo, err := os.Stat(valPath)
			require.NoError(t, err)
			require.NotNil(t, fileInfo)
			valName = strings.TrimSuffix(fileInfo.Name(), ".yaml")
			if len(valName) > 50 {
				valName = valName[:50]
			}
		}

		t.Run(fmt.Sprintf("Verify Agent conf:%s", valName), func(t *testing.T) {
			// Setup
			namespaceName := fmt.Sprintf("datadog-agent-%s", strings.ToLower(random.UniqueId()))
			kubectlOptions := k8s.NewKubectlOptions("", "", namespaceName)
			k8s.CreateNamespace(t, kubectlOptions, namespaceName)
			defer k8s.DeleteNamespace(t, kubectlOptions, namespaceName)

			cleanupSecrets := common.CreateSecretFromEnv(t, kubectlOptions, apiKeyEnv, appKeyEnv)
			defer cleanupSecrets()

			//	Install Datadog chart
			cleanUpDatadog := common.InstallChart(t, kubectlOptions, common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				Values:      []string{valPath},
				Overrides:   map[string]string{"datadog.operator.enabled": "false", "datadog-operator.datadogCRDs.crds.datadogMonitors": "false", "datadog-operator.datadogCRDs.crds.datadogDashboards": "false"},
			})
			defer cleanUpDatadog()

			datadogReleaseName := getHelmReleaseName(t, kubectlOptions, namespaceName, "datadog")

			valuesCmd := common.HelmCommand{
				ReleaseName: datadogReleaseName,
			}
			actualValuesFilePath := common.GetFullValues(t, valuesCmd, namespaceName)
			defer os.Remove(actualValuesFilePath)

			t.Log("GetFullValues created temp file:", actualValuesFilePath)

			// Install Operator chart
			cleanUpOperator := common.InstallChart(t, kubectlOptions, common.HelmCommand{
				ReleaseName: "operator",
				ChartPath:   "../../charts/datadog-operator",
			})
			defer cleanUpOperator()

			ddaName := verifyAgentConf(t, kubectlOptions, valPath, namespaceName)

			if ddaName != "" {
				k8s.RunKubectl(t, kubectlOptions, []string{"delete", "-f", ddaName}...)
				os.Remove(ddaName)
			}
		})
	}
}

func verifyAgentConf(t *testing.T, kubectlOptions *k8s.KubectlOptions, valuesPath string, namespace string) (ddaName string) {
	// Run mapper against values.yaml
	destFile, err := os.CreateTemp(".", ddaDestPath)
	require.NoError(t, err)

	mapConfig := mapper.MapConfig{
		MappingPath: mappingPath,
		SourcePath:  valuesPath,
		DestPath:    destFile.Name(),
		Namespace:   namespace,
		UpdateMap:   false,
		PrintOutput: false,
	}

	helmMapper := mapper.NewMapper(mapConfig)
	err = helmMapper.Run()
	require.NoError(t, err)

	outputBytes, err := os.ReadFile(destFile.Name())
	require.NoError(t, err)

	var ddaResult map[string]interface{}
	err = yaml.Unmarshal(outputBytes, &ddaResult)
	require.NoError(t, err)

	// Get agent conf from helm install
	helmAgentPods, err := k8s.ListPodsE(t, kubectlOptions, metav1.ListOptions{LabelSelector: "app.kubernetes.io/component=agent,app.kubernetes.io/managed-by=Helm"})
	require.NoError(t, err)
	assert.NotEmpty(t, helmAgentPods)
	k8s.WaitUntilPodAvailable(t, kubectlOptions, helmAgentPods[0].Name, 10, 15*time.Second)

	helmAgentConf, err := k8s.RunKubectlAndGetOutputE(t, kubectlOptions, []string{"exec", helmAgentPods[0].Name, "--", "agent", "config"}...)
	require.NoError(t, err)
	helmAgentConf = normalizeAgentConf(helmAgentConf)

	// Apply DDA from mapper
	err = k8s.RunKubectlE(t, kubectlOptions, []string{"apply", "-f", destFile.Name()}...)
	require.NoError(t, err)

	//time.Sleep(120 * time.Second)

	// Get agent conf from operator install
	operatorAgentPods, err := k8s.ListPodsE(t, kubectlOptions, metav1.ListOptions{LabelSelector: "agent.datadoghq.com/component=agent,app.kubernetes.io/managed-by=datadog-operator"})
	require.NoError(t, err)
	assert.NotEmpty(t, operatorAgentPods)
	k8s.WaitUntilPodAvailable(t, kubectlOptions, operatorAgentPods[0].Name, 10, 20*time.Second)

	operatorAgentConf, err := k8s.RunKubectlAndGetOutputE(t, kubectlOptions, []string{"exec", operatorAgentPods[0].Name, "--", "agent", "config"}...)
	require.NoError(t, err)
	operatorAgentConf = normalizeAgentConf(operatorAgentConf)

	// Check agent conf diff
	assert.EqualValues(t, helmAgentConf, operatorAgentConf)

	return destFile.Name()
}

func verifyConfigData(t *testing.T, kubectlOptions *k8s.KubectlOptions, valuesPath string, namespace string) (ddaManifestName string) {
	dda := verifyAgentConf(t, kubectlOptions, valuesPath, namespace)

	datadogReleaseName := getHelmReleaseName(t, kubectlOptions, namespace, "datadog")
	helmConfdCm, err := k8s.GetConfigMapE(t, kubectlOptions, fmt.Sprintf("%s-confd", datadogReleaseName))
	require.NoError(t, err)

	operatorConfdCm, err := k8s.GetConfigMapE(t, kubectlOptions, "nodeagent-extra-confd")

	require.NoError(t, err)
	assert.EqualValues(t, helmConfdCm.Data, operatorConfdCm.Data)

	return dda

}

func getHelmReleaseName(t *testing.T, kubectlOptions *k8s.KubectlOptions, namespace string, shortReleaseName string) string {
	t.Log("Finding Helm release name...")
	helmListOutput, err := helm.RunHelmCommandAndGetOutputE(t, &helm.Options{KubectlOptions: kubectlOptions}, "list", "-n", namespace, "--short")
	require.NoError(t, err, "failed to list helm releases")

	var releaseName string
	releaseNames := strings.Split(strings.TrimSpace(helmListOutput), "\n")
	for _, release := range releaseNames {
		release = strings.TrimSpace(release)
		if strings.HasPrefix(release, shortReleaseName+"-") {
			releaseName = release
			break
		}
	}
	require.NotEmpty(t, releaseName, fmt.Sprintf("could not find release %v", releaseName))
	t.Logf("Found %s release name: %s", shortReleaseName, releaseName)
	return releaseName
}

func validateEnv(t *testing.T) {
	context := common.CurrentContext(t)
	t.Log("Checking current context:", context)
	if strings.Contains(strings.ToLower(context), "staging") ||
		strings.Contains(strings.ToLower(context), "prod") {
		t.Fatal("Make sure context is pointing to local cluster")
	}

	require.NotEmpty(t, os.Getenv(apiKeyEnv), "API key can't be empty")
	require.NotEmpty(t, os.Getenv(appKeyEnv), "APP key can't be empty")
}
