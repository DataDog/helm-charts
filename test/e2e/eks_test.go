package e2e

import (
	"context"
	"github.com/DataDog/datadog-agent/test/new-e2e/runner"
	"github.com/DataDog/datadog-agent/test/new-e2e/utils/infra"
	"github.com/DataDog/test-infra-definitions/aws/scenarios/eks"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAgentOnEKS(t *testing.T) {
	//Create the stack
	config := SetupConfig()
	stackConfig := runner.ConfigMap{
		"ddinfra:aws/eks/linuxNodeGroup":             auto.ConfigValue{Value: "false"},
		"ddinfra:aws/eks/linuxARMNodeGroup":          auto.ConfigValue{Value: "false"},
		"ddinfra:aws/eks/linuxBottlerocketNodeGroup": auto.ConfigValue{Value: "false"},
		"ddinfra:aws/eks/windowsNodeGroup":           auto.ConfigValue{Value: "false"},
		"pulumi:disable-default-providers":           auto.ConfigValue{Value: "[]"},
		"ddagent:deploy":                             auto.ConfigValue{Value: "false"},
	}
	stackConfig.Merge(config)

	_, _, err := infra.GetStackManager().GetStack(context.Background(), "helm-charts-eks-cluster", stackConfig, eks.Run, false)

	require.NoError(t, err)
}
