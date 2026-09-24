package datadog

import (
	"context"
	"encoding/json"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const networkPathFiltersHelmValues = `
datadog:
  networkMonitoring:
    enabled: true
  networkPath:
    connectionsMonitoring:
      enabled: true
    collector:
      filters:
        - type: exclude
          match_domain: "*.example.com"
        - type: include
          match_domain: '^api-[0-9]+\.example\.com$'
          match_domain_strategy: regex
        - type: exclude
          match_ip: "10.0.0.0/8"
  traceroute:
    enabled: true
`

const (
	networkPathFiltersEnv               = "DD_NETWORK_PATH_COLLECTOR_FILTERS"
	networkPathConnectionsMonitoringEnv = "DD_NETWORK_PATH_CONNECTIONS_MONITORING_ENABLED"
)

type networkPathFilter struct {
	Type                string `json:"type"`
	MatchDomain         string `json:"match_domain"`
	MatchDomainStrategy string `json:"match_domain_strategy"`
	MatchIP             string `json:"match_ip"`
}

var expectedNetworkPathFilters = []networkPathFilter{
	{Type: "exclude", MatchDomain: "*.example.com"},
	{Type: "include", MatchDomain: `^api-[0-9]+\.example\.com$`, MatchDomainStrategy: "regex"},
	{Type: "exclude", MatchIP: "10.0.0.0/8"},
}

func systemProbeEnvValue(pod corev1.Pod, name string) (string, int, bool) {
	for _, container := range pod.Spec.Containers {
		if container.Name != "system-probe" {
			continue
		}

		value, count := "", 0
		for _, env := range container.Env {
			if env.Name == name {
				value = env.Value
				count++
			}
		}
		return value, count, true
	}
	return "", 0, false
}

func assertSystemProbeNetworkPathFilterEnv(c *assert.CollectT, pod corev1.Pod) bool {
	raw, count, found := systemProbeEnvValue(pod, networkPathFiltersEnv)
	if !assert.True(c, found, "system-probe container not found") {
		return false
	}
	if !assert.Equalf(c, 1, count, "%s count in system-probe", networkPathFiltersEnv) {
		return false
	}

	var got []networkPathFilter
	if !assert.NoErrorf(c, json.Unmarshal([]byte(raw), &got), "invalid filter JSON in system-probe: %q", raw) {
		return false
	}
	if !assert.Equal(c, expectedNetworkPathFilters, got, "filters in system-probe") {
		return false
	}

	enabled, count, _ := systemProbeEnvValue(pod, networkPathConnectionsMonitoringEnv)
	return assert.Equalf(c, 1, count, "%s count in system-probe", networkPathConnectionsMonitoringEnv) &&
		assert.Equal(c, "true", enabled, "Connections Monitoring activation in system-probe")
}

// Verifies that system-probe receives the enabled Network Path collector and its filters in a live cluster.
func (s *k8sSuite) testNetworkPathCollectorFilters() {
	s.T().Log("Verifying Network Path collector filters")

	var verifiedPod corev1.Pod
	assert.EventuallyWithTf(s.T(), func(c *assert.CollectT) {
		pods, err := s.Env().KubernetesCluster.Client().CoreV1().Pods("datadog").List(context.TODO(), metav1.ListOptions{})
		if !assert.NoError(c, err) {
			return
		}
		pod, running := assertRunningPod(c, pods.Items, "Agent", isLinuxNodeAgentPod)
		if !running {
			return
		}

		if !assertSystemProbeNetworkPathFilterEnv(c, pod) {
			return
		}

		verifiedPod = pod
	}, 5*time.Minute, 15*time.Second, "Network Path collector filters did not reach system-probe")

	for _, container := range verifiedPod.Spec.Containers {
		if container.Name != "system-probe" {
			continue
		}
		for _, status := range verifiedPod.Status.ContainerStatuses {
			if status.Name == container.Name {
				s.T().Logf("Network Path filter verification used %s image=%s imageID=%s", container.Name, container.Image, status.ImageID)
				break
			}
		}
	}
}
