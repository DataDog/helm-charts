//go:build e2e_autopilot

package datadog

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DataDog/datadog-agent/test/e2e-framework/components/datadog/kubernetesagentparams"
	"github.com/DataDog/datadog-agent/test/e2e-framework/components/kubernetes/k8sapply"
	"github.com/DataDog/datadog-agent/test/e2e-framework/scenarios/gcp/gke"
	"github.com/DataDog/datadog-agent/test/e2e-framework/testing/e2e"
	gcpkubernetes "github.com/DataDog/datadog-agent/test/e2e-framework/testing/provisioners/gcp/kubernetes"
	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		assert.NoError(c, err)
		if err != nil {
			return
		}
		if _, ok := assertRunningPod(c, res.Items, "Agent", isLinuxNodeAgentPod); !ok {
			return
		}

		if _, ok := assertRunningPod(c, res.Items, "Cluster Agent", isClusterAgentPod); !ok {
			return
		}
	}, 5*time.Minute, 30*time.Second, "GKE Autopilot readiness timed out")
}

func (s *gkeAutopilotSuite) TestGenericK8sAutopilot() {
	s.testGenericK8sAutopilot()
}
