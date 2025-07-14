//go:build e2e_autopilot

package datadog

import (
	"context"
	"fmt"
	"github.com/DataDog/datadog-agent/test/fakeintake/aggregator"
	"github.com/DataDog/datadog-agent/test/fakeintake/client"
	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/runner"
	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/require"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DataDog/test-infra-definitions/components/datadog/kubernetesagentparams"
	"github.com/DataDog/test-infra-definitions/scenarios/gcp/gke"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gcpkubernetes "github.com/DataDog/datadog-agent/test/new-e2e/pkg/provisioners/gcp/kubernetes"

	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/e2e"
	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/environments"
)

var (
	matchTags = []*regexp.Regexp{regexp.MustCompile("kube_container_name:.*")}
	matchOpts = []client.MatchOpt[*aggregator.MetricSeries]{client.WithMatchingTags[*aggregator.MetricSeries](matchTags)}
)

type gkeAutopilotSuite struct {
	e2e.BaseSuite[environments.Kubernetes]
	defaultConfig runner.ConfigMap
}

func (s *gkeAutopilotSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
	config, err := common.SetupConfig()
	if err != nil {
		s.Error(err)
	}
	s.defaultConfig = config

}

func TestGKEAutopilotSuite(t *testing.T) {
	config, err := common.SetupConfig()
	if err != nil {
		t.Skipf("Skipping test, problem setting up stack config: %s", err)
	}

	e2e.Run(t, &gkeAutopilotSuite{}, e2e.WithProvisioner(gcpkubernetes.GKEProvisioner(gcpkubernetes.WithGKEOptions(gke.WithAutopilot()), gcpkubernetes.WithAgentOptions(kubernetesagentparams.WithGKEAutopilot()), gcpkubernetes.WithExtraConfigParams(config))))
}

func (s *gkeAutopilotSuite) TestGKEAutopilot() {
	s.T().Log("Running GKE Autopilot test")
	assert.EventuallyWithTf(s.T(), func(c *assert.CollectT) {
		res, _ := s.Env().KubernetesCluster.Client().CoreV1().Pods("datadog").List(context.TODO(), metav1.ListOptions{})

		var agent corev1.Pod
		containsAgent := false
		for _, pod := range res.Items {
			if strings.Contains(pod.Name, "dda-linux-datadog") && !strings.Contains(pod.Name, "cluster-agent") {
				containsAgent = true
				agent = pod
				break
			}
		}
		assert.True(s.T(), containsAgent, "Agent not found")
		assert.Equal(s.T(), corev1.PodPhase("Running"), agent.Status.Phase, fmt.Sprintf("Agent is not running: %s", agent.Status.Phase))

		var clusterAgent corev1.Pod
		containsClusterAgent := false
		for _, pod := range res.Items {
			if strings.Contains(pod.Name, "cluster-agent") {
				containsClusterAgent = true
				clusterAgent = pod
				break
			}
		}
		assert.True(s.T(), containsClusterAgent, "Cluster Agent not found")
		assert.Equal(s.T(), corev1.PodPhase("Running"), clusterAgent.Status.Phase, fmt.Sprintf("Cluster Agent is not running: %s", clusterAgent.Status.Phase))
	}, 5*time.Minute, 30*time.Second, "GKE Autopilot readiness timed out")
}

func (s *gkeAutopilotSuite) TestGenericK8s() {
	s.T().Run("Kubelet check works", func(t *testing.T) {
		s.Assert().EventuallyWithT(func(c *assert.CollectT) {
			kubeletCheckRun, err := s.Env().FakeIntake.Client().GetCheckRun("kubernetes.kubelet.check")
			assert.NoError(c, err)
			assert.NotEmpty(c, kubeletCheckRun)
			assert.Equal(c, 0, kubeletCheckRun[0].Status, "kubelet check status should be running")

			kubeletMetricSeries, err := s.Env().FakeIntake.Client().FilterMetrics("kubernetes.cpu.usage.total", matchOpts...)
			s.Assert().NoError(err)
			s.Assert().NotEmptyf(kubeletMetricSeries, fmt.Sprintf("expected Kubelet check series to not be empty: %s", err))

		}, 1*time.Minute, 15*time.Second, "could not validate kubelet check in time")

		s.Assert().EventuallyWithT(func(c *assert.CollectT) {
			s.verifyKSMCheck(c)
		}, 1*time.Minute, 15*time.Second, "could not validate KSM check in time")

	})

	s.T().Run("KSM core check works", func(t *testing.T) {
		s.Assert().EventuallyWithT(func(c *assert.CollectT) {
			s.verifyKSMCheck(c)
		}, 1*time.Minute, 15*time.Second, "could not validate KSM check in time")

	})

	s.T().Run("KSM check works cluster check runner", func(t *testing.T) {
		s.UpdateEnv(gcpkubernetes.GKEProvisioner(gcpkubernetes.WithGKEOptions(gke.WithAutopilot()), gcpkubernetes.WithAgentOptions(
			kubernetesagentparams.WithGKEAutopilot(),
			kubernetesagentparams.WithHelmValues(`
datadog:
  kubeStateMetricsCore:
    useClusterCheckRunners: true
clusterChecksRunner:
  enabled: true
`),
		),
			gcpkubernetes.WithExtraConfigParams(s.defaultConfig)))

		err := s.Env().FakeIntake.Client().FlushServerAndResetAggregators()
		s.Assert().NoError(err)

		s.Assert().EventuallyWithTf(func(c *assert.CollectT) {
			res, _ := s.Env().KubernetesCluster.Client().CoreV1().Pods("datadog").List(context.TODO(), metav1.ListOptions{})

			var agent corev1.Pod
			containsCCR := false
			for _, pod := range res.Items {
				s.T().Log("CHECKING POD: ", pod.Name)
				if strings.Contains(pod.Name, "dda-linux-datadog-cluster-check-runner") {
					containsCCR = true
					agent = pod
					break
				}
			}
			assert.True(s.T(), containsCCR, "Agent cluster check runner not found")
			assert.Equal(s.T(), corev1.PodPhase("Running"), agent.Status.Phase, fmt.Sprintf("Agent cluster check runner is not running: %s", agent.Status.Phase))

			s.verifyKSMCheck(c)
		}, 10*time.Minute, 15*time.Second, "could not validate kubernetes_state_core (cluster check on CCR) check in time")
	})

	//s.T().Run("Autodiscovery works", func(t *testing.T) {
	//	ddaConfigPath, err := common.GetAbsPath(common.DdaMinimalPath)
	//	assert.NoError(s.T(), err)
	//
	//	ddaOpts := []agentwithoperatorparams.Option{
	//		agentwithoperatorparams.WithDDAConfig(agentwithoperatorparams.DDAConfig{Name: "dda-autodiscovery", YamlFilePath: ddaConfigPath}),
	//	}
	//	ddaOpts = append(ddaOpts, defaultDDAOpts...)
	//
	//	provisionerOptions := []provisioners.KubernetesProvisionerOption{
	//		provisioners.WithTestName("e2e-operator-autodiscovery"),
	//		provisioners.WithDDAOptions(ddaOpts...),
	//		provisioners.WithYAMLWorkload(provisioners.YAMLWorkload{Name: "nginx", Path: strings.Join([]string{common.ManifestsPath, "autodiscovery-annotation.yaml"}, "/")}),
	//		provisioners.WithLocal(s.local),
	//	}
	//	provisionerOptions = append(provisionerOptions, defaultProvisionerOpts...)
	//
	//	// Add nginx with annotations
	//	s.UpdateEnv(provisioners.KubernetesProvisioner(provisionerOptions...))
	//
	//	err = s.Env().FakeIntake.Client().FlushServerAndResetAggregators()
	//	s.Assert().NoError(err)
	//
	//	s.Assert().EventuallyWithTf(func(c *assert.CollectT) {
	//		utils.VerifyNumPodsForSelector(s.T(), c, common.NamespaceName, s.Env().KubernetesCluster.Client(), 1, "app=nginx")
	//		utils.VerifyAgentPods(s.T(), c, common.NamespaceName, s.Env().KubernetesCluster.Client(), common.NodeAgentSelector+",agent.datadoghq.com/name=dda-autodiscovery")
	//		s.verifyHTTPCheck(c)
	//	}, 5*time.Minute, 15*time.Second, "could not validate http_check in time")
	//})
	//
	//s.T().Run("Logs collection works", func(t *testing.T) {
	//	ddaConfigPath, err := common.GetAbsPath(filepath.Join(common.ManifestsPath, "datadog-agent-logs.yaml"))
	//	assert.NoError(s.T(), err)
	//
	//	ddaOpts := []agentwithoperatorparams.Option{
	//		agentwithoperatorparams.WithDDAConfig(agentwithoperatorparams.DDAConfig{
	//			Name:         "datadog-agent-logs",
	//			YamlFilePath: ddaConfigPath,
	//		}),
	//	}
	//	ddaOpts = append(ddaOpts, defaultDDAOpts...)
	//
	//	provisionerOptions := []provisioners.KubernetesProvisionerOption{
	//		provisioners.WithTestName("e2e-operator-logs-collection"),
	//		provisioners.WithK8sVersion(common.K8sVersion),
	//		provisioners.WithOperatorOptions(defaultOperatorOpts...),
	//		provisioners.WithDDAOptions(ddaOpts...),
	//		provisioners.WithLocal(s.local),
	//	}
	//
	//	s.UpdateEnv(provisioners.KubernetesProvisioner(provisionerOptions...))
	//
	//	// Verify logs collection on agent pod
	//	s.Assert().EventuallyWithTf(func(c *assert.CollectT) {
	//		utils.VerifyAgentPods(s.T(), c, common.NamespaceName, s.Env().KubernetesCluster.Client(), "app.kubernetes.io/instance=datadog-agent-logs-agent")
	//
	//		agentPods, err := s.Env().KubernetesCluster.Client().CoreV1().Pods(common.NamespaceName).List(context.TODO(), metav1.ListOptions{LabelSelector: "app.kubernetes.io/instance=datadog-agent-logs-agent"})
	//		assert.NoError(c, err)
	//
	//		for _, pod := range agentPods.Items {
	//			output, _, err := s.Env().KubernetesCluster.KubernetesClient.PodExec(common.NamespaceName, pod.Name, "agent", []string{"agent", "status", "logs agent", "-j"})
	//			assert.NoError(c, err)
	//			utils.VerifyAgentPodLogs(c, output)
	//		}
	//
	//		s.verifyAPILogs()
	//	}, 5*time.Minute, 15*time.Second, "could not valid logs collection in time")
	//})
	//
	//s.T().Run("APM hostPort k8s service UDP works", func(t *testing.T) {
	//
	//	// Cleanup to avoid potential lingering DatadogAgent
	//	// Avoid race with the new Agent not being able to bind to the hostPort
	//	withoutDDAProvisionerOptions := []provisioners.KubernetesProvisionerOption{
	//		provisioners.WithTestName("e2e-operator-apm"),
	//		provisioners.WithoutDDA(),
	//		provisioners.WithLocal(s.local),
	//	}
	//	withoutDDAProvisionerOptions = append(withoutDDAProvisionerOptions, defaultProvisionerOpts...)
	//	s.UpdateEnv(provisioners.KubernetesProvisioner(withoutDDAProvisionerOptions...))
	//
	//	var apmAgentSelector = ",agent.datadoghq.com/name=datadog-agent-apm"
	//	ddaConfigPath, err := common.GetAbsPath(filepath.Join(common.ManifestsPath, "apm", "datadog-agent-apm.yaml"))
	//	assert.NoError(s.T(), err)
	//
	//	ddaOpts := []agentwithoperatorparams.Option{
	//		agentwithoperatorparams.WithDDAConfig(agentwithoperatorparams.DDAConfig{
	//			Name:         "datadog-agent-apm",
	//			YamlFilePath: ddaConfigPath,
	//		}),
	//	}
	//	ddaOpts = append(ddaOpts, defaultDDAOpts...)
	//
	//	ddaProvisionerOptions := []provisioners.KubernetesProvisionerOption{
	//		provisioners.WithTestName("e2e-operator-apm"),
	//		provisioners.WithDDAOptions(ddaOpts...),
	//		provisioners.WithYAMLWorkload(provisioners.YAMLWorkload{
	//			Name: "tracegen-deploy",
	//			Path: strings.Join([]string{common.ManifestsPath, "apm", "tracegen-deploy.yaml"}, "/"),
	//		}),
	//		provisioners.WithLocal(s.local),
	//	}
	//	ddaProvisionerOptions = append(ddaProvisionerOptions, defaultProvisionerOpts...)
	//
	//	// Deploy APM DatadogAgent and tracegen
	//	s.UpdateEnv(provisioners.KubernetesProvisioner(ddaProvisionerOptions...))
	//
	//	// Verify traces collection on agent pod
	//	s.EventuallyWithTf(func(c *assert.CollectT) {
	//		// Verify tracegen deployment is running
	//		utils.VerifyNumPodsForSelector(s.T(), c, common.NamespaceName, s.Env().KubernetesCluster.Client(), 1, "app=tracegen-tribrid")
	//
	//		// Verify agent pods are running
	//		utils.VerifyAgentPods(s.T(), c, common.NamespaceName, s.Env().KubernetesCluster.Client(), common.NodeAgentSelector+apmAgentSelector)
	//		agentPods, err := s.Env().KubernetesCluster.Client().CoreV1().Pods(common.NamespaceName).List(context.TODO(), metav1.ListOptions{LabelSelector: common.NodeAgentSelector + apmAgentSelector, FieldSelector: "status.phase=Running"})
	//		assert.NoError(c, err)
	//
	//		// This works because we have a single Agent pod (so located on same node as tracegen)
	//		// Otherwise, we would need to deploy tracegen on the same node as the Agent pod / as a DaemonSet
	//		for _, pod := range agentPods.Items {
	//
	//			output, _, err := s.Env().KubernetesCluster.KubernetesClient.PodExec(common.NamespaceName, pod.Name, "agent", []string{"agent", "status", "apm agent", "-j"})
	//			assert.NoError(c, err)
	//
	//			utils.VerifyAgentTraces(c, output)
	//		}
	//
	//		// Verify traces collection ingestion by fakeintake
	//		s.verifyAPITraces(c)
	//	}, 5*time.Minute, 15*time.Second, "could not validate traces on agent pod") // TODO: check duration
	//})
}

func (s *gkeAutopilotSuite) verifyAPILogs() {
	logs, err := s.Env().FakeIntake.Client().FilterLogs("agent")
	s.Assert().NoError(err)
	s.Assert().NotEmptyf(logs, fmt.Sprintf("Expected fake intake-ingested logs to not be empty: %s", err))
}

func (s *gkeAutopilotSuite) verifyAPITraces(c *assert.CollectT) {
	traces, err := s.Env().FakeIntake.Client().GetTraces()
	assert.NoError(c, err)
	assert.NotEmptyf(c, traces, fmt.Sprintf("Expected fake intake-ingested traces to not be empty: %s", err))
}

func (s *gkeAutopilotSuite) verifyKSMCheck(c *assert.CollectT) {
	ksmCheckRun, err := s.Env().FakeIntake.Client().GetCheckRun("kubernetes_state.node.ready")
	assert.NoError(c, err)
	require.NotEmpty(c, ksmCheckRun)
	assert.Equal(c, 0, ksmCheckRun[0].Status, "KSM check status should be running")

	metricNames, err := s.Env().FakeIntake.Client().GetMetricNames()
	assert.NoError(c, err)
	assert.Contains(c, metricNames, "kubernetes_state.container.running")

	metrics, err := s.Env().FakeIntake.Client().FilterMetrics("kubernetes_state.container.running", matchOpts...)
	assert.NoError(c, err)
	assert.NotEmptyf(c, metrics, fmt.Sprintf("expected metric series to not be empty: %s", err))
}

func (s *gkeAutopilotSuite) verifyHTTPCheck(c *assert.CollectT) {
	httpCheckRun, err := s.Env().FakeIntake.Client().GetCheckRun("http.can_connect")
	assert.NoError(c, err)
	require.NotEmpty(c, httpCheckRun)
	assert.Equal(c, 0, httpCheckRun[0].Status, "HTTP check status should be running")

	metricNames, err := s.Env().FakeIntake.Client().GetMetricNames()
	assert.NoError(c, err)
	assert.Contains(c, metricNames, "network.http.can_connect")
	metrics, err := s.Env().FakeIntake.Client().FilterMetrics("network.http.can_connect")
	assert.NoError(c, err)
	assert.Greater(c, len(metrics), 0)
	for _, metric := range metrics {
		for _, points := range metric.Points {
			assert.Greater(c, points.Value, float64(0))
		}
	}
}
