package datadog

import (
	"fmt"
	"testing"

	"github.com/DataDog/datadog-agent/test/new-e2e/runner"
	"github.com/DataDog/helm-charts/test/common"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const namespace = "datadog"

var k8sClient *kubernetes.Clientset
var restConfig *rest.Config

func Test_E2E_AgentOnEKS(t *testing.T) {
	// Create pulumi EKS stack with latest version of the datadog/datadog helm chart
	config, err := common.SetupConfig()
	if err != nil {
		t.Skipf("Skipping test, problem setting up stack config: %s", err)
	}

	stackConfig := runner.ConfigMap{
		"ddtestworkload:deploy":                      auto.ConfigValue{Value: "false"},
		"ddinfra:aws/eks/linuxBottlerocketNodeGroup": auto.ConfigValue{Value: "false"},
		"ddinfra:aws/eks/windowsNodeGroup":           auto.ConfigValue{Value: "false"},
	}

	stackConfig.Merge(config)

	eksEnv, err := common.NewEKStack(stackConfig, common.DestroyStacks)
	defer common.TeardownE2EStack(eksEnv, common.PreserveStacks)

	if eksEnv != nil {
		if common.DestroyStacks {
			common.PreserveStacks = false
			t.Skipf("Skipping test, tearing down stack")
		}
		kubeconfig := eksEnv.StackOutput.Outputs["kubeconfig"]
		agentChartInstallName := eksEnv.StackOutput.Outputs["agent-linux-helm-install-name"].Value.(string)
		agentChartInstallStatus := eksEnv.StackOutput.Outputs["agent-linux-helm-install-status"].Value.(map[string]interface{})
		if kubeconfig.Value != nil {
			kc := kubeconfig.Value.(map[string]interface{})
			_, restConfig, k8sClient, err = common.NewClientFromKubeconfig(kc)
			if err == nil {
				verifyPods(t)
				assertLatestAgentChart(t, agentChartInstallName, agentChartInstallStatus)
			}
		} else {
			err = fmt.Errorf("could not create Kubernetes client, cluster kubeconfig is nil")
		}
	}
	if err != nil {
		t.Skipf("Skipping test. Encountered problem creating or updating E2E stack: %s", err)
	}
}

func verifyPods(t *testing.T) {
	nodes, err := common.ListNodes(namespace, k8sClient)
	require.NoError(t, err)

	ddaPodList, err := common.ListPods(namespace, "app=dda-datadog", k8sClient)
	require.NoError(t, err)
	dcaPodList, err := common.ListPods(namespace, "app=dda-datadog-cluster-agent", k8sClient)
	require.NoError(t, err)
	ccPodList, err := common.ListPods(namespace, "app=dda-datadog-clusterchecks", k8sClient)
	require.NoError(t, err)

	assert.EqualValues(t, len(nodes.Items), len(ddaPodList.Items), "There should be 1 datadog-agent pod per node.")
	assert.EqualValues(t, 1, len(dcaPodList.Items), "There should be 1 datadog-cluster-agent pod by default.")
	assert.EqualValues(t, 2, len(ccPodList.Items), "There should be 2 datadog-cluster-check pods by default.")

	podExec := common.K8sExec{
		ClientSet:  k8sClient,
		RestConfig: restConfig,
	}

	assertPodStatus(t, podExec, ddaPodList, "agent")
	assertPodStatus(t, podExec, dcaPodList, "cluster-agent")
	assertPodStatus(t, podExec, ccPodList, "agent")

}

func assertPodStatus(t *testing.T, podExec common.K8sExec, podList *v1.PodList, containerName string) {
	for _, pod := range podList.Items {
		assert.True(t, pod.Status.Phase == "Running")
		_, _, err := podExec.K8sExec(namespace, pod.Name, containerName, []string{"agent", "status"})
		require.NoError(t, err)
	}
}

func assertLatestAgentChart(t *testing.T, chartInstallName string, chartInstallStatus map[string]interface{}) {
	assert.EqualValues(t, chartInstallName, "dda", "Agent helm chart install name should be `dda`")
	assert.EqualValues(t, chartInstallStatus["chart"], "datadog", "Agent helm chart name should be `datadog`")
	assert.EqualValues(t, chartInstallStatus["version"], "3.32.4", "Agent helm chart version should be `3.32.4` (latest)")
}
