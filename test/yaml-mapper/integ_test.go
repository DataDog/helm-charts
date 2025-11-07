// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

//go:build integration

package yaml_mapper

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DataDog/datadog-operator/cmd/yaml-mapper/mapper"
	"github.com/DataDog/helm-charts/test/common"
	"github.com/google/go-cmp/cmp"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	mappingPath = "../../tools/yaml-mapper/mapping_datadog_helm_to_datadogagent_crd.yaml"
	ddaDestPath = "tempDDADest.yaml"
	apiKeyEnv   = "API_KEY"
	appKeyEnv   = "APP_KEY"
)

type AssertionFunc func(t *testing.T, kubectlOptions *k8s.KubectlOptions, valuesPath string, namespace string, cleanup *CleanupRegistry)

func Test(t *testing.T) {
	// Prerequisites
	validateEnv(t)

	tests := []struct {
		name       string
		valuesPath string
		assertion  AssertionFunc
	}{
		{
			name:       "Minimal default values",
			valuesPath: "../../charts/datadog/ci/default-values.yaml",
			assertion:  verifyAgentConf,
		},
		{
			name:       "Agent confd - equal agent config",
			valuesPath: "../../charts/datadog/ci/agent-confd-values.yaml",
			assertion:  verifyAgentConf,
		},
		{
			name:       "Agent confd - equal confd configMap",
			valuesPath: "../../charts/datadog/ci/agent-confd-values.yaml",
			assertion:  verifyConfigData,
		},
		{
			name:       "Dogstatsd with UDS",
			valuesPath: "../../charts/datadog/ci/dogstastd-socket-values.yaml",
			assertion:  verifyAgentConf,
		},
		{
			name:       "APM with local service",
			valuesPath: "../../charts/datadog/ci/agent-apm-use-local-service-values.yaml",
			assertion:  verifyAgentConf,
		},
		{
			name:       "Admission controller w/apm disabled",
			valuesPath: "../../charts/datadog/ci/apm-disabled-admission-controller-values.yaml",
			assertion:  verifyAgentConf,
		},
		{
			name:       "Admission controller w/apm portEnabled",
			valuesPath: "../../charts/datadog/ci/apm-port-enabled-admission-controller-values.yaml",
			assertion:  verifyAgentConf,
		},
		{
			name:       "Admission controller w/apm socket and port enabled",
			valuesPath: "../../charts/datadog/ci/apm-socket-and-port-admission-controller-values.yaml",
			assertion:  verifyAgentConf,
		},
		{
			name:       "Admission controller w/apm socket enabled",
			valuesPath: "../../charts/datadog/ci/apm-socket-enabled-admission-controller-values.yaml",
			assertion:  verifyAgentConf,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			cleanupRegistry := &CleanupRegistry{}

			namespaceName := fmt.Sprintf("datadog-agent-%s", strings.ToLower(random.UniqueId()))
			kubectlOptions := k8s.NewKubectlOptions("", "", namespaceName)
			k8s.CreateNamespace(t, kubectlOptions, namespaceName)
			defer k8s.DeleteNamespace(t, kubectlOptions, namespaceName)

			if os.Getenv(apiKeyEnv) != "" && os.Getenv(appKeyEnv) != "" {
				cleanupSecrets := common.CreateSecretFromEnv(t, kubectlOptions, apiKeyEnv, appKeyEnv)
				defer cleanupSecrets()
			}

			//	Install Datadog chart
			agentInstallCmd := common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				Values:      []string{tt.valuesPath},
			}

			cleanUpDatadog := common.InstallChart(t, kubectlOptions, agentInstallCmd)
			cleanupRegistry.AddDatadog(cleanUpDatadog)

			datadogReleaseName := getHelmReleaseName(t, kubectlOptions, namespaceName, agentInstallCmd.ReleaseName)

			valuesCmd := common.HelmCommand{
				ReleaseName: datadogReleaseName,
			}
			actualValuesFilePath := common.GetFullValues(t, valuesCmd, namespaceName)

			t.Log("GetFullValues created temp file:", actualValuesFilePath)

			// Install Operator chart
			operatorInstallCmd := common.HelmCommand{
				ReleaseName: "operator",
				ChartPath:   "../../charts/datadog-operator",
			}
			cleanUpOperator := common.InstallChart(t, kubectlOptions, operatorInstallCmd)
			cleanupRegistry.AddOperator(cleanUpOperator)

			tt.assertion(t, kubectlOptions, actualValuesFilePath, namespaceName, cleanupRegistry)

			t.Cleanup(func() {
				for _, dda := range cleanupRegistry.GetFiles() {
					k8s.RunKubectl(t, kubectlOptions, []string{"delete", "-f", dda}...)
					os.Remove(dda)
				}
				if cleanupRegistry.operator != nil {
					cleanupRegistry.operator()
				}
				if cleanupRegistry.datadog != nil {
					cleanupRegistry.datadog()
				}
				os.Remove(actualValuesFilePath)
			})
		})
	}
}

// verifyAgentConf validates the agent runtime config from the operator-installed agent against the
// helm-installed agent
// Note: errors must be handled with an empty return so that the kind environment can be cleaned up between test cases.
func verifyAgentConf(t *testing.T, kubectlOptions *k8s.KubectlOptions, valuesPath string, namespace string, cleanup *CleanupRegistry) {
	// Run mapper against values.yaml
	destFilePath := runMapper(t, valuesPath, namespace, cleanup)

	helmAgentPods := k8s.ListPods(t, kubectlOptions, metav1.ListOptions{LabelSelector: "app.kubernetes.io/component=agent,app.kubernetes.io/managed-by=Helm"})
	require.NotEmpty(t, helmAgentPods)
	err := k8s.WaitUntilPodAvailableE(t, kubectlOptions, helmAgentPods[0].Name, 10, 15*time.Second)
	if err != nil {
		return
	}

	// Get agent conf from helm install
	helmAgentConf, err := k8s.RunKubectlAndGetOutputE(t, kubectlOptions, []string{"exec", helmAgentPods[0].Name, "--", "agent", "config", "--all"}...)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	helmAgentConf = normalizeAgentConf(helmAgentConf)
	cleanup.datadog()
	cleanup.UnsetDatadog()

	// Apply mapped DDA
	err = k8s.RunKubectlE(t, kubectlOptions, []string{"apply", "-f", destFilePath}...)
	if err != nil {
		return
	}

	expectedPods := expectedDsCount(t, kubectlOptions)
	err = k8s.WaitUntilNumPodsCreatedE(t, kubectlOptions, metav1.ListOptions{LabelSelector: "agent.datadoghq.com/component=agent,app.kubernetes.io/managed-by=datadog-operator", FieldSelector: "status.phase=Running"}, expectedPods, 10, 15*time.Second)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	operatorAgentPods := k8s.ListPods(t, kubectlOptions, metav1.ListOptions{LabelSelector: "agent.datadoghq.com/component=agent,app.kubernetes.io/managed-by=datadog-operator", FieldSelector: "status.phase=Running"})
	require.NotEmpty(t, operatorAgentPods)

	err = k8s.WaitUntilPodAvailableE(t, kubectlOptions, operatorAgentPods[0].Name, 5, 15*time.Second)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	// Get agent conf from operator install
	operatorAgentConf, err := k8s.RunKubectlAndGetOutputE(t, kubectlOptions, []string{"exec", operatorAgentPods[0].Name, "--", "agent", "config", "--all"}...)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	operatorAgentConf = normalizeAgentConf(operatorAgentConf)

	// Check agent conf diff
	assert.True(t, cmp.Equal(helmAgentConf, operatorAgentConf), cmp.Diff(helmAgentConf, operatorAgentConf))
	return
}

func verifyConfigData(t *testing.T, kubectlOptions *k8s.KubectlOptions, valuesPath string, namespace string, cleanup *CleanupRegistry) {
	destFilePath := runMapper(t, valuesPath, namespace, cleanup)
	err := k8s.RunKubectlE(t, kubectlOptions, []string{"apply", "-f", destFilePath}...)
	require.NoError(t, err)

	datadogReleaseName := getHelmReleaseName(t, kubectlOptions, namespace, "datadog")
	helmConfdCm, err := k8s.GetConfigMapE(t, kubectlOptions, fmt.Sprintf("%s-confd", datadogReleaseName))
	require.NoError(t, err)

	operatorConfdName := "nodeagent-extra-confd"
	k8s.WaitUntilConfigMapAvailable(t, kubectlOptions, operatorConfdName, 5, 15*time.Second)
	operatorConfdCm, err := k8s.GetConfigMapE(t, kubectlOptions, operatorConfdName)

	require.NoError(t, err)
	require.EqualValues(t, helmConfdCm.Data, operatorConfdCm.Data)
}

func runMapper(t *testing.T, valuesPath string, namespace string, cleanup *CleanupRegistry) string {
	destFile, err := os.CreateTemp("", ddaDestPath)
	require.NoError(t, err)
	defer func() {
		if destFile != nil && destFile.Name() != "" {
			cleanup.AddDDA(destFile.Name())
		}
	}()

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

	return destFile.Name()
}
