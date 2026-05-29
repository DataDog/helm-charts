package datadog

import (
	"fmt"
	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v3"
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
	"/var/run/datadog":                  nil,
	"/var/lib/docker/containers":        nil,
	"/var/run/containerd":               nil,
	"/sys/fs/cgroup":                    nil,
	"/var/log/containers":               nil,
	"/proc":                             nil,
	"/etc/passwd":                       nil,
	"/var/autopilot/addon/datadog/logs": nil,
	"/var/log/pods":                     nil,
	"/etc/os-release":                   nil,
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
				requireContainerNames(t, ds, "agent", "system-probe")
				verifyAutopilotWorkloadAllowlistConstraints(t, manifest)

				// Without featureGates the workload must match v1.0.1-v1.0.3 (which use
				// `pointerdir` and do not require hostPID=true).
				assert.False(t, ds.Spec.Template.Spec.HostPID, "hostPID should remain false on Autopilot without featureGates")
				volNames := common.GetVolumeNames(ds)
				assert.True(t, common.Contains("pointerdir", volNames), "v1.0.1-v1.0.3 expect the volume to be named `pointerdir`, got: %v", volNames)
				assert.False(t, common.Contains("datadogrun", volNames), "`datadogrun` is reserved for the v1.0.5 path (featureGates enabled)")
			},
		},
		{
			name: "with agent-data-plane",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.envDict.HELM_FORCE_RENDER":   "true",
					"datadog.apiKeyExistingSecret":        "datadog-secret",
					"datadog.appKeyExistingSecret":        "datadog-secret",
					"providers.gke.autopilot":             "true",
					"datadog.dataPlane.enabled":           "true",
					"datadog.dataPlane.dogstatsd.enabled": "true",
				},
			},
			assertions: func(t *testing.T, manifest string) {
				var ds appsv1.DaemonSet
				common.Unmarshal(t, manifest, &ds)
				requireContainerNames(t, ds, "agent", "system-probe", "agent-data-plane")
				verifyAutopilotWorkloadAllowlistConstraints(t, manifest)
			},
		},
		{
			// OTAGENT-980: when datadog.otelCollector.featureGates is set, the rendered
			// DaemonSet must still satisfy the WorkloadAllowlist (no extra hostPaths,
			// no hostPorts, no disallowed capabilities). The synchronizer is expected
			// to reference v1.0.5 for this configuration.
			name: "with otel-agent featureGates",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.envDict.HELM_FORCE_RENDER":  "true",
					"datadog.apiKeyExistingSecret":       "datadog-secret",
					"datadog.appKeyExistingSecret":       "datadog-secret",
					"providers.gke.autopilot":            "true",
					"datadog.otelCollector.enabled":      "true",
					"datadog.otelCollector.featureGates": "service.profilesSupport",
				},
			},
			assertions: func(t *testing.T, manifest string) {
				var ds appsv1.DaemonSet
				common.Unmarshal(t, manifest, &ds)
				requireContainerNames(t, ds, "agent", "system-probe", "otel-agent")
				verifyAutopilotWorkloadAllowlistConstraints(t, manifest)

				// OTAGENT-980: v1.0.5 requires hostPID=true and the logs-pointer volume named "datadogrun".
				assert.True(t, ds.Spec.Template.Spec.HostPID, "v1.0.5 matchingCriteria requires hostPID=true")
				volNames := common.GetVolumeNames(ds)
				assert.True(t, common.Contains("datadogrun", volNames), "v1.0.5 requires the volume to be named `datadogrun`, got: %v", volNames)
				assert.False(t, common.Contains("pointerdir", volNames), "`pointerdir` must NOT be present when featureGates is set; v1.0.5 expects `datadogrun`")
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
					"datadog.envDict.HELM_FORCE_RENDER":        "true",
					"datadog.apiKeyExistingSecret":             "datadog-secret",
					"datadog.appKeyExistingSecret":             "datadog-secret",
					"providers.gke.autopilot":                  "true",
					"datadog.networkMonitoring.enabled":        "true",
					"datadog.serviceMonitoring.enabled":        "true",
					"datadog.systemProbe.enableTCPQueueLength": "true",
					"datadog.systemProbe.enableOOMKill":        "true",
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

// Test_autopilotAllowlistSynchronizerPaths verifies the AllowlistSynchronizer references
// the right Datadog WorkloadAllowlist exemption versions:
//   - v1.0.4 (permits agent-data-plane) only when ADP is enabled (#2605).
//   - v1.0.5 (permits the otel-agent container when feature gates are configured) only
//     when datadog.otelCollector.featureGates is set (OTAGENT-980).
//
// Uses --api-versions to simulate a GKE cluster that supports WorkloadAllowlist CRDs
// (>= 1.32.1-gke.1729000).
func Test_autopilotAllowlistSynchronizerPaths(t *testing.T) {
	gkeCRDArgs := []string{
		"--api-versions", "auto.gke.io/v1/AllowlistSynchronizer",
		"--api-versions", "auto.gke.io/v1/WorkloadAllowlist",
		"--kube-version", "v1.33.0",
	}

	const (
		v104Path = "Datadog/datadog/datadog-datadog-daemonset-exemption-v1.0.4.yaml"
		v105Path = "Datadog/datadog/datadog-datadog-daemonset-exemption-v1.0.5.yaml"
	)

	tests := []struct {
		name       string
		overrides  map[string]string
		expectV104 bool
		expectV105 bool
	}{
		{
			name: "default (autopilot, no feature toggles)",
			overrides: map[string]string{
				"datadog.apiKeyExistingSecret": "datadog-secret",
				"datadog.appKeyExistingSecret": "datadog-secret",
				"providers.gke.autopilot":      "true",
			},
			expectV104: false,
			expectV105: false,
		},
		{
			name: "with ADP enabled",
			overrides: map[string]string{
				"datadog.apiKeyExistingSecret":        "datadog-secret",
				"datadog.appKeyExistingSecret":        "datadog-secret",
				"providers.gke.autopilot":             "true",
				"datadog.dataPlane.enabled":           "true",
				"datadog.dataPlane.dogstatsd.enabled": "true",
			},
			expectV104: true,
			expectV105: false,
		},
		{
			name: "with otelCollector enabled (no featureGates)",
			overrides: map[string]string{
				"datadog.apiKeyExistingSecret":  "datadog-secret",
				"datadog.appKeyExistingSecret":  "datadog-secret",
				"providers.gke.autopilot":       "true",
				"datadog.otelCollector.enabled": "true",
			},
			expectV104: false,
			expectV105: false,
		},
		{
			name: "with otelCollector featureGates",
			overrides: map[string]string{
				"datadog.apiKeyExistingSecret":       "datadog-secret",
				"datadog.appKeyExistingSecret":       "datadog-secret",
				"providers.gke.autopilot":            "true",
				"datadog.otelCollector.enabled":      "true",
				"datadog.otelCollector.featureGates": "service.profilesSupport",
			},
			expectV104: false,
			expectV105: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest, err := common.RenderChart(t, common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/gke_autopilot_allowlist_synchronizer.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides:   tt.overrides,
				ExtraArgs:   gkeCRDArgs,
			})
			assert.Nil(t, err, "couldn't render template")

			var synchronizer struct {
				Spec struct {
					AllowlistPaths []string `yaml:"allowlistPaths"`
				} `yaml:"spec"`
			}
			assert.NoError(t, yaml.Unmarshal([]byte(manifest), &synchronizer))

			hasV104 := common.Contains(v104Path, synchronizer.Spec.AllowlistPaths)
			assert.Equal(t, tt.expectV104, hasV104, "v1.0.4 exemption path presence (gated on dataPlane.enabled)")

			hasV105 := common.Contains(v105Path, synchronizer.Spec.AllowlistPaths)
			assert.Equal(t, tt.expectV105, hasV105, "v1.0.5 exemption path presence (gated on otelCollector.featureGates)")
		})
	}
}

// Test_autopilotAllowlistWaitJob verifies that the wait Job + RBAC are rendered only
// when datadog.otelCollector.featureGates is set on GKE Autopilot (and not GDC), so the
// DaemonSet rollout is gated on v1.0.5 being installed by the AllowlistSynchronizer.
// See OTAGENT-980.
func Test_autopilotAllowlistWaitJob(t *testing.T) {
	gkeCRDArgs := []string{
		"--api-versions", "auto.gke.io/v1/AllowlistSynchronizer",
		"--api-versions", "auto.gke.io/v1/WorkloadAllowlist",
		"--kube-version", "v1.33.0",
	}

	tests := []struct {
		name          string
		overrides     map[string]string
		expectRender  bool
	}{
		{
			name: "wait job is NOT rendered without featureGates",
			overrides: map[string]string{
				"datadog.apiKeyExistingSecret":  "datadog-secret",
				"datadog.appKeyExistingSecret":  "datadog-secret",
				"providers.gke.autopilot":       "true",
				"datadog.otelCollector.enabled": "true",
			},
			expectRender: false,
		},
		{
			name: "wait job is NOT rendered when GDC",
			overrides: map[string]string{
				"datadog.apiKeyExistingSecret":       "datadog-secret",
				"datadog.appKeyExistingSecret":       "datadog-secret",
				"providers.gke.autopilot":            "true",
				"providers.gke.gdc":                  "true",
				"datadog.otelCollector.enabled":      "true",
				"datadog.otelCollector.featureGates": "service.profilesSupport",
			},
			expectRender: false,
		},
		{
			name: "wait job is rendered when featureGates is set",
			overrides: map[string]string{
				"datadog.apiKeyExistingSecret":       "datadog-secret",
				"datadog.appKeyExistingSecret":       "datadog-secret",
				"providers.gke.autopilot":            "true",
				"datadog.otelCollector.enabled":      "true",
				"datadog.otelCollector.featureGates": "service.profilesSupport",
			},
			expectRender: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest, err := common.RenderChart(t, common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/gke_autopilot_allowlist_wait_job.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides:   tt.overrides,
				ExtraArgs:   gkeCRDArgs,
			})

			if !tt.expectRender {
				// helm template returns an error when --show-only matches no rendered content.
				assert.Error(t, err, "expected template to be empty / not rendered")
				return
			}

			assert.NoError(t, err)
			assert.Contains(t, manifest, "kind: Job", "wait Job should be rendered")
			assert.Contains(t, manifest, "kind: ServiceAccount", "ServiceAccount should be rendered")
			assert.Contains(t, manifest, "kind: ClusterRole", "ClusterRole should be rendered")
			assert.Contains(t, manifest, "kind: ClusterRoleBinding", "ClusterRoleBinding should be rendered")
			assert.Contains(t, manifest, "allowlistsynchronizer/datadog-synchronizer", "wait should target the synchronizer")
			assert.Contains(t, manifest, "--for=condition=Ready", "wait should gate on Ready condition")
			// Hook ordering: SA/Role/Binding at -2, Job at 0, after the synchronizer at -1.
			assert.Contains(t, manifest, `"helm.sh/hook-weight": "-2"`, "RBAC should run before the synchronizer")
			assert.Contains(t, manifest, `"helm.sh/hook-weight": "0"`, "Job should run after the synchronizer")
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
