package datadog

import (
	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"testing"
)

var allowedHostPaths = []string{
	"/var/datadog/logs",
	"/var/log/pods",
	"/var/log/containers",
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
					"datadog.apiKeyExistingSecret": "datadog-secret",
					"datadog.appKeyExistingSecret": "datadog-secret",
					"datadog.clusterName":          "test-gdce",
					"datadog.logs.enabled":         "true",
					"agents.image.doNotCheckTag":   "true",
					"providers.gke.gdc":            "true",
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

	var validHostPaths = true
	for _, volume := range ds.Spec.Template.Spec.Volumes {
		if volume.HostPath != nil {
			for _, path := range allowedHostPaths {
				if volume.HostPath.Path != path {
					validHostPaths = false
					break
				}
			}
		}
	}
	assert.True(t, validHostPaths, "Daemonset has restricted hostPath mounted")

	validPorts := true
	for _, container := range ds.Spec.Template.Spec.Containers {
		for _, port := range container.Ports {
			if port.HostPort > 0 {
				validPorts = false
				break
			}
		}
	}
	assert.True(t, validPorts, "Daemonset has restricted hostPort mounted.")
}
