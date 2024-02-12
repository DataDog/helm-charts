#!/bin/bash
set -euo pipefail

KUBECONFORM_VERSION="v0.6.4"
# https://github.com/yannh/kubeconform/issues/51
CRD_SPEC_URL="https://raw.githubusercontent.com/kubernetes/kubernetes/master/api/openapi-spec/v3/apis__apiextensions.k8s.io__v1_openapi.json"
# Remove after v1.16 support / testing is dropped
LEGACY_SCHEMA_URL="https://github.com/instrumenta/kubernetes-json-schema"
OS=$(uname)

CHANGED_CHARTS=${CHANGED_CHARTS:-${1:-}}
if [ -n "$CHANGED_CHARTS" ];
then
  CHART_DIRS=$CHANGED_CHARTS
else
  CHART_DIRS=$(ls -d charts/*)
fi

# install kubeconform
curl --silent --show-error --fail --location --output /tmp/kubeconform.tar.gz "https://github.com/yannh/kubeconform/releases/download/${KUBECONFORM_VERSION}/kubeconform-${OS}-amd64.tar.gz"
tar -xf /tmp/kubeconform.tar.gz kubeconform

# validate charts
for CHART_DIR in ${CHART_DIRS}; do
  echo "Running kubeconform for folder: '$CHART_DIR'"

  # Note: -ignore-missing-schemas could be added if needed, but not currently
  # needed since we have the schema necessary to validate the CRDs themselves.
  #
  # Also, if at some point we needed to validate things _using_ these CRDs,
  # they're available via
  # https://github.com/datreeio/CRDs-catalog/tree/main/datadoghq.com
  helm dep up "${CHART_DIR}" && helm template --kube-version "${KUBERNETES_VERSION#v}" \
        --values "${CHART_DIR}/ci/kubeconform.yaml" "${CHART_DIR}" \
    | ./kubeconform -strict -schema-location default -schema-location "$CRD_SPEC_URL" \
        -schema-location $LEGACY_SCHEMA_URL -output pretty \
        -verbose -kubernetes-version "${KUBERNETES_VERSION#v}" -
done
