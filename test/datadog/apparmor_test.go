package datadog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"

	"github.com/DataDog/helm-charts/test/common"
)

func TestApparmor(t *testing.T) {
	t.Run("<1.30", func(t *testing.T) {
		manifest, err := common.RenderChart(t, common.HelmCommand{
			ReleaseName: "datadog",
			ChartPath:   "../../charts/datadog",
			ShowOnly:    []string{"templates/daemonset.yaml"},
			Values:      []string{"../../charts/datadog/values.yaml"},
			Overrides: map[string]string{
				"datadog.apiKeyExistingSecret":        "datadog-secret",
				"datadog.appKeyExistingSecret":        "datadog-secret",
				"datadog.sbom.containerImage.enabled": "true",
				"datadog.networkMonitoring.enabled":   "true",
			},
			ExtraArgs: []string{"--kube-version=1.29.8"},
		})
		require.NoError(t, err)
		var deployment appsv1.DaemonSet
		common.Unmarshal(t, manifest, &deployment)
		coreAgentContainer, ok := getContainer(t, deployment.Spec.Template.Spec.Containers, "agent")
		if assert.True(t, ok, "has agent container") {
			assert.Nil(t, coreAgentContainer.SecurityContext.AppArmorProfile, "agent apparmor profile")
		}
		systemProbeContainer, ok := getContainer(t, deployment.Spec.Template.Spec.Containers, "system-probe")
		if assert.True(t, ok, "has system-probe container") {
			assert.Nil(t, systemProbeContainer.SecurityContext.AppArmorProfile, "system-probe apparmor profile")
		}
		assert.Contains(t, deployment.Spec.Template.Annotations, "container.apparmor.security.beta.kubernetes.io/agent")
		assert.Contains(t, deployment.Spec.Template.Annotations, "container.apparmor.security.beta.kubernetes.io/system-probe")
	})
	t.Run("1.30+", func(t *testing.T) {
		manifest, err := common.RenderChart(t, common.HelmCommand{
			ReleaseName: "datadog",
			ChartPath:   "../../charts/datadog",
			ShowOnly:    []string{"templates/daemonset.yaml"},
			Values:      []string{"../../charts/datadog/values.yaml"},
			Overrides: map[string]string{
				"datadog.apiKeyExistingSecret":        "datadog-secret",
				"datadog.appKeyExistingSecret":        "datadog-secret",
				"datadog.sbom.containerImage.enabled": "true",
				"datadog.networkMonitoring.enabled":   "true",
			},
			ExtraArgs: []string{"--kube-version=1.30.4"},
		})
		require.NoError(t, err)
		var deployment appsv1.DaemonSet
		common.Unmarshal(t, manifest, &deployment)
		coreAgentContainer, ok := getContainer(t, deployment.Spec.Template.Spec.Containers, "agent")
		if assert.True(t, ok, "has agent container") {
			profile := coreAgentContainer.SecurityContext.AppArmorProfile
			if assert.NotNil(t, profile, "agent apparmor profile") {
				assert.Equal(t, v1.AppArmorProfileTypeUnconfined, profile.Type, "agent apparmor profile type")
			}
		}
		systemProbeContainer, ok := getContainer(t, deployment.Spec.Template.Spec.Containers, "system-probe")
		if assert.True(t, ok, "has system-probe container") {
			profile := systemProbeContainer.SecurityContext.AppArmorProfile
			if assert.NotNil(t, profile, "system-probe apparmor profile") {
				assert.Equal(t, v1.AppArmorProfileTypeUnconfined, profile.Type, "system-probe apparmor profile type")
			}
		}
		assert.NotContains(t, deployment.Spec.Template.Annotations, "container.apparmor.security.beta.kubernetes.io/agent")
		assert.NotContains(t, deployment.Spec.Template.Annotations, "container.apparmor.security.beta.kubernetes.io/system-probe")
	})
}
