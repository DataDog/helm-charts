# Datadog Private Action Runner

![Version: 0.7.0](https://img.shields.io/badge/Version-0.7.0-informational?style=flat-square) ![AppVersion: v0.0.1-alpha27](https://img.shields.io/badge/AppVersion-v0.0.1--alpha27-informational?style=flat-square)

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
2. Set up a new Private Action runner by following the Kubernetes instructions. When you reach step 4, instead of running `helm install`, make the following changes to the Helm chart.
3. Download the chart locally.
```bash
helm pull datadog/private-action-runner --untar
```
4. Add connection credential json file to `templates/secrets.yaml` in the format corresponding to the credential and action types you want to use.

HTTP Basic Auth:
```
{
   auth_type: 'Basic Auth',
   credentials: [
      {
         username: 'USERNAME',
         password: 'PASSWORD',
      },
   ],
}
```
HTTP Token Auth:
```
{
   auth_type: 'Token Auth',
   credentials: [
      {
         tokenName: 'TOKEN1',
         tokenValue: 'VALUE1',
      },
   ],
}
```
Jenkins:
```
{
   auth_type: 'Token Auth',
   credentials: [
      {
         username: 'USERNAME',
         token: 'TOKEN',
         domain: 'DOMAIN',
      },
   ],
}
```
Postgres:
```
{
   auth_type: 'Token Auth',
   credentials: [
      {
         tokenName: 'connectionUri',
         tokenValue: 'postgres://usr:password@example_host:5432/example_db',
      },
   ],
}
```
5. Install the chart locally.
```bash
helm install <RELEASE_NAME> ./private-action-runner -f ./config.yaml
```

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
| common.image | object | `{"repository":"us-east4-docker.pkg.dev/datadog-sandbox/apps-on-prem/onprem-runner","tag":"v0.0.1-alpha27"}` | Current Datadog Private Action Runner image |
| runners[0].config | object | `{"actionsAllowlist":["com.datadoghq.kubernetes.core.listPod"],"appBuilder":{"port":9016},"ddBaseURL":"https://app.datadoghq.com","modes":["workflowAutomation","appBuilder"],"privateKey":"PRIVATE_KEY_FROM_CONFIG","urn":"URN_FROM_CONFIG"}` | Configuration for the Datadog Private Action Runner |
| runners[0].config.actionsAllowlist | list | `["com.datadoghq.kubernetes.core.listPod"]` | List of actions that the Datadog Private Action Runner is allowed to execute |
| runners[0].config.appBuilder.port | int | `9016` | Required port for App Builder Mode |
| runners[0].config.ddBaseURL | string | `"https://app.datadoghq.com"` | Base URL of the Datadog app |
| runners[0].config.modes | list | `["workflowAutomation","appBuilder"]` | Modes that the runner can run in |
| runners[0].config.privateKey | string | `"PRIVATE_KEY_FROM_CONFIG"` | The runner's privateKey from the enrollment page |
| runners[0].config.urn | string | `"URN_FROM_CONFIG"` | The runner's URN from the enrollment page |
| runners[0].kubernetesPermissions | list | `[{"apiGroups":[""],"resources":["pods"],"verbs":["list","get"]},{"apiGroups":["apps"],"resources":["deployments"],"verbs":["list","get"]}]` | List of Kubernetes permissions that the Datadog Private Action Runner has |
| runners[0].name | string | `"default"` | Name of the Datadog Private Action Runner |
| runners[0].replicas | int | `1` | Number of pod instances for the Datadog Private Action Runner |
