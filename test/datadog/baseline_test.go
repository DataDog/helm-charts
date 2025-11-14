package datadog

import (
	"bufio"
	"io"
	"os"
	"regexp"
	"strconv"
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
			valuesFile := "./baseline/values/" + file.Name()
			manifest, err := common.RenderChart(t, common.HelmCommand{
				ReleaseName: "datadog",
				ChartPath:   "../../charts/datadog",
				Values:      []string{valuesFile},
			})
			assert.NoError(t, err, "couldn't render template")

			if err != nil && strings.Contains(err.Error(), "Use --debug flag to render out invalid YAML") {
				getInvalidYamlLocation(t, valuesFile)
			}

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

func getInvalidYamlLocation(t *testing.T, valuesFile string) {
	manifest, err := common.RenderChart(t, common.HelmCommand{
		ReleaseName: "datadog",
		ChartPath:   "../../charts/datadog",
		Values:      []string{valuesFile},
		ExtraArgs:   []string{"--debug"},
	})

	manifestFiles := make(map[string][]string)
	currentFile := ""
	sourceRegex := regexp.MustCompile(`^# Source: ([^ ]+)$`)
	for _, line := range strings.Split(manifest, "\n") {
		matches := sourceRegex.FindStringSubmatch(line)
		if len(matches) > 1 {
			currentFile = matches[1]
			continue
		}

		manifestFiles[currentFile] = append(manifestFiles[currentFile], line)
	}

	// sample parse error: Error: YAML parse error on datadog/templates/daemonset.yaml
	targetFileRegexp := regexp.MustCompile(`Error: YAML parse error on ([^:]+):`)
	lineRegexp := regexp.MustCompile(`line (\d+):`)
	targetFileMatches := targetFileRegexp.FindStringSubmatch(err.Error())
	lineMatches := lineRegexp.FindStringSubmatch(err.Error())

	var fileToShow []string
	var targetFile string
	var minLine, targetLine int
	if len(targetFileMatches) > 1 {
		targetFile = targetFileMatches[1]
		fileToShow = manifestFiles[targetFile]
	}

	linesAroundError := 5
	if os.Getenv("LINES_AROUND_ERROR") != "" {
		linesAroundError, err = strconv.Atoi(os.Getenv("LINES_AROUND_ERROR"))
		assert.NoError(t, err, "couldn't convert lines around error env var to int")
	}

	if len(lineMatches) > 1 {
		var err error
		targetLine, err = strconv.Atoi(lineMatches[1])
		assert.NoError(t, err, "couldn't convert line to int")

		// indexes from helm are 1-based, we work in 0-based
		targetLine = targetLine - 1

		minLine = max(0, targetLine-linesAroundError)
		maxLine := min(len(fileToShow), targetLine+linesAroundError)

		fileToShow = fileToShow[minLine:maxLine]
	}

	if len(fileToShow) == 0 {
		return
	}

	if targetLine > 0 {
		t.Logf("Invalid YAML reported on line %d of rendered file %s, showing rendered content around the error (lines around error: %d, change with env var LINES_AROUND_ERROR)", targetLine, targetFile, linesAroundError)
	} else {
		t.Logf("Invalid YAML reported on rendered file %s, showing rendered content", targetFile)
	}

	for lineIdx, line := range fileToShow {
		t.Logf("%d: %s", lineIdx+minLine+1, line)
	}
}
