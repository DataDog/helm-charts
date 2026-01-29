# Migrating from Helm to Datadog Operator Demo

## Standard Migration

1. **Configure `datadog-values.yaml` to enable migration preview**.
    
    Add the following to your `datadog-values.yaml` file:

   ```yaml
   datadog:
      operator:
         enabled: true
         migration:
            preview: true
   ```

2. **Upgrade your Helm release and provide the file path to your updated `datadog-values.yaml` file using --set-file**.

   Run:

   ```shell
   helm upgrade datadog \
      --set-file datadog.operator.migration.userValues="charts/datadog/demo/migration-standard.yaml" \
      -f "charts/datadog/demo/migration-standard.yaml" \
      ./charts/datadog
   ```

3. **Review the migration job pod logs**

4. **Configure `datadog-values.yaml` to enable migration**.

   Add the following to your `datadog-values.yaml` file:

   ```yaml
   datadog:
      operator:
         enabled: true
         migration:
            enabled: true
   
   operator:
      image:
         tag: 1.22.0
      datadogCRDs:
         keepCrds: true
   ```

6. **Confirm Datadog Agent installation**.

## Install Datadog Operator Helm chart

1. Run:

   ```shell
   helm install operator \
      --set apiKeyExistingSecret=datadog-secret \
      --set appKeyExistingSecret=datadog-secret \
      --take-ownership \
      datadog/datadog-operator
   ```

2. Verify that the Datadog Operator pod is reporting

## Uninstall Datadog Helm chart

After you install the Datadog Operator Helm chart, uninstall the Datadog Helm chart.

1. Run:

   ```shell
   helm uninstall datadog
   ```

---

## Unsupported Values

1. **Configure `datadog-values.yaml` to enable migration preview**.
    
    Add the following to your `datadog-values.yaml` file:

   ```yaml
   datadog:
      operator:
         enabled: true
         migration:
            preview: true
   ```

2. **Upgrade your Helm release and provide the file path to your updated `datadog-values.yaml` file using --set-file**.

   Run:

   ```shell
   helm upgrade datadog \
      --set-file datadog.operator.migration.userValues="charts/datadog/demo/migration-unsupported.yaml" \
      -f "charts/datadog/demo/migration-unsupported.yaml" \
      ./charts/datadog
   ```