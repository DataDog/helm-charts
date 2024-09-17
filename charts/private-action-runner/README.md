# Datadog Private Action Runner

![Version: 0.12.0](https://img.shields.io/badge/Version-0.12.0-informational?style=flat-square) ![AppVersion: v0.1.0-beta](https://img.shields.io/badge/AppVersion-v0.1.0--beta-informational?style=flat-square)

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
3. Configure [connection credentials](https://docs.datadoghq.com/service_management/workflows/private_actions/private_action_credentials) for the selected private actions via `values.yaml`.

## To use Kubernetes actions
1. Go to the [Workflow connections page](https://app.datadoghq.com/workflow/connections).
2. Create a new connection, select your private action runner, and use **Service account authentication**.
3. Enable the actions you want in the Chart values using `kubernetesActions` (see [the example file](examples/values.yaml)).
4. Create a new workflow and use a Kubernetes action like **List pod** or **List deployment**.

## Going further
* Learn more about [Kubernetes RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac).
* Deploy several runners with different permissions or create different connections according to your needs.
* Learn more about [Private actions](https://docs.datadoghq.com/service_management/app_builder/private_actions).

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| common.image | object | `{"repository":"us-east4-docker.pkg.dev/datadog-sandbox/apps-on-prem/onprem-runner","tag":"v0.1.0-beta"}` | Current Datadog Private Action Runner image |
| credentialFiles | list | `[]` | List of credential files to be used by the Datadog Private Action Runner |
| runners[0].config | object | `{"actionsAllowlist":[],"ddBaseURL":"https://app.datadoghq.com","modes":["workflowAutomation","appBuilder"],"port":9016,"privateKey":"CHANGE_ME_PRIVATE_KEY_FROM_CONFIG","urn":"CHANGE_ME_URN_FROM_CONFIG"}` | Configuration for the Datadog Private Action Runner |
| runners[0].config.actionsAllowlist | list | `[]` | List of actions that the Datadog Private Action Runner is allowed to execute |
| runners[0].config.ddBaseURL | string | `"https://app.datadoghq.com"` | Base URL of the Datadog app |
| runners[0].config.modes | list | `["workflowAutomation","appBuilder"]` | Modes that the runner can run in |
| runners[0].config.port | int | `9016` | Port for HTTP server liveness checks and App Builder mode |
| runners[0].config.privateKey | string | `"CHANGE_ME_PRIVATE_KEY_FROM_CONFIG"` | The runner's privateKey from the enrollment page |
| runners[0].config.urn | string | `"CHANGE_ME_URN_FROM_CONFIG"` | The runner's URN from the enrollment page |
| runners[0].kubernetesActions | object | `{"configMaps":[],"controllerRevisions":[],"cronJobs":[],"customObjects":[],"customResourceDefinitions":[],"daemonSets":[],"deployments":[],"endpoints":[],"events":[],"jobs":[],"limitRanges":[],"namespaces":[],"nodes":[],"persistentVolumeClaims":[],"persistentVolumes":[],"podTemplates":[],"pods":["get","list"],"replicaSets":[],"replicationControllers":[],"resourceQuotas":[],"serviceAccounts":[],"services":[],"statefulSets":[]}` | Add Kubernetes actions to the `config.actionsAllowlist` and corresponding permissions for the service account |
| runners[0].kubernetesActions.configMaps | list | `[]` | Actions related to configMaps (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runners[0].kubernetesActions.controllerRevisions | list | `[]` | Actions related to controllerRevisions (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runners[0].kubernetesActions.cronJobs | list | `[]` | Actions related to cronJobs (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runners[0].kubernetesActions.customObjects | list | `[]` | Actions related to customObjects (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple"). You also need to add appropriate `kubernetesPermissions`. |
| runners[0].kubernetesActions.customResourceDefinitions | list | `[]` | Actions related to customResourceDefinitions (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runners[0].kubernetesActions.daemonSets | list | `[]` | Actions related to daemonSets (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runners[0].kubernetesActions.deployments | list | `[]` | Actions related to deployments (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple", "restart") |
| runners[0].kubernetesActions.endpoints | list | `[]` | Actions related to endpoints (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runners[0].kubernetesActions.events | list | `[]` | Actions related to events (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runners[0].kubernetesActions.jobs | list | `[]` | Actions related to jobs (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runners[0].kubernetesActions.limitRanges | list | `[]` | Actions related to limitRanges (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runners[0].kubernetesActions.namespaces | list | `[]` | Actions related to namespaces (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runners[0].kubernetesActions.nodes | list | `[]` | Actions related to nodes (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runners[0].kubernetesActions.persistentVolumeClaims | list | `[]` | Actions related to persistentVolumeClaims (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runners[0].kubernetesActions.persistentVolumes | list | `[]` | Actions related to persistentVolumes (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runners[0].kubernetesActions.podTemplates | list | `[]` | Actions related to podTemplates (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runners[0].kubernetesActions.pods | list | `["get","list"]` | Actions related to pods (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runners[0].kubernetesActions.replicaSets | list | `[]` | Actions related to replicaSets (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runners[0].kubernetesActions.replicationControllers | list | `[]` | Actions related to replicationControllers (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runners[0].kubernetesActions.resourceQuotas | list | `[]` | Actions related to resourceQuotas (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runners[0].kubernetesActions.serviceAccounts | list | `[]` | Actions related to serviceAccounts (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runners[0].kubernetesActions.services | list | `[]` | Actions related to services (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runners[0].kubernetesActions.statefulSets | list | `[]` | Actions related to statefulSets (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runners[0].kubernetesPermissions | list | `[]` | Kubernetes permissions to provide in addition to the one that will be inferred from `kubernetesActions` (useful for customObjects) |
| runners[0].name | string | `"default"` | Name of the Datadog Private Action Runner |
| runners[0].replicas | int | `1` | Number of pod instances for the Datadog Private Action Runner |
