//go:build e2e

package datadog

import (
	"context"
	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gcpkubernetes "github.com/DataDog/datadog-agent/test/new-e2e/pkg/provisioners/gcp/kubernetes"

	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/e2e"
	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/environments"
)

type gkeSuite struct {
	e2e.BaseSuite[environments.Kubernetes]
}

func TestGKESuite(t *testing.T) {
	runnerConfig, err := common.SetupConfig()
	if err != nil {
		t.Skipf("Skipping test, problem setting up stack config: %s", err)
	}

	e2e.Run(t, &gkeSuite{}, e2e.WithProvisioner(gcpkubernetes.GKEProvisioner(gcpkubernetes.WithExtraConfigParams(runnerConfig))), e2e.WithSkipDeleteOnFailure(), e2e.WithDevMode())
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
	}, 5*time.Minute, 30*time.Second, "GKE Autopilot readiness timed out")
}
