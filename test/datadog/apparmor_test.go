package datadog

import (
	"fmt"
	"strings"
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
	t.Run("1.30+_USM_GKE_COS", func(t *testing.T) {
		manifest, err := common.RenderChart(t, common.HelmCommand{
			ReleaseName: "datadog",
			ChartPath:   "../../charts/datadog",
			ShowOnly:    []string{"templates/daemonset.yaml"},
			Values:      []string{"../../charts/datadog/values.yaml"},
			Overrides: map[string]string{
				"datadog.apiKeyExistingSecret":      "datadog-secret",
				"datadog.appKeyExistingSecret":      "datadog-secret",
				"datadog.serviceMonitoring.enabled": "true",
				"providers.gke.cos":                 "true",
			},
			ExtraArgs: []string{"--kube-version=1.33.4-gke.1172000"},
		})
		require.NoError(t, err)
		var deployment appsv1.DaemonSet
		common.Unmarshal(t, manifest, &deployment)
		_, ok := getContainer(t, deployment.Spec.Template.Spec.Containers, "agent")
		assert.True(t, ok, "has agent container") // This configuration does not require the agent container container to be unconfined
		systemProbeContainer, ok := getContainer(t, deployment.Spec.Template.Spec.Containers, "system-probe")
		if assert.True(t, ok, "has system-probe container") {
			if assert.NotNil(t, systemProbeContainer.SecurityContext, "system-probe securityContext not found") {
				profile := systemProbeContainer.SecurityContext.AppArmorProfile
				if assert.NotNil(t, profile, "system-probe apparmor profile not found") {
					assert.Equal(t, v1.AppArmorProfileTypeUnconfined, profile.Type, "system-probe apparmor profile type")
				}
			}
		}
		assert.NotContains(t, deployment.Spec.Template.Annotations, "container.apparmor.security.beta.kubernetes.io/agent")
		assert.NotContains(t, deployment.Spec.Template.Annotations, "container.apparmor.security.beta.kubernetes.io/system-probe")
	})
	t.Run("1.30+_USM_SBOM_GKE_COS", func(t *testing.T) {
		manifest, err := common.RenderChart(t, common.HelmCommand{
			ReleaseName: "datadog",
			ChartPath:   "../../charts/datadog",
			ShowOnly:    []string{"templates/daemonset.yaml"},
			Values:      []string{"../../charts/datadog/values.yaml"},
			Overrides: map[string]string{
				"datadog.apiKeyExistingSecret":                          "datadog-secret",
				"datadog.appKeyExistingSecret":                          "datadog-secret",
				"datadog.serviceMonitoring.enabled":                     "true",
				"datadog.sbom.containerImage.enabled":                   "true",
				"datadog.sbom.containerImage.uncompressedLayersSupport": "true",
				"providers.gke.cos":                                     "true",
			},
			ExtraArgs: []string{"--kube-version=1.33.4-gke.1172000"},
		})
		require.NoError(t, err)
		var deployment appsv1.DaemonSet
		common.Unmarshal(t, manifest, &deployment)
		coreAgentContainer, ok := getContainer(t, deployment.Spec.Template.Spec.Containers, "agent")
		if assert.True(t, ok, "has agent container") {
			if assert.NotNil(t, coreAgentContainer.SecurityContext, "agent securityContext not found") {
			profile := coreAgentContainer.SecurityContext.AppArmorProfile
			if assert.NotNil(t, profile, "agent apparmor profile not found") {
					assert.Equal(t, v1.AppArmorProfileTypeUnconfined, profile.Type, "agent apparmor profile type")
				}
			}
		}
		systemProbeContainer, ok := getContainer(t, deployment.Spec.Template.Spec.Containers, "system-probe")
		if assert.True(t, ok, "has system-probe container") {
			if assert.NotNil(t, systemProbeContainer.SecurityContext, "system-probe securityContext not found") {
				profile := systemProbeContainer.SecurityContext.AppArmorProfile
				if assert.NotNil(t, profile, "system-probe apparmor profile not found") {
					assert.Equal(t, v1.AppArmorProfileTypeUnconfined, profile.Type, "system-probe apparmor profile type")
				}
			}
		}
		assert.NotContains(t, deployment.Spec.Template.Annotations, "container.apparmor.security.beta.kubernetes.io/agent")
		assert.NotContains(t, deployment.Spec.Template.Annotations, "container.apparmor.security.beta.kubernetes.io/system-probe")
	})

}

func innerTestApparmorGKE(t *testing.T, kubeVersion string, requiresSysprobeUnconfined bool, requiresAgentUnconfined bool) {
	helmCommand := common.HelmCommand{
		ReleaseName: "datadog",
		ChartPath:   "../../charts/datadog",
		ShowOnly:    []string{"templates/daemonset.yaml"},
		Values:      []string{"../../charts/datadog/values.yaml"},
		Overrides: map[string]string{
			"datadog.apiKeyExistingSecret": "datadog-secret",
			"datadog.appKeyExistingSecret": "datadog-secret",
			"providers.gke.cos":            "true",
		},
		ExtraArgs: []string{"--kube-version=" + kubeVersion},
	}
	if requiresSysprobeUnconfined {
		helmCommand.Overrides["datadog.serviceMonitoring.enabled"] = "true"
	}
	if requiresAgentUnconfined {
		helmCommand.Overrides["datadog.sbom.containerImage.enabled"] = "true"
		helmCommand.Overrides["datadog.sbom.containerImage.uncompressedLayersSupport"] = "true"
	}
	manifest, err := common.RenderChart(t, helmCommand)
	require.NoError(t, err)
	var deployment appsv1.DaemonSet
	common.Unmarshal(t, manifest, &deployment)

	usesAnnotation := strings.HasPrefix(kubeVersion, "1.29")
	checkApparmorConfinement := func(containerName string) {
		container, ok := getContainer(t, deployment.Spec.Template.Spec.Containers, containerName)

		if assert.True(t, ok, "has "+containerName+" container") {
			if assert.NotNil(t, container.SecurityContext, containerName+" securityContext not found") {
				profile := container.SecurityContext.AppArmorProfile
				if assert.NotNil(t, profile, containerName+" apparmor profile not found") {
					assert.Equal(t, v1.AppArmorProfileTypeUnconfined, profile.Type, containerName+" apparmor profile type")
				}
			}
		}
	}

	checkApparmorAnnotation := func(containerName string) {
		assert.NotNil(t, deployment.Spec.Template.Annotations, "annotations not found")
		assert.Equal(t, "unconfined", deployment.Spec.Template.Annotations["container.apparmor.security.beta.kubernetes.io/"+containerName], containerName+" apparmor profile type")
	}

	checkFunc := checkApparmorConfinement
	if usesAnnotation {
		checkFunc = checkApparmorAnnotation
	}

	if requiresAgentUnconfined {
		checkFunc("agent")
	}
	if requiresSysprobeUnconfined {
		checkFunc("system-probe")
	}
}

func TestApparmorGKE(t *testing.T) {
	// 1.29 should use annotation for the apparmor profile, 1.30+ should use securityContext
	kubeVersions := []string{"1.29.4-gke.1172000", "1.33.4-gke.1172000"}
	for _, kubeVersion := range kubeVersions {
		t.Run(kubeVersion, func(t *testing.T) {
			for _, requiresSysprobeUnconfined := range []bool{true, false} {
				for _, requiresAgentUnconfined := range []bool{true, false} {
					sysprobeConfinementStr := "Confined"
					if requiresSysprobeUnconfined {
						sysprobeConfinementStr = "Unconfined"
					}
					agentConfinementStr := "Confined"
					if requiresAgentUnconfined {
						agentConfinementStr = "Unconfined"
					}
					testName := fmt.Sprintf("agent%s_sysprobe%s", agentConfinementStr, sysprobeConfinementStr)
					t.Run(testName, func(t *testing.T) {
						innerTestApparmorGKE(t, kubeVersion, requiresSysprobeUnconfined, requiresAgentUnconfined)
					})
				}
			}
		})
	}
}
