// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"

	"helm.sh/helm/v3/pkg/chartutil"
)

const defaultDDAMappingPath = "mapping_datadog_helm_to_datadogagent_crd.yaml"

func main() {

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "help":
			fmt.Println("Helper binary to convert a YAML file into another YAML file using a provided mapping.")
			fmt.Println("Flags (all optional):")
			fmt.Println("  -printOutput (bool)")
			fmt.Println("  -mappingFile (string)")
			fmt.Println("  -sourceFile (string)")
			fmt.Println("  -destFile (string)")
			fmt.Println("  -prefixFile (string)")
			fmt.Println("  -updateMap (bool)")
			return
		}
	}

	printPtr := flag.Bool("printOutput", true, "print output to stdout")

	var mappingFile string
	var sourceFile string
	var destFile string
	var prefixFile string
	var updateMap bool
	flag.StringVar(&mappingFile, "mappingFile", "", "Path to mapping YAML file. Example: mapping.yaml")
	flag.StringVar(&sourceFile, "sourceFile", "", "Path to source YAML file. Example: source.yaml")
	flag.StringVar(&destFile, "destFile", "destination.yaml", "Path to destination YAML file.")
	flag.StringVar(&prefixFile, "prefixFile", "example_prefix.yaml", "Path to prefix YAML file. The content in this file will be prepended to the output.")
	flag.BoolVar(&updateMap, "updateMap", false, fmt.Sprintf("Update 'mappingFile' with provided 'sourceFile'. (default false) If set to 'true', default mappingFile is %s and default sourceFile is latest published Datadog chart values.yaml.", defaultDDAMappingPath))

	flag.Parse()

	fmt.Println("mappingFile:", mappingFile)
	fmt.Println("sourceFile:", sourceFile)
	fmt.Println("destFile:", destFile)
	fmt.Println("prefixFile:", prefixFile)
	fmt.Println("updateMap:", updateMap)
	fmt.Println("printOutput:", *printPtr)
	fmt.Println("")

	// If updating mapping:
	// Use latest datadog chart values.yaml as sourceFile if none provided
	// Use default mappingFile if none provided
	tmpSourceFile := ""
	if updateMap {
		if sourceFile == "" {
			tmpSourceFile = getLatestValuesFile()
			sourceFile = tmpSourceFile
		}
		if mappingFile == "" {
			mappingFile = defaultDDAMappingPath
		}
	}

	// Read mapping file
	mapping, err := os.ReadFile(mappingFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	mappingValues, err := chartutil.ReadValues(mapping)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Read source yaml file
	source, err := os.ReadFile(sourceFile)

	// Cleanup tmpSourceFile after it's been read
	if tmpSourceFile != "" {
		defer os.Remove(tmpSourceFile)
	}
	if err != nil {
		fmt.Println(err)
		return
	}
	sourceValues, err := chartutil.ReadValues(source)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Create an interim map that that has period-delimited destination key as the key, and the value from the source.yaml for the value
	var pathVal interface{}
	var destKey interface{}
	var ok bool
	interim := make(map[string]interface{})

	if updateMap {
		// Populate interim map with keys from latest chart's values.yaml
		interim = parseValues(sourceValues, make(map[string]interface{}), "")
		// Add back existing key values from mapping file
		for sourceKey, sourceVal := range mappingValues {
			if sourceVal == nil {
				interim[sourceKey] = ""
			} else {
				interim[sourceKey] = sourceVal
			}
		}
		newMapYaml, e := chartutil.Values(interim).YAML()
		if e != nil {
			fmt.Println(e)
			return
		}
		if mappingFile == defaultDDAMappingPath || tmpSourceFile != "" {
			newMapYaml = `# This file maps keys from the Datadog Helm chart (YAML) to the DatadogAgent CustomResource spec (YAML).
` + newMapYaml
		}

		if *printPtr {
			fmt.Println("")
			fmt.Println(newMapYaml)
		}

		e = os.WriteFile(mappingFile, []byte(newMapYaml), 0660)
		if e != nil {
			fmt.Printf("Error updating mapping yaml. %v", e)
			return
		}

		fmt.Printf("Mapping file, %s, successfully updated", mappingFile)
		return
	}
	// Map values.yaml => DDA
	for sourceKey := range mappingValues {
		pathVal, _ = sourceValues.PathValue(sourceKey)
		// If there is no corresponding key in the destination, then the pathVal will be nil
		if pathVal == nil {
			continue
		}

		destKey, ok = mappingValues[sourceKey]
		rt := reflect.TypeOf(destKey)
		if !ok || destKey == "" || destKey == nil {
			fmt.Printf("WARNING: source key with value '%s' has no mapping defined or maps to an empty value. This key will be skipped in the migration. Review your configuration if this key should be included.\n", sourceKey)
			// Continue through loop
		} else if rt.Kind() == reflect.Slice {
			// Provide support for the case where one source key may map to multiple destination keys
			fmt.Printf("WARNING: the source key with value '%s' maps to %d destination keys. Please verify that migrating this value across multiple fields is intended in your configuration.\n", sourceKey, len(destKey.([]interface{})))
			for _, v := range destKey.([]interface{}) {
				interim[v.(string)] = pathVal
			}
		} else {
			interim[destKey.(string)] = pathVal
		}
	}

	// Create final mapping with properly nested map keys (converted from period-delimited keys)
	result := make(map[string]interface{})
	for k, v := range interim {
		result = makeTable(k, v, result)
	}

	// Pretty print to YAML format
	out, err := chartutil.Values(result).YAML()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Read prefix yaml file
	var prefix []byte
	if prefixFile != "" {
		prefix, err = os.ReadFile(prefixFile)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	if len(prefix) > 0 {
		out = string(prefix) + out
	}

	if *printPtr {
		fmt.Println("")
		fmt.Println(out)
	}

	err = os.WriteFile(destFile, []byte(out), 0660)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("YAML file successfully written to", destFile)

	return
}

func makeTable(path string, val interface{}, mapName map[string]interface{}) map[string]interface{} {
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

	mergeMaps(mapName, res)

	return mapName
}

// Inspired by: https://stackoverflow.com/a/60420264
func mergeMaps(map1, map2 map[string]interface{}) map[string]interface{} {
	for key, rightVal := range map2 {
		if leftVal, found := map1[key]; found {
			// Recurse on the found key
			map1[key] = mergeMaps(leftVal.(map[string]interface{}), rightVal.(map[string]interface{}))
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
		fmt.Println(fmt.Printf("Error loading Chart.yaml: %s", err))
	}
	return ddChart.Version
}

func downloadYaml(url string, name string) string {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error fetching yaml file: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Failed to fetch yaml file %s: %v\n", url, resp.Status)
		return ""
	}

	tmpFile, err := os.CreateTemp("", fmt.Sprintf("%s.yaml.*", name))
	if err != nil {
		fmt.Printf("Error creating temporary file: %v\n", err)
		return ""
	}
	defer tmpFile.Close()

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		fmt.Printf("Error saving file: %v\n", err)
		return ""
	}

	// fmt.Printf("File downloaded and saved to temporary file: %s\n", tmpFile.Name())
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
