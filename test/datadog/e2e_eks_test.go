package datadog

import (
	"context"
	"strings"
	"testing"

	"github.com/DataDog/datadog-agent/test/new-e2e/runner"
	"github.com/DataDog/datadog-agent/test/new-e2e/utils/infra"
	"github.com/DataDog/helm-charts/test/common"
	"github.com/DataDog/test-infra-definitions/scenarios/aws/eks"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	ddaPrefix = "dda-datadog"
	dcaPrefix = "dda-datadog-cluster-agent"
	ccPrefix  = "dda-datadog-clusterchecks"
)

func Test_E2E_AgentOnEKS(t *testing.T) {
	// Create pulumi EKS stack
	config := common.SetupConfig()
	stackConfig := runner.ConfigMap{
		"ddinfra:aws/eks/linuxBottlerocketNodeGroup": auto.ConfigValue{Value: "false"},
		"ddinfra:aws/eks/windowsNodeGroup":           auto.ConfigValue{Value: "false"},
		"pulumi:disable-default-providers":           auto.ConfigValue{Value: "[]"},
		"aws:skipCredentialsValidation":              auto.ConfigValue{Value: "true"},
		"aws:skipMetadataApiCheck":                   auto.ConfigValue{Value: "false"},
	}
	stackConfig.Merge(config)

	_, stackOutput, err := infra.GetStackManager().GetStack(context.Background(), "eks-e2e", stackConfig, eks.Run, false)
	defer common.TeardownSuite(PreserveStacks)

	if stackOutput.Outputs["kubeconfig"].Value != nil {
		kc := stackOutput.Outputs["kubeconfig"].Value.(map[string]interface{})

		_, restConfig, clientSet, err := common.NewClientFromKubeconfig(kc)
		require.NoError(t, err)

		namespace := "datadog"
		nodes, err := common.ListNodes(namespace, clientSet)
		require.NoError(t, err)

		pods, err := common.ListPods(namespace, clientSet)
		ddaPods := 0
		dcaPods := 0
		ccPods := 0

		for _, pod := range pods.Items {
			containerName := "agent"
			switch {
			case strings.HasPrefix(pod.Name, dcaPrefix) == true:
				dcaPods++
				containerName = "cluster-agent"
			case strings.HasPrefix(pod.Name, ccPrefix) == true:
				ccPods++
			case strings.HasPrefix(pod.Name, ddaPrefix) == true:
				ddaPods++
			}
			assert.True(t, pod.Status.Phase == "Running")

			podExec := common.NewK8sExec(clientSet, restConfig, pod.Name, containerName, namespace)

			_, _, err := podExec.Exec([]string{"agent", "status"})
			require.NoError(t, err)
		}

		assert.EqualValues(t, ddaPods, len(nodes.Items), "There should be 1 datadog-agent pod per node.")
		assert.EqualValues(t, dcaPods, 1, "There should be 2 datadog-cluster-agent pod by default.")
		assert.EqualValues(t, ccPods, 2, "There should be 2 datadog-cluster-check pods by default.")
		require.NoError(t, err)
	}

	require.NoError(t, err)
}
