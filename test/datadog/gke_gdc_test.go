package datadog

import (
	"fmt"
	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	"testing"
)

// hostPaths permitted in GKE Distributed Cloud (GDC) environments.
// GDC is more restricted than GKE Autopilot: /proc, /sys/fs/cgroup, and most
// system-level paths are not available.
var allowedGDCHostPaths = map[string]interface{}{
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
					"providers.gke.gdc":                      "true",
				},
			},
			assertions: verifyDaemonsetGDCConstraints,
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

// verifyDaemonsetGDCConstraints checks that the rendered DaemonSet complies with GDC
// constraints: only 1 container, no forbidden volumes, all hostPaths within the allowed
// set, no hostPorts, and all volumeMounts reference defined volumes.
func verifyDaemonsetGDCConstraints(t *testing.T, manifest string) {
	var ds appsv1.DaemonSet
	common.Unmarshal(t, manifest, &ds)

	assert.Equal(t, 1, len(ds.Spec.Template.Spec.Containers), "GDC only supports the core agent container")

	volumeNames := common.GetVolumeNames(ds)

	for _, volume := range ds.Spec.Template.Spec.Volumes {
		if volume.HostPath != nil {
			_, allowed := allowedGDCHostPaths[volume.HostPath.Path]
			assert.True(t, allowed, fmt.Sprintf("volume %q uses hostPath %q which is not allowed in GDC", volume.Name, volume.HostPath.Path))
		}
	}

	for _, container := range ds.Spec.Template.Spec.Containers {
		for _, port := range container.Ports {
			assert.Equal(t, int32(0), port.HostPort,
				fmt.Sprintf("container %q has hostPort %d which is not allowed in GDC", container.Name, port.HostPort))
		}
		for _, vm := range container.VolumeMounts {
			assert.True(t, common.Contains(vm.Name, volumeNames),
				fmt.Sprintf("container %q has volumeMount %q with no matching volume", container.Name, vm.Name))
		}
	}

	for _, initContainer := range ds.Spec.Template.Spec.InitContainers {
		for _, port := range initContainer.Ports {
			assert.Equal(t, int32(0), port.HostPort,
				fmt.Sprintf("init container %q has hostPort %d which is not allowed in GDC", initContainer.Name, port.HostPort))
		}
		for _, vm := range initContainer.VolumeMounts {
			assert.True(t, common.Contains(vm.Name, volumeNames),
				fmt.Sprintf("init container %q has volumeMount %q with no matching volume", initContainer.Name, vm.Name))
		}
	}
}
