package datadog

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/DataDog/helm-charts/test/common"
)

const (
	pcapVolumeName = "sysprobe-capture-socket-dir"
	pcapSocketDir  = "/var/run/sysprobe-capture"
	pcapSocketPath = "/var/run/sysprobe-capture/capture.sock"
	pcapSocketEnv  = "DD_PACKET_CAPTURE_SOCKET"
)

var pcapCapabilities = []corev1.Capability{"SYS_ADMIN", "SYS_RESOURCE", "NET_ADMIN", "NET_RAW", "IPC_LOCK"}

// pcapOverrides enables the node Private Action Runner with packet capture and turns off discovery,
// the only system-probe feature enabled by default, so system-probe renders for packet capture alone.
func pcapOverrides(tag string) map[string]string {
	return map[string]string{
		"agents.image.tag":                                  tag,
		"datadog.apiKeyExistingSecret":                      "datadog-secret",
		"datadog.discovery.enabled":                         "false",
		"datadog.privateActionRunner.enabled":               "true",
		"datadog.privateActionRunner.selfEnroll":            "true",
		"datadog.privateActionRunner.packetCapture.enabled": "true",
	}
}

func renderPcapDaemonSet(t *testing.T, overrides map[string]string) appsv1.DaemonSet {
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

// renderPcapSystemProbeConfig returns the parsed system-probe.yaml, or nil when the ConfigMap is not rendered.
func renderPcapSystemProbeConfig(t *testing.T, overrides map[string]string) map[string]any {
	t.Helper()
	manifest, err := common.RenderChart(t, common.HelmCommand{
		ReleaseName: "datadog",
		ChartPath:   "../../charts/datadog",
		ShowOnly:    []string{"templates/system-probe-configmap.yaml"},
		Values:      []string{"../../charts/datadog/values.yaml"},
		Overrides:   overrides,
	})
	if err != nil && strings.Contains(err.Error(), "could not find template") {
		return nil
	}
	require.NoError(t, err)
	var cm corev1.ConfigMap
	common.Unmarshal(t, manifest, &cm)
	raw, ok := cm.Data["system-probe.yaml"]
	if !ok {
		return nil
	}
	cfg := map[string]any{}
	require.NoError(t, yaml.Unmarshal([]byte(raw), &cfg))
	return cfg
}

func pcapMount(c corev1.Container) (corev1.VolumeMount, bool) {
	for _, m := range c.VolumeMounts {
		if m.Name == pcapVolumeName {
			return m, true
		}
	}
	return corev1.VolumeMount{}, false
}

func pcapVolume(ds appsv1.DaemonSet) (corev1.Volume, bool) {
	for _, v := range ds.Spec.Template.Spec.Volumes {
		if v.Name == pcapVolumeName {
			return v, true
		}
	}
	return corev1.Volume{}, false
}

// assertPcapAbsent checks that no packet capture wiring is rendered on any container or volume.
func assertPcapAbsent(t *testing.T, ds appsv1.DaemonSet) {
	t.Helper()
	_, found := pcapVolume(ds)
	assert.False(t, found, "capture socket volume must not be rendered")
	for _, c := range ds.Spec.Template.Spec.Containers {
		_, found := pcapMount(c)
		assert.False(t, found, "container %s must not mount the capture socket", c.Name)
		env := getEnvVarMap(c.Env)
		assert.NotContains(t, env, pcapSocketEnv, "container %s", c.Name)
	}
}

func TestPrivateActionRunnerPacketCapture_DisabledByDefault(t *testing.T) {
	overrides := pcapOverrides("7.86.0")
	delete(overrides, "datadog.privateActionRunner.packetCapture.enabled")
	ds := renderPcapDaemonSet(t, overrides)

	assertPcapAbsent(t, ds)
	_, found := getContainer(t, ds.Spec.Template.Spec.Containers, "system-probe")
	assert.False(t, found, "system-probe must not be added when packet capture is disabled")
	_, found = getContainer(t, ds.Spec.Template.Spec.Containers, "private-action-runner")
	assert.True(t, found)
}

func TestPrivateActionRunnerPacketCapture_Enabled(t *testing.T) {
	for _, test := range []struct {
		name  string
		extra map[string]string
	}{
		{name: "default runner"},
		{name: "split mode runner", extra: map[string]string{"datadog.privateActionRunner.splitEnabled": "true"}},
		{
			// Other containers are rendered so the test can check none of them gets the capture socket.
			name: "with security agent and trace agent",
			extra: map[string]string{
				"datadog.securityAgent.compliance.enabled": "true",
				"datadog.apm.enabled":                      "true",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			overrides := pcapOverrides("7.86.0")
			for k, v := range test.extra {
				overrides[k] = v
			}
			ds := renderPcapDaemonSet(t, overrides)

			vol, found := pcapVolume(ds)
			require.True(t, found, "capture socket volume")
			assert.NotNil(t, vol.EmptyDir)
			assert.Nil(t, vol.HostPath)

			sp, found := getContainer(t, ds.Spec.Template.Spec.Containers, "system-probe")
			require.True(t, found, "system-probe must run for packet capture alone")
			mount, found := pcapMount(sp)
			require.True(t, found)
			assert.Equal(t, pcapSocketDir, mount.MountPath)
			assert.False(t, mount.ReadOnly, "system-probe creates the capture socket")
			require.NotNil(t, sp.SecurityContext)
			require.NotNil(t, sp.SecurityContext.Capabilities)
			for _, capability := range pcapCapabilities {
				assert.Contains(t, sp.SecurityContext.Capabilities.Add, capability)
			}
			hasSocketDir := false
			for _, m := range sp.VolumeMounts {
				if m.Name == "sysprobe-socket-dir" {
					hasSocketDir = true
					assert.False(t, m.ReadOnly)
				}
			}
			assert.True(t, hasSocketDir, "system-probe needs a writable socket dir for its main socket")

			par, found := getContainer(t, ds.Spec.Template.Spec.Containers, "private-action-runner")
			require.True(t, found)
			mount, found = pcapMount(par)
			require.True(t, found)
			assert.Equal(t, pcapSocketDir, mount.MountPath)
			assert.True(t, mount.ReadOnly, "the runner only dials the capture socket")
			assert.Equal(t, pcapSocketPath, getEnvVarMap(par.Env)[pcapSocketEnv])
			for _, m := range par.VolumeMounts {
				assert.NotEqual(t, "sysprobe-socket-dir", m.Name, "the full system-probe socket must not reach the runner")
				assert.NotEqual(t, "/var/run/sysprobe", m.MountPath)
			}

			// Only system-probe and the runner mount the capture socket, and only the runner gets the socket
			// env var: system-probe takes its settings from system-probe.yaml.
			for _, c := range ds.Spec.Template.Spec.Containers {
				if c.Name == "private-action-runner" {
					continue
				}
				env := getEnvVarMap(c.Env)
				assert.NotContains(t, env, pcapSocketEnv, "container %s", c.Name)
				if c.Name == "system-probe" {
					continue
				}
				_, found := pcapMount(c)
				assert.False(t, found, "container %s must not mount the capture socket", c.Name)
			}

			cfg := renderPcapSystemProbeConfig(t, overrides)
			require.NotNil(t, cfg)
			pc, ok := cfg["packet_capture"].(map[string]any)
			require.True(t, ok, "system-probe.yaml must have a packet_capture section")
			assert.Equal(t, true, pc["enabled"])
			assert.Equal(t, pcapSocketPath, pc["socket"])
		})
	}
}

func TestPrivateActionRunnerPacketCapture_AddsCapabilitiesToCustomSecurityContext(t *testing.T) {
	manifest, err := common.RenderChart(t, common.HelmCommand{
		ReleaseName: "datadog",
		ChartPath:   "../../charts/datadog",
		ShowOnly:    []string{"templates/daemonset.yaml"},
		Values:      []string{"../../charts/datadog/values.yaml"},
		Overrides:   pcapOverrides("7.86.0"),
		OverridesJson: map[string]string{
			"agents.containers.systemProbe.securityContext": `{"privileged":false,"capabilities":{"add":["BPF","NET_RAW"]}}`,
		},
	})
	require.NoError(t, err)
	var ds appsv1.DaemonSet
	common.Unmarshal(t, manifest, &ds)

	sp, found := getContainer(t, ds.Spec.Template.Spec.Containers, "system-probe")
	require.True(t, found)
	require.NotNil(t, sp.SecurityContext)
	require.NotNil(t, sp.SecurityContext.Capabilities)
	assert.ElementsMatch(t,
		[]corev1.Capability{"BPF", "NET_RAW", "SYS_ADMIN", "SYS_RESOURCE", "NET_ADMIN", "IPC_LOCK"},
		sp.SecurityContext.Capabilities.Add,
		"user capabilities are kept, missing packet capture capabilities added once")
}

func TestPrivateActionRunnerPacketCapture_AgentVersion(t *testing.T) {
	for _, test := range []struct {
		name          string
		tag           string
		doNotCheckTag bool
		enabled       bool
	}{
		{name: "7.85 is too old", tag: "7.85.0", enabled: false},
		{name: "latest resolves to the chart default", tag: "latest", enabled: false},
		{name: "7 resolves to the chart default", tag: "7", enabled: false},
		{name: "7.86 release candidate", tag: "7.86.0-rc.1", enabled: true},
		{name: "7.86", tag: "7.86.0", enabled: true},
		{name: "7.86 with build suffix", tag: "7.86.1-jmx", enabled: true},
		{name: "non-semver tag", tag: "main-abc1234", doNotCheckTag: true, enabled: true},
		{name: "old tag with doNotCheckTag", tag: "7.80.0", doNotCheckTag: true, enabled: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			overrides := pcapOverrides(test.tag)
			if test.doNotCheckTag {
				overrides["agents.image.doNotCheckTag"] = "true"
			}
			ds := renderPcapDaemonSet(t, overrides)
			_, spFound := getContainer(t, ds.Spec.Template.Spec.Containers, "system-probe")
			if !test.enabled {
				assertPcapAbsent(t, ds)
				assert.False(t, spFound, "system-probe must not be added for a skipped packet capture")
				assert.Nil(t, renderPcapSystemProbeConfig(t, overrides))
				return
			}
			assert.True(t, spFound)
			par, found := getContainer(t, ds.Spec.Template.Spec.Containers, "private-action-runner")
			require.True(t, found)
			assert.Equal(t, pcapSocketPath, getEnvVarMap(par.Env)[pcapSocketEnv])
		})
	}
}

// With discovery left at its default, packet capture is added to system-probe.yaml and the discovery
// settings stay exactly as they render without packet capture.
func TestPrivateActionRunnerPacketCapture_WithDefaultDiscovery(t *testing.T) {
	overrides := pcapOverrides("7.86.0")
	delete(overrides, "datadog.discovery.enabled")
	cfg := renderPcapSystemProbeConfig(t, overrides)
	require.NotNil(t, cfg)
	pc, ok := cfg["packet_capture"].(map[string]any)
	require.True(t, ok, "system-probe.yaml must have a packet_capture section")
	assert.Equal(t, true, pc["enabled"])
	assert.Equal(t, pcapSocketPath, pc["socket"])

	without := copyMap(overrides)
	delete(without, "datadog.privateActionRunner.packetCapture.enabled")
	baseline := renderPcapSystemProbeConfig(t, without)
	require.NotNil(t, baseline, "discovery alone renders system-probe by default")
	assert.NotContains(t, baseline, "packet_capture")
	assert.Equal(t, baseline["discovery"], cfg["discovery"])
	assert.Equal(t, baseline["system_probe_config"], cfg["system_probe_config"])

	ds := renderPcapDaemonSet(t, overrides)
	sp, found := getContainer(t, ds.Spec.Template.Spec.Containers, "system-probe")
	require.True(t, found)
	assert.Equal(t, []string{"system-probe", "--config=/etc/datadog-agent/system-probe.yaml"}, sp.Command)
}

func TestPrivateActionRunnerPacketCapture_RequiresPrivateActionRunner(t *testing.T) {
	overrides := pcapOverrides("7.86.0")
	overrides["datadog.privateActionRunner.enabled"] = "false"
	ds := renderPcapDaemonSet(t, overrides)

	assertPcapAbsent(t, ds)
	_, found := getContainer(t, ds.Spec.Template.Spec.Containers, "system-probe")
	assert.False(t, found)
	_, found = getContainer(t, ds.Spec.Template.Spec.Containers, "private-action-runner")
	assert.False(t, found)
}

func TestPrivateActionRunnerPacketCapture_KeepsOtherSystemProbeFeatures(t *testing.T) {
	overrides := pcapOverrides("7.86.0")
	overrides["datadog.networkMonitoring.enabled"] = "true"
	ds := renderPcapDaemonSet(t, overrides)

	sp, found := getContainer(t, ds.Spec.Template.Spec.Containers, "system-probe")
	require.True(t, found)
	_, found = pcapMount(sp)
	assert.True(t, found)
	agent, found := getContainer(t, ds.Spec.Template.Spec.Containers, "agent")
	require.True(t, found)
	_, found = pcapMount(agent)
	assert.False(t, found)
}

// The Private Action Runner itself is rejected on Windows, GKE Autopilot and GKE GDC, so packet capture
// cannot be enabled there either.
func TestPrivateActionRunnerPacketCapture_UnsupportedPlatforms(t *testing.T) {
	for _, test := range []struct {
		name     string
		override map[string]string
		errText  string
	}{
		{name: "windows", override: map[string]string{"targetSystem": "windows"}, errText: "Private Action Runner is not supported on Windows"},
		{name: "gke autopilot", override: map[string]string{"providers.gke.autopilot": "true"}, errText: "Private Action Runner is not supported on GKE Autopilot"},
		{name: "gke autopilot with workload allowlist", override: map[string]string{"providers.gke.autopilot": "true", "datadog.envDict.HELM_FORCE_RENDER": "true"}, errText: "Private Action Runner is not supported on GKE Autopilot"},
		{name: "gke gdc", override: map[string]string{"providers.gke.gdc": "true"}, errText: "Private Action Runner is not supported on GKE Autopilot / GDC"},
	} {
		t.Run(test.name, func(t *testing.T) {
			overrides := pcapOverrides("7.86.0")
			for k, v := range test.override {
				overrides[k] = v
			}
			_, err := common.RenderChart(t, common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides:   overrides,
			})
			require.Error(t, err)
			assert.Contains(t, err.Error(), test.errText)
		})
	}

	// With the runner off, enabling packet capture on Windows renders nothing for it.
	overrides := pcapOverrides("7.86.0")
	overrides["datadog.privateActionRunner.enabled"] = "false"
	overrides["targetSystem"] = "windows"
	ds := renderPcapDaemonSet(t, overrides)
	assertPcapAbsent(t, ds)
	_, found := getContainer(t, ds.Spec.Template.Spec.Containers, "system-probe")
	assert.False(t, found)
}
