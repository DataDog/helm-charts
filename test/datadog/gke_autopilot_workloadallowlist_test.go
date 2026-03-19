package datadog

import (
	"fmt"
	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"testing"
)

// Capabilities allowed by the Datadog WorkloadAllowlist (system-probe securityContext).
// Keep in sync with the Datadog WorkloadAllowlist.
var workloadAllowlistAllowedCapabilities = map[corev1.Capability]struct{}{
	"BPF":             {},
	"CHOWN":           {},
	"DAC_READ_SEARCH": {},
	"IPC_LOCK":        {},
	"NET_ADMIN":       {},
	"NET_BROADCAST":   {},
	"NET_RAW":         {},
	"SYS_ADMIN":       {},
	"SYS_PTRACE":      {},
	"SYS_RESOURCE":    {},
}

// hostPaths exempted by the Datadog WorkloadAllowlist.
// Keep in sync with the Datadog WorkloadAllowlist.
var workloadAllowlistExemptedHostPaths = map[string]interface{}{
	// agent / process-agent / trace-agent
	"/var/run/datadog":                   nil,
	"/var/lib/docker/containers":         nil,
	"/var/run/containerd":                nil,
	"/sys/fs/cgroup":                     nil,
	"/var/log/containers":                nil,
	"/proc":                              nil,
	"/etc/passwd":                        nil,
	"/var/autopilot/addon/datadog/logs":  nil,
	"/var/log/pods":                      nil,
	"/etc/os-release":                    nil,
	// system-probe
	"/sys/kernel/debug":                                  nil,
	"/var/tmp/datadog-agent/system-probe/build":          nil,
	"/var/tmp/datadog-agent/system-probe/kernel-headers": nil,
	"/var/lib/kubelet/seccomp":                           nil,
	"/":                                                  nil,
	"/lib/modules":                                       nil,
	"/sys/fs/bpf":                                        nil,
	// runtime compilation / package management
	"/etc/apt":         nil,
	"/etc/yum.repos.d": nil,
	"/etc/zypp":        nil,
	"/etc/pki":         nil,
	"/etc/yum/vars":    nil,
	"/etc/dnf/vars":    nil,
	"/etc/rhsm":        nil,
}

// Test_autopilotWorkloadAllowlistConfigs tests GKE Autopilot with WorkloadAllowlist.
// HELM_FORCE_RENDER=true simulates a cluster with WorkloadAllowlist CRDs available
// (GKE >= 1.32.1-gke.1729000). On real clusters the CRDs are detected automatically.
func Test_autopilotWorkloadAllowlistConfigs(t *testing.T) {
	tests := []struct {
		name       string
		command    common.HelmCommand
		assertions func(t *testing.T, manifest string)
	}{
		{
			name: "default",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.envDict.HELM_FORCE_RENDER": "true",
					"datadog.apiKeyExistingSecret":      "datadog-secret",
					"datadog.appKeyExistingSecret":      "datadog-secret",
					"providers.gke.autopilot":           "true",
				},
			},
			assertions: func(t *testing.T, manifest string) {
				var ds appsv1.DaemonSet
				common.Unmarshal(t, manifest, &ds)
				requireContainerNames(t, ds, "agent")
				verifyAutopilotWorkloadAllowlistConstraints(t, manifest)
			},
		},
		{
			// Exercises system-probe features to catch hostPath and capability violations
			// when npm/usm/enforcement are enabled (e.g. KILL from CWS enforcement).
			name: "with system-probe",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.envDict.HELM_FORCE_RENDER":                 "true",
					"datadog.apiKeyExistingSecret":                      "datadog-secret",
					"datadog.appKeyExistingSecret":                      "datadog-secret",
					"providers.gke.autopilot":                           "true",
					"datadog.networkMonitoring.enabled":                 "true",
					"datadog.serviceMonitoring.enabled":                 "true",
					"datadog.systemProbe.enableTCPQueueLength":          "true",
					"datadog.systemProbe.enableOOMKill":                 "true",
					"datadog.securityAgent.runtime.enforcement.enabled": "true",
				},
			},
			assertions: func(t *testing.T, manifest string) {
				var ds appsv1.DaemonSet
				common.Unmarshal(t, manifest, &ds)
				requireContainerNames(t, ds, "agent", "process-agent", "system-probe")
				verifyAutopilotWorkloadAllowlistConstraints(t, manifest)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest, err := common.RenderChart(t, tt.command)
			assert.Nil(t, err, "couldn't render template")
			tt.assertions(t, manifest)
		})
	}
}

// requireContainerNames asserts that exactly the expected container names are present.
func requireContainerNames(t *testing.T, ds appsv1.DaemonSet, expected ...string) {
	t.Helper()
	names := make([]string, 0, len(ds.Spec.Template.Spec.Containers))
	for _, c := range ds.Spec.Template.Spec.Containers {
		names = append(names, c.Name)
	}
	for _, name := range expected {
		assert.True(t, common.Contains(name, names),
			fmt.Sprintf("expected container %q to be present, got: %v", name, names))
	}
	assert.Equal(t, len(expected), len(names),
		fmt.Sprintf("unexpected containers present: %v", names))
}

// verifyAutopilotWorkloadAllowlistConstraints checks that the rendered DaemonSet
// complies with the Datadog WorkloadAllowlist: all hostPaths and capabilities are
// within the allowed sets, no forbidden volumes, no hostPorts, and all volumeMounts
// reference defined volumes.
func verifyAutopilotWorkloadAllowlistConstraints(t *testing.T, manifest string) {
	var ds appsv1.DaemonSet
	common.Unmarshal(t, manifest, &ds)

	volumeNames := common.GetVolumeNames(ds)

	for _, volume := range ds.Spec.Template.Spec.Volumes {
		if volume.HostPath != nil {
			_, allowed := workloadAllowlistExemptedHostPaths[volume.HostPath.Path]
			assert.True(t, allowed, fmt.Sprintf("volume %q uses hostPath %q not in the Datadog WorkloadAllowlist", volume.Name, volume.HostPath.Path))
		}
	}

	for _, container := range ds.Spec.Template.Spec.Containers {
		for _, port := range container.Ports {
			assert.Equal(t, int32(0), port.HostPort,
				fmt.Sprintf("container %q has hostPort %d which is not allowed", container.Name, port.HostPort))
		}
		for _, vm := range container.VolumeMounts {
			assert.True(t, common.Contains(vm.Name, volumeNames),
				fmt.Sprintf("container %q has volumeMount %q with no matching volume", container.Name, vm.Name))
		}
		if container.SecurityContext != nil && container.SecurityContext.Capabilities != nil {
			for _, cap := range container.SecurityContext.Capabilities.Add {
				_, allowed := workloadAllowlistAllowedCapabilities[cap]
				assert.True(t, allowed,
					fmt.Sprintf("container %q adds capability %q not in the Datadog WorkloadAllowlist", container.Name, cap))
			}
		}
	}

	for _, initContainer := range ds.Spec.Template.Spec.InitContainers {
		for _, port := range initContainer.Ports {
			assert.Equal(t, int32(0), port.HostPort,
				fmt.Sprintf("init container %q has hostPort %d which is not allowed", initContainer.Name, port.HostPort))
		}
		for _, vm := range initContainer.VolumeMounts {
			assert.True(t, common.Contains(vm.Name, volumeNames),
				fmt.Sprintf("init container %q has volumeMount %q with no matching volume", initContainer.Name, vm.Name))
		}
	}
}
