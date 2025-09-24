// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/DataDog/helm-charts/tools/yaml-mapper/pkg/yamlmapper"
)

const defaultDDAMappingPath = "mapping_datadog_helm_to_datadogagent_crd.yaml"

func main() {

	var mappingFile string
	var sourceFile string
	var destFile string
	var ddaName string
	var namespace string
	var updateMap bool
	var printPtr bool

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `helm-mapper: migrate Datadog Helm values to the DatadogAgent CRD
Usage:
	helm-mapper -sourceFile=<FILE> -destFile=<DEST_FILE> -mappingFile=<MAPPING_FILE>

Options:
`)
		flag.PrintDefaults()
	}

	flag.StringVar(&mappingFile, "mappingFile", "", "Path to mapping YAML file. Example: mapping.yaml")
	flag.StringVar(&sourceFile, "sourceFile", "", "Path to source YAML file. Example: source.yaml")
	flag.StringVar(&destFile, "destFile", "destination.yaml", "Path to destination YAML file.")
	flag.StringVar(&ddaName, "ddaName", "", "Name to use for the destination DDA manifest.")
	flag.StringVar(&namespace, "namespace", "", "Namespace to use in destination DDA manifest.")
	flag.BoolVar(&updateMap, "updateMap", false, fmt.Sprintf("Update 'mappingFile' with provided 'sourceFile'. (default false) If set to 'true', default mappingFile is %s and default sourceFile is latest published Datadog chart values.yaml.", defaultDDAMappingPath))
	flag.BoolVar(&printPtr, "printOutput", true, "print output to stdout")

	flag.Parse()

	yamlmapper.MapYaml(mappingFile, sourceFile, destFile, ddaName, namespace, updateMap, printPtr)
}
