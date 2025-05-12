package datadog

import (
	"bufio"
	"io"
	"os"
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
	files, err := os.ReadDir("./baseline/values")
	assert.Nil(t, err, "couldn't read baseline values directory")
	for _, file := range files {
		t.Run(file.Name(), func(t *testing.T) {
			manifest, err := common.RenderChart(t, common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				Values:      []string{"./baseline/values/" + file.Name()},
			})
			assert.Nil(t, err, "couldn't render template")

			// // Print Kubernetes version from the rendered manifest
			// t.Logf("Kubernetes version from manifest capabilities: %s", common.GetKubeVersionFromManifest(manifest))

			manifest, err = common.FilterYamlKeysMultiManifest(manifest, FilterKeys)

			if err != nil {
				t.Fatalf("couldn't filter yaml keys: %v", err)
			}

			t.Log("update baselines", common.UpdateBaselines)
			if common.UpdateBaselines {
				common.WriteToFile(t, "./baseline/manifests/"+file.Name(), manifest)
			}

			verifyUntypedResources(t, "./baseline/manifests/"+file.Name(), manifest)
		})
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
