package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/DataDog/datadog-agent/test/new-e2e/runner"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"os"
)

func SetupConfig() runner.ConfigMap {
	res := runner.ConfigMap{}
	config := os.Getenv("PULUMI_CONFIGS")
	if config != "" {
		var result map[string]map[string]string
		err := json.Unmarshal([]byte(config), &result)
		if err != nil {
			return res
		} else {
			configs := result["config"]
			for key, value := range configs {
				res[key] = auto.ConfigValue{Value: value}
			}
		}
	}
	return res
}

func ListPods(namespace string, client kubernetes.Interface) (*v1.PodList, error) {
	fmt.Println("Get Kubernetes Pods")
	pods, err := client.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		err = fmt.Errorf("error getting pods: %v\n", err)
		return nil, err
	}
	return pods, nil
}
