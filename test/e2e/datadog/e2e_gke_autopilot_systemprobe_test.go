//go:build e2e_autopilot_systemprobe

package datadog

import (
	"context"
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
		res, err := v.Env().KubernetesCluster.Client().CoreV1().Pods("datadog").List(context.TODO(), metav1.ListOptions{})
		assert.NoError(c, err)
		if err != nil {
			return
		}

		agent, ok := assertRunningPod(c, res.Items, "Agent", isLinuxNodeAgentPod)
		if !ok {
			return
		}

		var systemProbeStatus *corev1.ContainerStatus
		containsSystemProbe := false
		for i, status := range agent.Status.ContainerStatuses {
			if strings.Contains(status.Name, "system-probe") {
				containsSystemProbe = true
				systemProbeStatus = &agent.Status.ContainerStatuses[i]
				break
			}
		}
		assert.True(c, containsSystemProbe, "System probe container not found")
		assert.NotNil(c, systemProbeStatus, "System probe container status is nil")
		// corev1.ContainerStateRunning is non-nil if the container is running
		if systemProbeStatus != nil {
			assert.NotNil(c, systemProbeStatus.State.Running, "System probe container is not running")
		}

		if _, ok := assertRunningPod(c, res.Items, "Cluster Agent", isClusterAgentPod); !ok {
			return
		}
	}, 5*time.Minute, 30*time.Second, "GKE Autopilot readiness timed out")
}
