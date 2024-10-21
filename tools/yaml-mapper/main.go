// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package main

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"

	"helm.sh/helm/v3/pkg/chartutil"
)

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
			return
		}
	}

	printPtr := flag.Bool("printOutput", true, "print output to stdout")

	var mappingFile string
	var sourceFile string
	var destFile string
	var prefixFile string
	flag.StringVar(&mappingFile, "mappingFile", "mapping.yaml", "path to mapping YAML file")
	flag.StringVar(&sourceFile, "sourceFile", "source.yaml", "path to source YAML file")
	flag.StringVar(&destFile, "destFile", "destination.yaml", "path to destination YAML file")
	flag.StringVar(&prefixFile, "prefixFile", "", "path to prefix YAML file. The content in this file will be prepended to the output")

	flag.Parse()

	fmt.Println("mappingFile:", mappingFile)
	fmt.Println("sourceFile:", sourceFile)
	fmt.Println("destFile:", destFile)
	fmt.Println("prefixFile:", prefixFile)
	fmt.Println("printOutput:", *printPtr)
	fmt.Println("")

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
	if err != nil {
		fmt.Println(err)
		return
	}
	sourceValues, err := chartutil.ReadValues(source)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Create an interim map that that has period-delimited destination key as the key,
	// and the value from the source.yaml for the value
	var pathVal interface{}
	var destKey interface{}
	var ok bool
	interim := make(map[string]interface{})
	for sourceKey := range mappingValues {
		pathVal, _ = sourceValues.PathValue(sourceKey)
		// If there is no corresponding key in the destination, then the pathVal will be nil
		if pathVal == nil {
			continue
		}

		destKey, ok = mappingValues[sourceKey]
		rt := reflect.TypeOf(destKey)
		if !ok || destKey == "" || destKey == nil {
			fmt.Printf("Warning: key not found: %s\n", sourceKey)
			// Continue through loop
		} else if rt.Kind() == reflect.Slice {
			// Provide support for the case where one source key may map to multiple destination keys
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
