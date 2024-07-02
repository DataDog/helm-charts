package datadog

import (
	"bufio"
	"io"
	"strings"
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func Test_baseline_manifests(t *testing.T) {
	tests := []struct {
		name                 string
		command              common.HelmCommand
		baselineManifestPath string
		assertions           func(t *testing.T, baselineManifestPath, manifest string)
	}{
		{
			name: "Daemonset default",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/daemonset.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret": "datadog-secret",
					"datadog.appKeyExistingSecret": "datadog-secret",
				},
			},
			baselineManifestPath: "./baseline/daemonset_default.yaml",
			assertions:           verifyDaemonset,
		},
		{
			name: "DCA Deployment default",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/cluster-agent-deployment.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides:   map[string]string{},
			},
			baselineManifestPath: "./baseline/cluster-agent-deployment_default.yaml",
			assertions:           verifyDeployment,
		},
		{
			name: "DCA Deployment default with minimal AC sidecar injection",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/cluster-agent-deployment.yaml"},
				Values: []string{"../../charts/datadog/values.yaml",
					"./manifests/dca_AC_sidecar_fargateMinimal.yaml"},
				Overrides: map[string]string{},
			},
			baselineManifestPath: "./baseline/cluster-agent-deployment_default_minimal_AC_injection.yaml",
			assertions:           verifyDeployment,
		},
		{
			name: "DCA Deployment default with advanced AC sidecar injection",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/cluster-agent-deployment.yaml"},
				Values: []string{"../../charts/datadog/values.yaml",
					"./manifests/dca_AC_sidecar_advanced.yaml"},
				Overrides: map[string]string{},
			},
			baselineManifestPath: "./baseline/cluster-agent-deployment_default_advanced_AC_injection.yaml",
			assertions:           verifyDeployment,
		},
		{
			name: "CLC Deployment default",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{"templates/agent-clusterchecks-deployment.yaml"},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret":                        "datadog-secret",
					"datadog.appKeyExistingSecret":                        "datadog-secret",
					"datadog.kubeStateMetricsCore.useClusterCheckRunners": "true",
					"datadog.clusterChecks.enabled":                       "true",
					"clusterChecksRunner.enabled":                         "true",
				}},
			baselineManifestPath: "./baseline/agent-clusterchecks-deployment_default.yaml",
			assertions:           verifyDeployment,
		},
		{
			name: "Other resources, skips Deployment, DaemonSet, Secret; creates PDBs",
			command: common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				ShowOnly:    []string{},
				Values:      []string{"../../charts/datadog/values.yaml"},
				Overrides: map[string]string{
					"datadog.apiKeyExistingSecret":                        "datadog-secret",
					"datadog.appKeyExistingSecret":                        "datadog-secret",
					"datadog.kubeStateMetricsCore.useClusterCheckRunners": "true",
					"datadog.clusterChecks.enabled":                       "true",
					"clusterChecksRunner.enabled":                         "true",
					// Create PDB for DCA and CLC
					"clusterAgent.createPodDisruptionBudget":        "true",
					"clusterChecksRunner.createPodDisruptionBudget": "true",
				}},
			baselineManifestPath: "./baseline/other_default.yaml",
			assertions:           verifyUntypedResources,
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

func verifyDaemonset(t *testing.T, baselineManifestPath, manifest string) {
	verifyBaseline(t, baselineManifestPath, manifest, appsv1.DaemonSet{}, appsv1.DaemonSet{})
}

func verifyDeployment(t *testing.T, baselineManifestPath, manifest string) {
	verifyBaseline(t, baselineManifestPath, manifest, appsv1.Deployment{}, appsv1.Deployment{})
}

func verifyBaseline[T any](t *testing.T, baselineManifestPath, manifest string, baseline, actual T) {
	common.Unmarshal(t, manifest, &actual)
	common.LoadFromFile(t, baselineManifestPath, &baseline)

	// Exclude
	// - "helm.sh/chart" label
	// - checksum annotations
	// - Image
	// to avoid frequent baseline update and CI failures.
	ops := make(cmp.Options, 0)
	ops = append(ops, cmpopts.IgnoreMapEntries(func(k, v string) bool {
		return k == "helm.sh/chart" || k == "checksum/clusteragent_token" || strings.Contains(k, "checksum")
	}))
	ops = append(ops, cmpopts.IgnoreFields(corev1.Container{}, "Image"))

	assert.True(t, cmp.Equal(baseline, actual, ops), cmp.Diff(baseline, actual))
}

func verifyUntypedResources(t *testing.T, baselineManifestPath, actual string) {
	baselineManifest := common.ReadFile(t, baselineManifestPath)

	rB := bufio.NewReader(strings.NewReader(baselineManifest))
	baselineReader := yaml.NewYAMLReader(rB)
	rA := bufio.NewReader(strings.NewReader(actual))
	expectedReader := yaml.NewYAMLReader(rA)

	for {
		baselineResource, errB := baselineReader.Read()
		actualResource, errA := expectedReader.Read()
		if errB == io.EOF || errA == io.EOF {
			break
		}
		require.NoError(t, errB, "couldn't read resource from manifest", baselineManifest)
		require.NoError(t, errA, "couldn't read resource from manifest", actual)

		// unmarshal as map since this can be any resource
		var expected, actual map[string]interface{}
		yaml.Unmarshal(baselineResource, &expected)
		yaml.Unmarshal(actualResource, &actual)

		assert.Equal(t, expected["kind"], actual["kind"])
		kind := expected["kind"]
		if kind == "Deployment" || kind == "DaemonSet" || kind == "Secret" {
			continue
		}

		ops := make(cmp.Options, 0)
		ops = append(ops, cmpopts.IgnoreMapEntries(func(k string, v any) bool {
			// skip these as these change frequently
			t.Log(k, v)
			return k == "helm.sh/chart" || k == "token" || strings.Contains(k, "checksum") ||
				k == "Image" || k == "install_id" || k == "install_time"
		}))

		assert.True(t, cmp.Equal(expected, actual, ops), cmp.Diff(expected, actual))
	}
}
