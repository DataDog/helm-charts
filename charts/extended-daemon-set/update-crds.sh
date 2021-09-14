#!/bin/bash
set -euo pipefail

ROOT=$(git rev-parse --show-toplevel)

DATADOG_EXTENDED_DAEMON_SET_REPO=Datadog/extendeddaemonset

DATADOG_EXTENDED_DAEMON_SET_TAG=main
if [[ $# -eq 1 ]] ; then
   DATADOG_EXTENDED_DAEMON_SET_TAG=$1
fi

download_crd() {
    repo=$1
    tag=$2
    name=$3
    version=$4

    inFile=datadoghq.com_$name.yaml
    # shellcheck disable=SC2154
    outFile=datadoghq.com_"$name"_"$version".yaml
    path=$ROOT/charts/extended-daemon-set/templates/crds/$outFile
    echo "Download CRD \"$inFile\" version \"$version\" from repo \"$repo\" tag \"$tag\""
    curl --silent --show-error --fail --location --output "$path" "https://raw.githubusercontent.com/$repo/$tag/config/crd/bases/$version/$inFile"

    ifCondition="{{- if and .Values.installCRDs (not (.Capabilities.APIVersions.Has \"apiextensions.k8s.io/v1/CustomResourceDefinition\")) }}"
    if [ "$version" = "v1" ]; then
        ifCondition="{{- if and .Values.installCRDs (.Capabilities.APIVersions.Has \"apiextensions.k8s.io/v1/CustomResourceDefinition\") }}"
        cp "$path" "$ROOT/crds/datadoghq.com_$name.yaml"
    fi

    VALUE="'{{ include \"extendeddaemonset.chart\" . }}'" \
    yq eval '.metadata.labels."helm.sh/chart" = env(VALUE)'                              -i "$path"
    yq eval '.metadata.labels."app.kubernetes.io/managed-by" = "{{ .Release.Service }}"' -i "$path"
    VALUE="'{{ include \"extendeddaemonset.name\" . }}'" \
    yq eval '.metadata.labels."app.kubernetes.io/name" = env(VALUE)'                     -i "$path"
    yq eval '.metadata.labels."app.kubernetes.io/instance" = "{{ .Release.Name }}"'      -i "$path"

    { echo "$ifCondition"; cat "$path"; } > tmp.file
    mv tmp.file "$path"
    echo '{{- end }}' >> "$path"
}

eds_crds=(extendeddaemonsetreplicasets extendeddaemonsets extendeddaemonsetsettings)
for eds_crd in "${eds_crds[@]}"
do
  download_crd "$DATADOG_EXTENDED_DAEMON_SET_REPO" "$DATADOG_EXTENDED_DAEMON_SET_TAG" "$eds_crd" v1beta1
  download_crd "$DATADOG_EXTENDED_DAEMON_SET_REPO" "$DATADOG_EXTENDED_DAEMON_SET_TAG" "$eds_crd" v1
done
