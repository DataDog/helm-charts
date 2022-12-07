#!/bin/bash
set -euo pipefail

KUBEVAL_VERSION="v0.16.1"
SCHEMA_LOCATION="https://raw.githubusercontent.com/yannh/kubernetes-json-schema/master"
OS=$(uname)

CHANGED_CHARTS=${CHANGED_CHARTS:-${1:-}}
if [ -n "$CHANGED_CHARTS" ];
then
  CHART_DIRS=$CHANGED_CHARTS
else
  CHART_DIRS=$(ls -d charts/*)
fi

# install kubeval
curl --silent --show-error --fail --location --output /tmp/kubeval.tar.gz "https://github.com/instrumenta/kubeval/releases/download/${KUBEVAL_VERSION}/kubeval-${OS}-amd64.tar.gz"
tar -xf /tmp/kubeval.tar.gz kubeval

# validate charts
for CHART_DIR in ${CHART_DIRS}; do
  echo "Running kubeval for folder: '$CHART_DIR'"
  helm dep up "${CHART_DIR}" && helm template --kube-version "${KUBERNETES_VERSION#v}" --values "${CHART_DIR}"/ci/kubeval-values.yaml "${CHART_DIR}" | ./kubeval --strict --ignore-missing-schemas --kubernetes-version "${KUBERNETES_VERSION#v}" --schema-location "${SCHEMA_LOCATION}"
done
