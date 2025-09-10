#!/bin/bash

helm dependency update ./charts/datadog-operator &>/dev/null
helm dependency build ./charts/datadog-operator &>/dev/null

helm dependency update ./charts/datadog &>/dev/null
helm dependency build ./charts/datadog &>/dev/null

echo "Dependencies rebuilt successfully"