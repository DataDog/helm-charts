//go:build e2e_autopilot

package datadog

import (
	"context"
	"fmt"
	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/e2e"
	gcpkubernetes "github.com/DataDog/datadog-agent/test/new-e2e/pkg/provisioners/gcp/kubernetes"
	"github.com/DataDog/helm-charts/test/common"
	"github.com/DataDog/test-infra-definitions/components/datadog/kubernetesagentparams"
	"github.com/DataDog/test-infra-definitions/components/kubernetes/k8sapply"
	"github.com/DataDog/test-infra-definitions/scenarios/gcp/gke"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"strings"
	"testing"
	"time"
)

type gkeAutopilotSuite struct {
	k8sSuite
}

func TestGKEAutopilotSuite(t *testing.T) {
	config, err := common.SetupConfig()
	if err != nil {
		t.Skipf("Skipping test, problem setting up stack config: %s", err)
	}
	assert.NoError(t, err)
	currentDir, err := os.Getwd()
	assert.NoError(t, err)

	e2e.Run(t, &gkeAutopilotSuite{}, e2e.WithProvisioner(gcpkubernetes.GKEProvisioner(
		gcpkubernetes.WithGKEOptions(gke.WithAutopilot()),
		gcpkubernetes.WithWorkloadApp(
			k8sapply.K8sAppDefinition(k8sapply.YAMLWorkload{Name: "nginx", Path: strings.Join([]string{currentDir, "manifests", "autodiscovery-annotation.yaml"}, "/")})),
		gcpkubernetes.WithExtraConfigParams(config),
		gcpkubernetes.WithAgentOptions(
			kubernetesagentparams.WithGKEAutopilot(),
			kubernetesagentparams.WithHelmRepoURL(""),
			kubernetesagentparams.WithHelmChartPath(datadogChartPath()),
			kubernetesagentparams.WithHelmValues(`
datadog:
  kubelet:
    useApiServer: true
    tlsVerify: false
  logs:
    enabled: true
    containerCollectAll: true
clusterChecksRunner:
  enabled: false
`)))))
}

func (s *gkeAutopilotSuite) TestGKEAutopilot() {
	s.T().Log("Running GKE Autopilot test")
	assert.EventuallyWithTf(s.T(), func(c *assert.CollectT) {
		res, err := s.Env().KubernetesCluster.Client().CoreV1().Pods("datadog").List(context.TODO(), metav1.ListOptions{})
		s.Assert().NoError(err)
		var agent corev1.Pod
		containsAgent := false
		for _, pod := range res.Items {
			if strings.Contains(pod.Name, "dda-linux-datadog") && !strings.Contains(pod.Name, "cluster-agent") {
				containsAgent = true
				agent = pod
				break
			}
		}
		assert.True(c, containsAgent, "Agent not found")
		assert.Equal(c, corev1.PodPhase("Running"), agent.Status.Phase, fmt.Sprintf("Agent is not running: %s", agent.Status.Phase))

		var clusterAgent corev1.Pod
		containsClusterAgent := false
		for _, pod := range res.Items {
			if strings.Contains(pod.Name, "cluster-agent") {
				containsClusterAgent = true
				clusterAgent = pod
				break
			}
		}
		assert.True(c, containsClusterAgent, "Cluster Agent not found")
		assert.Equal(c, corev1.PodPhase("Running"), clusterAgent.Status.Phase, fmt.Sprintf("Cluster Agent is not running: %s", clusterAgent.Status.Phase))
	}, 5*time.Minute, 30*time.Second, "GKE Autopilot readiness timed out")
}

func (s *gkeAutopilotSuite) TestGenericK8sAutopilot() {
	s.testGenericK8sAutopilot()
}
