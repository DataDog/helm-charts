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
			interim[destKey.(string)] = pathVal
		}
	}

	// Create final mapping with properly nested map keys (converted from period-delimited keys)
	result := make(map[string]interface{})
	for k, v := range interim {
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
