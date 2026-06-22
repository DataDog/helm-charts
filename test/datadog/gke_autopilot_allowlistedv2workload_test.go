package datadog

import (
	"fmt"
	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	"testing"
)

// hostPaths permitted in GKE Autopilot AllowlistedV2Workload mode (GKE < 1.32.1-gke.1729000).
// The allowlist also permits process-agent and trace-agent, but the current chart runs
// these in-process inside the core agent container, so only 1 container is rendered.
// system-probe and otel-agent are not permitted by the allowlist.
var allowlistedV2WorkloadExemptedHostPaths = map[string]interface{}{
	"/var/log/pods":                     nil,
	"/var/log/containers":               nil,
	"/var/autopilot/addon/datadog/logs": nil,
	"/var/lib/docker/containers":        nil,
	"/proc":                             nil,
	"/sys/fs/cgroup":                    nil,
	"/etc/passwd":                       nil,
	"/var/run/containerd":               nil,
}

// Test_autopilotAllowlistedV2WorkloadConfigs tests GKE Autopilot in AllowlistedV2Workload
// (legacy) mode. HELM_FORCE_RENDER=false simulates clusters without WorkloadAllowlist CRDs.
func Test_autopilotAllowlistedV2WorkloadConfigs(t *testing.T) {
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
					"datadog.envDict.HELM_FORCE_RENDER": "false",
					"datadog.apiKeyExistingSecret":      "datadog-secret",
					"datadog.appKeyExistingSecret":      "datadog-secret",
					"providers.gke.autopilot":           "true",
				},
			},
			assertions: verifyDaemonsetAutopilotAllowlistedV2WorkloadMinimal,
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

func verifyDaemonsetAutopilotAllowlistedV2WorkloadMinimal(t *testing.T, manifest string) {
	var ds appsv1.DaemonSet
	common.Unmarshal(t, manifest, &ds)

	assert.Equal(t, 1, len(ds.Spec.Template.Spec.Containers))
	assert.Equal(t, ds.Spec.Template.Spec.Containers[0].Name, "agent")

	volumeNames := common.GetVolumeNames(ds)

	for _, volume := range ds.Spec.Template.Spec.Volumes {
		if volume.HostPath != nil {
			_, allowed := allowlistedV2WorkloadExemptedHostPaths[volume.HostPath.Path]
			assert.True(t, allowed, fmt.Sprintf("volume %q uses hostPath %q which is not allowed in AllowlistedV2Workload mode", volume.Name, volume.HostPath.Path))
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
