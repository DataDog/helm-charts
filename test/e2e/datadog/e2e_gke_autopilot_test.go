//go:build e2e_autopilot

package datadog

import (
	"context"
	"fmt"
	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/runner"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"os"
	"strings"
	"testing"

	"github.com/DataDog/test-infra-definitions/components/datadog/kubernetesagentparams"
	"github.com/DataDog/test-infra-definitions/scenarios/gcp/gke"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gcpkubernetes "github.com/DataDog/datadog-agent/test/new-e2e/pkg/provisioners/gcp/kubernetes"

	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/e2e"
	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/environments"
)

type gkeAutopilotSuite struct {
	e2e.BaseSuite[environments.Kubernetes]
}

func TestGKEAutopilotSuite(t *testing.T) {
	gcpPrivateKeyPassword := os.Getenv("E2E_GCP_PRIVATE_KEY_PASSWORD")

	config := runner.ConfigMap{
		"ddinfra:kubernetesVersion":             auto.ConfigValue{Value: "1.32"},
		"ddinfra:env":                           auto.ConfigValue{Value: "gcp/agent-qa"},
		"ddinfra:gcp/defaultPrivateKeyPassword": auto.ConfigValue{Value: gcpPrivateKeyPassword},
	}
	e2e.Run(t, &gkeAutopilotSuite{}, e2e.WithProvisioner(gcpkubernetes.GKEProvisioner(gcpkubernetes.WithGKEOptions(gke.WithAutopilot()), gcpkubernetes.WithAgentOptions(kubernetesagentparams.WithGKEAutopilot()), gcpkubernetes.WithExtraConfigParams(config))), e2e.WithDevMode())
}

func (v *gkeAutopilotSuite) TestGKEAutopilot() {
	v.T().Log("Running GKE test")
	res, _ := v.Env().KubernetesCluster.Client().CoreV1().Pods("datadog").List(context.TODO(), v1.ListOptions{})

	var agent corev1.Pod
	containsAgent := false
	for _, pod := range res.Items {
		if strings.Contains(pod.Name, "agent") {
			containsAgent = true
			agent = pod
			break
		}
	}
	assert.True(v.T(), containsAgent, "Agent not found")
	assert.Equal(v.T(), corev1.PodPhase("Running"), agent.Status.Phase, fmt.Sprintf("Agent is not running: %s", agent.Status.Phase))

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
