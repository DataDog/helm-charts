#!/bin/bash

set -euox pipefail

ROOT=$(git rev-parse --show-toplevel)

DATADOG_OPERATOR_REPO=Datadog/datadog-operator

DATADOG_OPERATOR_TAG=main
if [[ $# -eq 1 ]]; then
    DATADOG_OPERATOR_TAG=$1
fi

download_crd() {
    repo=$1
    tag=$2
    name=$3
    installOption=$4 # Name of the option to install the CRD (defined in values.yaml)
    version=$5

    inFile=datadoghq.com_$name.yaml
    # shellcheck disable=SC2154
    outFile=datadoghq.com_"$name"_"$version".yaml
    path=$ROOT/charts/datadog-crds/templates/$outFile
    echo "Download CRD \"$inFile\" version \"$version\" from repo \"$repo\" tag \"$tag\""
    curl --silent --show-error --fail --location --output "$path" "https://raw.githubusercontent.com/$repo/$tag/config/crd/bases/$version/$inFile"

    if [ "$name" = "datadogagents" ]; then
        yq -i eval 'del(.. | select(has("defaultOverride")).defaultOverride.properties)' "$path"
        yq -i eval 'del(.. | select(has("description")).description)' "$path"
    fi

    ifCondition="{{- if and .Values.crds.$installOption (not (.Capabilities.APIVersions.Has \"apiextensions.k8s.io/v1/CustomResourceDefinition\")) }}"
    if [ "$version" = "v1" ]; then
        ifCondition="{{- if and .Values.crds.$installOption (.Capabilities.APIVersions.Has \"apiextensions.k8s.io/v1/CustomResourceDefinition\") }}"
        cp "$path" "$ROOT/crds/datadoghq.com_$name.yaml"
    fi

    VALUE="'{{ include \"datadog-crds.chart\" . }}'" \
    yq eval '.metadata.labels."helm.sh/chart" = env(VALUE)'                              -i "$path"
    yq eval '.metadata.labels."app.kubernetes.io/managed-by" = "{{ .Release.Service }}"' -i "$path"
    VALUE="'{{ include \"datadog-crds.name\" . }}'" \
    yq eval '.metadata.labels."app.kubernetes.io/name" = env(VALUE)'                     -i "$path"
    yq eval '.metadata.labels."app.kubernetes.io/instance" = "{{ .Release.Name }}"'      -i "$path"

    { echo "$ifCondition"; cat "$path"; } > tmp.file
    mv tmp.file "$path"
    echo '{{- end }}' >> "$path"
}

mkdir -p "$ROOT/crds"
download_crd "$DATADOG_OPERATOR_REPO" "$DATADOG_OPERATOR_TAG" datadogmetrics datadogMetrics v1beta1
download_crd "$DATADOG_OPERATOR_REPO" "$DATADOG_OPERATOR_TAG" datadogmetrics datadogMetrics v1
download_crd "$DATADOG_OPERATOR_REPO" "$DATADOG_OPERATOR_TAG" datadogagents datadogAgents v1beta1
download_crd "$DATADOG_OPERATOR_REPO" "$DATADOG_OPERATOR_TAG" datadogagents datadogAgents v1
download_crd "$DATADOG_OPERATOR_REPO" "$DATADOG_OPERATOR_TAG" datadogmonitors datadogMonitors v1beta1
download_crd "$DATADOG_OPERATOR_REPO" "$DATADOG_OPERATOR_TAG" datadogmonitors datadogMonitors v1
