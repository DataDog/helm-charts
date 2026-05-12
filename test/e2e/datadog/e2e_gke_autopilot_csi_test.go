//go:build e2e_autopilot_csi

package datadog

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
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

	e2e.Run(t, &gkeAutopilotCSISuite{}, e2e.WithProvisioner(gcpkubernetes.GKEProvisioner(
		gcpkubernetes.WithGKEOptions(gke.WithAutopilot()),
		gcpkubernetes.WithAgentOptions(
			kubernetesagentparams.WithGKEAutopilot(),
			kubernetesagentparams.WithHelmRepoURL(""),
			kubernetesagentparams.WithHelmChartPath(datadogChartPath()),
		),
		gcpkubernetes.WithExtraConfigParams(config))))
}

func (v *gkeAutopilotCSISuite) TestGKEAutopilotCSI() {
	v.T().Log("Running GKE Autopilot CSI driver test")

	kubeconfigPath := v.writeKubeconfig()
	defer os.Remove(kubeconfigPath)

	v.logClusterInfo()

	chartPath, err := filepath.Abs("../../../charts/datadog-csi-driver")
	require.NoError(v.T(), err, "Failed to get chart path")

	v.helmInstall(chartPath, kubeconfigPath)
	v.waitForPodsReady()
}

// writeKubeconfig writes the cluster kubeconfig to a temp file and returns the path.
func (v *gkeAutopilotCSISuite) writeKubeconfig() string {
	kubeconfigFile, err := os.CreateTemp("", "gke-kubeconfig-")
	require.NoError(v.T(), err, "Failed to create kubeconfig temp file")

	kubeconfig := v.Env().KubernetesCluster.KubeConfig
	require.NoError(v.T(), os.WriteFile(kubeconfigFile.Name(), []byte(kubeconfig), 0600), "Failed to write kubeconfig")
	require.NoError(v.T(), kubeconfigFile.Close(), "Failed to close kubeconfig file")

	return kubeconfigFile.Name()
}

// logClusterInfo logs the Kubernetes server version for diagnosing which
// Autopilot detection path the helm chart takes.
func (v *gkeAutopilotCSISuite) logClusterInfo() {
	serverVersion, err := v.Env().KubernetesCluster.Client().Discovery().ServerVersion()
	if err != nil {
		v.T().Logf("Failed to get server version: %v", err)
		return
	}
	v.T().Logf("Kubernetes server version: %s (Major=%s Minor=%s)", serverVersion.GitVersion, serverVersion.Major, serverVersion.Minor)
}

// helmInstall installs the CSI driver chart via helm.
func (v *gkeAutopilotCSISuite) helmInstall(chartPath, kubeconfigPath string) {
	helmCmd := exec.Command("helm", "install", "datadog-csi-driver", chartPath,
		"--kubeconfig", kubeconfigPath,
		"--namespace", "datadog-agent", "--create-namespace")
	output, err := helmCmd.CombinedOutput()
	v.T().Logf("Helm install output: %s", string(output))
	require.NoError(v.T(), err, "Helm install failed")
}

// waitForPodsReady polls for CSI driver pods and asserts that every container
// in every pod is ready with zero restarts.
func (v *gkeAutopilotCSISuite) waitForPodsReady() {
	assert.EventuallyWithTf(v.T(), func(c *assert.CollectT) {
		listOptions := metav1.ListOptions{LabelSelector: "app=datadog-csi-driver-node-server"}
		res, err := v.Env().KubernetesCluster.Client().CoreV1().Pods("datadog-agent").List(context.TODO(), listOptions)
		require.NoError(c, err)

		assert.True(c, len(res.Items) > 0, "No CSI driver pods found")

		allReady := true
		for _, pod := range res.Items {
			assert.NotEmpty(c, pod.Status.ContainerStatuses, "pod %s has no container statuses yet", pod.Name)
			for _, cs := range pod.Status.ContainerStatuses {
				if !cs.Ready || cs.RestartCount > 0 {
					allReady = false
					v.T().Logf("Pod %s container %s: ready=%v restarts=%d", pod.Name, cs.Name, cs.Ready, cs.RestartCount)
					v.logContainerState(cs)
				}
				assert.True(c, cs.Ready, "container %s in pod %s is not ready", cs.Name, pod.Name)
				assert.Zero(c, cs.RestartCount, "container %s in pod %s has restarted %d times", cs.Name, pod.Name, cs.RestartCount)
			}
		}

		if !allReady {
			v.logFailureDiagnostics(res.Items)
		}
	}, 5*time.Minute, 30*time.Second, "CSI Driver readiness timed out")
}

// logContainerState logs the waiting/terminated state of an unhealthy container.
func (v *gkeAutopilotCSISuite) logContainerState(cs corev1.ContainerStatus) {
	if cs.State.Waiting != nil {
		v.T().Logf("  state: waiting reason=%s message=%s", cs.State.Waiting.Reason, cs.State.Waiting.Message)
	}
	if cs.State.Terminated != nil {
		v.T().Logf("  state: terminated reason=%s exitCode=%d", cs.State.Terminated.Reason, cs.State.Terminated.ExitCode)
	}
	if cs.LastTerminationState.Terminated != nil {
		v.T().Logf("  lastTermination: reason=%s exitCode=%d", cs.LastTerminationState.Terminated.Reason, cs.LastTerminationState.Terminated.ExitCode)
	}
}

// logFailureDiagnostics collects container logs and namespace events when pods
// are unhealthy, to help diagnose the root cause.
func (v *gkeAutopilotCSISuite) logFailureDiagnostics(pods []corev1.Pod) {
	tailLines := int64(30)
	for _, pod := range pods {
		for _, container := range pod.Spec.Containers {
			for _, previous := range []bool{false, true} {
				label := "current"
				if previous {
					label = "previous"
				}
				logs, err := v.Env().KubernetesCluster.Client().CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
					Container: container.Name,
					TailLines: &tailLines,
					Previous:  previous,
				}).Do(context.TODO()).Raw()
				if err != nil {
					continue
				}
				if len(logs) > 0 {
					v.T().Logf("%s logs %s/%s:\n%s", label, pod.Name, container.Name, string(logs))
				}
			}
		}
	}

	events, err := v.Env().KubernetesCluster.Client().CoreV1().Events("datadog-agent").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return
	}
	for _, event := range events.Items {
		if event.Type != "Normal" {
			v.T().Logf("Event %s %s/%s: %s - %s", event.Type, event.InvolvedObject.Kind, event.InvolvedObject.Name, event.Reason, event.Message)
		}
	}
}
