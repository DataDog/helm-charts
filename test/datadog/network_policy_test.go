package datadog

import (
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v3"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestClusterChecksRunnerKubernetesNetworkPolicyAllowsStatsCollection(t *testing.T) {
	for name, runnerOverrides := range clusterChecksRunnerDeploymentOverrides() {
		t.Run(name, func(t *testing.T) {
			overrides := clusterChecksNetworkPolicyOverrides("kubernetes", runnerOverrides)

			runnerManifest, err := common.RenderChart(t, common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/agent-clusterchecks-network-policy.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides:   overrides,
			})
			require.NoError(t, err)

			var runnerPolicy networkingv1.NetworkPolicy
			common.Unmarshal(t, runnerManifest, &runnerPolicy)
			require.Len(t, runnerPolicy.Spec.Ingress, 1)
			assertKubernetesNetworkPolicyPeerRule(t, runnerPolicy.Spec.Ingress[0].From, runnerPolicy.Spec.Ingress[0].Ports, "datadog-cluster-agent")

			clusterAgentManifest, err := common.RenderChart(t, common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/cluster-agent-network-policy.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides:   overrides,
			})
			require.NoError(t, err)

			var clusterAgentPolicy networkingv1.NetworkPolicy
			common.Unmarshal(t, clusterAgentManifest, &clusterAgentPolicy)
			foundRunnerEgress := false
			for _, rule := range clusterAgentPolicy.Spec.Egress {
				if kubernetesNetworkPolicyRuleTargetsApp(rule.To, "datadog-clusterchecks") {
					assertKubernetesNetworkPolicyPeerRule(t, rule.To, rule.Ports, "datadog-clusterchecks")
					foundRunnerEgress = true
					break
				}
			}
			require.True(t, foundRunnerEgress, "expected Cluster Agent egress rule to target Cluster Checks Runners")

			foundRunnerIngress := false
			for _, rule := range clusterAgentPolicy.Spec.Ingress {
				if kubernetesNetworkPolicyRuleTargetsApp(rule.From, "datadog-clusterchecks") {
					assertKubernetesNetworkPolicyPorts(t, rule.Ports)
					foundRunnerIngress = true
					break
				}
			}
			require.True(t, foundRunnerIngress, "expected Cluster Agent ingress rule to allow Cluster Checks Runners")
		})
	}
}

func TestClusterChecksRunnerCiliumNetworkPolicyAllowsStatsCollection(t *testing.T) {
	for name, runnerOverrides := range clusterChecksRunnerDeploymentOverrides() {
		t.Run(name, func(t *testing.T) {
			overrides := clusterChecksNetworkPolicyOverrides("cilium", runnerOverrides)

			runnerPolicy := renderCiliumNetworkPolicy(t, "templates/agent-clusterchecks-cilium-network-policy.yaml", overrides)
			runnerSpec := findCiliumPolicySpec(t, runnerPolicy.Specs, "Ingress from cluster agent for runner stats collection")
			assert.Equal(t, "datadog-clusterchecks", runnerSpec.EndpointSelector.MatchLabels["app"])
			require.Len(t, runnerSpec.Ingress, 1)
			assertCiliumNetworkPolicyPeerRule(t, runnerSpec.Ingress[0].FromEndpoints, runnerSpec.Ingress[0].ToPorts, "datadog-cluster-agent")

			clusterAgentPolicy := renderCiliumNetworkPolicy(t, "templates/cluster-agent-cilium-network-policy.yaml", overrides)
			clusterAgentSpec := findCiliumPolicySpec(t, clusterAgentPolicy.Specs, "Egress to cluster checks runners for runner stats collection")
			assert.Equal(t, "datadog-cluster-agent", clusterAgentSpec.EndpointSelector.MatchLabels["app"])
			require.Len(t, clusterAgentSpec.Egress, 1)
			assertCiliumNetworkPolicyPeerRule(t, clusterAgentSpec.Egress[0].ToEndpoints, clusterAgentSpec.Egress[0].ToPorts, "datadog-clusterchecks")

			clusterAgentIngressSpec := findCiliumPolicySpec(t, clusterAgentPolicy.Specs, "Ingress from cluster workers")
			require.Len(t, clusterAgentIngressSpec.Ingress, 1)
			assertCiliumNetworkPolicyPeerRule(t, clusterAgentIngressSpec.Ingress[0].FromEndpoints, clusterAgentIngressSpec.Ingress[0].ToPorts, "datadog-clusterchecks")
		})
	}
}

func TestClusterChecksRunnerCiliumNetworkPolicyAllowsHostNetworkClusterAgent(t *testing.T) {
	overrides := clusterChecksNetworkPolicyOverrides("cilium", map[string]string{
		"clusterChecksRunner.enabled": "true",
	})
	overrides["clusterAgent.useHostNetwork"] = "true"

	runnerPolicy := renderCiliumNetworkPolicy(t, "templates/agent-clusterchecks-cilium-network-policy.yaml", overrides)
	runnerSpec := findCiliumPolicySpec(t, runnerPolicy.Specs, "Ingress from cluster agent for runner stats collection")
	require.Len(t, runnerSpec.Ingress, 1)
	assert.Empty(t, runnerSpec.Ingress[0].FromEndpoints)
	assert.Equal(t, []string{"host", "remote-node"}, runnerSpec.Ingress[0].FromEntities)
	assertCiliumNetworkPolicyPorts(t, runnerSpec.Ingress[0].ToPorts)
}

func clusterChecksRunnerDeploymentOverrides() map[string]map[string]string {
	return map[string]map[string]string{
		"explicit runners": {
			"clusterChecksRunner.enabled": "true",
		},
		"kube state metrics runners": {
			"datadog.kubeStateMetricsCore.useClusterCheckRunners": "true",
		},
	}
}

func clusterChecksNetworkPolicyOverrides(flavor string, runnerOverrides map[string]string) map[string]string {
	overrides := map[string]string{
		"datadog.apiKeyExistingSecret": "datadog-secret",
		"datadog.appKeyExistingSecret": "datadog-secret",
		"datadog.networkPolicy.create": "true",
		"datadog.networkPolicy.flavor": flavor,
	}
	for key, value := range runnerOverrides {
		overrides[key] = value
	}
	return overrides
}

func assertKubernetesNetworkPolicyPeerRule(t *testing.T, peers []networkingv1.NetworkPolicyPeer, ports []networkingv1.NetworkPolicyPort, app string) {
	t.Helper()
	require.Len(t, peers, 1)
	require.NotNil(t, peers[0].PodSelector)
	assert.Equal(t, app, peers[0].PodSelector.MatchLabels["app"])
	assertKubernetesNetworkPolicyPorts(t, ports)
}

func assertKubernetesNetworkPolicyPorts(t *testing.T, ports []networkingv1.NetworkPolicyPort) {
	t.Helper()
	require.Len(t, ports, 1)
	require.NotNil(t, ports[0].Port)
	assert.Equal(t, intstr.FromInt32(5005), *ports[0].Port)
}

func kubernetesNetworkPolicyRuleTargetsApp(peers []networkingv1.NetworkPolicyPeer, app string) bool {
	for _, peer := range peers {
		if peer.PodSelector != nil && peer.PodSelector.MatchLabels["app"] == app {
			return true
		}
	}
	return false
}

type ciliumNetworkPolicy struct {
	Specs []ciliumPolicySpec `yaml:"specs"`
}

type ciliumPolicySpec struct {
	Description      string                 `yaml:"description"`
	EndpointSelector ciliumEndpointSelector `yaml:"endpointSelector"`
	Ingress          []ciliumIngressRule    `yaml:"ingress"`
	Egress           []ciliumEgressRule     `yaml:"egress"`
}

type ciliumEndpointSelector struct {
	MatchLabels map[string]string `yaml:"matchLabels"`
}

type ciliumIngressRule struct {
	FromEndpoints []ciliumEndpointSelector `yaml:"fromEndpoints"`
	FromEntities  []string                 `yaml:"fromEntities"`
	ToPorts       []ciliumPortRule         `yaml:"toPorts"`
}

type ciliumEgressRule struct {
	ToEndpoints []ciliumEndpointSelector `yaml:"toEndpoints"`
	ToPorts     []ciliumPortRule         `yaml:"toPorts"`
}

type ciliumPortRule struct {
	Ports []ciliumPortProtocol `yaml:"ports"`
}

type ciliumPortProtocol struct {
	Port     string `yaml:"port"`
	Protocol string `yaml:"protocol"`
}

func renderCiliumNetworkPolicy(t *testing.T, template string, overrides map[string]string) ciliumNetworkPolicy {
	t.Helper()
	manifest, err := common.RenderChart(t, common.HelmCommand{
		ReleaseName: "datadog",
		ChartPath:   "../../charts/datadog",
		ShowOnly:    []string{template},
		Values:      []string{"../../charts/datadog/values.yaml"},
		Overrides:   overrides,
	})
	require.NoError(t, err)

	var policy ciliumNetworkPolicy
	require.NoError(t, yaml.Unmarshal([]byte(manifest), &policy))
	return policy
}

func findCiliumPolicySpec(t *testing.T, specs []ciliumPolicySpec, description string) ciliumPolicySpec {
	t.Helper()
	for _, spec := range specs {
		if spec.Description == description {
			return spec
		}
	}
	t.Fatalf("expected Cilium policy spec %q", description)
	return ciliumPolicySpec{}
}

func assertCiliumNetworkPolicyPeerRule(t *testing.T, peers []ciliumEndpointSelector, ports []ciliumPortRule, app string) {
	t.Helper()
	require.Len(t, peers, 1)
	assert.Equal(t, app, peers[0].MatchLabels["app"])
	assertCiliumNetworkPolicyPorts(t, ports)
}

func assertCiliumNetworkPolicyPorts(t *testing.T, ports []ciliumPortRule) {
	t.Helper()
	require.Len(t, ports, 1)
	require.Len(t, ports[0].Ports, 1)
	assert.Equal(t, "5005", ports[0].Ports[0].Port)
	assert.Equal(t, "TCP", ports[0].Ports[0].Protocol)
}
