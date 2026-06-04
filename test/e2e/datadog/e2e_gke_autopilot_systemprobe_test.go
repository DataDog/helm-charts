//go:build e2e_autopilot_systemprobe

package datadog

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/DataDog/helm-charts/test/common"

	"github.com/DataDog/datadog-agent/test/e2e-framework/components/datadog/kubernetesagentparams"
	"github.com/DataDog/datadog-agent/test/e2e-framework/scenarios/gcp/gke"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gcpkubernetes "github.com/DataDog/datadog-agent/test/e2e-framework/testing/provisioners/gcp/kubernetes"

	"github.com/DataDog/datadog-agent/test/e2e-framework/testing/e2e"
)

type gkeAutopilotSystemProbeSuite struct {
	k8sSuite
}

func TestGKEAutopilotSystemProbeSuite(t *testing.T) {
	config, err := common.SetupConfig()
	if err != nil {
		t.Skipf("Skipping test, problem setting up stack config: %s", err)
	}

	helmValues := `
datadog:
  kubelet:
    tlsVerify: false
  systemProbe:
    enableTCPQueueLength: true
    enableOOMKill: true
`
	// Override the default stack name to keep the Pulumi stack name short enough
	// (<= 63 chars) for GCP resource labels. e2e-framework's DefaultResourceTags
	// adds a `stack` label valued with Ctx().Stack(), and GCP rejects label
	// values > 63 bytes. The default `e2e-<TypeName>-<hash>` combined with the
	// CI namePrefix `ci-${CI_PIPELINE_ID}-${CI_JOB_ID}-` overflows this limit
	// for this suite. Remove this override once the upstream fix lands (cf.
	// pending PR on DataDog/datadog-agent to truncate GCP label values).
	e2e.Run(t, &gkeAutopilotSystemProbeSuite{},
		e2e.WithStackName("gke-ap-sysprobe"),
		e2e.WithProvisioner(gcpkubernetes.GKEProvisioner(
			gcpkubernetes.WithGKEOptions(gke.WithAutopilot()),
			gcpkubernetes.WithAgentOptions(
				kubernetesagentparams.WithGKEAutopilot(),
				kubernetesagentparams.WithHelmRepoURL(""),
				kubernetesagentparams.WithHelmChartPath(datadogChartPath()),
				kubernetesagentparams.WithHelmValues(helmValues),
			),
			gcpkubernetes.WithExtraConfigParams(config))))
}

func (v *gkeAutopilotSystemProbeSuite) TestGKEAutopilotSystemProbe() {
	v.T().Log("Running GKE Autopilot with system-probe test")
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

		var systemProbeStatus *corev1.ContainerStatus
		containsSystemProbe := false
		for i, status := range agent.Status.ContainerStatuses {
			if strings.Contains(status.Name, "system-probe") {
				containsSystemProbe = true
				systemProbeStatus = &agent.Status.ContainerStatuses[i]
				break
			}
		}
		assert.True(v.T(), containsSystemProbe, "System probe container not found")
		assert.NotNil(v.T(), systemProbeStatus, "System probe container status is nil")
		// corev1.ContainerStateRunning is non-nil if the container is running
		assert.NotNil(v.T(), systemProbeStatus.State.Running, "System probe container is not running")

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
	}, 5*time.Minute, 30*time.Second, "GKE Autopilot readiness timed out")
}
