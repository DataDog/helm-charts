package datadog

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/DataDog/helm-charts/test/common"
)

var hostProfilerBaseOverrides = map[string]string{
	"datadog.apiKeyExistingSecret": "datadog-secret",
	"datadog.hostProfiler.enabled": "true",
	"datadog.hostProfiler.image":   "myreg/host-profiler:v1.2.3",
}

func renderHostProfilerDaemonSet(t *testing.T, overrides map[string]string) appsv1.DaemonSet {
	t.Helper()
	manifest, err := common.RenderChart(t, common.HelmCommand{
		ReleaseName: "datadog",
		ChartPath:   "../../charts/datadog",
		ShowOnly:    []string{"templates/daemonset.yaml"},
		Values:      []string{"../../charts/datadog/values.yaml"},
		Overrides:   overrides,
	})
	require.NoError(t, err)
	var ds appsv1.DaemonSet
	common.Unmarshal(t, manifest, &ds)
	return ds
}

func TestHostProfilerSeccomp(t *testing.T) {
	ds := renderHostProfilerDaemonSet(t, hostProfilerBaseOverrides)

	// Container uses a hashed localhost seccomp profile.
	hpContainer, ok := getContainer(t, ds.Spec.Template.Spec.Containers, "host-profiler")
	require.True(t, ok, "host-profiler container should be present")
	require.NotNil(t, hpContainer.SecurityContext)
	require.NotNil(t, hpContainer.SecurityContext.SeccompProfile)
	assert.Equal(t, corev1.SeccompProfileTypeLocalhost, hpContainer.SecurityContext.SeccompProfile.Type)
	profileRef := *hpContainer.SecurityContext.SeccompProfile.LocalhostProfile
	assert.Regexp(t, `^host-profiler-[0-9a-f]{8}$`, profileRef)

	// seccomp-root volume must be present.
	var seccompVolume *corev1.Volume
	for i := range ds.Spec.Template.Spec.Volumes {
		if ds.Spec.Template.Spec.Volumes[i].Name == "host-profiler-seccomp-root" {
			seccompVolume = &ds.Spec.Template.Spec.Volumes[i]
			break
		}
	}
	require.NotNil(t, seccompVolume, "host-profiler-seccomp-root volume should be present when seccomp is enabled")
	require.NotNil(t, seccompVolume.HostPath, "host-profiler-seccomp-root volume should be a hostPath volume")
	assert.Equal(t, "/var/lib/kubelet/seccomp", seccompVolume.HostPath.Path)

	// Init container copies to the matching hashed filename.
	initContainer, ok := getContainer(t, ds.Spec.Template.Spec.InitContainers, "host-profiler-seccomp-setup")
	require.True(t, ok, "host-profiler-seccomp-setup init container should be present when seccomp is enabled")
	assert.Equal(t, "myreg/host-profiler:v1.2.3", initContainer.Image)
	assert.True(t, containsString(initContainer.Command, "/host/var/lib/kubelet/seccomp/"+profileRef),
		"init container cp destination should match the seccomp profile name; command: %v", initContainer.Command)

}

func TestHostProfilerSeccompDisabled(t *testing.T) {
	overrides := copyMap(hostProfilerBaseOverrides)
	overrides["datadog.hostProfiler.seccomp.enabled"] = "false"

	ds := renderHostProfilerDaemonSet(t, overrides)

	// Container must run Unconfined, but other hardening still applies.
	hpContainer, ok := getContainer(t, ds.Spec.Template.Spec.Containers, "host-profiler")
	require.True(t, ok, "host-profiler container should be present")
	require.NotNil(t, hpContainer.SecurityContext)
	require.NotNil(t, hpContainer.SecurityContext.SeccompProfile,
		"host-profiler should carry a seccomp profile when datadog.hostProfiler.seccomp.enabled=false")
	assert.Equal(t, corev1.SeccompProfileTypeUnconfined, hpContainer.SecurityContext.SeccompProfile.Type,
		"host-profiler should run Unconfined when datadog.hostProfiler.seccomp.enabled=false")
	assert.Nil(t, hpContainer.SecurityContext.SeccompProfile.LocalhostProfile,
		"Unconfined profile should not reference a localhost profile")

	// Seccomp setup init container must be absent.
	_, ok = getContainer(t, ds.Spec.Template.Spec.InitContainers, "host-profiler-seccomp-setup")
	assert.False(t, ok, "host-profiler-seccomp-setup init container should be absent when seccomp is disabled")

	// seccomp-root volume must be absent.
	for _, v := range ds.Spec.Template.Spec.Volumes {
		assert.NotEqual(t, "host-profiler-seccomp-root", v.Name,
			"host-profiler-seccomp-root volume should be absent when seccomp is disabled")
	}
}

func TestHostProfilerSeccompDifferentImages(t *testing.T) {
	overridesV1 := copyMap(hostProfilerBaseOverrides)
	overridesV2 := copyMap(hostProfilerBaseOverrides)
	overridesV2["datadog.hostProfiler.image"] = "myreg/host-profiler:v1.2.4"

	dsV1 := renderHostProfilerDaemonSet(t, overridesV1)
	dsV2 := renderHostProfilerDaemonSet(t, overridesV2)

	c1, _ := getContainer(t, dsV1.Spec.Template.Spec.Containers, "host-profiler")
	c2, _ := getContainer(t, dsV2.Spec.Template.Spec.Containers, "host-profiler")

	profile1 := *c1.SecurityContext.SeccompProfile.LocalhostProfile
	profile2 := *c2.SecurityContext.SeccompProfile.LocalhostProfile
	assert.NotEqual(t, profile1, profile2, "different images should produce different seccomp profile names")
}

func TestHostProfilerSCC(t *testing.T) {
	overrides := copyMap(hostProfilerBaseOverrides)
	overrides["agents.podSecurity.securityContextConstraints.create"] = "true"
	overrides["providers.openshift.enabled"] = "true"

	// Resolve the expected profile name from the DaemonSet render.
	ds := renderHostProfilerDaemonSet(t, overrides)
	hpContainer, ok := getContainer(t, ds.Spec.Template.Spec.Containers, "host-profiler")
	require.True(t, ok)
	profileRef := *hpContainer.SecurityContext.SeccompProfile.LocalhostProfile

	// Render the SCC and assert the hashed profile is in the allowlist.
	manifest, err := common.RenderChart(t, common.HelmCommand{
		ReleaseName: "datadog",
		ChartPath:   "../../charts/datadog",
		ShowOnly:    []string{"templates/agent-scc.yaml"},
		Values:      []string{"../../charts/datadog/values.yaml"},
		Overrides:   overrides,
	})
	require.NoError(t, err)
	assert.Contains(t, manifest, "localhost/"+profileRef,
		"SCC should allow the hashed seccomp profile")
}

func TestHostProfilerLoggingSeccomp(t *testing.T) {
	overrides := copyMap(hostProfilerBaseOverrides)
	overrides["datadog.hostProfiler.loggingSeccomp"] = "true"
	ds := renderHostProfilerDaemonSet(t, overrides)
	initContainer, ok := getContainer(t, ds.Spec.Template.Spec.InitContainers, "host-profiler-seccomp-setup")
	require.True(t, ok)

	// Prefer the logging profile, falling back to the default if the image predates it.
	cmd := strings.Join(initContainer.Command, " ")
	assert.Contains(t, cmd, "if [ -f /etc/dd-host-profiler/logging-seccomp.json ]",
		"init container should guard the logging profile copy; command: %v", initContainer.Command)
	assert.Contains(t, cmd, "cp /etc/dd-host-profiler/logging-seccomp.json")
	assert.Contains(t, cmd, "cp /etc/dd-host-profiler/seccomp.json", "should fall back to the default profile")
}

func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if strings.Contains(v, s) {
			return true
		}
	}
	return false
}

func copyMap(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
