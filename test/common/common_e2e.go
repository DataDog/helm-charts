package common

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/DataDog/datadog-agent/test/new-e2e/utils/infra"

	"github.com/DataDog/datadog-agent/test/new-e2e/runner"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

var defaultLocalPulumiConfigs = runner.ConfigMap{
	"ddinfra:aws/defaultKeyPairName": auto.ConfigValue{Value: os.Getenv("AWS_KEYPAIR_NAME")},
}

var defaultCIPulumiConfigs = runner.ConfigMap{}

func TeardownSuite(preserveStacks bool) {
	if !preserveStacks {
		fmt.Fprintf(os.Stderr, "Cleaning up E2E stacks. ")
		errs := infra.GetStackManager().Cleanup(context.Background())
		for _, err := range errs {
			fmt.Fprint(os.Stderr, err.Error())
		}
	} else {
		fmt.Fprintf(os.Stderr, "Preserving E2E stacks. ")
	}
}

func SetupConfig() (runner.ConfigMap, error) {
	res := runner.ConfigMap{}
	configs := parseE2EConfigParams()
	if os.Getenv("E2E_PROFILE") == "ci" {
		res.Merge(defaultCIPulumiConfigs)
	} else {
		// use "local" E2E profile for local testing
		// fast-fail if missing required env vars
		_, e2eApiKeyBool := os.LookupEnv("E2E_API_KEY")
		_, e2eAppKeyBool := os.LookupEnv("E2E_APP_KEY")
		_, e2eAwsKeypairNameBool := os.LookupEnv("AWS_KEYPAIR_NAME")
		if !e2eApiKeyBool || !e2eAppKeyBool || !e2eAwsKeypairNameBool {
			return nil, fmt.Errorf("missing required environment variables. Must set `E2E_API_KEY`, `E2E_APP_KEY`, and `AWS_KEYPAIR_NAME` for the local E2E profile")
		} else {
			res.Merge(defaultLocalPulumiConfigs)
		}
	}

	if len(configs) > 0 {
		for _, config := range configs {
			kv := strings.Split(config, "=")
			_, exists := res[kv[0]]
			if !exists {
				res[kv[0]] = auto.ConfigValue{Value: kv[1]}
			} else {
				log.Printf("Config param %s used more than once. Value: %s", kv[0], kv[1])
			}
		}
	}
	log.Printf("Setting up Pulumi E2E stack with configs: %v", res)
	return res, nil
}

func parseE2EConfigParams() []string {
	// "key1=val1 key2=val2"
	configParams := os.Getenv("E2E_CONFIG_PARAMS")
	if len(configParams) < 1 {
		return []string{}
	}
	// ["key1=val1", "key2=val2"]
	configKVs := strings.Split(configParams, " ")
	return configKVs
}

func ListPods(namespace string, client kubernetes.Interface) (*corev1.PodList, error) {
	pods, err := client.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Panicf("error getting pods: %v", err)
		return nil, err
	}
	return pods, nil
}

func ListNodes(namespace string, client kubernetes.Interface) (*corev1.NodeList, error) {
	nodes, err := client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		log.Panicf("error getting pods: %v", err)
	}
	return nodes, nil
}

func NewClientFromKubeconfig(kc map[string]interface{}) (clientcmd.ClientConfig, *rest.Config, *kubernetes.Clientset, error) {
	kubeconfig, err := json.Marshal(kc)
	if err != nil {
		log.Printf("Error encoding kubeconfig json. %v", err)
	}
	clientConfig, err := clientcmd.NewClientConfigFromBytes(kubeconfig)
	if err != nil {
		log.Printf("Error creating client config from kubeconfig. %v", err)
	}
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		log.Printf("Error creating rest config. %v", err)
	}

	clientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		log.Printf("Error creating clientset from rest config. %v", err)
	}

	return clientConfig, restConfig, clientSet, err
}

type K8sExec struct {
	ClientSet     kubernetes.Interface
	RestConfig    *rest.Config
	PodName       string
	ContainerName string
	Namespace     string
}

func (k8s *K8sExec) Exec(command []string) ([]byte, []byte, error) {
	req := k8s.ClientSet.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(k8s.PodName).
		Namespace(k8s.Namespace).
		SubResource("exec")

	req.VersionedParams(&corev1.PodExecOptions{
		Container: k8s.ContainerName,
		Command:   command,
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(k8s.RestConfig, "POST", req.URL())
	if err != nil {
		log.Printf("Failed to exec:%v", err)
		return []byte{}, []byte{}, err
	}
	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		log.Printf("Failed to get result:%v", err)
		return []byte{}, []byte{}, err
	}
	return stdout.Bytes(), stderr.Bytes(), nil
}

func NewK8sExec(clientSet *kubernetes.Clientset, restConfig *rest.Config, podName string, containerName string, namespace string) K8sExec {
	k8sExec := K8sExec{
		ClientSet:     clientSet,
		RestConfig:    restConfig,
		PodName:       podName,
		ContainerName: containerName,
		Namespace:     namespace,
	}
	return k8sExec
}
