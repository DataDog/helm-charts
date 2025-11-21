//go:build e2e_autopilot_csi

package datadog

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/DataDog/helm-charts/test/common"

	"github.com/DataDog/test-infra-definitions/components/datadog/kubernetesagentparams"
	"github.com/DataDog/test-infra-definitions/scenarios/gcp/gke"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	e2e.Run(t, &gkeAutopilotCSISuite{}, e2e.WithProvisioner(gcpkubernetes.GKEProvisioner(gcpkubernetes.WithGKEOptions(gke.WithAutopilot()), gcpkubernetes.WithAgentOptions(kubernetesagentparams.WithGKEAutopilot()), gcpkubernetes.WithExtraConfigParams(config))))
}

func (v *gkeAutopilotCSISuite) TestGKEAutopilotCSI() {

	v.T().Log("Running GKE Autopilot CSI driver test")

	// Write kubeconfig to temp file
	kubeconfigFile, err := os.CreateTemp("", "gke-kubeconfig-")
	if err != nil {
		v.T().Fatalf("Failed to create kubeconfig temp file: %v", err)
	}
	defer os.Remove(kubeconfigFile.Name())

	kubeconfig := v.Env().KubernetesCluster.KubeConfig
	if err := os.WriteFile(kubeconfigFile.Name(), []byte(kubeconfig), 0600); err != nil {
		v.T().Fatalf("Failed to write kubeconfig: %v", err)
	}
	if err := kubeconfigFile.Close(); err != nil {
		v.T().Fatalf("Failed to close kubeconfig file: %v", err)
	}

	// Installing the datadog repository
	helmCmd := exec.Command("helm", "repo", "add", "datadog", "https://helm.datadoghq.com")
	output, err := helmCmd.CombinedOutput()
	v.T().Logf("Helm output: %s", string(output))
	if err != nil {
		v.T().Fatalf("Helm repo add failed: %v", err)
	}
	v.T().Log("Datadog repository added")

	// Installing the csi driver via helm
	v.T().Log("Installing CSI driver")
	helmCmd = exec.Command("helm", "install", "datadog-csi-driver", "datadog/datadog-csi-driver",
		"--kubeconfig", kubeconfigFile.Name(),
		"--namespace", "datadog-agent", "--create-namespace")

	output, err = helmCmd.CombinedOutput()
	v.T().Logf("Helm output: %s", string(output))
	if err != nil {
		v.T().Fatalf("Helm install failed: %v", err)
	}
	v.T().Log("CSI driver installed")
	
	// Check if CSI driver pods exist
	assert.EventuallyWithTf(v.T(), func(c *assert.CollectT) {
		listOptions := metav1.ListOptions{LabelSelector: "app=datadog-csi-driver-node-server"}
		res, err := v.Env().KubernetesCluster.Client().CoreV1().Pods("datadog-agent").List(context.TODO(), listOptions)
		require.NoError(c, err)

		assert.True(c, len(res.Items) > 0, "No CSI driver pods found")

		runningPods := 0
		for _, pod := range res.Items {
			if pod.Status.Phase == corev1.PodPhase("Running") {
				runningPods++
			}
		}

		assert.Equal(c, len(res.Items), runningPods, "All CSI driver pods are not running")

	}, 5*time.Minute, 30*time.Second, "CSI Driver readiness timed out")

}
