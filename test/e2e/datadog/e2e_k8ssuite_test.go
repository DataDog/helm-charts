package datadog

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/DataDog/datadog-agent/pkg/util/testutil/flake"
	"github.com/DataDog/datadog-agent/test/e2e-framework/components/datadog/kubernetesagentparams"
	"github.com/DataDog/datadog-agent/test/e2e-framework/testing/e2e"
	"github.com/DataDog/datadog-agent/test/e2e-framework/testing/environments"
	gcpkubernetes "github.com/DataDog/datadog-agent/test/e2e-framework/testing/provisioners/gcp/kubernetes"
	"github.com/DataDog/datadog-agent/test/e2e-framework/testing/runner"
	"github.com/DataDog/datadog-agent/test/fakeintake/aggregator"
	"github.com/DataDog/datadog-agent/test/fakeintake/client"
	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	s.Require().NoError(err)
	s.DefaultConfig = config
	s.Assert().NotEmpty(datadogChartPath())
}

func datadogChartPath() string {
	currentDir, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("datadogChartPath: os.Getwd(): %v", err))
	}
	chartPath, err := filepath.Abs(filepath.Join(currentDir, "..", "..", "..", "charts", "datadog"))
	if err != nil {
		panic(fmt.Sprintf("datadogChartPath: filepath.Abs(): %v", err))
	}
	return chartPath
}

func isLinuxNodeAgentPod(pod corev1.Pod) bool {
	return strings.Contains(pod.Name, "dda-linux-datadog") &&
		!strings.Contains(pod.Name, "cluster-agent") &&
		!strings.Contains(pod.Name, "clusterchecks")
}

func isClusterAgentPod(pod corev1.Pod) bool {
	return strings.Contains(pod.Name, "cluster-agent")
}

func isClusterChecksPod(pod corev1.Pod) bool {
	return strings.Contains(pod.Name, "dda-linux-datadog-clusterchecks")
}

func isNginxPod(pod corev1.Pod) bool {
	return strings.Contains(pod.Name, "nginx")
}

func assertRunningPod(c *assert.CollectT, pods []corev1.Pod, label string, match func(corev1.Pod) bool) (corev1.Pod, bool) {
	var firstMatch corev1.Pod
	found := false
	for _, pod := range pods {
		if !match(pod) {
			continue
		}
		if pod.Status.Phase == corev1.PodRunning {
			return pod, true
		}
		if !found {
			firstMatch = pod
			found = true
		}
	}

	assert.Truef(c, found, "%s not found. Pods: %s", label, podPhaseSummary(pods))
	if found {
		assert.Equalf(c, corev1.PodRunning, firstMatch.Status.Phase, "%s is not running: %s. Pods: %s", label, firstMatch.Status.Phase, podPhaseSummary(pods))
	}
	return corev1.Pod{}, false
}

func podPhaseSummary(pods []corev1.Pod) string {
	phases := make([]string, 0, len(pods))
	for _, pod := range pods {
		phases = append(phases, fmt.Sprintf("%s=%s", pod.Name, pod.Status.Phase))
	}
	sort.Strings(phases)
	return strings.Join(phases, ", ")
}

func (s *k8sSuite) testGenericK8sAutopilot() {
	s.testGenericK8sKubeletCheck()
	s.testGenericK8sAutodiscovery()
	s.testGenericK8sLogs()
	s.testGenericK8sKSMCore()
}

func (s *k8sSuite) testGenericK8s() {
	s.testGenericK8sKubeletCheck()
	s.testGenericK8sAutodiscovery()
	s.testGenericK8sLogs()
	s.testGenericK8sKSMCore()
	s.testGenericK8sKSMCoreCCR()
}

func (s *k8sSuite) testGenericK8sKubeletCheck() {
	s.Run("Kubelet check works", func() {
		flake.Mark(s.T())
		s.Assert().EventuallyWithT(func(c *assert.CollectT) {
			kubeletCheckRun, err := s.Env().FakeIntake.Client().GetCheckRun("kubernetes.kubelet.check")
			assert.NoError(c, err)
			assert.NotEmpty(c, kubeletCheckRun)

			kubeletMetricSeries, err := s.Env().FakeIntake.Client().FilterMetrics("kubernetes.cpu.usage.total", matchOpts...)
			assert.NoError(c, err)
			assert.NotEmptyf(c, kubeletMetricSeries, "expected Kubelet check series to not be empty: %s", err)

		}, 5*time.Minute, 15*time.Second, "could not validate kubelet check in time")

		s.Assert().EventuallyWithT(func(c *assert.CollectT) {
			s.verifyKSMCheck(c)
		}, 1*time.Minute, 15*time.Second, "could not validate KSM check in time")
	})
}

func (s *k8sSuite) testGenericK8sLogs() {
	s.Run("Logs collection works", func() {
		// Verify logs collection on agent pod
		s.Assert().EventuallyWithT(func(c *assert.CollectT) {
			agentPods, err := s.Env().KubernetesCluster.Client().CoreV1().Pods("datadog").List(context.TODO(), metav1.ListOptions{})
			assert.NoError(c, err)
			if err != nil {
				return
			}

			_, ok := assertRunningPod(c, agentPods.Items, "Agent", isLinuxNodeAgentPod)
			if !ok {
				return
			}

			s.verifyAPILogs(c)
		}, 5*time.Minute, 15*time.Second, "could not validate logs collection in time")
	})
}

func (s *k8sSuite) testGenericK8sAutodiscovery() {
	s.Run("Autodiscovery works", func() {
		err := s.Env().FakeIntake.Client().FlushServerAndResetAggregators()
		s.Assert().NoError(err)

		s.Assert().EventuallyWithTf(func(c *assert.CollectT) {
			res, err := s.Env().KubernetesCluster.Client().CoreV1().Pods("datadog").List(context.TODO(), metav1.ListOptions{})
			assert.NoError(c, err)
			if err != nil {
				return
			}

			if _, ok := assertRunningPod(c, res.Items, "Nginx", isNginxPod); !ok {
				return
			}
			if _, ok := assertRunningPod(c, res.Items, "Agent", isLinuxNodeAgentPod); !ok {
				return
			}

			s.verifyHTTPCheck(c)
		}, 5*time.Minute, 15*time.Second, "could not validate http_check in time")
	})
}

func (s *k8sSuite) testGenericK8sKSMCore() {
	s.Run("KSM core check works", func() {
		flake.Mark(s.T())
		s.Assert().EventuallyWithT(func(c *assert.CollectT) {
			s.verifyKSMCheck(c)
		}, 1*time.Minute, 15*time.Second, "could not validate KSM check in time")

	})
}

func (s *k8sSuite) testGenericK8sKSMCoreCCR() {
	s.Run("KSM check works cluster check runner", func() {
		flake.Mark(s.T())
		agentOpts := []kubernetesagentparams.Option{
			kubernetesagentparams.WithHelmRepoURL(""),
			kubernetesagentparams.WithHelmChartPath(datadogChartPath()),
			kubernetesagentparams.WithHelmValues(`
datadog:
  kubelet:
    useApiServer: true
    tlsVerify: false
  kubeStateMetricsCore:
    useClusterCheckRunners: true
providers:
  gke:
    cos: true
clusterChecksRunner:
  enabled: true`)}

		s.UpdateEnv(gcpkubernetes.GKEProvisioner(
			gcpkubernetes.WithExtraConfigParams(s.DefaultConfig),
			gcpkubernetes.WithAgentOptions(agentOpts...)))

		err := s.Env().FakeIntake.Client().FlushServerAndResetAggregators()
		s.Assert().NoError(err)

		s.Assert().EventuallyWithTf(func(c *assert.CollectT) {
			res, err := s.Env().KubernetesCluster.Client().CoreV1().Pods("datadog").List(context.TODO(), metav1.ListOptions{})
			assert.NoError(c, err)
			if err != nil {
				return
			}

			if _, ok := assertRunningPod(c, res.Items, "Agent cluster check runner", isClusterChecksPod); !ok {
				return
			}

			s.verifyKSMCheck(c)
		}, 10*time.Minute, 15*time.Second, "could not validate kubernetes_state_core (cluster check on CCR) check in time")
	})
}

func (s *k8sSuite) verifyAPILogs(c *assert.CollectT) {
	logs, err := s.Env().FakeIntake.Client().FilterLogs("agent")
	assert.NoError(c, err)
	assert.NotEmptyf(c, logs, "Expected fake intake-ingested logs to not be empty: %s", err)
}

func (s *k8sSuite) verifyKSMCheck(c *assert.CollectT) {
	ksmCheckRun, err := s.Env().FakeIntake.Client().GetCheckRun("kubernetes_state.node.ready")
	assert.NoError(c, err)
	require.NotEmpty(c, ksmCheckRun)
	assert.Equalf(c, 0, ksmCheckRun[0].Status, "KSM check status should be running: %s", ksmCheckRun[0].Message)

	metricNames, err := s.Env().FakeIntake.Client().GetMetricNames()
	assert.NoError(c, err)
	assert.Contains(c, metricNames, "kubernetes_state.container.running")

	metrics, err := s.Env().FakeIntake.Client().FilterMetrics("kubernetes_state.container.running", matchOpts...)
	assert.NoError(c, err)
	assert.NotEmptyf(c, metrics, "expected metric series to not be empty: %s", err)
}

func (s *k8sSuite) verifyHTTPCheck(c *assert.CollectT) {
	httpCheckRun, err := s.Env().FakeIntake.Client().GetCheckRun("http.can_connect")
	assert.NoError(c, err)
	require.NotEmpty(c, httpCheckRun)
	assert.Equalf(c, 0, httpCheckRun[0].Status, "HTTP check status should be running: %s", httpCheckRun[0].Message)

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
