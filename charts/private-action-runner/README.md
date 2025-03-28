# Datadog Private Action Runner

![Version: 0.20.1](https://img.shields.io/badge/Version-0.20.1-informational?style=flat-square) ![AppVersion: v1.1.1](https://img.shields.io/badge/AppVersion-v1.1.1-informational?style=flat-square)

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

### Use a kubernetes secret to store the runner's identity

If you want to store the runner's identity outside of the Helm chart, you can create a kubernetes secret and use it in the `values.yaml` file.
```bash
# Create a secret with runner's private key and urn
kubectl create secret generic <secret-name> --from-literal RUNNER_URN=<CHANGE_ME_URN_FROM_CONFIG> --from-literal RUNNER_PRIVATE_KEY=<CHANGE_ME_PRIVATE_KEY_FROM_CONFIG>
# Alternatively you can only store the private key in the secret and keep the URN in the values.yaml
kubectl create secret generic <secret-name>  --from-literal RUNNER_PRIVATE_KEY=<CHANGE_ME_PRIVATE_KEY_FROM_CONFIG>
```
Update the `values.yaml` file with the secret name
```yaml
runners:
  -
  # ... other fields
    runnerIdentitySecret: <secret-name>
  # ... other fields
    config:
      # you can get rid of the values in `config`, the secret will take precedence
      # urn: "STORED_IN_A_SECRET"
      # privateKey: "STORED_IN_A_SECRET"
```

### Use kubernetes secrets to store credentials

If you want to store the credentials outside of the Helm chart, you can create a kubernetes secret and use it in the `values.yaml` file.
```bash
# Create a secret with the credentials files
kubectl create secret generic <secret-name> --from-literal jenkins_token.json='{"auth_type": "Token Auth", "credentials": [{"tokenName": "username", "tokenValue": "USERNAME"}, {"tokenName": "token", "tokenValue": "TOKEN"}, {"tokenName": "domain", "tokenValue": "DOMAIN" }]}' --from-literal gitlab_token.json='{"auth_type": "Token Auth", "credentials": [{"tokenName": "baseURL", "tokenValue": "GITLAB_BASE_URL"}, {"tokenName": "gitlabApiToken", "tokenValue": "GITLAB_API_TOKEN"}]}'

# Alternatively create one secret per credentials file
kubectl create secret generic <secret-name-jenkins> --from-literal jenkins_token.json='{"auth_type": "Token Auth", "credentials": [{"tokenName": "username", "tokenValue": "USERNAME"}, {"tokenName": "token", "tokenValue": "TOKEN"}, {"tokenName": "domain", "tokenValue": "DOMAIN" }]}'
kubectl create secret generic <secret-name-gitlab> --from-literal gitlab_token.json='{"auth_type": "Token Auth", "credentials": [{"tokenName": "baseURL", "tokenValue": "GITLAB_BASE_URL"}, {"tokenName": "gitlabApiToken", "tokenValue": "GITLAB_API_TOKEN"}]}'
```

Update the `values.yaml` file with the secrets names and the directory names
```yaml
credentialSecrets:
  # your credentials files will be located at /etc/dd-action-runner/credentials/jenkins_token.json and /etc/dd-action-runner/credentials/gitlab_token.json
  - secretName: <secret-name>
    directoryName: ""
  # your credentials file will be located at /etc/dd-action-runner/credentials/jenkins/jenkins_token.json
  - secretName: <secret-name-jenkins>
    directoryName: "jenkins"
  # your credentials file will be located at /etc/dd-action-runner/credentials/gitlab/gitlab_token.json
  - secretName: <secret-name-gitlab>
    directoryName: "gitlab"
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| common.image | object | `{"repository":"gcr.io/datadoghq/private-action-runner","tag":"v1.1.1"}` | Current Datadog Private Action Runner image |
| credentialFiles | list | `[]` | List of credential files to be used by the Datadog Private Action Runner |
| credentialSecrets | list | `[]` | References to kubernetes secrets that contain credentials to be used by the Datadog Private Action Runner |
| runners[0].config | object | `{"actionsAllowlist":[],"ddBaseURL":"https://app.datadoghq.com","modes":["workflowAutomation","appBuilder"],"port":9016,"privateKey":"CHANGE_ME_PRIVATE_KEY_FROM_CONFIG","urn":"CHANGE_ME_URN_FROM_CONFIG"}` | Configuration for the Datadog Private Action Runner |
| runners[0].config.actionsAllowlist | list | `[]` | List of actions that the Datadog Private Action Runner is allowed to execute |
| runners[0].config.ddBaseURL | string | `"https://app.datadoghq.com"` | Base URL of the Datadog app |
| runners[0].config.modes | list | `["workflowAutomation","appBuilder"]` | Modes that the runner can run in |
| runners[0].config.port | int | `9016` | Port for HTTP server liveness checks and App Builder mode |
| runners[0].config.privateKey | string | `"CHANGE_ME_PRIVATE_KEY_FROM_CONFIG"` | The runner's privateKey from the enrollment page |
| runners[0].config.urn | string | `"CHANGE_ME_URN_FROM_CONFIG"` | The runner's URN from the enrollment page |
| runners[0].env | list | `[]` | Environment variables to be passed to the Datadog Private Action Runner |
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
| runners[0].roleType | string | `"Role"` | Type of kubernetes role to create (either "Role" or "ClusterRole") |
| runners[0].runnerIdentitySecret | string | `""` | Reference to a kubernetes secrets that contains the runner identity |
