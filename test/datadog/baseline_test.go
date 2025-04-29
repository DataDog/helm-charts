package datadog

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DataDog/helm-charts/test/common"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v3"
	yaml2 "k8s.io/apimachinery/pkg/util/yaml"
)

var FilterKeys = map[string]interface{}{
	"helm.sh/chart":                   nil,
	"checksum/clusteragent_token":     nil,
	"checksum/clusteragent-configmap": nil,
	"checksum/install_info":           nil,
	"checksum":                        nil,
	"checksum/autoconf-config":        nil,
	"checksum/checksd-config":         nil,
	"checksum/confd-config":           nil,
	"checksum/otel-config":            nil,
	"checksum/api_key":                nil,
	"checksum/application_key":        nil,
	// ServiceAccount
	"chart": nil,
	// ConfigMap
	"install_id":   nil,
	"install_time": nil,
	// Secret
	"token": nil,
	// install info CM, it contains chart version
	// TODO: we are dropping everything; instead could we have a mapper/function for these keys or separate for coverage.
	"install_info": nil,
}

func Test_baseline_inputs(t *testing.T) {
	tests := []struct {
		name             string
		showOnly         []string
		valuesSubPath    string
		manifestsSubPath string
	}{
		{
			name:             "Agent DaemonSets",
			showOnly:         []string{"templates/daemonset.yaml"},
			valuesSubPath:    "./baseline/values/agent_daemonset",
			manifestsSubPath: "./baseline/manifests/agent_daemonset",
		},
		{
			name:             "Cluster Agent Deployments",
			showOnly:         []string{"templates/cluster-agent-deployment.yaml"},
			valuesSubPath:    "./baseline/values/cluster-agent_deployment",
			manifestsSubPath: "./baseline/manifests/cluster-agent_deployment",
		},
	}

	for _, tt := range tests {
		files, err := os.ReadDir(tt.valuesSubPath)
		assert.Nil(t, err, "couldn't read baseline values directory")
		for _, file := range files {
			t.Run(file.Name(), func(t *testing.T) {
				valuesPath := filepath.Join(tt.valuesSubPath, file.Name())
				manifest, err := common.RenderChart(t, common.HelmCommand{
					ReleaseName: "datadog",
					ChartPath:   "../../charts/datadog",
					ShowOnly:    tt.showOnly,
					Values:      []string{valuesPath},
				})
				assert.Nil(t, err, "couldn't render template")

				manifest, err = common.FilterYamlKeysMultiManifest(manifest, FilterKeys)

				if err != nil {
					t.Fatalf("couldn't filter yaml keys: %v", err)
				}

				containerManifests, err := common.ExtractContainersManifests(t, manifest, valuesPath)
				if err != nil {
					t.Fatalf("couldn't get container manifests: %v", err)
				}

				t.Log("update baselines", common.UpdateBaselines)
				for containerName, containerManifest := range containerManifests {
					containerManifestPath := filepath.Join(tt.manifestsSubPath, containerName, fmt.Sprintf("%s_%s", containerName, file.Name()))
					if common.UpdateBaselines {
						common.WriteToFile(t, containerManifestPath, containerManifest)
					}
					verifyUntypedResources(t, containerManifestPath, containerManifest)
					verifyVolumeMounts(t, containerManifest, manifest)
				}
			})
		}
	}
}

func verifyUntypedResources(t *testing.T, baselineManifestPath, actual string) {
	baselineManifest := common.ReadFile(t, baselineManifestPath)

	rB := bufio.NewReader(strings.NewReader(baselineManifest))
	baselineReader := yaml2.NewYAMLReader(rB)
	rA := bufio.NewReader(strings.NewReader(actual))
	expectedReader := yaml2.NewYAMLReader(rA)

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

		assert.True(t, cmp.Equal(expected, actual), cmp.Diff(expected, actual))
	}
}

// verifyVolumeMounts checks that all volume mounts in the container manifest have a corresponding volume in the DaemonSet or deployment manifest
func verifyVolumeMounts(t *testing.T, containerManifest, fullManifest string) {
	// Decode container and full (ds or deployment) manifests
	containerManifestDecoder := yaml.NewDecoder(strings.NewReader(containerManifest))
	fullManifestDecoder := yaml.NewDecoder(strings.NewReader(fullManifest))

	var containerResource map[string]interface{}
	if err := containerManifestDecoder.Decode(&containerResource); err != nil {
		t.Fatalf("couldn't decode container manifest: %s", err)
	}

	var manifestResource map[string]interface{}
	if err := fullManifestDecoder.Decode(&manifestResource); err != nil {
		t.Fatalf("couldn't decode full manifest: %s", err)
	}

	// Get volumes from the full ds/deployment manifest
	assert.NotNil(t, manifestResource)
	spec, _ := manifestResource["spec"].(map[string]interface{})
	template, _ := spec["template"].(map[string]interface{})
	templateSpec := template["spec"].(map[string]interface{})
	manifestVolumes := templateSpec["volumes"].([]interface{})
	assert.Greater(t, len(manifestVolumes), 0)

	// Get volumeMounts from the container manifest
	containerVolumeMounts := containerResource["volumeMounts"].([]interface{})
	assert.Greater(t, len(containerVolumeMounts), 0)

	// Verify that each volumeMount in the container has a corresponding volume in the full manifest
	for _, volume := range containerVolumeMounts {
		volumeName := volume.(map[string]interface{})["name"].(string)
		assert.Truef(t, common.VolumeExists(manifestVolumes, volumeName), "volumeMount %s not found in manifest volumes", volumeName)
	}
}
