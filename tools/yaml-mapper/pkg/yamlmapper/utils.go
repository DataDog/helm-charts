package yamlmapper

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"

	"helm.sh/helm/v3/pkg/chartutil"
)

func MakeTable(path string, val interface{}, mapName map[string]interface{}) map[string]interface{} {
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

	MergeMaps(mapName, res)

	return mapName
}

// Inspired by: https://stackoverflow.com/a/60420264
func MergeMaps(map1, map2 map[string]interface{}) map[string]interface{} {
	for key, rightVal := range map2 {
		if leftVal, found := map1[key]; found && reflect.TypeOf(leftVal).Kind().String() == "map" {
			// Recurse on the found key
			map1[key] = MergeMaps(leftVal.(map[string]interface{}), rightVal.(map[string]interface{}))
		} else {
			// Key is not in map1, add it
			map1[key] = rightVal
		}
	}
	return map1
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
