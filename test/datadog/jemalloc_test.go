package datadog

import (
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

const jemallocPreload = "/opt/lib/nosys.so /opt/datadog-agent/embedded/lib/libjemalloc.so"
const jemallocConfig = "background_thread:true,dirty_decay_ms:5000"

func jemallocPod(t *testing.T, template string, values, jsonValues map[string]string) corev1.PodSpec {
	t.Helper()
	overrides := map[string]string{"datadog.apiKeyExistingSecret": "datadog-secret", "clusterChecksRunner.enabled": "true"}
	for k, v := range values {
		overrides[k] = v
	}
	manifest, err := common.RenderChart(t, common.HelmCommand{
		ReleaseName: "datadog", ChartPath: "../../charts/datadog", ShowOnly: []string{"templates/" + template + ".yaml"},
		Overrides: overrides, OverridesJson: jsonValues,
	})
	require.NoError(t, err)
	var workload struct {
		Spec struct{ Template corev1.PodTemplateSpec }
	}
	common.Unmarshal(t, manifest, &workload)
	return workload.Spec.Template.Spec
}

func allocatorEnv(container corev1.Container) map[string]corev1.EnvVar {
	env := map[string]corev1.EnvVar{}
	for _, e := range container.Env {
		if e.Name == "LD_PRELOAD" || e.Name == "MALLOC_CONF" {
			env[e.Name] = e // Match the existing last-entry-wins ordering.
		}
	}
	return env
}

func TestJemallocEnablement(t *testing.T) {
	for _, tc := range []struct {
		name, enabled, nodeTag, runnerTag string
		bypass, wantNode, wantRunner      bool
	}{
		{"omitted", "", "7.84.0", "7.84.0", false, false, false},
		{"disabled", "false", "7.84.0", "7.84.0", false, false, false},
		{"minimum", "true", "7.84.0", "7.84.0", false, true, true},
		{"older node", "true", "7.83.2", "7.84.0", false, false, true},
		{"older runner", "true", "7.84.0", "7.83.2", false, true, false},
		{"newer", "true", "7.85.0", "7.85.0", false, true, true},
		{"prerelease and variant", "true", "7.84.0-rc.1", "7.84.0-jmx", false, true, true},
		{"floating tags", "true", "7", "latest", false, true, true},
		{"custom images", "true", "custom", "custom", true, true, true},
		{"disabled custom images", "false", "custom", "custom", true, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			values := map[string]string{"agents.image.tag": tc.nodeTag, "clusterChecksRunner.image.tag": tc.runnerTag}
			if tc.enabled != "" {
				values["datadog.jemalloc.enabled"] = tc.enabled
			}
			if tc.bypass {
				values["agents.image.doNotCheckTag"], values["clusterChecksRunner.image.doNotCheckTag"] = "true", "true"
			}
			for template, want := range map[string]bool{"daemonset": tc.wantNode, "agent-clusterchecks-deployment": tc.wantRunner} {
				pod := jemallocPod(t, template, values, nil)
				require.NotEmpty(t, pod.Containers)
				env := allocatorEnv(pod.Containers[0])
				if want {
					assert.Equal(t, jemallocPreload, env["LD_PRELOAD"].Value, template)
					assert.Equal(t, jemallocConfig, env["MALLOC_CONF"].Value, template)
				} else {
					assert.Empty(t, env, template)
				}
			}
		})
	}
}

func TestJemallocContainerScope(t *testing.T) {
	for _, tc := range []struct {
		name, template string
		values         map[string]string
		want           []string
	}{
		{"node sidecars", "daemonset", map[string]string{"datadog.processAgent.processCollection": "true", "datadog.processAgent.runInCoreAgent": "false", "datadog.securityAgent.compliance.enabled": "true", "datadog.otelCollector.enabled": "true"}, []string{"agent", "trace-agent", "system-probe"}},
		{"core only", "daemonset", map[string]string{"datadog.apm.socketEnabled": "false", "datadog.discovery.enabled": "false"}, []string{"agent"}},
		{"runner", "agent-clusterchecks-deployment", nil, []string{"agent"}},
		{"cluster agent", "cluster-agent-deployment", nil, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			values := copyMap(tc.values)
			before := jemallocPod(t, tc.template, values, nil)
			values["datadog.jemalloc.enabled"] = "true"
			after := jemallocPod(t, tc.template, values, nil)
			var affected []string
			for i, c := range after.Containers {
				if len(allocatorEnv(c)) == 0 {
					continue
				}
				affected = append(affected, c.Name)
				var restored []corev1.EnvVar
				for _, e := range c.Env {
					if e.Name != "LD_PRELOAD" && e.Name != "MALLOC_CONF" {
						restored = append(restored, e)
					}
				}
				assert.Len(t, c.Env, len(restored)+2)
				after.Containers[i].Env = restored
			}
			assert.ElementsMatch(t, tc.want, affected)
			assert.Equal(t, before, after, "only the selected runtime environments should change")
		})
	}
}

func TestJemallocOverrides(t *testing.T) {
	json := map[string]string{
		"datadog.envDict":                      `{"LD_PRELOAD":"/global.so","MALLOC_CONF":"narenas:4"}`,
		"agents.containers.agent.env":          `[{"name":"EXTRA_LIB","value":"/core.so"},{"name":"LD_PRELOAD","value":"$(EXTRA_LIB)"}]`,
		"agents.containers.traceAgent.envDict": `{"LD_PRELOAD":"","MALLOC_CONF":{"valueFrom":{"secretKeyRef":{"name":"allocator","key":"config"}}}}`,
		"clusterChecksRunner.env":              `[{"name":"LD_PRELOAD","value":"/list.so"}]`,
		"clusterChecksRunner.envDict":          `{"LD_PRELOAD":"/runner.so"}`,
	}
	for _, template := range []string{"daemonset", "agent-clusterchecks-deployment"} {
		t.Run(template, func(t *testing.T) {
			values := map[string]string{"datadog.jemalloc.enabled": "true"}
			pod := jemallocPod(t, template, values, json)
			for _, c := range pod.Containers {
				env := allocatorEnv(c)
				switch {
				case template == "agent-clusterchecks-deployment":
					assert.Equal(t, "/runner.so", env["LD_PRELOAD"].Value)
					assert.Equal(t, jemallocConfig, env["MALLOC_CONF"].Value, "runners do not inherit global overrides")
				case c.Name == "agent":
					assert.Equal(t, "$(EXTRA_LIB)", env["LD_PRELOAD"].Value)
				case c.Name == "trace-agent":
					assert.Empty(t, env["LD_PRELOAD"].Value)
					require.NotNil(t, env["MALLOC_CONF"].ValueFrom)
					assert.Equal(t, "allocator", env["MALLOC_CONF"].ValueFrom.SecretKeyRef.Name)
				default:
					assert.Equal(t, "/global.so", env["LD_PRELOAD"].Value)
					assert.Equal(t, "narenas:4", env["MALLOC_CONF"].Value)
				}
			}
			values["datadog.jemalloc.enabled"] = "false"
			assert.Equal(t, jemallocPod(t, template, nil, json), jemallocPod(t, template, values, json), "disabling preserves explicit overrides")
		})
	}
}
