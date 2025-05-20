//go:build e2e

package datadog

import (
	"context"
	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/e2e"
	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/environments"
	awskubernetes "github.com/DataDog/datadog-agent/test/new-e2e/pkg/provisioners/aws/kubernetes"
	"github.com/DataDog/test-infra-definitions/scenarios/aws/eks"
	"strings"
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type eksSuite struct {
	e2e.BaseSuite[environments.Kubernetes]
}

func TestEKSSuite(t *testing.T) {
	// Create pulumi EKS stack with latest version of the datadog/datadog helm chart
	config, err := common.SetupConfig()
	if err != nil {
		t.Skipf("Skipping test, problem setting up stack config: %s", err)
	}

	e2e.Run(t, &eksSuite{}, e2e.WithProvisioner(awskubernetes.EKSProvisioner(
		awskubernetes.WithEKSOptions(
			eks.WithLinuxNodeGroup(),
		),
		awskubernetes.WithExtraConfigParams(config),
	)))
}

func (s *eksSuite) TestEKS() {
	s.T().Log("Running EKS test")

	res, _ := s.Env().KubernetesCluster.Client().CoreV1().Pods("datadog").List(context.TODO(), metav1.ListOptions{})

	var agent corev1.Pod
	containsAgent := false
	for _, pod := range res.Items {
		if strings.Contains(pod.Name, "dda-linux-datadog") && !strings.Contains(pod.Name, "cluster-agent") {
			containsAgent = true
			agent = pod
			break
		}
	}
	assert.True(s.T(), containsAgent, "Agent not found")

	stdout, stderr, err := s.Env().KubernetesCluster.KubernetesClient.
		PodExec("datadog", agent.Name, "agent", []string{"agent", "status"})
	require.NoError(s.T(), err)
	assert.Empty(s.T(), stderr)
	assert.NotEmpty(s.T(), stdout)

	var clusterAgent corev1.Pod
	containsClusterAgent := false
	for _, pod := range res.Items {
		if strings.Contains(pod.Name, "cluster-agent") {
			containsClusterAgent = true
			clusterAgent = pod
			break
		}
	}
	assert.True(s.T(), containsClusterAgent, "Cluster Agent not found")

	stdout, stderr, err = s.Env().KubernetesCluster.KubernetesClient.
		PodExec("datadog", clusterAgent.Name, "cluster-agent", []string{"agent", "status"})
	require.NoError(s.T(), err)
	assert.Empty(s.T(), stderr)
	assert.NotEmpty(s.T(), stdout)
}
