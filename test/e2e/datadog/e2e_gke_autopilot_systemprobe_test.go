//go:build e2e_autopilot

package datadog

import (
	"context"
	"fmt"
	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/runner"
	"github.com/DataDog/helm-charts/test/common"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DataDog/test-infra-definitions/components/datadog/kubernetesagentparams"
	"github.com/DataDog/test-infra-definitions/scenarios/gcp/gke"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gcpkubernetes "github.com/DataDog/datadog-agent/test/new-e2e/pkg/provisioners/gcp/kubernetes"

	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/e2e"
	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/environments"
)

type gkeAutopilotSystemProbeSuite struct {
	e2e.BaseSuite[environments.Kubernetes]
}

func TestGKEAutopilotSystemProbeSuite(t *testing.T) {
	config, err := common.SetupConfig()
	if err != nil {
		t.Skipf("Skipping test, problem setting up stack config: %s", err)
	}

	gcpPrivateKeyPassword := os.Getenv("E2E_GCP_PRIVATE_KEY_PASSWORD")

	runnerConfig := runner.ConfigMap{
		"ddinfra:kubernetesVersion":             auto.ConfigValue{Value: "1.32"},
		"ddinfra:env":                           auto.ConfigValue{Value: "gcp/agent-qa"},
		"ddinfra:gcp/defaultPrivateKeyPassword": auto.ConfigValue{Value: gcpPrivateKeyPassword},
	}
	runnerConfig.Merge(config)

	helmValues := `
datadog:
  kubelet:
    tlsVerify: false
  systemProbe:
    enableTCPQueueLength: true
    enableOOMKill: true
`
	e2e.Run(t, &gkeAutopilotSystemProbeSuite{}, e2e.WithProvisioner(gcpkubernetes.GKEProvisioner(gcpkubernetes.WithGKEOptions(gke.WithAutopilot()), gcpkubernetes.WithAgentOptions(kubernetesagentparams.WithGKEAutopilot(), kubernetesagentparams.WithHelmValues(helmValues)), gcpkubernetes.WithExtraConfigParams(runnerConfig))), e2e.WithDevMode(), e2e.WithSkipDeleteOnFailure())
}

func (v *gkeAutopilotSystemProbeSuite) TestGKEAutopilotSystemProbe() {
	v.T().Log("Running GKE Autopilot with system-probe test")
	res, _ := v.Env().KubernetesCluster.Client().CoreV1().Pods("datadog").List(context.TODO(), metav1.ListOptions{})

	var agent corev1.Pod
	//var agentPodName string
	containsAgent := false
	for _, pod := range res.Items {
		v.T().Log("Checking pod: ", pod.Name)
		if strings.Contains(pod.Name, "dda-linux-datadog-") && !strings.Contains(pod.Name, "cluster-agent") {
			containsAgent = true
			agent = pod
			//agentPodName = pod.Name
			break
		}
	}
	assert.True(v.T(), containsAgent, "Agent not found")
	assert.Equal(v.T(), corev1.PodPhase("Running"), agent.Status.Phase, fmt.Sprintf("Agent is not running: %s", agent.Status.Phase))

	assert.EventuallyWithTf(v.T(), func(c *assert.CollectT) {
		//agent, err := v.Env().KubernetesCluster.Client().CoreV1().Pods("datadog").Get(context.TODO(), agentPodName, metav1.GetOptions{})
		//assert.NoError(v.T(), err)

		var systemProbeState *corev1.ContainerStatus
		containsSystemProbe := false
		for i, status := range agent.Status.ContainerStatuses {
			if strings.Contains(status.Name, "system-probe") {
				containsSystemProbe = true
				systemProbeState = &agent.Status.ContainerStatuses[i]
				break
			}
		}
		assert.True(v.T(), containsSystemProbe, "System probe container not found")
		assert.NotNil(v.T(), systemProbeState, "System probe container status is nil")
		v.T().Log("WHAT IS THE SYSTEM PROBE RUNNING STATE: ", systemProbeState.State.Running.String())
		// corev1.ContainerStateRunning is non-nil if the container is running
		assert.NotNil(v.T(), systemProbeState.State.Running, "System probe container is not running")
	}, 5*time.Minute, 30*time.Second, "system-probe readiness timed out")

	var clusterAgent corev1.Pod
	containsClusterAgent := false
	for _, pod := range res.Items {
		if strings.Contains(pod.Name, "cluster-agent") {
			containsClusterAgent = true
			clusterAgent = pod
			break
		}
	}
	assert.True(v.T(), containsClusterAgent, "Cluster Agent not found")
	assert.Equal(v.T(), corev1.PodPhase("Running"), clusterAgent.Status.Phase, fmt.Sprintf("Cluster Agent is not running: %s", clusterAgent.Status.Phase))

}
