package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/DataDog/datadog-agent/test/new-e2e/runner"
	"github.com/DataDog/datadog-agent/test/new-e2e/utils/infra"
	"github.com/DataDog/test-infra-definitions/scenarios/aws/eks"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeconfig   []byte
	clientConfig clientcmd.ClientConfig
	restConfig   *rest.Config
	clientSet    *kubernetes.Clientset
	pods         *corev1.PodList
)

func TestAgentOnEKS(t *testing.T) {
	// Create pulumi EKS stack
	config := SetupConfig()
	stackConfig := runner.ConfigMap{
		// "ddinfra:aws/eks/windowsNodeGroup": auto.ConfigValue{Value: "false"},
		"pulumi:disable-default-providers": auto.ConfigValue{Value: "[]"},
		"aws:skipCredentialsValidation":    auto.ConfigValue{Value: "true"},
		"aws:skipMetadataApiCheck":         auto.ConfigValue{Value: "false"},
	}
	stackConfig.Merge(config)

	_, stackOutput, err := infra.GetStackManager().GetStack(context.Background(), "eks-e2e", stackConfig, eks.Run, false)
	defer debugCI()

	if stackOutput.Outputs["kubeconfig"].Value != nil {
		kc := stackOutput.Outputs["kubeconfig"].Value.(map[string]interface{})
		fmt.Println("KUBECONFIG: ", kc)
		kubeconfig, err = json.Marshal(kc)
		clientConfig, err = clientcmd.NewClientConfigFromBytes(kubeconfig)
		restConfig, err = clientConfig.ClientConfig()
		clientSet, err = kubernetes.NewForConfig(restConfig)

		namespace := "datadog"
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
