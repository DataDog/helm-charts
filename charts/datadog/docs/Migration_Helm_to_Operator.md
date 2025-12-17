# Datadog Helm chart to Datadog Operator Migration Guide

## Overview

This guide breaks down the process for migrating from the Datadog Helm chart to the Datadog Operator for managing the Datadog Agent in Kubernetes. Using the Datadog Operator offers the following advantages:

* Operator configuration is more flexible for future enhancements.
* Validation for your Agent configurations.
* Orchestration for creating and updating Agent resources. 
* As a Kubernetes Operator, the Datadog Operator is treated as a first-class resource by the Kubernetes API. 
* Unlike the Helm chart, the Operator is included in the Kubernetes reconciliation loop.

Learn more about the [Datadog Operator][1] and its benefits.

## Prerequisites

* Datadog Helm chart version X.X.X+
* Helm version 3.17.0+

## Migrate Existing Datadog Helm Release

To migrate Datadog Agent workloads deployed by an existing Datadog Helm release, use the built-in migration tooling available in Datadog Helm chart version X.X.X and later:

1. **Configure `datadog-values.yaml`**
    
    Add the following to your `datadog-values.yaml` file:

    ```yaml
    datadog:
      operator:
        enabled: true
        migration:
          enabled: true
    ```
2. **Upgrade your Helm release with the above configuration file**

   Run:

   ```shell
   helm upgrade datadog-agent -f datadog-values.yaml datadog/datadog
   ```
3. **Confirm Agent installation**

   Verify that Agent pods (tagged with `app.kubernetes.io/component:agent` and `app.kubernetes.io/managed-by: datadog-operator`) are reporting on the Containers page in Datadog. Agent pods are detected within a few minutes of deployment.

For more advanced configurations and use-cases, refer to [Datadog Helm chart to Datadog Operator Advanced Migration Guide][2].

[1]: https://docs.datadoghq.com/containers/datadog_operator/#why-use-the-datadog-operator-instead-of-a-helm-chart-or-daemonset
[2]: https://docs.datadoghq.com/containers/datadog_operator/migration_advanced