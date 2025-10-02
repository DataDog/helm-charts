// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package yamlmapper

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/chartutil"
)

const (
	//defaultDDAMappingPath = "mapping_datadog_helm_to_datadogagent_crd_v2.yaml"
	defaultDDAMappingPath = "/Users/fanny.jiang/go/src/github.com/DataDog/helm-charts/tools/yaml-mapper/mapping_datadog_helm_to_datadogagent_crd_v2.yaml"
)

var defaultFilePrefix = map[string]interface{}{
	"apiVersion": "datadoghq.com/v2alpha1",
	"kind":       "DatadogAgent",
	"metadata":   map[string]interface{}{},
}

func MapYaml(mappingFile string, sourceFile string, destFile string, prefixFile string, ddaName string, namespace string, updateMap bool, printPtr bool) {

	log.Println("Mapper Config: ")
	log.Println("mappingFile:", mappingFile)
	log.Println("sourceFile:", sourceFile)
	log.Println("destFile:", destFile)
	log.Println("ddaName:", ddaName)
	log.Println("namespace:", namespace)
	log.Println("updateMap:", updateMap)
	log.Println("printOutput:", printPtr)
	log.Println("")

	// If updating mapping:
	// Use latest datadog chart values.yaml as sourceFile if none provided
	// Use default mappingFile if none provided
	tmpSourceFile := ""
	if updateMap {
		if sourceFile == "" {
			tmpSourceFile = getLatestValuesFile()
			sourceFile = tmpSourceFile
		}
	}

	if mappingFile == "" {
		mappingFile = defaultDDAMappingPath
	}

	_, err := os.Open(mappingFile)
	if err != nil {
		mappingFile, err = getDatadogMapping()
	}

	// Read mapping file
	mapping, err := os.ReadFile(mappingFile)
	if err != nil {
		log.Println(err)
	}
	mappingValues, err := chartutil.ReadValues(mapping)
	if err != nil {
		log.Println(err)
	}

	// Read source yaml file
	source, err := os.ReadFile(sourceFile)

	// Cleanup tmpSourceFile after it's been read
	if tmpSourceFile != "" {
		defer os.Remove(tmpSourceFile)
	}
	if err != nil {
		log.Println(err)
	}
	sourceValues, err := chartutil.ReadValues(source)
	if err != nil {
		log.Println(err)
	}

	// Create an interim map that that has period-delimited destination key as the key, and the value from the source.yaml for the value
	var pathVal interface{}
	var interim = map[string]interface{}{}

	if prefixFile == "" {
		interim = defaultFilePrefix
		metadata := interim["metadata"].(map[string]interface{})
		if ddaName == "" {
			ddaName = "datadog"
		}
		metadata["name"] = ddaName

		if namespace != "" {
			metadata := interim["metadata"].(map[string]interface{})
			metadata["namespace"] = namespace
		}
	}

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
			log.Println(e)
		}
		if mappingFile == defaultDDAMappingPath || tmpSourceFile != "" {
			newMapYaml = `# This file maps keys from the Datadog Helm chart (YAML) to the DatadogAgent CustomResource spec (YAML).
` + newMapYaml
		}

		if printPtr {
			log.Println("")
			log.Println(newMapYaml)
		}

		e = os.WriteFile(mappingFile, []byte(newMapYaml), 0660)
		if e != nil {
			log.Printf("Error updating mapping yaml. %v", e)
		}

		log.Printf("Mapping file, %s, successfully updated", mappingFile)
	}
	// Map values.yaml => DDA
	for sourceKey := range mappingValues {
		pathVal, _ = sourceValues.PathValue(sourceKey)
		mapVal := sourceValues[sourceKey]

		// Source val might be a value at the end of path or a map
		if pathVal == nil {
			if mapVal == nil {
				continue
			}
			if m, ok := mapVal.(map[string]interface{}); ok && m != nil {
				pathVal = mapVal
			}
		}

		destKey, ok := mappingValues[sourceKey]
		rt := reflect.TypeOf(destKey)
		if !ok || destKey == "" || destKey == nil {
			//log.Printf("Warning: key not found: %s\n", sourceKey)
			// Continue through loop
		} else if rt.Kind() == reflect.Slice {
			// Provide support for the case where one source key may map to multiple destination keys
			for _, v := range destKey.([]interface{}) {
				interim[v.(string)] = pathVal
			}
		} else if rt.Kind() == reflect.Map {
			// Perform type remapping
			newPath := destKey.(map[string]interface{})["newPath"].(string)
			newType, newTypeOk := destKey.(map[string]interface{})["newType"].(string)
			// if values type is different from new type, convert it
			pathValType := reflect.TypeOf(pathVal).Kind().String()
			var newPathVal []byte
			if newTypeOk && newType != "" && pathValType != newType {
				switch {
				case newType == "string":
					switch {
					case pathValType == "slice":
						newPathVal, err = yaml.Marshal(pathVal)
						if err != nil {
							log.Println(err)
						}
						interim[newPath] = string(newPathVal)
					}
				case newType == "integer":
					switch {
					case pathValType == "string":
						interim[newPath], err = strconv.Atoi(pathVal.(string))
						if err != nil {
							log.Println(err)
						}
					}
				}
			}

			//	Use custom map func
			if mapFunc, ok := destKey.(map[string]interface{})["mapFunc"].(string); ok {
				var mapFuncArgs []interface{}
				if args, ok := destKey.(map[string]interface{})["args"].([]interface{}); ok {
					mapFuncArgs = args
				}
				customMapFuncs[mapFunc](interim, newPath, pathVal, mapFuncArgs)
			}

		} else if destKey.(string) == "metadata.name" {
			name := pathVal
			if len(name.(string)) > 63 {
				name = name.(string)[:63]
			}
			metadata, ok := interim["metadata"].(map[string]interface{})
			if !ok {
				interim["metadata"] = map[string]interface{}{
					"name": name,
				}
			} else {
				metadata["name"] = name
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
		log.Println(err)
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

	if printPtr {
		log.Println("")
		log.Println(out)
	}

	// Create destination file if it doesn't exist
	_, err = os.Open(destFile)
	if err != nil {
		file, err := os.Create(fmt.Sprintf("dda.yaml.%s", time.Now().Format("20060102-150405")))
		if err != nil {
			log.Println(err)
		}
		destFile = file.Name()
	}

	err = os.WriteFile(destFile, []byte(out), 0660)
	if err != nil {
		log.Println(err)
	}

	log.Println("YAML file successfully written to", destFile)
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
		if leftVal, found := map1[key]; found && reflect.TypeOf(leftVal).Kind().String() == "map" {
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

var customMapFuncs = map[string]customMapFunc{
	"mapApiSecretKey":        mapApiSecretKey,
	"mapAppSecretKey":        mapAppSecretKey,
	"mapTokenSecretKey":      mapTokenSecretKey,
	"mapSeccompProfile":      mapSeccompProfile,
	"mapSystemProbeAppArmor": mapSystemProbeAppArmor,
	"mapLocalServiceName":    mapLocalServiceName,
	"mapAppendEnvVar":        mapAppendEnvVar,
	"mapMergeEnvs":           mapMergeEnvs,
}

type customMapFunc func(values map[string]interface{}, newPath string, pathVal interface{}, args []interface{})

func mapApiSecretKey(interim map[string]interface{}, newPath string, pathVal interface{}, args []interface{}) {
	//	if existing apikey secret, need to add key-name
	interim[newPath] = pathVal
	interim["spec.global.credentials.apiSecret.keyName"] = "api-key"
}

func mapAppSecretKey(interim map[string]interface{}, newPath string, pathVal interface{}, args []interface{}) {
	interim[newPath] = pathVal
	interim["spec.global.credentials.appSecret.keyName"] = "app-key"
}

func mapTokenSecretKey(interim map[string]interface{}, newPath string, pathVal interface{}, args []interface{}) {
	interim[newPath] = pathVal
	interim["spec.global.clusterAgentTokenSecret.keyName"] = "token"
}

func mapSeccompProfile(interim map[string]interface{}, newPath string, pathVal interface{}, args []interface{}) {
	seccompValue, err := pathVal.(string)
	if !err {
		return
	}

	if strings.HasPrefix(seccompValue, "localhost/") {
		profileName := strings.TrimPrefix(seccompValue, "localhost/")
		interim[newPath+".type"] = "Localhost"
		interim[newPath+".localhostProfile"] = profileName

	} else if seccompValue == "runtime/default" {
		interim[newPath+".type"] = "RuntimeDefault"

	} else if seccompValue == "unconfined" {
		interim[newPath+".type"] = "Unconfined"

	}
}

func mapSystemProbeAppArmor(interim map[string]interface{}, newPath string, pathVal interface{}, args []interface{}) {
	appArmorValue, err := pathVal.(string)
	if !err || appArmorValue == "" {
		// must be set to non-empty string
		return
	}

	systemProbeFeatures := []string{
		"spec.features.cws.enabled",            // datadog.securityAgent.runtime.enabled
		"spec.features.npm.enabled",            // datadog.networkMonitoring.enabled
		"spec.features.tcpQueueLength.enabled", // datadog.systemProbe.enableTCPQueueLength
		"spec.features.oomKill.enabled",        // datadog.systemProbe.enableOOMKill
		"spec.features.usm.enabled",            // datadog.serviceMonitoring.enabled
	}

	hasSystemProbeFeature := false
	for _, feature := range systemProbeFeatures {
		if val, exists := interim[feature]; exists {
			if enabled, ok := val.(bool); ok && enabled {
				hasSystemProbeFeature = true
				break
			}
		}
	}

	if !hasSystemProbeFeature {
		gpuEnabled, gpuExists := interim["spec.features.gpu.enabled"]
		gpuPrivileged, privExists := interim["spec.features.gpu.privilegedMode"]
		if gpuExists && privExists {
			if gpuEnabledBool, ok := gpuEnabled.(bool); ok && gpuEnabledBool {
				if gpuPrivilegedBool, ok := gpuPrivileged.(bool); ok && gpuPrivilegedBool {
					hasSystemProbeFeature = true
				}
			}
		}
	}

	if hasSystemProbeFeature {
		// must be set to non-empty string
		interim[newPath] = appArmorValue
	}
}

func mapLocalServiceName(interim map[string]interface{}, newPath string, pathVal interface{}, args []interface{}) {
	nameOverride, ok := pathVal.(string)
	if !ok || nameOverride == "" {
		return
	}
	interim[newPath] = nameOverride
}

// mapAppendEnvVar appends environment variables to a specified path in the interim configuration.
// It takes a list of environment variable definitions in the format []map[string]interface{}{{"name": "VAR_NAME"}}
// and creates new environment variable entries with the provided pathVal as the value.
// The new variables are added to the interim map at the specified newPath.
// Example:
//   - mapFuncArgs: []interface{}{map[string]interface{}{"name": "DD_LOG_LEVEL"}}
//   - pathVal: "debug"
//   - Result: Appends {"name": "DD_LOG_LEVEL", "value": "debug"} to newPath in interim
func mapAppendEnvVar(interim map[string]interface{}, newPath string, pathVal interface{}, args []interface{}) {
	if len(args) != 1 {
		return
	}

	envMap, ok := args[0].(map[string]interface{})
	if !ok {
		//log.Printf("expected map[string]interface{} for env var map definition, got %T", args[0])
		return
	}

	newEnvVar := map[string]interface{}{
		"name":  envMap["name"],
		"value": pathVal,
	}

	// Create the interim[newPath] if it doesn't exist yet
	if _, exists := interim[newPath]; !exists {
		interim[newPath] = []interface{}{newEnvVar}
		return
	}

	existing, ok := interim[newPath].([]interface{})
	if !ok {
		//log.Printf("Error: expected []interface{} at path %s, got %T", newPath, interim[newPath])
		return
	}

	interim[newPath] = append(existing, newEnvVar)
}

// mapMergeEnvs merges lists of environment variables at the specified path.
// It takes a slice of environment variable maps and merges them with any existing
// environment variables at the target path.
// Example:
//   - pathVal: []map[string]interface{}{{"name": "VAR1", "value": "val1"}}
//   - Result: Merges the new env vars with any existing ones at newPath
func mapMergeEnvs(interim map[string]interface{}, newPath string, pathVal interface{}, args []interface{}) {
	newEnvs, ok := pathVal.([]interface{})
	if !ok {
		//log.Printf("Warning: expected []interface{} for pathVal, got %T", pathVal)
		return
	}

	// If the interim[newPath] doesn't exist yet, just set the new environment variables
	existingEnvs, exists := interim[newPath]
	if !exists {
		interim[newPath] = newEnvs
		return
	}

	existingEnvsSlice, ok := existingEnvs.([]interface{})
	if !ok {
		//log.Printf("Warning: expected []interface{} at path %s, got %T", newPath, existingEnvs)
		return
	}

	// Merge the slices, avoiding duplicates
	mergedEnvs := make([]interface{}, len(existingEnvsSlice))
	copy(mergedEnvs, existingEnvsSlice)

	// Add new envs that don't already exist
	for _, newEnv := range newEnvs {
		newEnvMap, ok := newEnv.(map[string]interface{})
		if !ok {
			//log.Printf("Warning: expected map[string]interface{} in newEnvs, got %T", newEnv)
			continue
		}

		exists := false
		newName, hasName := newEnvMap["name"].(string)
		if !hasName {
			continue
		}

		for _, existingEnv := range mergedEnvs {
			if existingMap, ok := existingEnv.(map[string]interface{}); ok {
				if existingName, ok := existingMap["name"].(string); ok && existingName == newName {
					exists = true
					break
				}
			}
		}

		if !exists {
			mergedEnvs = append(mergedEnvs, newEnv)
		}
	}

	interim[newPath] = mergedEnvs
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
