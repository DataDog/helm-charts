package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/DataDog/datadog-agent/test/new-e2e/runner"
	"github.com/DataDog/datadog-agent/test/new-e2e/utils/infra"
	"github.com/DataDog/test-infra-definitions/aws/scenarios/eks"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"testing"
)

var (
	kubeconfig   []byte
	clientConfig clientcmd.ClientConfig
	restConfig   *rest.Config
	clientSet    *kubernetes.Clientset
	pods         *corev1.PodList
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
		"ddagent:deploy":                             auto.ConfigValue{Value: "true"},
	}
	stackConfig.Merge(config)

	ctx := context.Background()

	stack, stackOutput, err := infra.GetStackManager().GetStack(ctx, "helm-charts-eks-cluster", stackConfig, eks.Run, false)
	defer stack.Destroy(ctx)

	if err != nil && stackOutput.Outputs["kubeconfig"].Value != nil {

		kc := stackOutput.Outputs["kubeconfig"].Value.(map[string]interface{})
		kubeconfig, err = json.Marshal(kc)
		clientConfig, err = clientcmd.NewClientConfigFromBytes(kubeconfig)
		restConfig, err = clientConfig.ClientConfig()
		clientSet, err = kubernetes.NewForConfig(restConfig)

		namespace := ""
		pods, err = ListPods(namespace, clientSet)

		for _, pod := range pods.Items {
			fmt.Printf("Pod name: %v\n", pod.Name)
			fmt.Printf("Pod namespace: %v\n", pod.Namespace)
			fmt.Printf("Pod status: %v\n", pod.Status.Phase)
		}
		message := fmt.Sprintf("Total Pods in namespace `%s`", namespace)
		fmt.Printf("%s %d\n", message, len(pods.Items))
	}

	require.NoError(t, err)
}
