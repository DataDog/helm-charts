package yamlmapper

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"helm.sh/helm/v3/pkg/chartutil"
)

func MakeTable(path string, val interface{}, destMap map[string]interface{}) map[string]interface{} {
	parts := parsePath(path)
	res := make(map[string]interface{})
	if len(parts) > 0 {
		// create innermost map using the input value
		res[parts[len(parts)-1]] = val
		// iterate backwards, skipping the last element (starting from i=1)
		for i := 1; i <= len(parts)-1; i++ {
			p := parts[len(parts)-(i+1)]
			// `t` is a placeholder map to carry over submaps between iterations
			t := make(map[string]interface{})
			t = res
			res = make(map[string]interface{})
			res[p] = t
		}
	}

	MergeMaps(destMap, res)

	return destMap
}

// MergeMaps recursively merges two maps, with values from map2 taking precedence over map1.
// It handles nil maps and type assertions safely.
// Inspired by: https://stackoverflow.com/a/60420264
func MergeMaps(map1, map2 map[string]interface{}) map[string]interface{} {
	if map1 == nil {
		map1 = make(map[string]interface{})
	}
	if map2 == nil {
		return map1
	}

	for key, rightVal := range map2 {
		if rightVal == nil {
			continue
		}

		leftVal, exists := map1[key]
		if !exists {
			// Key doesn't exist in map1, add it
			map1[key] = rightVal
			continue
		}

		// Both values are maps, merge them recursively
		leftMap, leftIsMap := leftVal.(map[string]interface{})
		rightMap, rightIsMap := rightVal.(map[string]interface{})

		if leftIsMap && rightIsMap {
			map1[key] = MergeMaps(leftMap, rightMap)
		} else {
			map1[key] = rightVal
		}
	}

	return map1
}

// setInterim sets a key in the interim map. If both the existing and new values are maps,
// it deep-merges them instead of overwriting. Otherwise it overwrites.
func setInterim(interim map[string]interface{}, key string, val interface{}) {
	if val == nil {
		return
	}
	if existing, exists := interim[key]; exists {
		if left, lok := toMap(existing); lok {
			if right, rok := toMap(val); rok {
				interim[key] = MergeMaps(left, right)
				return
			}
		}
	}
	interim[key] = val
}

// toMap tries to coerce supported map-like types into map[string]interface{}.
func toMap(v interface{}) (map[string]interface{}, bool) {
	switch t := v.(type) {
	case map[string]interface{}:
		return t, true
	case chartutil.Values:
		return map[string]interface{}(t), true
	default:
		return nil, false
	}
}

func parsePath(key string) []string { return strings.Split(key, ".") }

func getLatestValuesFile() string {
	chartVersion := getChartVersion()
	chartValuesFile := downloadYaml(fmt.Sprintf("https://raw.githubusercontent.com/DataDog/helm-charts/refs/tags/datadog-%s/charts/datadog/values.yaml", chartVersion), "datadog-values")

	return chartValuesFile
}

func getChartVersion() string {
	chartYamlPath := downloadYaml("https://raw.githubusercontent.com/DataDog/helm-charts/main/charts/datadog/Chart.yaml", "datadog-Chart")

	ddChart, err := chartutil.LoadChartfile(chartYamlPath)
	defer os.Remove(chartYamlPath)
	if err != nil {
		log.Printf("Error loading Chart.yaml: %s", err)
	}
	return ddChart.Version
}

func downloadYaml(url string, name string) string {
	resp, err := fetchUrl(context.TODO(), url)
	if err != nil {
		log.Printf("Error fetching yaml file: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to fetch yaml file %s: %v\n", url, resp.Status)
		return ""
	}

	tmpFile, err := os.CreateTemp("", fmt.Sprintf("%s.yaml.*", name))
	if err != nil {
		log.Printf("Error creating temporary file: %v\n", err)
		return ""
	}
	defer tmpFile.Close()

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		log.Printf("Error saving file: %v\n", err)
		return ""
	}

	// log.Printf("File downloaded and saved to temporary file: %s\n", tmpFile.Name())
	return tmpFile.Name()
}

func parseValues(sourceValues chartutil.Values, valuesMap map[string]interface{}, prefix string) map[string]interface{} {
	for key, value := range sourceValues {
		currentKey := prefix + key
		// If the value is a map, recursive call to get nested keys.
		if nestedMap, ok := value.(map[string]interface{}); ok {
			parseValues(nestedMap, valuesMap, currentKey+".")
		} else {
			valuesMap[currentKey] = ""
		}
	}
	return valuesMap
}

func fetchUrl(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}

func getDatadogMapping() (string, error) {
	//url := "https://raw.githubusercontent.com/DataDog/helm-charts/main/tools/yaml-mapper/mapping_datadog_helm_to_datadogagent_crd_v2.yaml"
	url := "https://raw.githubusercontent.com/DataDog/helm-charts/refs/heads/fanny/AGENTONB-2450/migration-mapper/tools/yaml-mapper/mapping_datadog_helm_to_datadogagent_crd_v2.yaml"

	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Printf("Error fetching Datadog mapping yaml file: %v\n", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Failed to fetch yaml file %s: %v\n", url, resp.Status)
	}

	tmpFile, err := os.CreateTemp("", defaultDDAMappingPath)
	if err != nil {
		log.Printf("Error creating temporary file: %v\n", err)
		return "", err
	}
	defer tmpFile.Close()

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		log.Printf("Error saving file: %v\n", err)
		return "", err
	}

	// log.Printf("File downloaded and saved to temporary file: %s\n", tmpFile.Name())
	return tmpFile.Name(), nil
}

// DepOp Operation to perform on deprecated keys
type DepOp int

const (
	DepBoolOr  DepOp = iota // boolean OR operation
	DepBoolNeg              // boolean ! operation
)

// DepRule describes how to fold deprecated keys into a standard key.
type DepRule struct {
	Deprecated []string
	Op         DepOp
}

// depRules describes how to fold deprecated keys into a standard key.
var depRules = map[string]DepRule{
	"datadog.apm.portEnabled": {
		[]string{"datadog.apm.enabled"},
		DepBoolOr,
	},
	"datadog.apm.socketEnabled": {
		[]string{"datadog.apm.useSocketVolume"},
		DepBoolOr,
	},
	"datadog.disableDefaultOsReleasePaths": {
		[]string{"datadog.systemProbe.enableDefaultOsReleasePaths"},
		DepBoolNeg,
	},
	"remoteConfiguration.enabled": {
		[]string{"datadog.remoteConfiguration.enabled"},
		DepBoolOr,
	},
	"datadog.useHostPID": {
		[]string{"datadog.dogstatsd.useHostPID"},
		DepBoolOr,
	},
	"datadog.securityAgent.compliance.host_benchmarks.enabled": {
		[]string{"datadog.securityAgent.compliance.xccdf"},
		DepBoolOr,
	},
	"datadog.networkPolicy.create": {
		[]string{
			"agents.networkPolicy.create",
			"clusterAgent.networkPolicy.create",
			"clusterChecksRunner.networkPolicy.create",
		},
		DepBoolOr,
	},
	"clusterAgent.pdb.create": {
		[]string{"clusterAgent.createPodDisruptionBudget"},
		DepBoolOr,
	},
	"clusterChecksRunner.pdb.create": {
		[]string{"clusterChecksRunner.createPodDisruptionBudget"},
		DepBoolOr,
	},
}

// FoldDeprecated maps “standard” key values by looking at their
// deprecated aliases according to depRules. It writes the effective
// value to sourceValues under the standard key.
func FoldDeprecated(sourceValues chartutil.Values) chartutil.Values {
	// chartutil.Values is a map[string]interface{}
	root := map[string]interface{}(sourceValues)

	for stdKey, depRule := range depRules {
		candidates := depRule.Deprecated
		// If the standard key is present in the source values, add it to the candidates
		if stdVal, err := sourceValues.PathValue(stdKey); stdVal != nil && err == nil {
			candidates = append(candidates, stdKey)
		}

		if len(candidates) == 0 {
			continue // nothing to do
		}

		val := false
		seen := false
		for _, c := range candidates {
			cVal, err := sourceValues.PathValue(c)
			if err != nil {
				continue
			}

			switch depRule.Op {
			case DepBoolOr:
				val = val || cVal.(bool)

			case DepBoolNeg:
				stdVal, err := sourceValues.PathValue(stdKey)
				if err != nil {
					val = !cVal.(bool)
				} else {
					val = stdVal.(bool)
				}
			default:
				continue
			}

			if c != stdKey {
				deletePath(root, c)
			}
			seen = true
		}
		if seen {
			root = MakeTable(stdKey, val, root)
		}
	}
	return root
}

func deletePath(root map[string]interface{}, dotted string) {
	parts := strings.Split(dotted, ".")
	if len(parts) == 0 {
		return
	}

	m := root
	for i := 0; i < len(parts)-1; i++ {
		next, ok := m[parts[i]].(map[string]interface{})
		if !ok {
			// Path doesn’t exist — nothing to delete
			return
		}
		m = next
	}

	delete(m, parts[len(parts)-1])
}
