package synthetics_private_location

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"

	"github.com/DataDog/helm-charts/test/common"
)

func render(t *testing.T, showOnly string, overrides map[string]string) (string, error) {
	values := map[string]string{"configSecret": "synthetics-pl-config"}
	for k, v := range overrides {
		values[k] = v
	}
	return common.RenderChart(t, common.HelmCommand{
		ReleaseName: "synthetics-pl",
		ChartPath:   chartPath,
		ShowOnly:    []string{showOnly},
		Values:      []string{chartPath + "/values.yaml"},
		Overrides:   values,
	})
}

func renderDeployment(t *testing.T, overrides map[string]string) appsv1.Deployment {
	manifest, err := render(t, "templates/deployment-multi-container.yaml", overrides)
	require.NoError(t, err)
	var deployment appsv1.Deployment
	common.Unmarshal(t, manifest, &deployment)
	return deployment
}

func containerByName(t *testing.T, containers []corev1.Container, name string) corev1.Container {
	for _, c := range containers {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("container %q not found", name)
	return corev1.Container{}
}

func envValue(c corev1.Container, name string) (string, bool) {
	for _, e := range c.Env {
		if e.Name == name {
			return e.Value, true
		}
	}
	return "", false
}

func Test_multiContainer_disabledByDefault(t *testing.T) {
	_, err := render(t, "templates/deployment-multi-container.yaml", nil)
	assert.ErrorContains(t, err, "could not find template")

	_, err = render(t, "templates/network-policy.yaml", map[string]string{"multiContainer.networkPolicy.enabled": "true"})
	assert.ErrorContains(t, err, "could not find template")
}

func Test_multiContainer_replacesLegacyDeployment(t *testing.T) {
	_, err := render(t, "templates/deployment.yaml", map[string]string{"multiContainer.enabled": "true"})
	assert.ErrorContains(t, err, "could not find template")
}

func Test_multiContainer_containersAreHardened(t *testing.T) {
	deployment := renderDeployment(t, map[string]string{"multiContainer.enabled": "true"})
	containers := deployment.Spec.Template.Spec.Containers

	names := []string{}
	for _, c := range containers {
		names = append(names, c.Name)
	}
	assert.Equal(t, []string{"synthetics-private-location", "deno-executor", "browser-pool", "traceroute-server"}, names)

	for _, c := range containers {
		sc := c.SecurityContext
		require.NotNil(t, sc, c.Name)
		assert.False(t, *sc.AllowPrivilegeEscalation, c.Name)
		assert.True(t, *sc.ReadOnlyRootFilesystem, c.Name)
		assert.True(t, *sc.RunAsNonRoot, c.Name)
		assert.NotZero(t, *sc.RunAsUser, c.Name)
		assert.Equal(t, []corev1.Capability{"ALL"}, sc.Capabilities.Drop, c.Name)
		assert.Equal(t, corev1.SeccompProfileTypeRuntimeDefault, sc.SeccompProfile.Type, c.Name)
		if c.Name == "traceroute-server" {
			assert.Equal(t, []corev1.Capability{"NET_RAW"}, sc.Capabilities.Add, c.Name)
		} else {
			assert.Empty(t, sc.Capabilities.Add, c.Name)
		}
	}
	assert.Equal(t, int64(502), *deployment.Spec.Template.Spec.SecurityContext.FSGroup)
	assert.False(t, *deployment.Spec.Template.Spec.AutomountServiceAccountToken)
	assert.Equal(t, int64(620), *deployment.Spec.Template.Spec.TerminationGracePeriodSeconds)
}

func Test_multiContainer_workerWiring(t *testing.T) {
	deployment := renderDeployment(t, map[string]string{"multiContainer.enabled": "true"})
	containers := deployment.Spec.Template.Spec.Containers
	worker := containerByName(t, containers, "synthetics-private-location")
	browserPool := containerByName(t, containers, "browser-pool")

	assert.Equal(t, []string{"/bin/bash", "/home/dog/scripts/multi-pod/start-worker.sh"}, worker.Command)
	for name, expected := range map[string]string{
		"DATADOG_WORKER_ENABLE_STATUS_PROBES": "true",
		"DENO_EXECUTOR_SOCKET_PATH":           "/run/deno-executor/deno-executor.sock",
		"DATADOG_BROWSER_POOL_URL":            "http://localhost:9444",
		"DATADOG_TRACEROUTE_USE_HTTP":         "true",
		"DATADOG_TRACEROUTE_HTTP_URL":         "http://localhost:3765",
		"DATADOG_PKI":                         "/tmp/pki",
	} {
		value, ok := envValue(worker, name)
		assert.True(t, ok, name)
		assert.Equal(t, expected, value, name)
	}

	for _, c := range []corev1.Container{worker, browserPool} {
		var token *corev1.EnvVar
		for i := range c.Env {
			if c.Env[i].Name == "BROWSER_POOL_ADMIN_TOKEN" {
				token = &c.Env[i]
			}
		}
		require.NotNil(t, token, c.Name)
		assert.Equal(t, "metadata.uid", token.ValueFrom.FieldRef.FieldPath, c.Name)
	}
}

func Test_multiContainer_withoutTracerouteServer(t *testing.T) {
	deployment := renderDeployment(t, map[string]string{
		"multiContainer.enabled":                  "true",
		"multiContainer.tracerouteServer.enabled": "false",
	})
	containers := deployment.Spec.Template.Spec.Containers
	assert.Len(t, containers, 3)
	for _, c := range containers {
		assert.NotEqual(t, "traceroute-server", c.Name)
		assert.Empty(t, c.SecurityContext.Capabilities.Add, c.Name)
	}
	worker := containerByName(t, containers, "synthetics-private-location")
	_, ok := envValue(worker, "DATADOG_TRACEROUTE_USE_HTTP")
	assert.False(t, ok)
}

func Test_multiContainer_workerResourcesFallback(t *testing.T) {
	deployment := renderDeployment(t, map[string]string{
		"multiContainer.enabled":    "true",
		"resources.limits.memory":   "3Gi",
		"resources.requests.memory": "3Gi",
	})
	worker := containerByName(t, deployment.Spec.Template.Spec.Containers, "synthetics-private-location")
	assert.Equal(t, "3Gi", worker.Resources.Limits.Memory().String())

	deployment = renderDeployment(t, map[string]string{
		"multiContainer.enabled":                        "true",
		"resources.limits.memory":                       "3Gi",
		"multiContainer.worker.resources.limits.memory": "1Gi",
	})
	worker = containerByName(t, deployment.Spec.Template.Spec.Containers, "synthetics-private-location")
	assert.Equal(t, "1Gi", worker.Resources.Limits.Memory().String())
}

func Test_networkPolicy(t *testing.T) {
	tests := []struct {
		name             string
		overrides        map[string]string
		expectedAllowed  []string
		expectedExceptV4 []string
		expectedExceptV6 []string
	}{
		{
			name:      "no blocked ranges by default",
			overrides: map[string]string{},
		},
		{
			name: "custom ranges only",
			overrides: map[string]string{
				"multiContainer.networkPolicy.allowedIPRanges.IPv4[0]": "10.1.0.0/16",
				"multiContainer.networkPolicy.allowedIPRanges.IPv6[0]": "fd00::/8",
				"multiContainer.networkPolicy.blockedIPRanges.IPv4[0]": "8.8.8.0/24",
			},
			expectedAllowed:  []string{"10.1.0.0/16", "fd00::/8"},
			expectedExceptV4: []string{"8.8.8.0/24"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			overrides := map[string]string{
				"multiContainer.enabled":               "true",
				"multiContainer.networkPolicy.enabled": "true",
			}
			for k, v := range tt.overrides {
				overrides[k] = v
			}
			manifest, err := render(t, "templates/network-policy.yaml", overrides)
			require.NoError(t, err)
			var policy networkingv1.NetworkPolicy
			common.Unmarshal(t, manifest, &policy)

			assert.Equal(t, []networkingv1.PolicyType{networkingv1.PolicyTypeEgress}, policy.Spec.PolicyTypes)
			assert.Equal(t, "synthetics-private-location", policy.Spec.PodSelector.MatchLabels["app.kubernetes.io/name"])

			rules := policy.Spec.Egress
			public := rules[len(rules)-1].To
			require.Len(t, public, 2)
			assert.Equal(t, "0.0.0.0/0", public[0].IPBlock.CIDR)
			assert.Equal(t, tt.expectedExceptV4, public[0].IPBlock.Except)
			assert.Equal(t, "::/0", public[1].IPBlock.CIDR)
			assert.Equal(t, tt.expectedExceptV6, public[1].IPBlock.Except)

			if tt.expectedAllowed == nil {
				assert.Len(t, rules, 3)
				return
			}
			require.Len(t, rules, 4)
			allowed := []string{}
			for _, peer := range rules[2].To {
				allowed = append(allowed, peer.IPBlock.CIDR)
			}
			assert.Equal(t, tt.expectedAllowed, allowed)
		})
	}
}

func Test_networkPolicy_defaultBlockedRanges(t *testing.T) {
	manifest, err := render(t, "templates/network-policy.yaml", map[string]string{
		"multiContainer.enabled":                                    "true",
		"multiContainer.networkPolicy.enabled":                      "true",
		"multiContainer.networkPolicy.enableDefaultBlockedIpRanges": "true",
		"multiContainer.networkPolicy.blockedIPRanges.IPv4[0]":      "8.8.8.0/24",
	})
	require.NoError(t, err)
	var policy networkingv1.NetworkPolicy
	common.Unmarshal(t, manifest, &policy)

	public := policy.Spec.Egress[len(policy.Spec.Egress)-1].To
	assert.Contains(t, public[0].IPBlock.Except, "10.0.0.0/8")
	assert.Contains(t, public[0].IPBlock.Except, "169.254.0.0/16")
	assert.Equal(t, "8.8.8.0/24", public[0].IPBlock.Except[len(public[0].IPBlock.Except)-1])
	assert.Contains(t, public[1].IPBlock.Except, "fc00::/7")
	assert.NotContains(t, public[1].IPBlock.Except, "::ffff:0:0/96")
}
