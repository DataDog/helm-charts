//go:build e2e

package datadog

import (
	"context"
	"fmt"
	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/runner"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DataDog/test-infra-definitions/components/datadog/kubernetesagentparams"
	"github.com/DataDog/test-infra-definitions/scenarios/gcp/gke"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gcpkubernetes "github.com/DataDog/datadog-agent/test/new-e2e/pkg/provisioners/gcp/kubernetes"

	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/e2e"
	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/environments"
)

type gkeAutopilotSystemProbeSuite struct {
	e2e.BaseSuite[environments.Kubernetes]
}

func TestGKEAutopilotSystemProbeSuite(t *testing.T) {
	gcpPrivateKeyPassword := os.Getenv("E2E_GCP_PRIVATE_KEY_PASSWORD")

	config := runner.ConfigMap{
		"ddinfra:kubernetesVersion":             auto.ConfigValue{Value: "1.32"},
		"ddinfra:env":                           auto.ConfigValue{Value: "gcp/agent-qa"},
		"ddinfra:gcp/defaultPrivateKeyPassword": auto.ConfigValue{Value: gcpPrivateKeyPassword},
	}

	helmValues := `
datadog:
  systemProbe:
    enableTCPQueueLength: true
    enableOOMKill: true
`
	e2e.Run(t, &gkeAutopilotSystemProbeSuite{}, e2e.WithProvisioner(gcpkubernetes.GKEProvisioner(gcpkubernetes.WithGKEOptions(gke.WithAutopilot()), gcpkubernetes.WithAgentOptions(kubernetesagentparams.WithGKEAutopilot(), kubernetesagentparams.WithHelmValues(helmValues)), gcpkubernetes.WithExtraConfigParams(config))), e2e.WithDevMode())
}

func (v *gkeAutopilotSystemProbeSuite) TestGKEAutopilotSystemProbe() {
	v.T().Log("Running GKE test")
	res, _ := v.Env().KubernetesCluster.Client().CoreV1().Pods("datadog").List(context.TODO(), v1.ListOptions{})

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

	assert.EventuallyWithTf(v.T(), func(c *assert.CollectT) {
		var systemProbe corev1.Pod
		containsSystemProbe := false
		for _, pod := range res.Items {
			v.T().Log("Checking pod: ", pod.Name)
			if strings.Contains(pod.Name, "system-probe") {
				containsSystemProbe = true
				systemProbe = pod
				break
			}
		}
		assert.True(v.T(), containsSystemProbe, "System probe container not found")
		assert.Equal(v.T(), corev1.PodPhase("Running"), systemProbe.Status.Phase, fmt.Sprintf("System probe container is not running: %s", systemProbe.Status.Phase))
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
