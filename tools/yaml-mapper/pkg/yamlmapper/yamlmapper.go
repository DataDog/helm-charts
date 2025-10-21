// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package yamlmapper

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"sort"
	"strconv"
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
		return
	}
	mappingValues, err := chartutil.ReadValues(mapping)
	if err != nil {
		log.Println(err)
		return
	}

	// Read source yaml file
	source, err := os.ReadFile(sourceFile)
	if err != nil {
		log.Println(err)
		return
	}

	// Cleanup tmpSourceFile after it's been read
	if tmpSourceFile != "" {
		defer os.Remove(tmpSourceFile)
	}

	sourceValues, err := chartutil.ReadValues(source)
	if err != nil {
		log.Println(err)
		return
	}

	// Create an interim map that that has period-delimited destination key as the key, and the value from the source.yaml for the value
	//var pathVal interface{}
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
	// Collect and sort mapping keys for deterministic processing order
	mappingKeys := make([]string, 0, len(mappingValues))
	for k := range mappingValues {
		mappingKeys = append(mappingKeys, k)
	}
	sort.Strings(mappingKeys)

	// Map values.yaml => DDA
	for _, sourceKey := range mappingKeys {
		pathVal, _ := sourceValues.PathValue(sourceKey)
		mapVal := sourceValues[sourceKey]
		tableVal, tableValErr := sourceValues.Table(sourceKey)

		// Source val might be a value at the end of path, a map, or a yaml subsection
		if pathVal == nil {
			if mapVal != nil {
				if m, ok := mapVal.(map[string]interface{}); ok && m != nil {
					pathVal = mapVal
					tableVal = nil
				}
			} else {
				if len(tableVal) == 1 && tableValErr == nil {
					pathVal = tableVal
					tableVal = nil
				}
				if tableVal != nil && tableValErr == nil && len(tableVal) == 1 {
					tableYaml, tableYamlErr := tableVal.YAML()
					if tableYamlErr != nil {
						continue
					}
					pathVal = tableYaml
				}
				if pathVal == nil {
					continue
				}
			}
		}

		destKey, ok := mappingValues[sourceKey]
		destKeyType := reflect.TypeOf(destKey)

		if !ok || destKey == "" || destKey == nil {
			log.Printf("Warning: DDA destination key not found: %s\n", sourceKey)
			continue
			// Continue through loop
		} else if destKeyType.Kind() == reflect.Slice {
			// Provide support for the case where one source key may map to multiple destination keys
			for _, v := range destKey.([]interface{}) {
				setInterim(interim, v.(string), pathVal)
			}
		} else if destKeyType.Kind() == reflect.Map {
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
						setInterim(interim, newPath, string(newPathVal))
					}
				case newType == "integer":
					switch {
					case pathValType == "string":
						convertedInt, convErr := strconv.Atoi(pathVal.(string))
						if convErr != nil {
							log.Println(convErr)
						} else {
							setInterim(interim, newPath, convertedInt)
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
				CustomMapFuncs[mapFunc](interim, newPath, pathVal, mapFuncArgs)
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
			if interim != nil {
				setInterim(interim, destKey.(string), pathVal)
			}
		}
	}

	// Sort interim keys to ensure deterministic nesting/merge order
	interimKeys := make([]string, 0, len(interim))
	for k := range interim {
		interimKeys = append(interimKeys, k)
	}
	sort.Strings(interimKeys)

	// Create final mapping with properly nested map keys (converted from period-delimited keys)
	result := make(map[string]interface{})
	for _, k := range interimKeys {
		v := interim[k]
		result = MakeTable(k, v, result)
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
