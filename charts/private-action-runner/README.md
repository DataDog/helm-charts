# Private action runner Helm chart

This Helm Chart will deploy the Datadog Private Action runner inside a Kubernetes cluster.
You will be able to use private actions from the Datadog Workflow and Datadog App Builder products.
When deploying this chart, you will be able to give permissions to the runner in order to be able to run Kubernetes actions.

## Requirements
* A Datadog account with private actions enabled
* The kubectl cli
* Helm
* Sufficient permissions to the kubernetes cluster

## In order to use this chart

* Go to the private action runner tab https://app.datadoghq.com/workflow/private-action-runners
* Create a new private action runner
* Follow the instructions and you should have a running docker container and `config/config.yaml` file
* Stop the docker container (`docker stop <name-of-the-container>` or `docker compose stop`)
* Replace the `URN_FROM_CONFIG` and the `PRIVATE_KEY_FROM_CONFIG` from the chart's `values.yaml` with the `urn` and the `privateKey` from the `config/config.yaml` of the docker container.
* Create the kubernetes namespace `kubectl create namespace private-action-runner`
* Install the Helm chart `helm install <name> .`
* Go to the workflow connections https://app.datadoghq.com/workflow/connections
* Create a new connection, select your private action runner and use `Service account authentication`
* Create a new workflow and use a kubernetes action like `List pod` / `List deployment`

## Going further
* Adjust the service account permissions according to your needs, you can find more informations about kubernetes RBAC here https://kubernetes.io/docs/reference/access-authn-authz/rbac
* Deploy several runners with different permissions / create different connections according to your needs
* Private actions documentation https://docs.datadoghq.com/service_management/app_builder/private_actions

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| common.image | string | `"us-east4-docker.pkg.dev/datadog-sandbox/apps-on-prem/onprem-runner:v0.0.1-alpha22"` | Current Datadog Private Action Runner image |
| common.namespace | string | `"private-action-runner"` | The namespace where the Datadog Private Action Runner will be deployed |
| runners[0] | object | `{"config":{"actionsAllowlist":["com.datadoghq.kubernetes.core.listPod"],"appBuilder":{"port":9016},"ddBaseURL":"https://app.datadoghq.com","modes":["workflowAutomation","appBuilder"],"privateKey":"PRIVATE_KEY_FROM_CONFIG","urn":"URN_FROM_CONFIG"},"kubernetesPermissions":[{"apiGroups":[""],"resources":["pods"],"verbs":["list","get"]},{"apiGroups":["apps"],"resources":["deployments"],"verbs":["list","get"]}],"name":"default","replicas":1}` | Name of the Datadog Private Action Runner |
| runners[0].config | object | `{"actionsAllowlist":["com.datadoghq.kubernetes.core.listPod"],"appBuilder":{"port":9016},"ddBaseURL":"https://app.datadoghq.com","modes":["workflowAutomation","appBuilder"],"privateKey":"PRIVATE_KEY_FROM_CONFIG","urn":"URN_FROM_CONFIG"}` | This is the default configuration for the Datadog Private Action Runner |
| runners[0].config.actionsAllowlist | list | `["com.datadoghq.kubernetes.core.listPod"]` | List of actions that the Datadog Private Action Runner is allowed to execute |
| runners[0].config.appBuilder.port | int | `9016` | Required port for App Builder Mode |
| runners[0].config.ddBaseURL | string | `"https://app.datadoghq.com"` | The base URL of the Datadog |
| runners[0].config.modes | list | `["workflowAutomation","appBuilder"]` | Modes that the runner can run in |
| runners[0].config.privateKey | string | `"PRIVATE_KEY_FROM_CONFIG"` | User to specify the runner's privateKey from the enrollment page |
| runners[0].config.urn | string | `"URN_FROM_CONFIG"` | User to specify the runner's URN from the enrollment page |
| runners[0].kubernetesPermissions | list | `[{"apiGroups":[""],"resources":["pods"],"verbs":["list","get"]},{"apiGroups":["apps"],"resources":["deployments"],"verbs":["list","get"]}]` | List of Kubernetes permissions that the Datadog Private Action Runner will have |
| runners[0].replicas | int | `1` | Number of instances of Datadog Private Action Runner |