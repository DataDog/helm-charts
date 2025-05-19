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
	assert.EventuallyWithTf(v.T(), func(c *assert.CollectT) {
		res, _ := v.Env().KubernetesCluster.Client().CoreV1().Pods("datadog").List(context.TODO(), metav1.ListOptions{})

		var agent corev1.Pod
		containsAgent := false
		for _, pod := range res.Items {
			v.T().Log("Checking pod: ", pod.Name)
			if strings.Contains(pod.Name, "agent") {
				containsAgent = true
				agent = pod
				break
			}
		}
		assert.True(v.T(), containsAgent, "Agent not found")
		assert.Equal(v.T(), corev1.PodPhase("Running"), agent.Status.Phase, fmt.Sprintf("Agent is not running: %s", agent.Status.Phase))

		var systemProbeState corev1.ContainerState
		containsSystemProbe := false
		for _, status := range agent.Status.ContainerStatuses {
			v.T().Log("Checking pod: ", status.Name)
			if strings.Contains(status.Name, "system-probe") {
				containsSystemProbe = true
				systemProbeState = status.State
				break
			}
		}
		assert.True(v.T(), containsSystemProbe, "System probe container not found")
		assert.Equal(v.T(), "Running", systemProbeState.Running.String(), fmt.Sprintf("System probe container is not running: %s", systemProbeState.Running.String()))

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
	}, 5*time.Minute, 30*time.Second, "system-probe readiness timed out")
}
