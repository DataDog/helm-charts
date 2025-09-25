# PodTplgen

PodTplgen goal is to generate Datadog Agent Daemonset manifests based from the `datadog/datadog` helm chart.
Each generated manifest is tune thanks to helm chart option
 
## How it works

1. Generate all options combination.
2. Use the `helm template` command + the `-s` option to generate only the `daemonset.yaml` file.
3. Patch result to remove unecessary information (with yq).
4. Write the result of each manifest in a file (tmp folder).
5. Calculate a hash for each generated manifest to remove duplicated result.
6. Write in `output` folder the unique manifests. 


## recommanded Options

list of options defined in our documentation: https://www.datadoghq.com/blog/gke-autopilot-monitoring/

```console
helm install <RELEASE_NAME> \
    --set datadog.apiKey=<DATADOG_API_KEY> \
    --set datadog.appKey=<DATADOG_APP_KEY> \
    --set clusterAgent.metricsProvider.enabled=true \
    --set providers.gke.autopilot=true \
    --set datadog.logs.enabled=true \
    --set datadog.apm.enabled=true \
    --set datadog.kubeStateMetricsEnabled=false \
    --set datadog.kubeStateMetricsCore.enabled=true \
    datadog/datadog
```

