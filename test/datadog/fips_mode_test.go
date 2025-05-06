package datadog

import (
	"fmt"
	"strings"
	"testing"

	"strconv"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func TestFIPSModeConditions(t *testing.T) {
	tests := []struct {
		name            string
		enableFIPSProxy bool
		enableFIPSAgent bool
		expectFIPSProxy bool
		expectFIPSAgent bool
		enableJMX       bool
	}{
		{
			name:            "neither fips proxy nor fips agent",
			enableFIPSProxy: false,
			enableFIPSAgent: false,
			expectFIPSProxy: false,
			expectFIPSAgent: false,
		},
		{
			name:            "fips proxy only",
			enableFIPSProxy: true,
			enableFIPSAgent: false,
			expectFIPSProxy: true,
			expectFIPSAgent: false,
		},
		{
			name:            "fips image only",
			enableFIPSProxy: false,
			enableFIPSAgent: true,
			expectFIPSProxy: false,
			expectFIPSAgent: true,
		},
		{
			name:            "fips proxy and fips image",
			enableFIPSProxy: true,
			enableFIPSAgent: true,
			expectFIPSProxy: false, // fips proxy should be disabled when fips agent is enabled
			expectFIPSAgent: true,
		},
		{
			name:            "fips image with JMX enabled",
			enableFIPSProxy: false,
			enableFIPSAgent: true,
			expectFIPSProxy: false,
			expectFIPSAgent: true,
			enableJMX:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := map[string]string{
				"useFIPSAgent":                 strconv.FormatBool(tt.enableFIPSAgent),
				"fips.enabled":                 strconv.FormatBool(tt.enableFIPSProxy),
				"datadog.apiKeyExistingSecret": "datadog-secret",
				"datadog.appKeyExistingSecret": "datadog-secret",
			}

			if tt.enableJMX {
				values["agents.image.tagSuffix"] = "jmx"
			}

			manifest, err := common.RenderChart(t, common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides:   values,
			})
			require.NoError(t, err, "couldn't render template")

			// Parse the manifest to find the should-enable-fips-proxy value and check image tags
			var daemonSet appsv1.DaemonSet
			common.Unmarshal(t, manifest, &daemonSet)

			// Checking that daemonSet contains or not fips-proxy container based on the fips proxy configuration
			checkFIPSProxy(t, daemonSet.Spec.Template.Spec.Containers, tt.expectFIPSProxy)

			// Checking that all containers have the fips image suffix if fips agent is enabled
			checkFIPSImage(t, daemonSet.Spec.Template.Spec.Containers, tt.expectFIPSAgent)
		})
	}
}

func checkFIPSProxy(t *testing.T, containers []corev1.Container, expectFIPSProxy bool) {
	hasFIPSProxy := false
	for _, container := range containers {
		if strings.Contains(container.Image, "fips-proxy") {
			hasFIPSProxy = true
			break
		}
	}
	if expectFIPSProxy {
		require.True(t, hasFIPSProxy, "fips proxy container should be present")
	} else {
		require.False(t, hasFIPSProxy, "fips proxy container should not be present")
	}
}

func checkFIPSImage(t *testing.T, containers []corev1.Container, expectFIPSImage bool) {
	if expectFIPSImage {
		for _, container := range containers {
			require.Contains(t, container.Image, "-fips", fmt.Sprintf("fips container %s should have the fips image suffix: %s", container.Name, container.Image))
		}
	} else {
		for _, container := range containers {
			require.NotContains(t, container.Image, "-fips", fmt.Sprintf("fips container %s should not have the fips image suffix: %s", container.Name, container.Image))
		}
	}
}
