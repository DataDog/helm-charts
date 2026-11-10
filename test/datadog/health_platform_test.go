package datadog

import (
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

const healthPlatformEnv = "DD_HEALTH_PLATFORM_ENABLED"

func healthPlatformContainer(t *testing.T, template, container string, values map[string]string) corev1.Container {
	t.Helper()
	overrides := map[string]string{
		"datadog.apiKeyExistingSecret":  "datadog-secret",
		"datadog.clusterChecks.enabled": "true",
		"clusterChecksRunner.enabled":   "true",
	}
	for k, v := range values {
		overrides[k] = v
	}
	manifest, err := common.RenderChart(t, common.HelmCommand{
		ReleaseName: "datadog", ChartPath: "../../charts/datadog", ShowOnly: []string{"templates/" + template + ".yaml"},
		Overrides: overrides,
	})
	require.NoError(t, err)
	var workload struct {
		Spec struct{ Template corev1.PodTemplateSpec }
	}
	common.Unmarshal(t, manifest, &workload)
	for _, c := range workload.Spec.Template.Spec.Containers {
		if c.Name == container {
			return c
		}
	}
	require.Failf(t, "container not found", "%s in %s", container, template)
	return corev1.Container{}
}

// effectiveEnv returns the value Kubernetes applies for name: the last entry wins.
func effectiveEnv(c corev1.Container, name string) (string, bool) {
	value, found := "", false
	for _, e := range c.Env {
		if e.Name == name {
			value, found = e.Value, true
		}
	}
	return value, found
}

func TestHealthPlatformEnv(t *testing.T) {
	targets := []struct{ template, container, envKey string }{
		{"daemonset", "agent", "agents.containers.agent.env"},
		{"cluster-agent-deployment", "cluster-agent", "clusterAgent.env"},
		{"agent-clusterchecks-deployment", "agent", "clusterChecksRunner.env"},
	}
	for _, tc := range []struct {
		name    string
		enabled string
		userEnv string
		want    string
	}{
		{name: "disabled by default", want: "false"},
		{name: "opt back in", enabled: "true", want: "true"},
		{name: "explicitly disabled", enabled: "false", want: "false"},
		{name: "user env wins", userEnv: "true", want: "true"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, target := range targets {
				values := map[string]string{}
				if tc.enabled != "" {
					values["datadog.healthPlatform.enabled"] = tc.enabled
				}
				if tc.userEnv != "" {
					values[target.envKey+"[0].name"] = healthPlatformEnv
					values[target.envKey+"[0].value"] = tc.userEnv
				}
				c := healthPlatformContainer(t, target.template, target.container, values)
				got, found := effectiveEnv(c, healthPlatformEnv)
				assert.True(t, found, target.template)
				assert.Equal(t, tc.want, got, target.template)
			}
		})
	}
}
