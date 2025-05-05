package datadog

import (
	"fmt"
	"strings"
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"strconv"
)

func TestFIPSModeConditions(t *testing.T) {
	tests := []struct {
		name                   string
		setFipsEnabledSetting  bool
		setUseFipsImageSetting bool
		expectFipsProxy        bool
		expectFipsImageSuffix  bool
	}{
		{
			name:                   "fips.useFipsImages true should not use fips-proxy and use fips image",
			setFipsEnabledSetting:  true,
			setUseFipsImageSetting: true,
			expectFipsProxy:        false,
			expectFipsImageSuffix:  true,
		},
		{
			name:                   "fips.useFipsImages false and fips.enabled true should use fips-proxy and not use fips image",
			setFipsEnabledSetting:  true,
			setUseFipsImageSetting: false,
			expectFipsProxy:        true,
			expectFipsImageSuffix:  false,
		},
		{
			name:                   "fips.useFipsImages false and fips.enabled false should not use fips-proxy or fips image",
			setFipsEnabledSetting:  false,
			setUseFipsImageSetting: true,
			expectFipsProxy:        false,
			expectFipsImageSuffix:  false,
		},
		{
			name:                   "fips.useFipsImages false and fips.enabled false should not use fips-proxy or fips image",
			setFipsEnabledSetting:  false,
			setUseFipsImageSetting: true,
			expectFipsProxy:        false,
			expectFipsImageSuffix:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := map[string]string{
				"fips.useFipsImages": strconv.FormatBool(tt.setUseFipsImageSetting),
				"fips.enabled":       strconv.FormatBool(tt.setFipsEnabledSetting),
				"datadog.apiKeyExistingSecret": "datadog-secret",
	            "datadog.appKeyExistingSecret": "datadog-secret",
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
			var configMap corev1.ConfigMap
			var daemonSet appsv1.DaemonSet

			common.Unmarshal(t, manifest, &configMap)
			common.Unmarshal(t, manifest, &daemonSet)

			fmt.Printf("configMap: %+v\n", configMap)

			// Check FIPS proxy setting
			if value, ok := configMap.Data["should-enable-fips-proxy"]; ok {
				fmt.Printf("should-enable-fips-proxy: %s\n", value)
				assert.Equal(t, tt.expectFipsProxy, value == "true", "should-enable-fips-proxy value is incorrect")
			}

			// Check FIPS image suffix
			for _, container := range daemonSet.Spec.Template.Spec.Containers {
				fmt.Printf("container.Image: %s\n", container.Image)
				assert.Equal(t, tt.expectFipsImageSuffix, strings.Contains(container.Image, "-fips"), "fips image suffix is incorrect")
			}

		})
	}
}
