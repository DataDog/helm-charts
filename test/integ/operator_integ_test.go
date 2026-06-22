//go:build integration
// +build integration

package integ

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	apiKeyEnv = "API_KEY"
	appKeyEnv = "APP_KEY"
)

func Test(t *testing.T) {
	// Prerequisites
	context := currentContext(t)
	t.Log("Checking current context:", context)
	if strings.Contains(strings.ToLower(context), "staging") ||
		strings.Contains(strings.ToLower(context), "prod") {
		t.Fatal("Make sure context is pointing to local cluster")

	}
	if os.Getenv(apiKeyEnv) == "" {
		err := os.Setenv(apiKeyEnv, "00000000000000000000000000000000")
		require.NoError(t, err)
	}

	if os.Getenv(appKeyEnv) == "" {
		err := os.Setenv(appKeyEnv, "0000000000000000000000000000000000000000")
		require.NoError(t, err)
	}
	require.NotEmpty(t, os.Getenv(apiKeyEnv), "API key can't be empty")
	require.NotEmpty(t, os.Getenv(appKeyEnv), "APP key can't be empty")

	tests := []struct {
		name                     string
		command                  common.HelmCommand
		datadogAgentManifestPath string
		operatorAssertions       func(t *testing.T, kubectlOptions *k8s.KubectlOptions)
		agentAssertions          func(t *testing.T, kubectlOptions *k8s.KubectlOptions)
	}{
		{
			name: "Datadog agent with default Operator Helm install and base manifest",
			command: common.HelmCommand{
				ReleaseName: "datadog-operator",
				ChartPath:   "../../charts/datadog-operator",
				Overrides:   map[string]string{},
			},
			datadogAgentManifestPath: "./manifests/default.yaml",
			operatorAssertions:       verifyOperator,
			agentAssertions:          verifyAgent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			namespaceName := fmt.Sprintf("datadog-agent-%s", strings.ToLower(random.UniqueId()))
			kubectlOptions := k8s.NewKubectlOptions("", "", namespaceName)
			k8s.CreateNamespace(t, kubectlOptions, namespaceName)
			defer k8s.DeleteNamespace(t, kubectlOptions, namespaceName)

			// Install Operator
			cleanupOperator := common.InstallChart(t, kubectlOptions, tt.command)
			defer cleanupOperator()
			time.Sleep(15 * time.Second)
			// Verify Operator
			verifyOperator(t, kubectlOptions)

			// Apply DatadogAgent Manifest
			cleanupSecrets := common.CreateSecretFromEnv(t, kubectlOptions, apiKeyEnv, appKeyEnv)
			defer cleanupSecrets()
			t.Log("Applying DatadogAgent manifest")
			k8s.KubectlApply(t, kubectlOptions, tt.datadogAgentManifestPath)
			defer k8s.KubectlDelete(t, kubectlOptions, tt.datadogAgentManifestPath)

			// Verify Agent Setup
			t.Log("Verifying agent pods are running")
			verifyAgent(t, kubectlOptions)

			// 'Pause' test for local debugging
			//t.Log("Sleeping for 2 minutes")
			//time.Sleep(120 * time.Second)
		})
	}
}

func verifyOperator(t *testing.T, kubectlOptions *k8s.KubectlOptions) {
	verifyNumPodsForSelector(t, kubectlOptions, 1, "app.kubernetes.io/name=datadog-operator")
}

func verifyAgent(t *testing.T, kubectlOptions *k8s.KubectlOptions) {
	verifyNumPodsForSelector(t, kubectlOptions, 1, "agent.datadoghq.com/component=agent")
	verifyNumPodsForSelector(t, kubectlOptions, 1, "agent.datadoghq.com/component=cluster-agent")
	verifyNumPodsForSelector(t, kubectlOptions, 1, "agent.datadoghq.com/component=cluster-checks-runner")
}

func verifyNumPodsForSelector(t *testing.T, kubectlOptions *k8s.KubectlOptions, numPods int, selector string) {
	t.Log("Waiting for number of pods created", "number", numPods, "selector", selector)
	k8s.WaitUntilNumPodsCreated(t, kubectlOptions, v1.ListOptions{
		LabelSelector: selector,
	}, numPods, 9, 10*time.Second)
}

func currentContext(t *testing.T) string {
	val, err := k8s.RunKubectlAndGetOutputE(t, k8s.NewKubectlOptions("", "", ""), "config", "current-context")
	require.Nil(t, err)
	return val
}
