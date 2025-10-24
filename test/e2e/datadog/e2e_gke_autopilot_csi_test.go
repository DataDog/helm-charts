//go:build e2e_autopilot_csi

package datadog

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/DataDog/helm-charts/test/common"

	"github.com/DataDog/test-infra-definitions/components/datadog/kubernetesagentparams"
	"github.com/DataDog/test-infra-definitions/scenarios/gcp/gke"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gcpkubernetes "github.com/DataDog/datadog-agent/test/new-e2e/pkg/provisioners/gcp/kubernetes"

	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/e2e"
	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/environments"
)

type gkeAutopilotCSISuite struct {
	e2e.BaseSuite[environments.Kubernetes]
}

func TestGKEAutopilotCSISuite(t *testing.T) {
	config, err := common.SetupConfig()
	if err != nil {
		t.Skipf("Skipping test, problem setting up stack config: %s", err)
	}

	helmValues := `
datadog:
  kubelet:
    tlsVerify: false
  csi:
    enabled: true
`
	e2e.Run(t, &gkeAutopilotCSISuite{}, e2e.WithProvisioner(gcpkubernetes.GKEProvisioner(gcpkubernetes.WithGKEOptions(gke.WithAutopilot()), gcpkubernetes.WithAgentOptions(kubernetesagentparams.WithGKEAutopilot(), kubernetesagentparams.WithHelmValues(helmValues)), gcpkubernetes.WithExtraConfigParams(config))))
}

func (v *gkeAutopilotCSISuite) TestGKEAutopilotCSI() {
	v.T().Log("Running GKE Autopilot CSI driver test")
	assert.EventuallyWithTf(v.T(), func(c *assert.CollectT) {
		res, _ := v.Env().KubernetesCluster.Client().CoreV1().Pods("datadog").List(context.TODO(), metav1.ListOptions{})

		var agent corev1.Pod
		containsAgent := false
		for _, pod := range res.Items {
			if strings.Contains(pod.Name, "dda-linux-datadog") && !strings.Contains(pod.Name, "cluster-agent") {
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

		var csiDriver corev1.Pod
		containsCsiDriver := false
		for _, pod := range res.Items {
			if strings.Contains(pod.Name, "csi-driver") {
				containsCsiDriver = true
				csiDriver = pod
				break
			}
		}
		assert.True(v.T(), containsCsiDriver, "CSI Driver not found")
		assert.Equal(v.T(), corev1.PodPhase("Running"), csiDriver.Status.Phase, fmt.Sprintf("CSI Driver is not running: %s", csiDriver.Status.Phase))

	}, 5*time.Minute, 30*time.Second, "GKE Autopilot readiness timed out")
}
