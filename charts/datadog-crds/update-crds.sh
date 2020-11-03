#!/bin/bash
set -euo pipefail

ROOT=$(git rev-parse --show-toplevel)

TAG=master
if [[ $# -eq 1 ]] ; then
    TAG=$1
fi

download_crd() {
    file=datadoghq.com_$2.yaml
    path=$ROOT/charts/datadog-crds/templates/$file
    echo "Download CRD \"$file\" from tag \"$1\""
    curl --silent --show-error --fail --location --output $path https://raw.githubusercontent.com/DataDog/datadog-operator/$1/bundle/manifests/$file

    cp $path $ROOT/crds/$file
    yq w -i $path 'metadata.labels."helm.sh/chart"' '{{ include "datadog-crds.chart" . }}'
    yq w -i $path 'metadata.labels."app.kubernetes.io/managed-by"' '{{ .Release.Service }}'
    yq w -i $path 'metadata.labels."app.kubernetes.io/name"' '{{ include "datadog-crds.name" . }}'
    yq w -i $path 'metadata.labels."app.kubernetes.io/instance"' '{{ .Release.Name }}'
    { echo "{{- if .Values.crds.$3 }}"; cat $path; } > tmp.file
    mv tmp.file $path
    echo '{{- end }}' >> $path
}

mkdir -p $ROOT/crds
download_crd $TAG datadogmetrics datadogMetrics
download_crd $TAG datadogagents datadogAgents