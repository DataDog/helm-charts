package datadog

import (
	"fmt"
	"testing"

	"github.com/DataDog/datadog-agent/test/new-e2e/runner"
	"github.com/DataDog/helm-charts/test/common"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const namespace = "datadog"

var k8sClient *kubernetes.Clientset
var restConfig *rest.Config

func Test_E2E_AgentOnEKS(t *testing.T) {
	// Create pulumi EKS stack
	config, err := common.SetupConfig()
	require.NoError(t, err)

	stackConfig := runner.ConfigMap{
		"ddinfra:aws/eks/linuxBottlerocketNodeGroup": auto.ConfigValue{Value: "false"},
		"ddinfra:aws/eks/windowsNodeGroup":           auto.ConfigValue{Value: "false"},
		// TODO: remove when upstream eks-pulumi bug is fixed https://github.com/pulumi/pulumi-eks/pull/886
		"pulumi:disable-default-providers": auto.ConfigValue{Value: "[]"},
	}
	stackConfig.Merge(config)

	eksEnv, err := common.NewEKStack(stackConfig)
	defer common.TeardownE2EStack(eksEnv, common.PreserveStacks)

	if err == nil {
		if common.DestroyStacks {
			err = common.TeardownE2EStack(eksEnv, false)
			require.NoError(t, err)
		}
		kubeconfig := eksEnv.StackOutput.Outputs["kubeconfig"]
		agentChartInstallName := eksEnv.StackOutput.Outputs["agent-linux-helm-install-name"].Value.(string)
		agentChartInstallStatus := eksEnv.StackOutput.Outputs["agent-linux-helm-install-status"].Value.(map[string]interface{})

		if kubeconfig.Value != nil {
			kc := kubeconfig.Value.(map[string]interface{})

			_, restConfig, k8sClient, err = common.NewClientFromKubeconfig(kc)
			require.NoError(t, err)

			verifyPods(t)
			assertLatestAgentChart(t, agentChartInstallName, agentChartInstallStatus)
		} else {
			err = fmt.Errorf("Error creating cluster")
		}
	}
	require.NoError(t, err)
}

func verifyPods(t *testing.T) {
	nodes, err := common.ListNodes(namespace, k8sClient)
	require.NoError(t, err)

	ddaPodsCount := assertPodsRunning(t, common.ExpDdaPods)
	dcaPodsCount := assertPodsRunning(t, common.ExpDcaPods)
	ccPodsCount := assertPodsRunning(t, common.ExpCcPods)

	assert.EqualValues(t, ddaPodsCount, len(nodes.Items), common.ExpDdaPods.Msg)
	assert.EqualValues(t, dcaPodsCount, common.ExpDcaPods.PodCount, common.ExpDcaPods.Msg)
	assert.EqualValues(t, ccPodsCount, common.ExpCcPods.PodCount, common.ExpCcPods.Msg)
}

func assertPodsRunning(t *testing.T, expPodType common.ExpectedPods) int {
	podCount := 0
	pods, err := common.ListPods(namespace, expPodType.PodLabelSelector, k8sClient)
	require.NoError(t, err)

	for _, pod := range pods.Items {
		podCount++
		assert.True(t, pod.Status.Phase == "Running")
		assertPodExec(t, pod.Name, expPodType.ContainerName)

	}
	return podCount
}

func assertPodExec(t *testing.T, podName string, containerName string) {
	podExec := common.NewK8sExec(k8sClient, restConfig, podName, containerName, namespace)

	_, _, err := podExec.K8sExec([]string{"agent", "status"})
	require.NoError(t, err)
}

func assertLatestAgentChart(t *testing.T, chartInstallName string, chartInstallStatus map[string]interface{}) {
	assert.EqualValues(t, chartInstallName, "dda", "Agent helm chart install name should be `dda`")
	assert.EqualValues(t, chartInstallStatus["chart"], "datadog", "Agent helm chart name should be `datadog`")
	assert.EqualValues(t, chartInstallStatus["version"], "3.32.4", "Agent helm chart version should be `3.32.4` (latest)")
}
