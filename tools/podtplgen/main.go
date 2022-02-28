package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/DataDog/helm-chart/tools/podtplgen/pkg/helm"
	"github.com/DataDog/helm-chart/tools/podtplgen/pkg/utils"
	"github.com/DataDog/helm-chart/tools/podtplgen/pkg/yq"
)

const (
	defaultTmpPath    = "./tmp"
	defaultOutputPath = "./output"
	defaultChartName  = "datadog/datadog"
)

func main() {
	var tmpPath, outputPath, chartName string
	flag.StringVar(&tmpPath, "tmp-path", defaultTmpPath, "tmp folder path")
	flag.StringVar(&outputPath, "output-path", defaultOutputPath, "output folder path")
	flag.StringVar(&chartName, "chart-name", defaultChartName, "chart name")
	flag.Parse()

	options := options{
		OutputPath:   outputPath,
		TmpPath:      tmpPath,
		ChartName:    chartName,
		ChartOptions: allowOptions,
		PatchPath:    removePaths,
	}

	if err := run(options); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}

type options struct {
	OutputPath   string
	TmpPath      string
	ChartName    string
	ChartOptions []string
	PatchPath    []string
}

func run(opts options) error {
	mandatoryOptions := []string{"providers.gke.autopilot=true"}

	// prepare tmp path
	os.MkdirAll(opts.TmpPath, 0755)

	combinations := utils.GetCombinations(opts.ChartOptions)
	// add default install
	combinations = append([][]string{{""}}, combinations...)

	csvfile, err := os.Create(fmt.Sprintf("%s/result.csv", opts.OutputPath))
	if err != nil {
		log.Fatal(err)
	}
	defer csvfile.Close()

	csvWriter := csv.NewWriter(csvfile)
	uniqueFiles := make(map[string]string)

	for id, combination := range combinations {
		combination = utils.CleanupUpSlice(combination)
		outputFile := fmt.Sprintf("%s/%d.yaml", opts.TmpPath, id)
		if err := renderChart(outputFile, opts.ChartName, mandatoryOptions, combination, opts.PatchPath); err != nil {
			fmt.Printf("error: %v\n", err)
		}

		md5result, err := utils.ComputeMd5(outputFile)
		if err != nil {
			return fmt.Errorf("unable to generate md5, err: %w", err)
		}
		if _, found := uniqueFiles[md5result]; !found {
			uniqueFiles[md5result] = outputFile
			err = utils.CopyFile(outputFile, fmt.Sprintf("%s/%s.yaml", opts.OutputPath, md5result))
			if err != nil {
				return fmt.Errorf("unable to copy file, err: %w", err)
			}
		}
		csvLine := []string{outputFile, md5result, strings.Join(combination, "|")}
		if err = csvWriter.Write(csvLine); err != nil {
			return fmt.Errorf("unable to write in csv file, err: %w", err)
		}
		fmt.Println(csvLine)
	}
	csvWriter.Flush()

	cleanup(opts.TmpPath)
	return nil
}

func cleanup(tmpPath string) error {
	return os.RemoveAll(tmpPath)
}

func renderChart(outputFile, chartName string, mandatoryOptions, argOptions []string, patchPath []string) error {
	renderOpts := helm.RenderOptions{
		MandatoryOptions: mandatoryOptions,
		Options:          argOptions,
		ChartName:        chartName,
	}
	renderResult, err := helm.Render(context.Background(), renderOpts)
	if err != nil {
		fmt.Printf("Error during render: %v\n", err)
		os.Exit(1)
	}

	fileOut, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	encoder := yaml.NewEncoder(fileOut)
	if err = encoder.Encode(renderResult); err != nil {
		return fmt.Errorf("unable to encode: %v", err)
	}
	fileOut.Close()

	patchOpts := yq.PatchOptions{
		Paths: patchPath,
	}
	yq.Patch(outputFile, renderResult, patchOpts)

	return nil
}
