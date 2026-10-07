package synthetics_private_location

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"

	"github.com/DataDog/helm-charts/test/common"
)

const chartPath = "../../charts/synthetics-private-location"

func Test_baseline_manifests(t *testing.T) {
	tests := []struct {
		name         string
		overrides    map[string]string
		snapshotName string
	}{
		{
			name:         "Single-container default",
			overrides:    map[string]string{},
			snapshotName: "default",
		},
		{
			name: "Multi-container",
			overrides: map[string]string{
				"multiContainer.enabled": "true",
			},
			snapshotName: "multi-container",
		},
		{
			name: "Multi-container without traceroute-server",
			overrides: map[string]string{
				"multiContainer.enabled":                  "true",
				"multiContainer.tracerouteServer.enabled": "false",
			},
			snapshotName: "multi-container-no-traceroute",
		},
		{
			name: "Multi-container with NetworkPolicy",
			overrides: map[string]string{
				"multiContainer.enabled":                                    "true",
				"multiContainer.networkPolicy.enabled":                      "true",
				"multiContainer.networkPolicy.enableDefaultBlockedIpRanges": "true",
				"multiContainer.networkPolicy.allowedIPRanges.IPv4[0]":      "10.1.0.0/16",
				"multiContainer.networkPolicy.blockedIPRanges.IPv4[0]":      "203.0.114.0/24",
				"multiContainer.networkPolicy.blockedIPRanges.IPv6[0]":      "2001:db9::/32",
			},
			snapshotName: "multi-container-network-policy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			overrides := map[string]string{"configSecret": "synthetics-pl-config"}
			for k, v := range tt.overrides {
				overrides[k] = v
			}
			manifest, err := common.RenderChart(t, common.HelmCommand{
				ReleaseName: "synthetics-pl",
				ChartPath:   chartPath,
				Values:      []string{chartPath + "/values.yaml"},
				Overrides:   overrides,
			})
			assert.Nil(t, err, "couldn't render template")
			if common.UpdateBaselines {
				helm.UpdateSnapshot(t, &helm.Options{}, manifest, tt.snapshotName)
			}
			diffCount := helm.DiffAgainstSnapshot(t, &helm.Options{}, manifest, tt.snapshotName)
			assert.Equal(t, 0, diffCount, "manifests are different")
		})
	}
}
