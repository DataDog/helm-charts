package e2e

import (
	"context"
	"fmt"
	"github.com/DataDog/datadog-agent/test/new-e2e/runner"
	"github.com/DataDog/helm-charts/test/e2e/pulumi_env"
	"github.com/DataDog/test-infra-definitions/aws/scenarios/eks"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestAgentOnEKS(t *testing.T) {
	//Create the stack
	config := pulumi_env.SetupConfig()
	stackConfig := runner.ConfigMap{
		"ddinfra:aws/eks/linuxNodeGroup":             auto.ConfigValue{Value: "false"},
		"ddinfra:aws/eks/linuxARMNodeGroup":          auto.ConfigValue{Value: "false"},
		"ddinfra:aws/eks/linuxBottlerocketNodeGroup": auto.ConfigValue{Value: "false"},
		"ddinfra:aws/eks/windowsNodeGroup":           auto.ConfigValue{Value: "false"},
		"pulumi:disable-default-providers":           auto.ConfigValue{Value: "[]"},
		"ddagent:deploy":                             auto.ConfigValue{Value: "false"},
	}
	stackConfig.Merge(config)

	_, stackOutput, err := pulumi_env.GetStackManager().GetStack(context.Background(), "helm-charts-eks-cluster", stackConfig, eks.Run, false)
	fmt.Println(stackOutput)
	errs := pulumi_env.TearDown()
	if errs != nil {
		for _, err := range errs {
			fmt.Fprint(os.Stderr, err.Error())
			require.NoError(t, err)
		}
	}
	require.NoError(t, err)
}
