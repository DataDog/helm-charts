# README.md

## Purpose

The purpose of this tool is to map a YAML file of a certain structure to another YAML file of a different structure. For instance, migrating a Helm chart `values.yaml` file to another `values.yaml` file after a significant chart update.

## Motivation

The motivation for creating this tool was to provide a way to support Datadog users who desire to switch from deploying the Datadog Agent using the `datadog` Helm chart to using the Datadog Operator controller. It is a potentially significant change that requires creating a new `DatadogAgent` custom resource specification. As a result, we are providing a way to map from a Helm chart `values.yaml` file to a `DatadogAgent` CRD spec, using a provided mapping.yaml file.

## How to install

```bash
go build -o yaml-mapper .
```

## How to use

### Mapping Helm YAML to DatadogAgent CRD Spec

If the desired conversion is between the `datadog` Helm chart and a `DatadogAgent` spec, use the provided `mapping_datadog_helm_to_datadogagent_crd.yaml` file as the mapping file. Otherwise, create your own using the following format:

```
source.key: destination.key
```

Both the key and value are period-delimited instead of nested or indented, as in a typical YAML file.

Pass the source file and mapping file to the command:

```bash
./yaml-mapper -sourceFile=source.yaml -mappingFile=mapping.yaml
```

The resulting file is written to `destination.yaml`. To specify a destination file, use flag `-destFile=[<FILENAME>.yaml]`.

Content from a file can be optionally prepended to the output. To specify the prefix file, use flag `-prefixFile=[<FILENAME>.yaml]`.

By default the output is also printed to STDOUT; to disable this use the flag `-printOutput=false`.

## Example usage (using provided files)

```bash
./yaml-mapper -sourceFile=<EXAMPLE_SOURCE>.yaml -mappingFile=mapping_datadog_helm_to_datadogagent_crd.yaml -prefixFile=<EXAMPLE_PREFIX>.yaml
```

The following command provides the example `destination.yaml` file in this directory. 
```bash
./yaml-mapper -sourceFile=example_source.yaml -mappingFile=mapping_datadog_helm_to_datadogagent_crd.yaml -prefixFile=example_prefix.yaml
```

### Updating Mapping File from a Source YAML

*When updating the mapper file, please be sure to add the [corresponding key!](#updating-mapping-keys)*

Below are different ways to update the mapping file based on your source:

1. **Local values.yaml from your branch**
If you have run into a CI error when adding a new field to values.yaml, run this command:

```bash
./yaml-mapper -updateMap -sourceFile=../../charts/datadog/values.yaml
```
2. **Latest published Datadog Helm chart values**
This pulls the latest values.yaml from the [latest published Helm chart](https://github.com/DataDog/helm-charts/releases/latest) and updates the default mapping file.

``` bash
./yaml-mapper -updateMap
```
3. **Update a custom mapping file with a custom source YAML**

```bash
./yaml-mapper -updateMap -sourceFile=<YOUR_SOURCE_FILE> -mappingFile=<YOUR_MAPPING_FILE>
```

### Updating Mapping Keys

Currently, this process is manual. To update a mapping key, search for it in the [operator configuration](https://github.com/DataDog/datadog-operator/blob/main/docs/configuration.v2alpha1.md).  When adding the corresponding operator value, be sure to prepend it with `spec.`.
 
If the key does not have a corresponding value in the Datadog Operator configuration, please leave the mapping as is with an empty string. 

Thank you for helping us keep the mapping accurate and up to date!