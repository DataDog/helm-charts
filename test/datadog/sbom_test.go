package datadog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/DataDog/helm-charts/test/common"
)

// TestSBOMContainerImageUseMount checks that the chart asks Agents older than
// 7.87.0 to mount the images of the containerd snapshotter, and that the Agent
// container keeps the SYS_ADMIN capability for every Agent version.
func TestSBOMContainerImageUseMount(t *testing.T) {
	for _, tc := range []struct {
		name          string
		tag           string
		doNotCheckTag bool
		useMount      bool
	}{
		{name: "7.86", tag: "7.86.0", useMount: true},
		{name: "floating tag", tag: "7", useMount: true},
		{name: "7.87 release candidate", tag: "7.87.0-rc.1"},
		{name: "7.87", tag: "7.87.0"},
		{name: "custom image", tag: "custom", doNotCheckTag: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			overrides := map[string]string{
				"agents.image.tag":                    tc.tag,
				"datadog.apiKeyExistingSecret":        "datadog-secret",
				"datadog.sbom.containerImage.enabled": "true",
			}
			if tc.doNotCheckTag {
				overrides["agents.image.doNotCheckTag"] = "true"
			}
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
			agent, found := getContainer(t, ds.Spec.Template.Spec.Containers, "agent")
			require.True(t, found)

			_, useMount := getEnvVarMap(agent.Env)["DD_SBOM_CONTAINER_IMAGE_USE_MOUNT"]
			assert.Equal(t, tc.useMount, useMount)
			require.NotNil(t, agent.SecurityContext)
			require.NotNil(t, agent.SecurityContext.Capabilities)
			assert.Contains(t, agent.SecurityContext.Capabilities.Add, corev1.Capability("SYS_ADMIN"))
		})
	}
}
