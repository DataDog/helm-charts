package datadog

import (
	"context"
	"testing"

	"github.com/DataDog/datadog-agent/test/new-e2e/runner"
	"github.com/DataDog/datadog-agent/test/new-e2e/utils/infra"
	"github.com/DataDog/helm-charts/test/common"
	"github.com/DataDog/test-infra-definitions/scenarios/aws/eks"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/stretchr/testify/require"
)

func Test_E2E_AgentOnEKS(t *testing.T) {
	// Create pulumi EKS stack
	config, err := common.SetupConfig()
	require.NoError(t, err)

	stackConfig := runner.ConfigMap{
		"ddinfra:aws/eks/linuxBottlerocketNodeGroup": auto.ConfigValue{Value: "false"},
		"ddinfra:aws/eks/windowsNodeGroup":           auto.ConfigValue{Value: "false"},
		"pulumi:disable-default-providers":           auto.ConfigValue{Value: "['*']"},
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

		common.VerifyPods(t, clientSet, restConfig)
	}

	require.NoError(t, err)
}
