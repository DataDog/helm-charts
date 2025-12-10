# Datadog Helm chart to Datadog Operator Migration Guide

## Overview

This guide breaks down the process for migrating from the Datadog Helm chart to the Datadog Operator for managing the Datadog Agent in Kubernetes. The migration process consists of 5 main steps:

1. Map Datadog Helm values to the DatadogAgent CRD format
2. Enable migration of Datadog Agent workload ownership to the Datadog Operator
3. Validate Datadog Agent workloads
4. Uninstall Datadog Helm chart
5. Install Operator Helm chart

Learn more about the [Datadog Operator][1] its benefits.

## Prerequisites

* Datadog Helm chart version X.X.X+
* Helm version 3.17.0+

## Map Datadog Helm values to DatadogAgent CRD

The [helm2dda][2] mapper CLI tool is now available in the Datadog Operator plugin for kubectl and can be used to map your Datadog Helm values to the DatadogAgent CRD format. You can use the tool directly by installing it using the Krew manager or in the Datadog Helm chart using the `datadog.operator.migration.preview` setting to preview the DDA manifest mapped from your Helm installation’s values.yaml.

The helm2dda CLI tool is intended to be a starting point for mapping your Datadog Helm release to the DatadogAgent CRD format. Some Helm configuration options might not be supported or may be added in a future chart release. Review the helm2dda logs and DatadogAgent manifest output to ensure accuracy, as some modifications to the manifest may be needed. Refer to the [DatadogAgent CRD documentation] for all available options.

### Map Datadog Helm values using Datadog kubectl plugin

1. [Install][2] the Datadog Operator plugin for kubectl using Krew.
2. Retrieve the full Helm values for your Datadog chart release and save contents to a file:
    ```
    helm get values <RELEASE_NAME> --namespace <NAMESPACE> --all > values.yaml
    ```
3. Run `helm2dda` subcommand:
    ```
    kubectl datadog helm2dda --sourcePath values.yaml --namespace <NAMESPACE>
    ```

### Map Datadog Helm values using the Datadog Helm chart

Alternatively, you can enable the `datadog.operator.migration.preview` option, which runs the Datadog `helm2dda` mapper as a Kubernetes Job and returns the projected DatadogAgent custom resource output. You can review this output and the mapper logs to determine if any manual changes are needed. 

```
helm upgrade <RELEASE_NAME> \
    --set datadog.apiKey=<DATADOG_API_KEY> \ 
    --set datadog.operator.enabled=true \
    --set datadog.operator.migration.preview=true \
    --namespace <NAMESPACE> \
    datadog/datadog 
```
Note: The Kubernetes Job is configured with `post-install` and `post-upgrade` [Helm chart hooks][4]. Run the `helm upgrade` command _without_ the `--no-hooks` flag.

__View migration job logs__

```
kubectl logs <MIGRATION_JOB_POD> --namespace <NAMESPACE>
```

## Migrate ownership of Datadog Agent resources

### Automatic migration via Helm (recommended)

Enable the `datadog.operator.migration.enabled` option to run a Kubernetes Job which executes the Datadog `helm2dda` mapper and applies the resulting DDA manifest to the active cluster. Your Datadog Agent pods will then update according to your Datadog chart's configured `updateStrategy`. Cluster Agent and Cluster Checks Runner pods will be terminated and re-created by the Datadog Operator.

```
helm upgrade <RELEASE_NAME> \
    --set datadog.apiKey=<DATADOG_API_KEY> \ 
    --set datadog.operator.enabled=true \
    --set datadog.operator.migration.enabled=true \
    --set datadog.operator.datadogCRDs.keepCrds=true \
    --namespace <NAMESPACE> \
    datadog/datadog 
```
Note: The Kubernetes Job is configured with `post-install` and `post-upgrade` [Helm chart hooks][4]. Run the `helm upgrade` command _without_ the `--no-hooks` flag.

__View migration job logs__

```
kubectl logs <MIGRATION_JOB_POD> --all-containers --namespace <NAMESPACE>
```

### Manual migration

It is recommended to use the [Automatic migration via Helm option](#automatic-migration-via-helm-recommended). If your mapped DatadogAgent manifest needs additional modification, confirm the following before proceeding:

1. Add required annotation to the DatadogAgent manifest:

    ```
    metadata:
      annotations:
        agent.datadoghq.com/helm-migration: true
    ```
2. DatadogAgent `metadata.name` exactly matches the `metadata.name` of the active agent DaemonSet. 

    ```
    kubectl get daemonsets --namespace <NAMESPACE>
    ```
3. You are running a supported Datadog Operator version 1.21.0+.   

When these items are confirmed, apply the DatadogAgent manifest to your cluster:

```
kubectl apply -f <datadog.yaml>
```

### Validation

After enabling migration, validate that the Datadog Operator has taken ownership of Datadog Agent resources and that data is still being collected.

1. Verify the Datadog Operator and DatadogAgent
    ```
    kubectl get deployment -n <NAMESPACE> datadog-operator
    kubectl get datadogagent -A
    kubectl describe datadogagent <DDA_NAME>
    ```

    **Expected:**
    * datadog-operator Deployment has all desired replicas available.
    * A DatadogAgent resource exists and has a status with no error or degraded conditions.

2. Verify Operator-managed workloads and rollout 

   Check that the Agent, Cluster Agent, and (if used) Cluster Checks Runner are healthy and Operator-managed.
    ```
    kubectl get daemonset -n <NAMESPACE> datadog-agent
    kubectl rollout status daemonset/datadog-agent -n <NAMESPACE>
    
    kubectl get deployment -n <NAMESPACE> datadog-cluster-agent
    kubectl rollout status deployment/datadog-cluster-agent -n <NAMESPACE>
    
    kubectl get deployment -n <NAMESPACE> datadog-cluster-checks-runner
    kubectl rollout status deployment/datadog-cluster-checks-runner -n <NAMESPACE>
    ```

    **Expected:**

    * `DESIRED`, `CURRENT`, and `READY` values match for the DaemonSet and Deployments.
    * `rollout status` reports a successful rollout. 
    * The Cluster Checks Runner Deployment is deleted and recreated; the Agent and Cluster Agent are updated with a `RollingUpdate` strategy.

    Confirm workloads are managed by the Operator.

    ```
    kubectl get daemonset datadog-agent -n <NAMESPACE> \
    -o jsonpath='{.metadata.labels.app\.kubernetes\.io/managed-by}{"\n"}'
    
    kubectl get deployment datadog-cluster-agent -n <NAMESPACE> \
    -o jsonpath='{.metadata.labels.app\.kubernetes\.io/managed-by}{"\n"}'
    
    kubectl get deployment datadog-cluster-checks-runner -n <NAMESPACE> \
    -o jsonpath='{.metadata.labels.app\.kubernetes\.io/managed-by}{"\n"}'
    ```

    **Expected:**

    * Output: `datadog-operator`

    Confirm there are no Helm-managed workloads remaining:

    ```
    kubectl get daemonset -n <NAMESPACE> | grep datadog
    kubectl get deployment -n <NAMESPACE> | grep datadog
    ```

    **Expected:**

    * Only the Operator-managed `datadog-agent`, `datadog-cluster-agent`, and (if applicable) `datadog-cluster-checks-runner` remain.

3. Check Operator and Agent logs

    Operator logs:
    ```
    kubectl logs deployment/datadog-operator -n <NAMESPACE>
    ```

    Agent and Cluster Agent logs (sample pods):

    ```
    kubectl logs <AGENT_POD_NAME> -n <NAMESPACE>
    kubectl logs <CLUSTER_AGENT_POD_NAME> -n <NAMESPACE>
    ```

    **Expected:**

    * No recurring errors about applying the DatadogAgent, RBAC, or missing CRDs. 
    * No recurring 401/403 errors or “failed to send” errors for metrics, logs, or traces.

4. Verify data in Datadog

    In the Datadog UI, verify that data continues to report after the migration time:

    * **Metrics:** Kubernetes / cluster / node dashboards show recent data; host/node count is consistent with pre-migration. 
    * **Logs (if enabled):** Kubernetes logs continue to appear with no significant gap at migration time. 
    * **APM / Traces (if enabled):** Services still emit traces after migration. 
    * **Other features:** Process, NPM, security signals, etc., continue to report data.

5. Agent configuration
   Compare key configuration between the original Helm values and the DatadogAgent spec:
   * Cluster name 
   * Site / intake URL 
   * Secret names for API/app keys 
   * Enabled features (logs, APM, process, NPM, security, admission controller)
   * Custom checks / Autodiscovery templates

If differences are found, update the DatadogAgent accordingly and re-check the validation steps above.

### Uninstall Datadog Chart
1. Confirm that `operator.datadogCRDs.keepCrds` is enabled. If not, enable it. This step instructs Helm to not delete the Datadog CRDs when uninstalling the Datadog chart and is __crucial__ for keeping DatadogAgent workloads running when the Operator chart is installed in the subsequent steps. 
2. Confirm that the `helm.sh/resource-policy: keep` annotation is present on the `datadogagents` CRD: 

    ```
    kubectl get crd datadogagents.datadoghq.com -o jsonpath='{.metadata.annotations}'
    ```
    ```
    > {"controller-gen.kubebuilder.io/version":"v0.17.3","helm.sh/resource-policy":"keep","meta.helm.sh/release-name":"datadog","meta.helm.sh/release-namespace":"default"}
    ```
3. Uninstall chart
```
helm uninstall <RELEASE_NAME> --namespace
```

### Install Datadog Operator Chart

```
helm repo update
helm install <RELEASE_NAME> datadog/datadog-operator --take-ownership
```

1. [Validate](#validation) that the DatadogAgent workloads remain running.
2. Validate that the new Datadog Operator leader pod is reconciling the DatadogAgent manifest:
    ```
    kubectl logs <DATADOG_OPERATOR_POD>
    ```
    ```
    {"level":"INFO","ts":"2025-12-10T21:21:27.074Z","logger":"controllers.DatadogAgent","msg":"Reconciling DatadogAgent","datadogagent":{"name":"datadog","namespace":"default"}}
    {"level":"INFO","ts":"2025-12-10T21:21:27.075Z","logger":"controllers.DatadogAgent","msg":"Helm-managed deployment has been migrated, checking for default deployment","component":"cluster-agent"}
    {"level":"INFO","ts":"2025-12-10T21:21:27.075Z","logger":"controllers.DatadogAgent","msg":"deployment is not found"}
    {"level":"INFO","ts":"2025-12-10T21:21:27.075Z","logger":"controllers.DatadogAgent","msg":"deployment is not found","datadogagent":{"name":"datadog","namespace":"default"},"component":"clusterAgent","provider":"","deployment.Namespace":"default","deployment.Name":"datadog-cluster-agent"}
    {"level":"INFO","ts":"2025-12-10T21:21:27.084Z","logger":"controllers.DatadogAgent","msg":"Creating Deployment","datadogagent":{"name":"datadog","namespace":"default"},"component":"clusterAgent","provider":"","deployment.Namespace":"default","deployment.Name":"datadog-cluster-agent"}
    {"level":"INFO","ts":"2025-12-10T21:21:27.084Z","logger":"controllers.DatadogAgent","msg":"Adding migration label to new operator daemonset as Helm migration has completed","datadogagent":{"name":"datadog","namespace":"default"}}
    {"level":"INFO","ts":"2025-12-10T21:21:27.085Z","logger":"controllers.DatadogAgent","msg":"daemonset is not found"}
    {"level":"INFO","ts":"2025-12-10T21:21:27.085Z","logger":"controllers.DatadogAgent","msg":"daemonset is not found","datadogagent":{"name":"datadog","namespace":"default"},"component":"nodeAgent","daemonset.Namespace":"default","daemonset.Name":"datadog-agent"}
    {"level":"INFO","ts":"2025-12-10T21:21:27.086Z","logger":"controllers.DatadogAgent","msg":"Creating Daemonset","datadogagent":{"name":"datadog","namespace":"default"},"component":"nodeAgent","daemonset.Namespace":"default","daemonset.Name":"datadog-agent"}
    {"level":"INFO","ts":"2025-12-10T21:21:27.091Z","logger":"controllers.DatadogAgent","msg":"Helm-managed deployment has been migrated, checking for default deployment","component":"cluster-checks-runner"}
    ```
   
### Troubleshooting

#### `helm2dda` mapper common logs

__*Warning: source value key X was not found in mapping.*__

1. The Helm key is actually a Helm value in YAML format and has been mapped successfully to the DatadogAgent. Confirm by looking for the value in the mapped DDA.
2. The Helm key hasn't been added to the mapping yet. If a corresponding config is available in the [DatadogAgent CRD][3], update your DDA manifest to use the config. Submit a support ticket to have the key added to the mapper.
3. The feature isn't supported in the DatadogAgent CRD either because it hasn’t been added yet, or is incompatible with the DatadogAgent CRD. Submit a [support ticket][5] to have the issue further investigated.

__*Warning: DDA destination key not found. Could not map: X*__

The Helm key is present in the mapping, but does not map to a corresponding DatadogAgent CRD config, either because it has not yet been added to the mapping schema or is not yet supported in the DatadogAgent CRD. Submit a [support ticket][5] to have the issue further investigated.

#### Migration job

Check the migration job logs:

```
kubectl logs <MIGRATION_JOB_POD> --all-containers    
```

*Migration job times out: __job datadog-migration-job failed: DeadlineExceeded__*
Check for mapper errors or Kubernetes RBAC errors in the job logs.

*Migration job fails: __StartError__*
Check the job pod resource: 

```
kubectl describe pod datadog-migration-job-XXXX -n <NAMESPACE> 
```

Submit a [support ticket][5] to have the issue further investigated.

[1]: https://docs.datadoghq.com/containers/datadog_operator/#why-use-the-datadog-operator-instead-of-a-helm-chart-or-daemonset
[2]: https://docs.datadoghq.com/containers/datadog_operator/kubectl_plugin/#install-the-plugin
[3]: https://docs.datadoghq.com/containers/datadog_operator/configuration/
[4]: https://helm.sh/docs/topics/charts_hooks/#helm
[5]: https://www.datadoghq.com/support/