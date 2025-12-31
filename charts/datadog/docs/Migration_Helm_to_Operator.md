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

To migrate Datadog Agent workloads deployed by an existing Datadog Helm release to the DatadogAgent custom resource definition, use the built-in migration tooling available in Datadog Helm chart version X.X.X and later:

1. Configure `datadog-values.yaml`
    
    Add the following to your `datadog-values.yaml` file:

    ```yaml
   datadog:
      operator:
         enabled: true
         migration:
         enabled: true
   
   operator:
      datadogCRDs:
         keepCrds: true
    ```

2. Upgrade your Helm release with the above configuration file

   Run:

   ```shell
   helm upgrade <RELEASE_NAME> -f datadog-values.yaml datadog/datadog
   ```

3. Confirm Agent installation

   Verify that Agent pods (tagged with `app.kubernetes.io/component:agent` and `app.kubernetes.io/managed-by: datadog-operator`) are updating according to the configured update strategy and reporting on the Containers page in Datadog. Agent pods are detected within a few minutes of deployment.

   Your Datadog Agent workloads are now managed by the DatadogAgent custom resource. To view and save the migrated DatadogAgent custom resource, run: 

   ```shell
   kubectl get datadogagents
   NAME      AGENT              CLUSTER-AGENT         CLUSTER-CHECKS-RUNNER   AGE
   datadog   Updating (5/0/0)   Progressing (1/0/1)                           5s

   kubectl get datadogagent datadog -oyaml > datadog.yaml
   ```

For more advanced configurations and use-cases, refer to [Datadog Helm chart to Datadog Operator Advanced Migration Guide][2].

## Uninstall Datadog Helm chart

After migrating your Datadog Agent workloads and validating that the Agent pods are reporting as expected, you can now uninstall the Datadog Helm chart in preparation for installing the Datadog Operator.

1. Run:

   ```shell
   helm uninstall <RELEASE_NAME>
   ```

Datadog Agent pods should remain unaffected and Datadog custom resource definitions (CRDs) should remain installed on the Kubernetes cluster. 

## Install Datadog Operator Helm chart

1. Run:

   ```shell
   helm install <RELEASE_NAME> \
      --set apiKeyExistingSecret=datadog-secret \
      --set appKeyExistingSecret=datadog-secret \
      --take-ownership \
      datadog/datadog-operator
   ```

2. Confirm Datadog Operator installation

   Verify that Datadog Operator pod is reporting on the Containers page in Datadog. 

To customize the Operator configuration, create a values.yaml file that can override the default Datadog Operator Helm chart [values][3]. 


## DatadogAgent custom resource configuration

Now, you can manage your Datadog Agent workloads using the DatadogAgent custom resource. To make updates to your Datadog Agent deployment, modify the configuration file containing your DatadogAgent spec and deploy it on your cluster:

```shell
kubectl apply -f datadog.yaml
```

For a full list of configuration options, see the DatadogAgent configuration [spec][4].

[1]: https://docs.datadoghq.com/containers/datadog_operator/#why-use-the-datadog-operator-instead-of-a-helm-chart-or-daemonset
[2]: https://docs.datadoghq.com/containers/datadog_operator/migration_advanced
[3]: https://github.com/DataDog/helm-charts/blob/main/charts/datadog-operator/values.yaml
[4]: https://docs.datadoghq.com/containers/datadog_operator/configuration/