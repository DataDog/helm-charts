//go:build e2e

package datadog

import (
	"context"
	"github.com/DataDog/helm-charts/test/common"
	"github.com/DataDog/test-infra-definitions/components/datadog/kubernetesagentparams"
	"github.com/DataDog/test-infra-definitions/components/kubernetes/k8sapply"
	"github.com/stretchr/testify/require"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gcpkubernetes "github.com/DataDog/datadog-agent/test/new-e2e/pkg/provisioners/gcp/kubernetes"

	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/e2e"
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
clusterChecksRunner:
  enabled: false
`)))))
}

func (v *gkeSuite) TestGKE() {
	v.T().Log("Running GKE test")
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

		stdout, stderr, err := v.Env().KubernetesCluster.KubernetesClient.
			PodExec("datadog", agent.Name, "agent", []string{"agent", "status"})
		require.NoError(v.T(), err)
		assert.Empty(v.T(), stderr)
		assert.NotEmpty(v.T(), stdout)

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

		stdout, stderr, err = v.Env().KubernetesCluster.KubernetesClient.
			PodExec("datadog", clusterAgent.Name, "cluster-agent", []string{"agent", "status"})
		require.NoError(v.T(), err)
		assert.Empty(v.T(), stderr)
		assert.NotEmpty(v.T(), stdout)
	}, 5*time.Minute, 30*time.Second, "GKE readiness timed out")
}

func (v *gkeSuite) TestGenericK8s() {
	v.testGenericK8s()
}
