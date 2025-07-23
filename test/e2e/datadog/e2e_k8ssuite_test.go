package datadog

import (
	"context"
	"fmt"
	"github.com/DataDog/datadog-agent/test/fakeintake/aggregator"
	"github.com/DataDog/datadog-agent/test/fakeintake/client"
	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/e2e"
	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/environments"
	gcpkubernetes "github.com/DataDog/datadog-agent/test/new-e2e/pkg/provisioners/gcp/kubernetes"
	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/runner"
	"github.com/DataDog/helm-charts/test/common"
	"github.com/DataDog/test-infra-definitions/components/datadog/kubernetesagentparams"
	"github.com/DataDog/test-infra-definitions/scenarios/gcp/gke"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var (
	matchTags = []*regexp.Regexp{regexp.MustCompile("kube_container_name:.*")}
	matchOpts = []client.MatchOpt[*aggregator.MetricSeries]{client.WithMatchingTags[*aggregator.MetricSeries](matchTags)}
)

type k8sSuite struct {
	e2e.BaseSuite[environments.Kubernetes]
	DefaultConfig runner.ConfigMap
}

func (s *k8sSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
	config, err := common.SetupConfig()
	if err != nil {
		s.Error(err)
	}
	s.DefaultConfig = config
	s.Assert().NotEmpty(datadogChartPath())
}

func datadogChartPath() string {
	currentDir, _ := os.Getwd()
	chartPath, _ := filepath.Abs(filepath.Join(currentDir, "..", "..", "..", "charts", "datadog"))
	return chartPath
}

func (s *k8sSuite) testGenericK8sAutopilot() {
	s.testGenericK8sKubeletCheck()
	s.testGenericK8sAutodiscovery()
	s.testGenericK8sLogs()
	s.testGenericK8sKSMCore()
	s.testGenericK8sKSMCoreCCR(true)
}

func (s *k8sSuite) testGenericK8s() {
	s.testGenericK8sKubeletCheck()
	s.testGenericK8sAutodiscovery()
	s.testGenericK8sLogs()
	s.testGenericK8sKSMCore()
	s.testGenericK8sKSMCoreCCR(false)
}

func (s *k8sSuite) testGenericK8sKubeletCheck() {
	s.Run("Kubelet check works", func() {
		s.Assert().EventuallyWithT(func(c *assert.CollectT) {
			kubeletCheckRun, err := s.Env().FakeIntake.Client().GetCheckRun("kubernetes.kubelet.check")
			assert.NoError(c, err)
			assert.NotEmpty(c, kubeletCheckRun)

			// Kubelet service check reports CRITICAL status 2 sometimes even though it's OK
			// assert.Equal(c, 0, kubeletCheckRun[0].Status, fmt.Sprintf("kubelet check status should be running: %s", kubeletCheckRun[0].Message))

			kubeletMetricSeries, err := s.Env().FakeIntake.Client().FilterMetrics("kubernetes.cpu.usage.total", matchOpts...)
			s.Assert().NoError(err)
			s.Assert().NotEmptyf(kubeletMetricSeries, fmt.Sprintf("expected Kubelet check series to not be empty: %s", err))

		}, 5*time.Minute, 15*time.Second, "could not validate kubelet check in time")

		s.Assert().EventuallyWithT(func(c *assert.CollectT) {
			s.verifyKSMCheck(c)
		}, 1*time.Minute, 15*time.Second, "could not validate KSM check in time")
	})
}

func (s *k8sSuite) testGenericK8sLogs() {
	s.Run("Logs collection works", func() {
		// Verify logs collection on agent pod
		s.Assert().EventuallyWithTf(func(c *assert.CollectT) {
			agentPods, err := s.Env().KubernetesCluster.Client().CoreV1().Pods("datadog").List(context.TODO(), metav1.ListOptions{})
			s.Assert().NoError(err)

			var agent corev1.Pod
			containsAgent := false
			for _, pod := range agentPods.Items {
				if strings.Contains(pod.Name, "dda-linux-datadog") && !strings.Contains(pod.Name, "cluster-agent") {
					containsAgent = true
					agent = pod
					break
				}
			}
			assert.True(c, containsAgent, "Agent not found")
			assert.Equal(c, corev1.PodPhase("Running"), agent.Status.Phase, fmt.Sprintf("Agent is not running: %s", agent.Status.Phase))

			assert.NoError(c, err)

			s.verifyAPILogs()
		}, 5*time.Minute, 15*time.Second, "could not valid logs collection in time")
	})
}

func (s *k8sSuite) testGenericK8sAutodiscovery() {
	s.Run("Autodiscovery works", func() {
		err := s.Env().FakeIntake.Client().FlushServerAndResetAggregators()
		s.Assert().NoError(err)

		s.Assert().EventuallyWithTf(func(c *assert.CollectT) {
			res, _ := s.Env().KubernetesCluster.Client().CoreV1().Pods("datadog").List(context.TODO(), metav1.ListOptions{})

			var nginx corev1.Pod
			containsNginx := false
			for _, pod := range res.Items {
				if strings.Contains(pod.Name, "nginx") {
					containsNginx = true
					nginx = pod
					break
				}
			}
			assert.True(c, containsNginx, "Nginx pod not found")
			assert.Equal(c, corev1.PodPhase("Running"), nginx.Status.Phase, fmt.Sprintf("Nginx is not running: %s", nginx.Status.Phase))

			var agent corev1.Pod
			containsAgent := false
			for _, pod := range res.Items {
				if strings.Contains(pod.Name, "dda-linux-datadog") && !strings.Contains(pod.Name, "cluster-agent") {
					containsAgent = true
					agent = pod
					break
				}
			}
			assert.True(c, containsAgent, "Agent not found")
			assert.Equal(c, corev1.PodPhase("Running"), agent.Status.Phase, fmt.Sprintf("Agent is not running: %s", agent.Status.Phase))

			s.verifyHTTPCheck(c)
		}, 5*time.Minute, 15*time.Second, "could not validate http_check in time")
	})
}

func (s *k8sSuite) testGenericK8sKSMCore() {
	s.Run("KSM core check works", func() {
		s.Assert().EventuallyWithT(func(c *assert.CollectT) {
			s.verifyKSMCheck(c)
		}, 1*time.Minute, 15*time.Second, "could not validate KSM check in time")

	})
}

func (s *k8sSuite) testGenericK8sKSMCoreCCR(withAutopilot bool) {
	s.Run("KSM check works cluster check runner", func() {
		s.UpdateEnv(gcpkubernetes.GKEProvisioner(
			gcpkubernetes.WithGKEOptions(s.getGKEOptions(withAutopilot)...),
			gcpkubernetes.WithExtraConfigParams(s.DefaultConfig),
			gcpkubernetes.WithAgentOptions(
				kubernetesagentparams.WithGKEAutopilot(),
				kubernetesagentparams.WithHelmRepoURL(""),
				kubernetesagentparams.WithHelmChartPath(datadogChartPath()),
				kubernetesagentparams.WithHelmValues(`
datadog:
  kubelet:
    useApiServer: true
    tlsVerify: false
  kubeStateMetricsCore:
    useClusterCheckRunners: true
clusterChecksRunner:
  enabled: true`))))

		err := s.Env().FakeIntake.Client().FlushServerAndResetAggregators()
		s.Assert().NoError(err)

		s.Assert().EventuallyWithTf(func(c *assert.CollectT) {
			res, _ := s.Env().KubernetesCluster.Client().CoreV1().Pods("datadog").List(context.TODO(), metav1.ListOptions{})

			var agent corev1.Pod
			containsCCR := false
			for _, pod := range res.Items {
				if strings.Contains(pod.Name, "dda-linux-datadog-clusterchecks") {
					containsCCR = true
					agent = pod
					break
				}
			}
			assert.True(c, containsCCR, "Agent cluster check runner not found")
			assert.Equal(c, corev1.PodPhase("Running"), agent.Status.Phase, fmt.Sprintf("Agent cluster check runner is not running: %s", agent.Status.Phase))

			s.verifyKSMCheck(c)
		}, 10*time.Minute, 15*time.Second, "could not validate kubernetes_state_core (cluster check on CCR) check in time")
	})
}

func (s *k8sSuite) verifyAPILogs() {
	logs, err := s.Env().FakeIntake.Client().FilterLogs("agent")
	s.Assert().NoError(err)
	s.Assert().NotEmptyf(logs, fmt.Sprintf("Expected fake intake-ingested logs to not be empty: %s", err))
}

func (s *k8sSuite) verifyKSMCheck(c *assert.CollectT) {
	ksmCheckRun, err := s.Env().FakeIntake.Client().GetCheckRun("kubernetes_state.node.ready")
	s.Assert().NoError(err)
	require.NotEmpty(c, ksmCheckRun)
	assert.Equal(c, 0, ksmCheckRun[0].Status, fmt.Sprintf("KSM check status should be running: %s", ksmCheckRun[0].Message))

	metricNames, err := s.Env().FakeIntake.Client().GetMetricNames()
	assert.NoError(c, err)
	assert.Contains(c, metricNames, "kubernetes_state.container.running")

	metrics, err := s.Env().FakeIntake.Client().FilterMetrics("kubernetes_state.container.running", matchOpts...)
	assert.NoError(c, err)
	assert.NotEmptyf(c, metrics, fmt.Sprintf("expected metric series to not be empty: %s", err))
}

func (s *k8sSuite) verifyHTTPCheck(c *assert.CollectT) {
	httpCheckRun, err := s.Env().FakeIntake.Client().GetCheckRun("http.can_connect")
	assert.NoError(c, err)
	require.NotEmpty(c, httpCheckRun)
	assert.Equal(c, 0, httpCheckRun[0].Status, fmt.Sprintf("HTTP check status should be running: %s", httpCheckRun[0].Message))

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

func (s *k8sSuite) getGKEOptions(autopilot bool) []gke.Option {
	if autopilot {
		return []gke.Option{
			gke.WithAutopilot(),
		}

	}
	return []gke.Option{}
}
