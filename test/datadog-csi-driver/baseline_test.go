package datadog_csi_driver

import (
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/DataDog/helm-charts/test/utils"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
)

func Test_baseline_manifests(t *testing.T) {
	tests := []struct {
		name                 string
		command              common.HelmCommand
		baselineManifestPath string
		assertions           func(t *testing.T, baselineManifestPath, manifest string)
	}{
		{
			name: "CSI Driver DaemonSet default",
			command: common.HelmCommand{
				ReleaseName: "datadog-csi-driver",
				ChartPath:   "../../charts/datadog-csi-driver",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog-csi-driver/values.yaml"},
				Overrides:   map[string]string{},
			},
			baselineManifestPath: "./baseline/CSI_Driver_default.yaml",
			assertions:           verifyCSIDriverDaemonSet,
		},
		{
			name: "CSI Driver with annotations and security context set",
			command: common.HelmCommand{
				ReleaseName: "datadog-csi-driver",
				ChartPath:   "../../charts/datadog-csi-driver",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values: []string{
					"../../charts/datadog-csi-driver/values.yaml",
					"./manifests/added_annotation_and_securitycontext.yaml",
				},
				Overrides: map[string]string{},
			},
			baselineManifestPath: "./baseline/CSI_Driver_annotation_and_securitycontext.yaml",
			assertions:           verifyCSIDriverDaemonSet,
		},
		{
			name: "CSI Driver with nodeSelector and nodeAffinity set",
			command: common.HelmCommand{
				ReleaseName: "datadog-csi-driver",
				ChartPath:   "../../charts/datadog-csi-driver",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values: []string{
					"../../charts/datadog-csi-driver/values.yaml",
					"./manifests/added_nodeselector_and_nodeaffinity.yaml",
				},
				Overrides: map[string]string{},
			},
			baselineManifestPath: "./baseline/CSI_Driver_nodeselector_and_nodeaffinity.yaml",
			assertions:           verifyCSIDriverDaemonSet,
		},
		{
			// Legacy GKE Autopilot (< 1.32.1-gke.1729000) exposes the
			// AllowlistedV2Workload CRD but not the newer WorkloadAllowlist /
			// AllowlistSynchronizer CRDs. The Datadog CSI driver allowlist published
			// for this mode (v1.0.1) does not exempt the `storage-dir` hostPath nor
			// the `DD_APM_ENABLED` env var, so the chart must omit them. This
			// baseline pins that behavior to avoid regressing legacy Autopilot
			// installability.
			name: "CSI Driver on legacy GKE Autopilot (AllowlistedV2Workload)",
			command: common.HelmCommand{
				ReleaseName: "datadog-csi-driver",
				ChartPath:   "../../charts/datadog-csi-driver",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog-csi-driver/values.yaml"},
				Overrides:   map[string]string{},
				ExtraArgs: []string{
					"--api-versions=allowlistedv2workloads.auto.gke.io/v1/AllowlistedV2Workload",
					"--kube-version=1.31.0-gke.0",
				},
			},
			baselineManifestPath: "./baseline/CSI_Driver_legacy_gke_autopilot.yaml",
			assertions:           verifyCSIDriverDaemonSet,
		},
		{
			// New GKE Autopilot (>= 1.32.1-gke.1729000) exposes the
			// WorkloadAllowlist / AllowlistSynchronizer CRDs. The chart installs
			// the v1.1.0 allowlist (via gke_autopilot_allowlist_synchronizer.yaml)
			// which exempts the `storage-dir` hostPath and the `DD_APM_ENABLED`
			// env var, so SSI is enabled. The DaemonSet must carry the matching
			// `cloud.google.com/matching-allowlist` label pointing at v1.1.0.
			name: "CSI Driver on GKE Autopilot (WorkloadAllowlist)",
			command: common.HelmCommand{
				ReleaseName: "datadog-csi-driver",
				ChartPath:   "../../charts/datadog-csi-driver",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog-csi-driver/values.yaml"},
				Overrides:   map[string]string{},
				ExtraArgs: []string{
					"--api-versions=auto.gke.io/v1/AllowlistSynchronizer",
					"--api-versions=auto.gke.io/v1/WorkloadAllowlist",
					"--kube-version=1.32.1-gke.1729000",
				},
			},
			baselineManifestPath: "./baseline/CSI_Driver_gke_autopilot_workloadallowlist.yaml",
			assertions:           verifyCSIDriverDaemonSet,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest, err := common.RenderChart(t, tt.command)
			assert.Nil(t, err, "couldn't render template")
			t.Log("update baselines", common.UpdateBaselines)
			if common.UpdateBaselines {
				common.WriteToFile(t, tt.baselineManifestPath, manifest)
			}

			tt.assertions(t, tt.baselineManifestPath, manifest)
		})
	}
}

func verifyCSIDriverDaemonSet(t *testing.T, baselineManifestPath, manifest string) {
	utils.VerifyBaseline(t, baselineManifestPath, manifest, appsv1.DaemonSet{}, appsv1.DaemonSet{})
}
