package datadog

import (
	"fmt"
	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"testing"
)

var allowedAutopilotHostPaths = map[string]interface{}{
	"/var/log/pods":                     nil,
	"/var/log/containers":               nil,
	"/var/autopilot/addon/datadog/logs": nil,
	"/var/lib/docker/containers":        nil,
	"/proc":                             nil,
	"/sys/fs/cgroup":                    nil,
	"/etc/passwd":                       nil,
	"/var/run/containerd":               nil,
}

func Test_autopilotConfigs(t *testing.T) {
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
					"datadog.apiKeyExistingSecret": "datadog-secret",
					"datadog.appKeyExistingSecret": "datadog-secret",
					"providers.gke.autopilot":      "true",
				},
			},
			assertions: verifyDaemonsetAutopilotMinimal,
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

func verifyDaemonsetAutopilotMinimal(t *testing.T, manifest string) {
	var ds appsv1.DaemonSet
	common.Unmarshal(t, manifest, &ds)
	agentContainer := &corev1.Container{}
	processAgentContainer := &corev1.Container{}

	assert.Equal(t, 2, len(ds.Spec.Template.Spec.Containers))

	for _, container := range ds.Spec.Template.Spec.Containers {
		if container.Name == "agent" {
			agentContainer = &container
		} else if container.Name == "process-agent" {
			processAgentContainer = &container
		}
	}

	assert.NotNil(t, agentContainer)
	assert.NotNil(t, processAgentContainer)

	var validHostPath = true
	for _, volume := range ds.Spec.Template.Spec.Volumes {
		if volume.HostPath != nil {
			_, validHostPath = allowedAutopilotHostPaths[volume.HostPath.Path]
			assert.True(t, validHostPath, fmt.Sprintf("DaemonSet has restricted hostPath mounted: %s ", volume.HostPath.Path))
		}
	}

	volumeNames := common.GetVolumeNames(ds)
	for _, container := range ds.Spec.Template.Spec.Containers {
		for _, volumeMount := range container.VolumeMounts {
			assert.True(t, common.Contains(volumeMount.Name, volumeNames),
				fmt.Sprintf("Found unexpected volumeMount `%s` in container `%s`", volumeMount.Name, container.Name))
		}
	}

	validPorts := true
	for _, container := range ds.Spec.Template.Spec.Containers {
		if container.Ports != nil {
			for _, port := range container.Ports {
				if port.HostPort > 0 {
					validPorts = false
					break
				}
			}
		}
	}
	assert.True(t, validPorts, "Daemonset has restricted hostPort mounted.")
}
