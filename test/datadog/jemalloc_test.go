package datadog

import (
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const jemallocDisabledPreload = "/opt/lib/nosys.so"

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

func TestJemallocDisablement(t *testing.T) {
	for _, tc := range []struct {
		name, disable, nodeTag, runnerTag string
		bypass, wantNode, wantRunner      bool
	}{
		{"omitted", "", "7.87.0", "7.87.0", false, false, false},
		{"false", "false", "7.87.0", "7.87.0", false, false, false},
		{"disabled", "true", "7.87.0", "7.87.0", false, true, true},
		{"older node", "true", "7.86.0", "7.87.0", false, true, true},
		{"older runner", "true", "7.87.0", "7.86.0", false, true, true},
		{"newer", "true", "7.88.0", "7.88.0", false, true, true},
		{"prerelease and variant", "true", "7.87.0-rc.1", "7.87.0-jmx", false, true, true},
		{"floating tags", "true", "7", "latest", false, true, true},
		{"custom images", "true", "custom", "custom", true, true, true},
		{"false with custom images", "false", "custom", "custom", true, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			values := map[string]string{"agents.image.tag": tc.nodeTag, "clusterChecksRunner.image.tag": tc.runnerTag}
			if tc.disable != "" {
				values["datadog.jemalloc.disable"] = tc.disable
			}
			if tc.bypass {
				values["agents.image.doNotCheckTag"], values["clusterChecksRunner.image.doNotCheckTag"] = "true", "true"
			}
			for template, want := range map[string]bool{"daemonset": tc.wantNode, "agent-clusterchecks-deployment": tc.wantRunner} {
				pod := jemallocPod(t, template, values, nil)
				require.NotEmpty(t, pod.Containers)
				env := allocatorEnv(pod.Containers[0])
				if want {
					assert.Equal(t, jemallocDisabledPreload, env["LD_PRELOAD"].Value, template)
					assert.NotContains(t, env, "MALLOC_CONF", template)
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
		{"node sidecars", "daemonset", map[string]string{"datadog.processAgent.processCollection": "true", "datadog.networkMonitoring.enabled": "true", "datadog.securityAgent.compliance.enabled": "true", "datadog.otelCollector.enabled": "true"}, []string{"agent", "trace-agent", "process-agent", "security-agent", "system-probe"}},
		{"core only", "daemonset", map[string]string{"datadog.apm.socketEnabled": "false", "datadog.discovery.enabled": "false"}, []string{"agent"}},
		{"runner", "agent-clusterchecks-deployment", nil, []string{"agent"}},
		{"cluster agent", "cluster-agent-deployment", nil, nil},
		{"windows", "daemonset", map[string]string{"targetSystem": "windows", "datadog.processAgent.processCollection": "true", "datadog.processAgent.runInCoreAgent": "false"}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			values := copyMap(tc.values)
			before := jemallocPod(t, tc.template, values, nil)
			values["datadog.jemalloc.disable"] = "true"
			after := jemallocPod(t, tc.template, values, nil)
			var affected []string
			for i, c := range after.Containers {
				if len(allocatorEnv(c)) == 0 {
					continue
				}
				affected = append(affected, c.Name)
				assert.Equal(t, jemallocDisabledPreload, allocatorEnv(c)["LD_PRELOAD"].Value, c.Name)
				assert.NotContains(t, allocatorEnv(c), "MALLOC_CONF", c.Name)
				var restored []corev1.EnvVar
				for _, e := range c.Env {
					if e.Name != "LD_PRELOAD" && e.Name != "MALLOC_CONF" {
						restored = append(restored, e)
					}
				}
				assert.Len(t, c.Env, len(restored)+1)
				after.Containers[i].Env = restored
			}
			assert.ElementsMatch(t, tc.want, affected)
			assert.Equal(t, before, after, "only the selected runtime environments should change")
		})
	}
}

func TestJemallocOverrides(t *testing.T) {
	json := map[string]string{
		"datadog.envDict":                         `{"LD_PRELOAD":"/global.so","MALLOC_CONF":"narenas:4"}`,
		"agents.containers.agent.env":             `[{"name":"EXTRA_LIB","value":"/core.so"},{"name":"LD_PRELOAD","value":"$(EXTRA_LIB)"}]`,
		"agents.containers.traceAgent.envDict":    `{"LD_PRELOAD":"","MALLOC_CONF":{"valueFrom":{"secretKeyRef":{"name":"allocator","key":"config"}}}}`,
		"agents.containers.processAgent.env":      `[{"name":"LD_PRELOAD","value":"/process.so"}]`,
		"agents.containers.securityAgent.envDict": `{"LD_PRELOAD":{"valueFrom":{"configMapKeyRef":{"name":"allocator","key":"preload"}}}}`,
		"clusterChecksRunner.env":                 `[{"name":"LD_PRELOAD","value":"/list.so"}]`,
		"clusterChecksRunner.envDict":             `{"LD_PRELOAD":"/runner.so"}`,
	}
	for _, template := range []string{"daemonset", "agent-clusterchecks-deployment"} {
		t.Run(template, func(t *testing.T) {
			values := map[string]string{
				"datadog.jemalloc.disable":                 "true",
				"datadog.processAgent.processCollection":   "true",
				"datadog.networkMonitoring.enabled":        "true",
				"datadog.securityAgent.compliance.enabled": "true",
			}
			pod := jemallocPod(t, template, values, json)
			for _, c := range pod.Containers {
				env := allocatorEnv(c)
				switch {
				case template == "agent-clusterchecks-deployment":
					assert.Equal(t, "/runner.so", env["LD_PRELOAD"].Value)
					assert.NotContains(t, env, "MALLOC_CONF", "runners do not inherit global overrides")
				case c.Name == "agent":
					assert.Equal(t, "$(EXTRA_LIB)", env["LD_PRELOAD"].Value)
				case c.Name == "process-agent":
					assert.Equal(t, "/process.so", env["LD_PRELOAD"].Value)
					assert.Equal(t, "narenas:4", env["MALLOC_CONF"].Value)
				case c.Name == "security-agent":
					require.NotNil(t, env["LD_PRELOAD"].ValueFrom)
					assert.Equal(t, "allocator", env["LD_PRELOAD"].ValueFrom.ConfigMapKeyRef.Name)
					assert.Equal(t, "narenas:4", env["MALLOC_CONF"].Value)
				case c.Name == "trace-agent":
					assert.Empty(t, env["LD_PRELOAD"].Value)
					require.NotNil(t, env["MALLOC_CONF"].ValueFrom)
					assert.Equal(t, "allocator", env["MALLOC_CONF"].ValueFrom.SecretKeyRef.Name)
				default:
					assert.Equal(t, "/global.so", env["LD_PRELOAD"].Value)
					assert.Equal(t, "narenas:4", env["MALLOC_CONF"].Value)
				}
			}
			values["datadog.jemalloc.disable"] = "false"
			withFalse := jemallocPod(t, template, values, json)
			delete(values, "datadog.jemalloc.disable")
			assert.Equal(t, jemallocPod(t, template, values, json), withFalse, "false preserves explicit overrides")
		})
	}
}

func TestJemallocGKE(t *testing.T) {
	for _, tc := range []struct {
		name, provider, forceRender, disable string
		wantError                            bool
	}{
		{"legacy Autopilot disable", "providers.gke.autopilot", "false", "true", true},
		{"WorkloadAllowlist Autopilot disable", "providers.gke.autopilot", "true", "true", true},
		{"legacy Autopilot default", "providers.gke.autopilot", "false", "", false},
		{"WorkloadAllowlist Autopilot default", "providers.gke.autopilot", "true", "", false},
		{"legacy Autopilot false", "providers.gke.autopilot", "false", "false", false},
		{"WorkloadAllowlist Autopilot false", "providers.gke.autopilot", "true", "false", false},
		{"GDC", "providers.gke.gdc", "false", "true", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			values := map[string]string{
				"datadog.apiKeyExistingSecret":      "datadog-secret",
				"agents.image.tag":                  "7.87.0",
				tc.provider:                         "true",
				"datadog.envDict.HELM_FORCE_RENDER": tc.forceRender,
			}
			if tc.disable != "" {
				values["datadog.jemalloc.disable"] = tc.disable
			}
			manifest, err := common.RenderChart(t, common.HelmCommand{
				ReleaseName: "datadog", ChartPath: "../../charts/datadog", ShowOnly: []string{"templates/daemonset.yaml"},
				Overrides: values,
			})
			if tc.wantError {
				require.ErrorContains(t, err, "Disabling jemalloc via LD_PRELOAD is not supported on GKE Autopilot")
				assert.ErrorContains(t, err, "Set 'datadog.jemalloc.disable' to 'false'")
				return
			}
			require.NoError(t, err)
			var ds appsv1.DaemonSet
			common.Unmarshal(t, manifest, &ds)
			require.NotEmpty(t, ds.Spec.Template.Spec.Containers)
			for _, c := range ds.Spec.Template.Spec.Containers {
				if c.Name == "agent" || c.Name == "trace-agent" || c.Name == "system-probe" {
					env := allocatorEnv(c)
					if tc.disable == "true" {
						assert.Equal(t, jemallocDisabledPreload, env["LD_PRELOAD"].Value, c.Name)
						assert.NotContains(t, env, "MALLOC_CONF", c.Name)
					} else {
						assert.Empty(t, env, c.Name)
					}
				}
			}
			if tc.forceRender == "true" {
				verifyAutopilotWorkloadAllowlistConstraints(t, manifest)
			}
		})
	}
}
