migrate_rbac.yaml basically taken from `datadog-operator` chart.

```sh
# vanilla install chart; this adds new Operator-specific labels to pod templates
# clusterAgent.token so it's not recreated
❯ helm install datadog ./charts/datadog -f charts/datadog/migration-poc1-step0.yaml

# flip the flag for migration; this will assume default DDA name `datadog` and migrate all three resource.
# CLC and DCA could be optional, since they could be a bit more complex due to associated service
❯ helm upgrade datadog ./charts/datadog -f charts/datadog/migration-poc1-step0.yaml --set operator.migration.migrateWorkloadSelector=true
```

This change doesn't capture other resources which may rely on selectors.