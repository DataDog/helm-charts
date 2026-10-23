package synthetics_private_location

import (
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

func multiContainerCommand(overrides map[string]string) common.HelmCommand {
	values := map[string]string{
		"image.tag":              `"1.77.0"`,
		"multiContainer.enabled": "true",
		"multiContainer.browserPool.adminTokenSecret.name": `"pool-token"`,
		"multiContainer.browserPool.adminTokenSecret.key":  `"admin"`,
		"multiContainer.networkPolicy.create":              "false",
	}
	for key, value := range overrides {
		values[key] = value
	}
	return common.HelmCommand{
		ReleaseName: "test", ChartPath: "../../charts/synthetics-private-location",
		OverridesJson: values,
	}
}

func renderDeployment(t *testing.T, command common.HelmCommand) appsv1.Deployment {
	t.Helper()
	command.ShowOnly = []string{"templates/deployment-multi-container.yaml"}
	manifest, err := common.RenderChart(t, command)
	require.NoError(t, err)
	var deployment appsv1.Deployment
	common.Unmarshal(t, manifest, &deployment)
	return deployment
}

func TestLegacyDeployment(t *testing.T) {
	manifest, err := common.RenderChart(t, common.HelmCommand{
		ReleaseName: "test", ChartPath: "../../charts/synthetics-private-location",
	})
	require.NoError(t, err)
	assert.NotContains(t, manifest, "kind: NetworkPolicy")
	assert.NotContains(t, manifest, "kind: SecurityContextConstraints")
	manifest, err = common.RenderChart(t, common.HelmCommand{
		ReleaseName: "test", ChartPath: "../../charts/synthetics-private-location",
		ShowOnly: []string{"templates/deployment.yaml"},
	})
	require.NoError(t, err)
	var deployment appsv1.Deployment
	common.Unmarshal(t, manifest, &deployment)
	require.Len(t, deployment.Spec.Template.Spec.Containers, 1)
	assert.Empty(t, deployment.Spec.Template.Spec.Containers[0].Command)
	multi := renderDeployment(t, multiContainerCommand(nil))
	assert.Equal(t, deployment.Name, multi.Name)
	assert.Equal(t, deployment.Spec.Selector, multi.Spec.Selector)
}

func TestSecretReferenceAndHardening(t *testing.T) {
	command := multiContainerCommand(nil)
	manifest, err := common.RenderChart(t, command)
	require.NoError(t, err)
	assert.NotContains(t, manifest, "kind: Secret\n")
	assert.NotContains(t, manifest, "kind: NetworkPolicy")
	deployment := renderDeployment(t, command)
	pod := deployment.Spec.Template.Spec
	require.NotNil(t, pod.AutomountServiceAccountToken)
	assert.False(t, *pod.AutomountServiceAccountToken)
	require.Len(t, pod.Containers, 4)
	for _, container := range pod.Containers {
		t.Run(container.Name, func(t *testing.T) {
			sc := container.SecurityContext
			require.NotNil(t, sc)
			assert.Equal(t, false, *sc.AllowPrivilegeEscalation)
			assert.Equal(t, true, *sc.ReadOnlyRootFilesystem)
			assert.Equal(t, true, *sc.RunAsNonRoot)
			assert.Equal(t, []corev1.Capability{"ALL"}, sc.Capabilities.Drop)
			var tokens []corev1.EnvVar
			for _, env := range container.Env {
				if env.Name == "BROWSER_POOL_ADMIN_TOKEN" {
					tokens = append(tokens, env)
				}
			}
			if container.Name == "synthetics-private-location" || container.Name == "browser-pool" {
				require.Len(t, tokens, 1)
				assert.Empty(t, tokens[0].Value)
				require.NotNil(t, tokens[0].ValueFrom)
				require.NotNil(t, tokens[0].ValueFrom.SecretKeyRef)
				assert.Equal(t, "pool-token", tokens[0].ValueFrom.SecretKeyRef.Name)
				assert.Equal(t, "admin", tokens[0].ValueFrom.SecretKeyRef.Key)
			} else {
				assert.Empty(t, tokens)
			}
			if container.Name == "deno-executor" {
				require.NotNil(t, container.LivenessProbe)
				require.NotNil(t, container.ReadinessProbe)
				assert.EqualValues(t, 15, container.LivenessProbe.TimeoutSeconds)
				assert.EqualValues(t, 15, container.ReadinessProbe.TimeoutSeconds)
				assert.Equal(t, "1Gi", container.Resources.Requests.Memory().String())
				assert.Equal(t, "1Gi", container.Resources.Limits.Memory().String())
			}
			if container.Name == "browser-pool" {
				assert.Equal(t, "/usr/local/bin/dumb-init", container.Command[0])
				assert.Contains(t, container.Command[len(container.Command)-1], "-name 'browser-*' -user nobody")
				assert.Contains(t, container.Command[len(container.Command)-1], "exec node")
			}
		})
	}
}

func TestInvalidMultiContainerConfiguration(t *testing.T) {
	tests := []struct {
		name      string
		overrides map[string]string
		message   string
	}{
		{"missing secret", map[string]string{"multiContainer.browserPool.adminTokenSecret.name": `""`}, "adminTokenSecret.name"},
		{"missing key", map[string]string{"multiContainer.browserPool.adminTokenSecret.key": `""`}, "adminTokenSecret.key"},
		{"missing create", map[string]string{"multiContainer.networkPolicy.create": "null"}, "must be set to true or false"},
		{"non-bool create", map[string]string{"multiContainer.networkPolicy.create": `"yes"`}, "must be set to true or false"},
		{"token override", map[string]string{"env": `[{"name":"BROWSER_POOL_ADMIN_TOKEN","value":"unsafe"}]`}, "not env"},
		{"missing DNS", map[string]string{"multiContainer.networkPolicy.create": "true"}, "must specify DNS"},
		{"unrestricted DNS peer", map[string]string{"multiContainer.networkPolicy.create": "true", "multiContainer.networkPolicy.dns.to": `[{}]`}, "each multiContainer.networkPolicy.dns.to entry"},
		{"duplicate sysctl", map[string]string{"podSecurityContext.sysctls": `[{"name":"net.ipv4.ping_group_range","value":"501 501"},{"name":"net.ipv4.ping_group_range","value":"0 65535"}]`}, "duplicate name net.ipv4.ping_group_range"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := common.RenderChart(t, multiContainerCommand(test.overrides))
			require.Error(t, err)
			var output *shell.ErrWithCmdOutput
			require.ErrorAs(t, err, &output)
			assert.Contains(t, output.Output.Stderr(), test.message)
		})
	}
}

func renderPolicy(t *testing.T, overrides map[string]string) networkingv1.NetworkPolicy {
	t.Helper()
	command := multiContainerCommand(map[string]string{
		"multiContainer.networkPolicy.create": "true",
		"multiContainer.networkPolicy.dns.to": `[{"ipBlock":{"cidr":"169.254.20.10/32"}}]`,
	})
	for key, value := range overrides {
		command.OverridesJson[key] = value
	}
	command.ShowOnly = []string{"templates/network-policy.yaml"}
	manifest, err := common.RenderChart(t, command)
	require.NoError(t, err)
	var policy networkingv1.NetworkPolicy
	common.Unmarshal(t, manifest, &policy)
	return policy
}

func TestScopedDNSAndDefaultBlocks(t *testing.T) {
	policy := renderPolicy(t, map[string]string{
		"multiContainer.networkPolicy.dns.to": `[{"namespaceSelector":{"matchLabels":{"kubernetes.io/metadata.name":"kube-system"}},"podSelector":{"matchLabels":{"k8s-app":"kube-dns"}}}]`,
	})
	require.Len(t, policy.Spec.Egress, 2)
	dns := policy.Spec.Egress[0]
	require.Len(t, dns.To, 1)
	require.NotNil(t, dns.To[0].NamespaceSelector)
	require.NotNil(t, dns.To[0].PodSelector)
	assert.Equal(t, "kube-system", dns.To[0].NamespaceSelector.MatchLabels["kubernetes.io/metadata.name"])
	assert.Equal(t, "kube-dns", dns.To[0].PodSelector.MatchLabels["k8s-app"])
	require.Len(t, dns.Ports, 2)
	for i, protocol := range []corev1.Protocol{corev1.ProtocolUDP, corev1.ProtocolTCP} {
		assert.Equal(t, protocol, *dns.Ports[i].Protocol)
		assert.EqualValues(t, 53, dns.Ports[i].Port.IntVal)
	}
	require.Len(t, policy.Spec.Egress[1].To, 2)
	assert.Contains(t, policy.Spec.Egress[1].To[0].IPBlock.Except, "10.0.0.0/8")
	assert.Contains(t, policy.Spec.Egress[1].To[1].IPBlock.Except, "fc00::/7")
}

func TestFullFamilyBlocksAndExceptions(t *testing.T) {
	overrides := map[string]string{
		"multiContainer.networkPolicy.blockedIPRanges":  `{"IPv4":["0.0.0.0/0"],"IPv6":["::/0"]}`,
		"multiContainer.networkPolicy.allowedIPRanges":  `{"IPv4":["10.0.0.1"],"IPv6":["fd00::1"]}`,
		"multiContainer.networkPolicy.additionalEgress": `[{"to":[{"ipBlock":{"cidr":"10.2.3.4/32"}}],"ports":[{"protocol":"TCP","port":3128}]}]`,
	}
	policy := renderPolicy(t, overrides)
	require.Len(t, policy.Spec.Egress, 3)
	assert.Equal(t, "169.254.20.10/32", policy.Spec.Egress[0].To[0].IPBlock.CIDR)
	proxy := policy.Spec.Egress[1]
	require.Len(t, proxy.To, 1)
	require.Len(t, proxy.Ports, 1)
	assert.Equal(t, "10.2.3.4/32", proxy.To[0].IPBlock.CIDR)
	assert.EqualValues(t, 3128, proxy.Ports[0].Port.IntVal)
	assert.Equal(t, corev1.ProtocolTCP, *proxy.Ports[0].Protocol)
	require.Len(t, policy.Spec.Egress[2].To, 2)
	assert.Equal(t, "10.0.0.1/32", policy.Spec.Egress[2].To[0].IPBlock.CIDR)
	assert.Equal(t, "fd00::1/128", policy.Spec.Egress[2].To[1].IPBlock.CIDR)
	overrides["multiContainer.networkPolicy.enableDefaultBlockedIpRanges"] = "false"
	overrides["multiContainer.networkPolicy.blockedIPRanges"] = `{"IPv4":["192.0.2.1"],"IPv6":[]}`
	policy = renderPolicy(t, overrides)
	require.Len(t, policy.Spec.Egress, 4)
	assert.Equal(t, &networkingv1.IPBlock{CIDR: "0.0.0.0/0", Except: []string{"192.0.2.1/32"}}, policy.Spec.Egress[3].To[0].IPBlock)
	assert.Equal(t, &networkingv1.IPBlock{CIDR: "::/0"}, policy.Spec.Egress[3].To[1].IPBlock)
}

func TestSysctlOverride(t *testing.T) {
	command := multiContainerCommand(map[string]string{"securityContext.runAsGroup": "1234"})
	deployment := renderDeployment(t, command)
	assert.Equal(t, []corev1.Sysctl{{Name: "net.ipv4.ping_group_range", Value: "1234 1234"}}, deployment.Spec.Template.Spec.SecurityContext.Sysctls)
	command.OverridesJson["podSecurityContext.sysctls"] = `[{"name":"net.ipv4.ping_group_range","value":"0 65535"},{"name":"net.ipv4.ip_local_port_range","value":"1024 65535"}]`
	deployment = renderDeployment(t, command)
	assert.Equal(t, []corev1.Sysctl{
		{Name: "net.ipv4.ping_group_range", Value: "0 65535"},
		{Name: "net.ipv4.ip_local_port_range", Value: "1024 65535"},
	}, deployment.Spec.Template.Spec.SecurityContext.Sysctls)
}

func TestTracerouteDisabledAndSCC(t *testing.T) {
	command := multiContainerCommand(map[string]string{
		"multiContainer.tracerouteServer.enabled":                      "false",
		"multiContainer.podSecurity.securityContextConstraints.create": "true",
	})
	deployment := renderDeployment(t, command)
	require.Len(t, deployment.Spec.Template.Spec.Containers, 3)
	for _, container := range deployment.Spec.Template.Spec.Containers {
		assert.NotEqual(t, "traceroute-server", container.Name)
		for _, env := range container.Env {
			assert.NotEqual(t, "DATADOG_TRACEROUTE_USE_HTTP", env.Name)
		}
	}
	command.ShowOnly = []string{"templates/scc.yaml"}
	manifest, err := common.RenderChart(t, command)
	require.NoError(t, err)
	var scc struct {
		AllowedCapabilities []string `json:"allowedCapabilities"`
	}
	common.Unmarshal(t, manifest, &scc)
	assert.Empty(t, scc.AllowedCapabilities)
}

func TestConfigSources(t *testing.T) {
	for _, source := range []string{"configSecret", "configConfigMap"} {
		t.Run(source, func(t *testing.T) {
			deployment := renderDeployment(t, multiContainerCommand(map[string]string{source: `"existing-config"`}))
			var config *corev1.Volume
			for i := range deployment.Spec.Template.Spec.Volumes {
				volume := &deployment.Spec.Template.Spec.Volumes[i]
				if volume.Name == "worker-config" {
					config = volume
				}
			}
			require.NotNil(t, config)
			if source == "configSecret" {
				require.NotNil(t, config.Secret)
				assert.Equal(t, "existing-config", config.Secret.SecretName)
			} else {
				require.NotNil(t, config.ConfigMap)
				assert.Equal(t, "existing-config", config.ConfigMap.Name)
			}
		})
	}
}
