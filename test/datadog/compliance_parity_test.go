package datadog

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

// complianceParityValues turns on every CSPM setting that changes what the
// container running the host checks needs, plus the platform settings
// (runtime socket, FIPS proxy, kubelet CA) those checks may depend on.
var complianceParityValues = map[string]string{
	"datadog.securityAgent.compliance.enabled":          "true",
	"datadog.securityAgent.compliance.configMap":        "custom-benchmarks",
	"datadog.securityAgent.compliance.checkInterval":    "5m",
	"datadog.securityAgent.compliance.containerInclude": "image:included",
	"datadog.securityAgent.compliance.containerExclude": "image:excluded",
	"datadog.securityAgent.compliance.xccdf.enabled":    "true",
	"datadog.dockerSocketPath":                          "/var/run/docker.sock",
	"datadog.kubelet.hostCAPath":                        "/etc/kubernetes/ca.crt",
	"fips.enabled":                                      "true",
}

// Security Agent settings that compliance in system-probe does not need. Every
// entry needs a reason: anything else the Security Agent gets must also reach
// system-probe, or moving compliance there silently drops it.
var (
	complianceParityExemptEnv = map[string]string{
		"DD_CLUSTER_AGENT_ENABLED":                 "pkg/compliance never talks to the Cluster Agent",
		"DD_CLUSTER_AGENT_KUBERNETES_SERVICE_NAME": "pkg/compliance never talks to the Cluster Agent",
		"DD_CLUSTER_AGENT_AUTH_TOKEN":              "pkg/compliance never talks to the Cluster Agent",
		"DD_RUNTIME_SECURITY_CONFIG_ENABLED":       "CWS, not compliance; set to false here",
		"DD_DOGSTATSD_SOCKET":                      "system-probe's statsd client sends over UDP to dogstatsd_port, shared by all its modules",
	}
	complianceParityExemptMounts = map[string]string{
		"dsdsocket":  "system-probe's statsd client sends over UDP to dogstatsd_port, shared by all its modules",
		"kubelet-ca": "pkg/compliance reads the kubelet's process and config files, it never calls the kubelet API",
	}
)

// complianceContainers renders the node Agent with compliance on one execution
// path and returns the container running the host checks and the Agent container.
func complianceContainers(t *testing.T, values map[string]string, inSystemProbe bool) (runner, agent corev1.Container) {
	t.Helper()
	v := copyMap(values)
	name := "security-agent"
	v["datadog.securityAgent.compliance.runInSystemProbe"] = "false"
	if inSystemProbe {
		name = "system-probe"
		v["datadog.securityAgent.compliance.runInSystemProbe"] = "true"
	}
	pod := jemallocPod(t, "daemonset", v, nil)
	var foundRunner, foundAgent bool
	for _, c := range pod.Containers {
		switch c.Name {
		case name:
			runner, foundRunner = c, true
		case "agent":
			agent, foundAgent = c, true
		}
	}
	require.True(t, foundRunner, "%s container", name)
	require.True(t, foundAgent, "agent container")
	return runner, agent
}

func envByName(c corev1.Container) map[string]corev1.EnvVar {
	env := map[string]corev1.EnvVar{}
	for _, e := range c.Env {
		env[e.Name] = e
	}
	return env
}

// TestComplianceExecutionPathParity checks that moving compliance from the
// Security Agent to system-probe keeps everything the host checks rely on:
// environment, mounts and capabilities.
func TestComplianceExecutionPathParity(t *testing.T) {
	sa, _ := complianceContainers(t, complianceParityValues, false)
	sp, _ := complianceContainers(t, complianceParityValues, true)

	spEnv := envByName(sp)
	for _, e := range sa.Env {
		if _, exempt := complianceParityExemptEnv[e.Name]; exempt {
			continue
		}
		if assert.Contains(t, spEnv, e.Name, "env var from security-agent missing in system-probe") {
			assert.Equal(t, e.Value, spEnv[e.Name].Value, "env var %s", e.Name)
		}
	}

	spMounts := map[string]corev1.VolumeMount{}
	for _, m := range sp.VolumeMounts {
		spMounts[m.Name] = m
	}
	for _, m := range sa.VolumeMounts {
		if _, exempt := complianceParityExemptMounts[m.Name]; exempt {
			continue
		}
		if assert.Contains(t, spMounts, m.Name, "volume mount from security-agent missing in system-probe") {
			assert.Equal(t, m.MountPath, spMounts[m.Name].MountPath, "volume mount %s", m.Name)
		}
	}

	require.NotNil(t, sa.SecurityContext.Capabilities)
	require.NotNil(t, sp.SecurityContext.Capabilities)
	for _, c := range sa.SecurityContext.Capabilities.Add {
		assert.Contains(t, sp.SecurityContext.Capabilities.Add, c, "capability from security-agent missing in system-probe")
	}
}

// TestComplianceExecutionPathParityConfigStream checks that under config
// streaming, where the host-check container drops its own environment, the
// Agent container carries the same compliance settings on both paths.
func TestComplianceExecutionPathParityConfigStream(t *testing.T) {
	values := copyMap(complianceParityValues)
	values["datadog.configStream.enabled"] = "true"
	complianceEnv := func(agent corev1.Container) map[string]string {
		env := map[string]string{}
		for _, e := range agent.Env {
			// These two legitimately differ: they select the execution path.
			if e.Name == "DD_COMPLIANCE_CONFIG_ENABLED" || e.Name == "DD_COMPLIANCE_CONFIG_RUN_IN_SYSTEM_PROBE" {
				continue
			}
			if strings.HasPrefix(e.Name, "DD_COMPLIANCE_CONFIG_") {
				env[e.Name] = e.Value
			}
		}
		return env
	}
	_, saAgent := complianceContainers(t, values, false)
	_, spAgent := complianceContainers(t, values, true)
	want := complianceEnv(saAgent)
	require.Contains(t, want, "DD_COMPLIANCE_CONFIG_CHECK_INTERVAL", "the Security Agent path must hoist compliance settings")
	assert.Equal(t, want, complianceEnv(spAgent))
}

// TestSystemProbeAuditCapabilities checks that the audit capabilities are only
// added when compliance runs in system-probe, and not duplicated when the user
// already grants them.
func TestSystemProbeAuditCapabilities(t *testing.T) {
	for _, tc := range []struct {
		name   string
		values map[string]string
		want   bool
	}{
		{"compliance in system-probe", map[string]string{"datadog.securityAgent.compliance.enabled": "true", "datadog.securityAgent.compliance.runInSystemProbe": "true"}, true},
		{"compliance in security-agent", map[string]string{"datadog.securityAgent.compliance.enabled": "true", "datadog.securityAgent.compliance.runInSystemProbe": "false"}, false},
		{"CWS only", map[string]string{"datadog.securityAgent.runtime.enabled": "true"}, false},
		{"user already grants AUDIT_READ", map[string]string{
			"datadog.securityAgent.compliance.enabled":          "true",
			"datadog.securityAgent.compliance.runInSystemProbe": "true",
			"agents.containers.systemProbe.securityContext.capabilities.add[0]": "SYS_ADMIN",
			"agents.containers.systemProbe.securityContext.capabilities.add[1]": "AUDIT_READ",
		}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pod := jemallocPod(t, "daemonset", tc.values, nil)
			var caps []corev1.Capability
			for _, c := range pod.Containers {
				if c.Name == "system-probe" {
					caps = c.SecurityContext.Capabilities.Add
				}
			}
			require.NotEmpty(t, caps)
			count := map[corev1.Capability]int{}
			for _, c := range caps {
				count[c]++
			}
			for _, c := range []corev1.Capability{"AUDIT_CONTROL", "AUDIT_READ"} {
				if tc.want {
					assert.Equal(t, 1, count[c], "%s in %v", c, caps)
				} else {
					assert.Zero(t, count[c], "%s in %v", c, caps)
				}
			}
		})
	}
}
