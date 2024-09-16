# Datadog Private Action Runner

![Version: 0.11.0](https://img.shields.io/badge/Version-0.11.0-informational?style=flat-square) ![AppVersion: v0.1.0-beta](https://img.shields.io/badge/AppVersion-v0.1.0--beta-informational?style=flat-square)

This Helm Chart deploys the Datadog Private Action runner inside a Kubernetes cluster. It allows you to use private actions from the Datadog Workflow and Datadog App Builder products. When deploying this chart, you can give permissions to the runner in order to be able to run Kubernetes actions.

## How to use Datadog Helm repository

You need to add this repository to your Helm repositories:

```
helm repo add datadog https://helm.datadoghq.com
helm repo update
```

## Requirements
* `kubectl` CLI is installed on my machine
* Helm is installed on my machine
* The permissions of my Kubernetes environment allow the Datadog Private Action Runner to read and write using a Kubernetes service account

## Use this chart
1. Go to the [Private Action Runner tab](https://app.datadoghq.com/workflow/private-action-runners).
2. Create a new Private Action Runner and follow the instructions for Kubernetes.

## Use this chart with connection credentials
1. Go to the [Private Action Runner tab](https://app.datadoghq.com/workflow/private-action-runners).
2. Create a new Private Action Runner and follow the instructions for Kubernetes.
3. Configure [connection credentials](https://docs.datadoghq.com/service_management/workflows/private_actions/private_action_credentials) for the selected private actions via `config.yaml`.

## To use Kubernetes actions
1. Go to the [Workflow connections page](https://app.datadoghq.com/workflow/connections).
2. Create a new connection, select your private action runner, and use **Service account authentication**.
3. Create a new workflow and use a Kubernetes action like **List pod** or **List deployment**.

## Going further
* Adjust the service account permissions according to your needs. Learn more about [Kubernetes RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac).
* Deploy several runners with different permissions or create different connections according to your needs.
* Learn more about [Private actions](https://docs.datadoghq.com/service_management/app_builder/private_actions).

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| common.image | object | `{"repository":"us-east4-docker.pkg.dev/datadog-sandbox/apps-on-prem/onprem-runner","tag":"v0.1.0-beta"}` | Current Datadog Private Action Runner image |
| connectionCredentials.basicAuth.credentials | list | `[]` | List of credentials for Basic Auth |
| connectionCredentials.jenkinsAuth.credentials | list | `[]` | List of credentials for Jenkins Auth |
| connectionCredentials.postgresAuth.credentials | list | `[]` | List of credentials for Postgres Auth |
| connectionCredentials.tokenAuth.credentials | list | `[]` | List of credentials for Token Auth |
| runners[0].config | object | `{"actionsAllowlist":["com.datadoghq.kubernetes.core.listPod"],"ddBaseURL":"https://app.datadoghq.com","modes":["workflowAutomation","appBuilder"],"port":9016,"privateKey":"PRIVATE_KEY_FROM_CONFIG","urn":"URN_FROM_CONFIG"}` | Configuration for the Datadog Private Action Runner |
| runners[0].config.actionsAllowlist | list | `["com.datadoghq.kubernetes.core.listPod"]` | List of actions that the Datadog Private Action Runner is allowed to execute |
| runners[0].config.ddBaseURL | string | `"https://app.datadoghq.com"` | Base URL of the Datadog app |
| runners[0].config.modes | list | `["workflowAutomation","appBuilder"]` | Modes that the runner can run in |
| runners[0].config.port | int | `9016` | Port for HTTP server liveness checks and App Builder mode |
| runners[0].config.privateKey | string | `"PRIVATE_KEY_FROM_CONFIG"` | The runner's privateKey from the enrollment page |
| runners[0].config.urn | string | `"URN_FROM_CONFIG"` | The runner's URN from the enrollment page |
| runners[0].kubernetesPermissions | list | `[{"apiGroups":[""],"resources":["pods"],"verbs":["list","get"]},{"apiGroups":["apps"],"resources":["deployments"],"verbs":["list","get"]}]` | List of Kubernetes permissions that the Datadog Private Action Runner has |
| runners[0].name | string | `"default"` | Name of the Datadog Private Action Runner |
| runners[0].replicas | int | `1` | Number of pod instances for the Datadog Private Action Runner |
