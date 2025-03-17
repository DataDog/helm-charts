package datadog

import (
	"fmt"
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

var allowedHostPaths = map[string]interface{}{
	"/var/datadog/logs":   nil,
	"/var/log/pods":       nil,
	"/var/log/containers": nil,
}

func Test_gdcConfigs(t *testing.T) {
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
					"datadog.apiKeyExistingSecret":            "datadog-secret",
					"datadog.appKeyExistingSecret":            "datadog-secret",
					"datadog.logs.enabled":                    "true",
					"agents.image.doNotCheckTag":              "true",
					"datadog.logs.containerCollectAll":        "true",
					"datadog.logs.containerCollectUsingFiles": "true",
					"datadog.logs.autoMultiLineDetection":     "true",
					"providers.gke.gdc":                       "true",
				},
			},
			assertions: verifyDaemonsetGDCMinimal,
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

func verifyDaemonsetGDCMinimal(t *testing.T, manifest string) {
	var ds appsv1.DaemonSet
	common.Unmarshal(t, manifest, &ds)
	agentContainer := &corev1.Container{}

	assert.Equal(t, 1, len(ds.Spec.Template.Spec.Containers))

	for _, container := range ds.Spec.Template.Spec.Containers {
		if container.Name == "agent" {
			agentContainer = &container
		}
	}

	assert.NotNil(t, agentContainer)

	var validHostPath = true
	for _, volume := range ds.Spec.Template.Spec.Volumes {
		if volume.HostPath != nil {
			_, validHostPath = allowedHostPaths[volume.HostPath.Path]
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
