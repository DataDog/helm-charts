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
	"github.com/gruntwork-io/terratest/modules/helm"
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

// INTEGRATION TEST

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
			time.Sleep(60 * time.Second)

			t.Log("Finding datadog release name...")
			helmListOutput, err := helm.RunHelmCommandAndGetOutputE(t, &helm.Options{KubectlOptions: kubectlOptions}, "list", "-n", namespaceName, "--short")
			require.NoError(t, err, "failed to list helm releases")

			var datadogReleaseName string
			releaseNames := strings.Split(strings.TrimSpace(helmListOutput), "\n")
			for _, releaseName := range releaseNames {
				releaseName = strings.TrimSpace(releaseName)
				if strings.HasPrefix(releaseName, tt.command.ReleaseName+"-") {
					datadogReleaseName = releaseName
					break
				}
			}
			require.NotEmpty(t, datadogReleaseName, "could not find datadog release")
			t.Log("Found datadog release name:", datadogReleaseName)

			valuesCmd := common.HelmCommand{
				ReleaseName: datadogReleaseName,
			}
			actualValuesFilePath := common.GetFullValues(t, valuesCmd, namespaceName)
			defer os.Remove(actualValuesFilePath)

			t.Log("GetFullValues created temp file:", actualValuesFilePath)

			cleanUpOperator := common.InstallChart(t, kubectlOptions, common.HelmCommand{
				ReleaseName: "datadog-operator",
				ChartPath:   "../../../../charts/datadog-operator",
			})
			defer cleanUpOperator()

			for _, assertion := range tt.assertions {
				assertion(t, kubectlOptions, actualValuesFilePath, namespaceName)
			}

		})
	}
}

func verifyAgentConf(t *testing.T, kubectlOptions *k8s.KubectlOptions, valuesPath string, namespace string) {
	// Run mapper against values.yaml
	destFile, err := os.CreateTemp(".", ddaDestPath)
	require.NoError(t, err)
	defer os.Remove(destFile.Name())

	MapYaml(mappingPath, valuesPath, destFile.Name(), "", "", namespace, false, false)

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
	operatorAgentPods, err := k8s.ListPodsE(t, kubectlOptions, metav1.ListOptions{LabelSelector: "agent.datadoghq.com/component=agent,app.kubernetes.io/managed-by=datadog-operator"})
	require.NoError(t, err)

	t.Logf("OPERATOR PODS: Found %d pods with operator labels", len(operatorAgentPods))
	for i, pod := range operatorAgentPods {
		t.Logf("  Operator Pod %d: %s (Labels: %v)", i, pod.Name, pod.Labels)
	}

	assert.NotEmpty(t, operatorAgentPods)
	operatorAgentConf, err := k8s.RunKubectlAndGetOutputE(t, kubectlOptions, []string{"exec", operatorAgentPods[0].Name, "--", "agent", "config"}...)
	require.NoError(t, err)
	operatorAgentConf = normalizeAgentConf(operatorAgentConf)

	// Check agent conf diff
	assert.EqualValues(t, helmAgentConf, operatorAgentConf)
}

var skipFields = map[string]interface{}{
	"install_id":              nil,
	"install_time":            nil,
	"install_type":            nil,
	"kubernetes_service_name": nil, // service name differs according to installation
	"kubernetes_kubelet_host": nil, // may also differ
	"token_name":              nil,
	"site":                    nil,
}

// normalizeAgentConf removes log lines that start with timestamps in the format "2006-01-02 15:04:05 UTC"
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
		// Skip lines that contain fields that should be skipped
		if shouldSkipField(line) {
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
	if len(line) < 23 { // Minimum length for "2006-01-02 15:04:05"
		return false
	}
	_, err := time.Parse("2006-01-02 15:04:05 MST", line[:23])
	if err == nil {
		return true
	}
	return false
}

func shouldSkipField(line string) bool {
	for field := range skipFields {
		if strings.Contains(line, field) {
			return true
		}
	}
	return false
}

// UNIT TESTS

func TestMergeMaps(t *testing.T) {
	tests := []struct {
		name     string
		map1     map[string]interface{}
		map2     map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "merge non-overlapping maps",
			map1: map[string]interface{}{
				"key1": "value1",
				"key2": 42,
			},
			map2: map[string]interface{}{
				"key3": "value3",
				"key4": []string{"a", "b"},
			},
			expected: map[string]interface{}{
				"key1": "value1",
				"key2": 42,
				"key3": "value3",
				"key4": []string{"a", "b"},
			},
		},
		{
			name: "merge overlapping maps with simple values (map2 overwrites map1)",
			map1: map[string]interface{}{
				"key1": "value1",
				"key2": 42,
			},
			map2: map[string]interface{}{
				"key1": "newvalue1",
				"key3": "value3",
			},
			expected: map[string]interface{}{
				"key1": "newvalue1",
				"key2": 42,
				"key3": "value3",
			},
		},
		{
			name: "merge nested maps",
			map1: map[string]interface{}{
				"config": map[string]interface{}{
					"database": map[string]interface{}{
						"host": "localhost",
						"port": 5432,
					},
					"cache": map[string]interface{}{
						"enabled": true,
					},
				},
				"version": "1.0",
			},
			map2: map[string]interface{}{
				"config": map[string]interface{}{
					"database": map[string]interface{}{
						"port":     3306,
						"username": "admin",
					},
					"logging": map[string]interface{}{
						"level": "debug",
					},
				},
				"environment": "production",
			},
			expected: map[string]interface{}{
				"config": map[string]interface{}{
					"database": map[string]interface{}{
						"host":     "localhost",
						"port":     3306,
						"username": "admin",
					},
					"cache": map[string]interface{}{
						"enabled": true,
					},
					"logging": map[string]interface{}{
						"level": "debug",
					},
				},
				"version":     "1.0",
				"environment": "production",
			},
		},
		{
			name: "one map is empty",
			map1: map[string]interface{}{
				"key1": "value1",
			},
			map2: map[string]interface{}{},
			expected: map[string]interface{}{
				"key1": "value1",
			},
		},
		{
			name:     "both maps are empty",
			map1:     map[string]interface{}{},
			map2:     map[string]interface{}{},
			expected: map[string]interface{}{},
		},
		{
			name: "mixed value types",
			map1: map[string]interface{}{
				"string":  "text",
				"number":  123,
				"boolean": true,
				"array":   []interface{}{1, 2, 3},
				"nested": map[string]interface{}{
					"inner": "value",
				},
			},
			map2: map[string]interface{}{
				"string": "newtext",
				"float":  3.14,
				"nested": map[string]interface{}{
					"additional": "data",
				},
			},
			expected: map[string]interface{}{
				"string":  "newtext",
				"number":  123,
				"boolean": true,
				"array":   []interface{}{1, 2, 3},
				"float":   3.14,
				"nested": map[string]interface{}{
					"inner":      "value",
					"additional": "data",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			map1Copy := make(map[string]interface{})
			for k, v := range tt.map1 {
				map1Copy[k] = v
			}
			map2Copy := make(map[string]interface{})
			for k, v := range tt.map2 {
				map2Copy[k] = v
			}

			result := mergeMaps(map1Copy, map2Copy)
			assert.Equal(t, tt.expected, result)

			assert.Equal(t, tt.expected, map1Copy)
		})
	}
}

func TestCustomMapFuncs(t *testing.T) {
	// Test that all custom map functions are properly registered
	t.Run("customMapFuncs_dict", func(t *testing.T) {
		expectedFuncs := []string{"mapApiSecretKey", "mapAppSecretKey", "mapTokenSecretKey", "mapSeccompProfile", "mapSystemProbeAppArmor", "mapLocalServiceName", "mapAppendEnvVar", "mapMergeEnvs"}

		for _, funcName := range expectedFuncs {
			t.Run(funcName+"_exists", func(t *testing.T) {
				_, exists := customMapFuncs[funcName]
				assert.True(t, exists, "Custom map function %s should be registered", funcName)
			})
		}

		assert.Equal(t, len(expectedFuncs), len(customMapFuncs), "Should have exactly %d custom map functions", len(expectedFuncs))
	})

	// Test individual functions through the dictionary
	tests := []struct {
		name        string
		funcName    string
		interim     map[string]interface{}
		newPath     string
		pathVal     interface{}
		mapFuncArgs []interface{}
		expectedMap map[string]interface{}
	}{
		// mapApiSecretKey tests
		{
			name:     "mapApiSecretKey_empty_map",
			funcName: "mapApiSecretKey",
			interim:  map[string]interface{}{},
			newPath:  "spec.global.credentials.apiSecret.secretName",
			pathVal:  "my-api-secret",
			expectedMap: map[string]interface{}{
				"spec.global.credentials.apiSecret.secretName": "my-api-secret",
				"spec.global.credentials.apiSecret.keyName":    "api-key",
			},
		},
		{
			name:     "mapApiSecretKey_existing_map",
			funcName: "mapApiSecretKey",
			interim: map[string]interface{}{
				"spec.global.site":      "datadoghq.com",
				"spec.agent.image.name": "datadog/agent",
			},
			newPath: "spec.global.credentials.apiSecret.secretName",
			pathVal: "datadog-api-secret",
			expectedMap: map[string]interface{}{
				"spec.global.site":                             "datadoghq.com",
				"spec.agent.image.name":                        "datadog/agent",
				"spec.global.credentials.apiSecret.secretName": "datadog-api-secret",
				"spec.global.credentials.apiSecret.keyName":    "api-key",
			},
		},
		{
			name:     "mapApiSecretKey_overwrite",
			funcName: "mapApiSecretKey",
			interim: map[string]interface{}{
				"spec.global.credentials.apiSecret.secretName": "old-secret",
				"spec.global.credentials.apiSecret.keyName":    "old-key",
			},
			newPath: "spec.global.credentials.apiSecret.secretName",
			pathVal: "new-api-secret",
			expectedMap: map[string]interface{}{
				"spec.global.credentials.apiSecret.secretName": "new-api-secret",
				"spec.global.credentials.apiSecret.keyName":    "api-key",
			},
		},
		// mapAppSecretKey tests
		{
			name:     "mapAppSecretKey_empty_map",
			funcName: "mapAppSecretKey",
			interim:  map[string]interface{}{},
			newPath:  "spec.global.credentials.appSecret.secretName",
			pathVal:  "my-app-secret",
			expectedMap: map[string]interface{}{
				"spec.global.credentials.appSecret.secretName": "my-app-secret",
				"spec.global.credentials.appSecret.keyName":    "app-key",
			},
		},
		{
			name:     "mapAppSecretKey_with_existing_api_secret",
			funcName: "mapAppSecretKey",
			interim: map[string]interface{}{
				"spec.global.credentials.apiSecret.secretName": "api-secret",
				"spec.global.credentials.apiSecret.keyName":    "api-key",
			},
			newPath: "spec.global.credentials.appSecret.secretName",
			pathVal: "datadog-app-secret",
			expectedMap: map[string]interface{}{
				"spec.global.credentials.apiSecret.secretName": "api-secret",
				"spec.global.credentials.apiSecret.keyName":    "api-key",
				"spec.global.credentials.appSecret.secretName": "datadog-app-secret",
				"spec.global.credentials.appSecret.keyName":    "app-key",
			},
		},
		{
			name:     "mapAppSecretKey_overwrite",
			funcName: "mapAppSecretKey",
			interim: map[string]interface{}{
				"spec.global.credentials.appSecret.secretName": "old-app-secret",
				"spec.global.credentials.appSecret.keyName":    "old-app-key",
			},
			newPath: "spec.global.credentials.appSecret.secretName",
			pathVal: "new-app-secret",
			expectedMap: map[string]interface{}{
				"spec.global.credentials.appSecret.secretName": "new-app-secret",
				"spec.global.credentials.appSecret.keyName":    "app-key",
			},
		},
		// mapTokenSecretKey tests
		{
			name:     "mapTokenSecretKey_empty_map",
			funcName: "mapTokenSecretKey",
			interim:  map[string]interface{}{},
			newPath:  "spec.global.clusterAgentTokenSecret.secretName",
			pathVal:  "my-token-secret",
			expectedMap: map[string]interface{}{
				"spec.global.clusterAgentTokenSecret.secretName": "my-token-secret",
				"spec.global.clusterAgentTokenSecret.keyName":    "token",
			},
		},
		{
			name:     "mapTokenSecretKey_with_existing_secrets",
			funcName: "mapTokenSecretKey",
			interim: map[string]interface{}{
				"spec.global.credentials.apiSecret.secretName": "api-secret",
				"spec.global.credentials.appSecret.secretName": "app-secret",
			},
			newPath: "spec.global.clusterAgentTokenSecret.secretName",
			pathVal: "cluster-agent-token",
			expectedMap: map[string]interface{}{
				"spec.global.credentials.apiSecret.secretName":   "api-secret",
				"spec.global.credentials.appSecret.secretName":   "app-secret",
				"spec.global.clusterAgentTokenSecret.secretName": "cluster-agent-token",
				"spec.global.clusterAgentTokenSecret.keyName":    "token",
			},
		},
		{
			name:     "mapTokenSecretKey_overwrite",
			funcName: "mapTokenSecretKey",
			interim: map[string]interface{}{
				"spec.global.clusterAgentTokenSecret.secretName": "old-token-secret",
				"spec.global.clusterAgentTokenSecret.keyName":    "old-token",
			},
			newPath: "spec.global.clusterAgentTokenSecret.secretName",
			pathVal: "new-token-secret",
			expectedMap: map[string]interface{}{
				"spec.global.clusterAgentTokenSecret.secretName": "new-token-secret",
				"spec.global.clusterAgentTokenSecret.keyName":    "token",
			},
		},
		// mapSeccompProfile tests
		{
			name:     "mapSeccompProfile_localhost",
			funcName: "mapSeccompProfile",
			interim:  map[string]interface{}{},
			newPath:  "spec.override.nodeAgent.containers.system-probe.securityContext.seccompProfile",
			pathVal:  "localhost/system-probe",
			expectedMap: map[string]interface{}{
				"spec.override.nodeAgent.containers.system-probe.securityContext.seccompProfile.type":             "Localhost",
				"spec.override.nodeAgent.containers.system-probe.securityContext.seccompProfile.localhostProfile": "system-probe",
			},
		},
		{
			name:     "mapSeccompProfile_runtime_default",
			funcName: "mapSeccompProfile",
			interim:  map[string]interface{}{},
			newPath:  "spec.override.nodeAgent.containers.system-probe.securityContext.seccompProfile",
			pathVal:  "runtime/default",
			expectedMap: map[string]interface{}{
				"spec.override.nodeAgent.containers.system-probe.securityContext.seccompProfile.type": "RuntimeDefault",
			},
		},
		{
			name:     "mapSeccompProfile_unconfined",
			funcName: "mapSeccompProfile",
			interim:  map[string]interface{}{},
			newPath:  "spec.override.nodeAgent.containers.system-probe.securityContext.seccompProfile",
			pathVal:  "unconfined",
			expectedMap: map[string]interface{}{
				"spec.override.nodeAgent.containers.system-probe.securityContext.seccompProfile.type": "Unconfined",
			},
		},
		// mapSystemProbeAppArmor tests
		{
			name:     "mapSystemProbeAppArmor_no_features_enabled",
			funcName: "mapSystemProbeAppArmor",
			interim: map[string]interface{}{
				"spec.features.cws.enabled": false,
				"spec.features.npm.enabled": false,
			},
			newPath: "spec.override.nodeAgent.containers.system-probe.appArmorProfile",
			pathVal: "unconfined",
			expectedMap: map[string]interface{}{
				"spec.features.cws.enabled": false,
				"spec.features.npm.enabled": false,
			},
		},
		{
			name:     "mapSystemProbeAppArmor_multiple_features_enabled",
			funcName: "mapSystemProbeAppArmor",
			interim: map[string]interface{}{
				"spec.features.cws.enabled":            true,
				"spec.features.npm.enabled":            false,
				"spec.features.tcpQueueLength.enabled": true,
			},
			newPath: "spec.override.nodeAgent.containers.system-probe.appArmorProfile",
			pathVal: "unconfined",
			expectedMap: map[string]interface{}{
				"spec.features.cws.enabled":                                       true,
				"spec.features.npm.enabled":                                       false,
				"spec.features.tcpQueueLength.enabled":                            true,
				"spec.override.nodeAgent.containers.system-probe.appArmorProfile": "unconfined",
			},
		},
		{
			name:     "mapSystemProbeAppArmor_gpu_enabled_privileged",
			funcName: "mapSystemProbeAppArmor",
			interim: map[string]interface{}{
				"spec.features.gpu.enabled":        true,
				"spec.features.gpu.privilegedMode": true,
			},
			newPath: "spec.override.nodeAgent.containers.system-probe.appArmorProfile",
			pathVal: "unconfined",
			expectedMap: map[string]interface{}{
				"spec.features.gpu.enabled":                                       true,
				"spec.features.gpu.privilegedMode":                                true,
				"spec.override.nodeAgent.containers.system-probe.appArmorProfile": "unconfined",
			},
		},
		{
			name:     "mapSystemProbeAppArmor_gpu_enabled_not_privileged",
			funcName: "mapSystemProbeAppArmor",
			interim: map[string]interface{}{
				"spec.features.gpu.enabled":        true,
				"spec.features.gpu.privilegedMode": false,
			},
			newPath: "spec.override.nodeAgent.containers.system-probe.appArmorProfile",
			pathVal: "unconfined",
			expectedMap: map[string]interface{}{
				"spec.features.gpu.enabled":        true,
				"spec.features.gpu.privilegedMode": false,
			},
		},
		{
			name:     "mapSystemProbeAppArmor_empty_apparmor_value",
			funcName: "mapSystemProbeAppArmor",
			interim: map[string]interface{}{
				"spec.features.cws.enabled": true,
			},
			newPath: "spec.override.nodeAgent.containers.system-probe.appArmorProfile",
			pathVal: "",
			expectedMap: map[string]interface{}{
				"spec.features.cws.enabled": true,
			},
		},
		{
			name:     "mapSystemProbeAppArmor_invalid_apparmor_type",
			funcName: "mapSystemProbeAppArmor",
			interim: map[string]interface{}{
				"spec.features.cws.enabled": true,
			},
			newPath: "spec.override.nodeAgent.containers.system-probe.appArmorProfile",
			pathVal: 123,
			expectedMap: map[string]interface{}{
				"spec.features.cws.enabled": true,
			},
		},
		// mapLocalServiceName tests
		{
			name:        "mapLocalServiceName_empty_name",
			funcName:    "mapLocalServiceName",
			interim:     map[string]interface{}{},
			newPath:     "spec.override.clusterAgent.config.external_metrics.local_service_name",
			pathVal:     "",
			expectedMap: map[string]interface{}{},
		},
		{
			name:        "mapLocalServiceName_invalid_type",
			funcName:    "mapLocalServiceName",
			interim:     map[string]interface{}{},
			newPath:     "spec.override.clusterAgent.config.external_metrics.local_service_name",
			pathVal:     123,
			expectedMap: map[string]interface{}{},
		},
		{
			name:     "mapLocalServiceName_overwrite_existing",
			funcName: "mapLocalServiceName",
			interim: map[string]interface{}{
				"spec.override.clusterAgent.config.external_metrics.local_service_name": "old-service",
			},
			newPath: "spec.override.clusterAgent.config.external_metrics.local_service_name",
			pathVal: "new-service",
			expectedMap: map[string]interface{}{
				"spec.override.clusterAgent.config.external_metrics.local_service_name": "new-service",
			},
		},
		{
			name:     "mapAppendEnvVar_add_env_var",
			funcName: "mapAppendEnvVar",
			interim:  map[string]interface{}{},
			newPath:  "spec.override.nodeAgent.containers.agent.env",
			pathVal:  "debug",
			mapFuncArgs: []interface{}{
				map[string]interface{}{
					"name": "DD_LOG_LEVEL",
				},
			},
			expectedMap: map[string]interface{}{
				"spec.override.nodeAgent.containers.agent.env": []interface{}{
					map[string]interface{}{
						"name":  "DD_LOG_LEVEL",
						"value": "debug",
					},
				},
			},
		},
		{
			name:     "mapAppendEnvVar_add_to_existing_env_vars",
			funcName: "mapAppendEnvVar",
			interim: map[string]interface{}{
				"spec.override.nodeAgent.containers.agent.env": []interface{}{
					map[string]interface{}{
						"name":  "EXISTING_VAR",
						"value": "existing_value",
					},
				},
			},
			newPath: "spec.override.nodeAgent.containers.agent.env",
			pathVal: "new_value",
			mapFuncArgs: []interface{}{
				map[string]interface{}{
					"name": "NEW_VAR",
				},
			},
			expectedMap: map[string]interface{}{
				"spec.override.nodeAgent.containers.agent.env": []interface{}{
					map[string]interface{}{
						"name":  "EXISTING_VAR",
						"value": "existing_value",
					},
					map[string]interface{}{
						"name":  "NEW_VAR",
						"value": "new_value",
					},
				},
			},
		},
		{
			name:     "mapMergeEnvs_add_new_envs",
			funcName: "mapMergeEnvs",
			interim:  map[string]interface{}{},
			newPath:  "spec.override.nodeAgent.containers.agent.env",
			pathVal: []interface{}{
				map[string]interface{}{
					"name":  "VAR1",
					"value": "value1",
				},
			},
			expectedMap: map[string]interface{}{
				"spec.override.nodeAgent.containers.agent.env": []interface{}{
					map[string]interface{}{
						"name":  "VAR1",
						"value": "value1",
					},
				},
			},
		},
		{
			name:     "mapMergeEnvs_add_to_existing_envs",
			funcName: "mapMergeEnvs",
			interim: map[string]interface{}{
				"spec.override.nodeAgent.containers.agent.env": []interface{}{
					map[string]interface{}{
						"name":  "EXISTING_VAR",
						"value": "existing_value",
					},
				},
			},
			newPath: "spec.override.nodeAgent.containers.agent.env",
			pathVal: []interface{}{
				map[string]interface{}{
					"name":  "NEW_VAR",
					"value": "new_value",
				},
			},
			expectedMap: map[string]interface{}{
				"spec.override.nodeAgent.containers.agent.env": []interface{}{
					map[string]interface{}{
						"name":  "EXISTING_VAR",
						"value": "existing_value",
					},
					map[string]interface{}{
						"name":  "NEW_VAR",
						"value": "new_value",
					},
				},
			},
		},
		{
			name:     "mapMergeEnvs_avoid_duplicates",
			funcName: "mapMergeEnvs",
			interim: map[string]interface{}{
				"spec.override.nodeAgent.containers.agent.env": []interface{}{
					map[string]interface{}{
						"name":  "EXISTING_VAR",
						"value": "existing_value",
					},
				},
			},
			newPath: "spec.override.nodeAgent.containers.agent.env",
			pathVal: []interface{}{
				map[string]interface{}{
					"name":  "EXISTING_VAR", // This should not be added again
					"value": "new_value",
				},
				map[string]interface{}{
					"name":  "NEW_VAR",
					"value": "new_value",
				},
			},
			expectedMap: map[string]interface{}{
				"spec.override.nodeAgent.containers.agent.env": []interface{}{
					map[string]interface{}{
						"name":  "EXISTING_VAR",
						"value": "existing_value", // Keeps the original value
					},
					map[string]interface{}{
						"name":  "NEW_VAR",
						"value": "new_value",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customFunc, exists := customMapFuncs[tt.funcName]
			require.True(t, exists, "Custom function %s should exist in registry", tt.funcName)
			customFunc(tt.interim, tt.newPath, tt.pathVal, tt.mapFuncArgs)

			assert.Equal(t, tt.expectedMap, tt.interim)
		})
	}

	t.Run("non_existent_function", func(t *testing.T) {
		_, exists := customMapFuncs["nonExistentFunc"]
		assert.False(t, exists, "Non-existent function should not be in registry")
	})
}

func TestMakeTable(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		val      interface{}
		mapName  map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name:    "simple single level path",
			path:    "key",
			val:     "value",
			mapName: map[string]interface{}{},
			expected: map[string]interface{}{
				"key": "value",
			},
		},
		{
			name:    "three level nested path",
			path:    "spec.global.site",
			val:     "datadoghq.com",
			mapName: map[string]interface{}{},
			expected: map[string]interface{}{
				"spec": map[string]interface{}{
					"global": map[string]interface{}{
						"site": "datadoghq.com",
					},
				},
			},
		},
		{
			name:    "deep nested path",
			path:    "spec.override.nodeAgent.containers.agent.resources.limits.memory",
			val:     "512Mi",
			mapName: map[string]interface{}{},
			expected: map[string]interface{}{
				"spec": map[string]interface{}{
					"override": map[string]interface{}{
						"nodeAgent": map[string]interface{}{
							"containers": map[string]interface{}{
								"agent": map[string]interface{}{
									"resources": map[string]interface{}{
										"limits": map[string]interface{}{
											"memory": "512Mi",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "merge with existing map - non-overlapping",
			path: "spec.global.site",
			val:  "datadoghq.com",
			mapName: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name": "datadog",
				},
			},
			expected: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name": "datadog",
				},
				"spec": map[string]interface{}{
					"global": map[string]interface{}{
						"site": "datadoghq.com",
					},
				},
			},
		},
		{
			name: "merge with existing map - overlapping paths",
			path: "spec.global.logLevel",
			val:  "debug",
			mapName: map[string]interface{}{
				"spec": map[string]interface{}{
					"global": map[string]interface{}{
						"site": "datadoghq.com",
					},
					"features": map[string]interface{}{
						"apm": map[string]interface{}{
							"enabled": true,
						},
					},
				},
			},
			expected: map[string]interface{}{
				"spec": map[string]interface{}{
					"global": map[string]interface{}{
						"site":     "datadoghq.com",
						"logLevel": "debug",
					},
					"features": map[string]interface{}{
						"apm": map[string]interface{}{
							"enabled": true,
						},
					},
				},
			},
		},
		{
			name: "overwrite existing value",
			path: "spec.global.site",
			val:  "datadoghq.eu",
			mapName: map[string]interface{}{
				"spec": map[string]interface{}{
					"global": map[string]interface{}{
						"site": "datadoghq.com",
					},
				},
			},
			expected: map[string]interface{}{
				"spec": map[string]interface{}{
					"global": map[string]interface{}{
						"site": "datadoghq.eu",
					},
				},
			},
		},
		{
			name:    "empty path",
			path:    "",
			val:     "",
			mapName: map[string]interface{}{},
			expected: map[string]interface{}{
				"": "",
			},
		},
		{
			name:    "different value types - integer",
			path:    "spec.override.clusterAgent.replicas",
			val:     3,
			mapName: map[string]interface{}{},
			expected: map[string]interface{}{
				"spec": map[string]interface{}{
					"override": map[string]interface{}{
						"clusterAgent": map[string]interface{}{
							"replicas": 3,
						},
					},
				},
			},
		},
		{
			name:    "different value types - boolean",
			path:    "spec.features.apm.enabled",
			val:     true,
			mapName: map[string]interface{}{},
			expected: map[string]interface{}{
				"spec": map[string]interface{}{
					"features": map[string]interface{}{
						"apm": map[string]interface{}{
							"enabled": true,
						},
					},
				},
			},
		},
		{
			name:    "different value types - slice",
			path:    "spec.global.tags",
			val:     []string{"env:prod", "team:backend"},
			mapName: map[string]interface{}{},
			expected: map[string]interface{}{
				"spec": map[string]interface{}{
					"global": map[string]interface{}{
						"tags": []string{"env:prod", "team:backend"},
					},
				},
			},
		},
		{
			name:    "different value types - map",
			path:    "spec.override.nodeAgent.resources",
			val:     map[string]interface{}{"limits": map[string]interface{}{"memory": "1Gi"}},
			mapName: map[string]interface{}{},
			expected: map[string]interface{}{
				"spec": map[string]interface{}{
					"override": map[string]interface{}{
						"nodeAgent": map[string]interface{}{
							"resources": map[string]interface{}{
								"limits": map[string]interface{}{
									"memory": "1Gi",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy of the input map to avoid modifying the test data
			mapNameCopy := make(map[string]interface{})
			for k, v := range tt.mapName {
				mapNameCopy[k] = v
			}

			result := makeTable(tt.path, tt.val, mapNameCopy)

			// Verify that the result matches expected
			assert.Equal(t, tt.expected, result)

			// Verify that the function modifies the input map in place
			assert.Equal(t, tt.expected, mapNameCopy)

			// Verify that the returned map is the same object as the input map
			assert.True(t, fmt.Sprintf("%p", result) == fmt.Sprintf("%p", mapNameCopy), "makeTable should return the same map object that was passed in")
		})
	}
}

func TestMakeTableEdgeCases(t *testing.T) {
	t.Run("nil_value", func(t *testing.T) {
		mapName := map[string]interface{}{}
		result := makeTable("spec.global.site", nil, mapName)

		expected := map[string]interface{}{
			"spec": map[string]interface{}{
				"global": map[string]interface{}{
					"site": nil,
				},
			},
		}
		assert.Equal(t, expected, result)
	})

	t.Run("path_with_multiple_dots", func(t *testing.T) {
		mapName := map[string]interface{}{}
		result := makeTable("a.b.c.d.e.f", "deep_value", mapName)

		expected := map[string]interface{}{
			"a": map[string]interface{}{
				"b": map[string]interface{}{
					"c": map[string]interface{}{
						"d": map[string]interface{}{
							"e": map[string]interface{}{
								"f": "deep_value",
							},
						},
					},
				},
			},
		}
		assert.Equal(t, expected, result)
	})

	t.Run("path_with_numeric_keys", func(t *testing.T) {
		mapName := map[string]interface{}{}
		result := makeTable("spec.containers.0.name", "agent", mapName)

		expected := map[string]interface{}{
			"spec": map[string]interface{}{
				"containers": map[string]interface{}{
					"0": map[string]interface{}{
						"name": "agent",
					},
				},
			},
		}
		assert.Equal(t, expected, result)
	})
}
