//go:build e2e

package datadog

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DataDog/datadog-agent/test/e2e-framework/components/datadog/kubernetesagentparams"
	"github.com/DataDog/datadog-agent/test/e2e-framework/components/kubernetes/k8sapply"
	"github.com/DataDog/helm-charts/test/common"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gcpkubernetes "github.com/DataDog/datadog-agent/test/e2e-framework/testing/provisioners/gcp/kubernetes"

	"github.com/DataDog/datadog-agent/test/e2e-framework/testing/e2e"
)

type gkeSuite struct {
	k8sSuite
}

func TestGKESuite(t *testing.T) {
	runnerConfig, err := common.SetupConfig()
	if err != nil {
		t.Skipf("Skipping test, problem setting up stack config: %s", err)
	}

	assert.NoError(t, err)
	currentDir, err := os.Getwd()
	assert.NoError(t, err)

	e2e.Run(t, &gkeSuite{}, e2e.WithProvisioner(gcpkubernetes.GKEProvisioner(
		gcpkubernetes.WithWorkloadApp(
			k8sapply.K8sAppDefinition(k8sapply.YAMLWorkload{Name: "nginx", Path: strings.Join([]string{currentDir, "manifests", "autodiscovery-annotation.yaml"}, "/")})),
		gcpkubernetes.WithExtraConfigParams(runnerConfig),
		gcpkubernetes.WithAgentOptions(
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
providers:
  gke:
    cos: true
clusterChecksRunner:
  enabled: false
`)))))
}

func (v *gkeSuite) TestGKE() {
	v.T().Log("Running GKE test")
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

		stdout, stderr, err := v.Env().KubernetesCluster.KubernetesClient.
			PodExec("datadog", agent.Name, "agent", []string{"agent", "status"})
		assert.NoError(c, err)
		if err != nil {
			return
		}
		assert.Empty(c, stderr)
		assert.NotEmpty(c, stdout)

		clusterAgent, ok := assertRunningPod(c, res.Items, "Cluster Agent", isClusterAgentPod)
		if !ok {
			return
		}

		stdout, stderr, err = v.Env().KubernetesCluster.KubernetesClient.
			PodExec("datadog", clusterAgent.Name, "cluster-agent", []string{"agent", "status"})
		assert.NoError(c, err)
		if err != nil {
			return
		}
		assert.Empty(c, stderr)
		assert.NotEmpty(c, stdout)
	}, 5*time.Minute, 30*time.Second, "GKE readiness timed out")
}

func (v *gkeSuite) TestGenericK8s() {
	v.testGenericK8s()
}
