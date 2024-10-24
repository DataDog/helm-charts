# README.md

## Purpose

The purpose of this tool is to map a YAML file of a certain structure to another YAML file of a different structure. For instance, migrating a Helm chart values.yaml file to another values.yaml file after a significant chart update.

## Motivation

The motivation for creating this tool was to provide a way to support Datadog users who desire to switch from deploying the Datadog Agent using the `datadog` Helm chart to using the Datadog Operator controller. It is a potentially significant change that requires creating a new `DatadogAgent` custom resource specification. As a result, we are providing a way to map from a Helm chart values.yaml file to a `DatadogAgent` CRD spec, using a provided mapping.yaml file.

## How to install

```bash

$ go build -o yaml-mapper .

```

## How to use

If the desired conversion is between the `datadog` Helm chart and a `DatadogAgent` spec, use the provided `mapping_datadog_helm_to_datadogagent_crd.yaml` file as the mapping file. Otherwise, create your own using the following format:

```
source.key: destination.key
```

Both the key and value are period-delimited instead of nested or indented, as in a typical YAML file.

Pass the source file and mapping file to the command:

```bash
$ ./yaml-mapper -sourceFile=source.yaml -mappingFile=mapping.yaml

```

The resulting file is written to `destination.yaml`. To specify a destination file, use flag `-destFile=[<FILENAME>.yaml]`.

Content from a file can be optionally prepended to the output. To specify the prefix file, use flag `-prefixFile=[<FILENAME>.yaml]`.

By default the output is also printed to STDOUT; to disable this use the flag `-printOutput=false`.

## Example usage (using provided files)

```bash
$ ./yaml-mapper -sourceFile=example_source.yaml -mappingFile=mapping_datadog_helm_to_datadogagent_crd.yaml -prefixFile=example_prefix.yaml
```