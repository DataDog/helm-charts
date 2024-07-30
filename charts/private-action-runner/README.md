# Private action runner Helm chart

This Helm Chart deploys the Datadog Private Action runner inside a Kubernetes cluster. It allows you to use private actions from the Datadog Workflow and Datadog App Builder products. When deploying this chart, you can give permissions to the runner in order to be able to run Kubernetes actions.

## Requirements
* A Datadog account with private actions enabled
* The `kubectl` cli
* Helm
* Sufficient permissions to the Kubernetes cluster

## Use this chart

1. Go to the [private action runner tab](https://app.datadoghq.com/workflow/private-action-runners).
2. Create a new private action runner.
3. Follow the instructions. You now have a running docker container and `config/config.yaml` file.
4. Stop the docker container (`docker stop <name-of-the-container>` or `docker compose stop`).
5. Create a `config.yaml` file with the appropriate values. An example `config.yaml` file is provided in the `examples` directory for you to copy.
    * Replace the `URN_FROM_CONFIG` and the `PRIVATE_KEY_FROM_CONFIG` in the example file with with the `urn` and the `privateKey` from the `config/config.yaml` of the docker container.
    * You can reconfigure other values or use the defaults in the example.
6. Add this repository to your Helm repositories:
    ```
    helm repo add datadog https://helm.datadoghq.com
    helm repo update
    ```
7. Install the Helm chart:
    ```bash
        helm install <RELEASE_NAME> datadog/private-action-runner -f ./config.yaml
    ```
8. Go to the [Workflow connections page](https://app.datadoghq.com/workflow/connections).
9. Create a new connection, select your private action runner, and use **Service account authentication**.
10. Create a new workflow and use a Kubernetes action like **List pod** or **List deployment**.

## Going further
* Adjust the service account permissions according to your needs. Learn more about [Kubernetes RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac).
* Deploy several runners with different permissions or create different connections according to your needs.
* Learn more about [Private actions](https://docs.datadoghq.com/service_management/app_builder/private_actions).

## Values

| Key                                  | Description                                                                  | Type   | Default                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
|--------------------------------------|------------------------------------------------------------------------------|--------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `common.image`                       | Current Datadog Private Action Runner image                                  | string | `"us-east4-docker.pkg.dev/datadog-sandbox/apps-on-prem/onprem-runner:v0.0.1-alpha22"`                                                                                                                                                                                                                                                                                                                                                                      |
| `runners[0]`                         | Name of the Datadog Private Action Runner                                    | object | `{"config":{"actionsAllowlist":["com.datadoghq.kubernetes.core.listPod"],"appBuilder":{"port":9016},"ddBaseURL":"https://app.datadoghq.com","modes":["workflowAutomation","appBuilder"],"privateKey":"PRIVATE_KEY_FROM_CONFIG","urn":"URN_FROM_CONFIG"},"kubernetesPermissions":[{"apiGroups":[""],"resources":["pods"],"verbs":["list","get"]},{"apiGroups":["apps"],"resources":["deployments"],"verbs":["list","get"]}],"name":"default","replicas":1}` |
| `runners[0].config`                  | Configuration for the Datadog Private Action Runner                          | object | `{"actionsAllowlist":["com.datadoghq.kubernetes.core.listPod"],"appBuilder":{"port":9016},"ddBaseURL":"https://app.datadoghq.com","modes":["workflowAutomation","appBuilder"],"privateKey":"PRIVATE_KEY_FROM_CONFIG","urn":"URN_FROM_CONFIG"}`                                                                                                                                                                                                             |
| `runners[0].config.actionsAllowlist` | List of actions that the Datadog Private Action Runner is allowed to execute | list   | `["com.datadoghq.kubernetes.core.listPod"]`                                                                                                                                                                                                                                                                                                                                                                                                                |
| `runners[0].config.appBuilder.port`  | Required port for App Builder Mode                                           | int    | `9016`                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| `runners[0].config.ddBaseURL`        | The base URL of the Datadog app                                              | string | `"https://app.datadoghq.com"`                                                                                                                                                                                                                                                                                                                                                                                                                              |
| `runners[0].config.modes`            | Modes that the runner can run in                                             | list   | `["workflowAutomation","appBuilder"]`                                                                                                                                                                                                                                                                                                                                                                                                                      |
| `runners[0].config.privateKey`       | The runner's privateKey from the enrollment page                             | string | `"PRIVATE_KEY_FROM_CONFIG"`                                                                                                                                                                                                                                                                                                                                                                                                                                |
| `runners[0].config.urn`              | The runner's URN from the enrollment page                                    | string | `"URN_FROM_CONFIG"`                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| `runners[0].kubernetesPermissions`   | List of Kubernetes permissions that the Datadog Private Action Runner has    | list   | `[{"apiGroups":[""],"resources":["pods"],"verbs":["list","get"]},{"apiGroups":["apps"],"resources":["deployments"],"verbs":["list","get"]}]`                                                                                                                                                                                                                                                                                                               |
| `runners[0].replicas`                | Number of instances of the Datadog Private Action Runner                     | int    | `1`                                                                                                                                                                                                                                                                                                                                                                                                                                                        |