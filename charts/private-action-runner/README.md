# Datadog Private Action Runner

![Version: 1.2.0](https://img.shields.io/badge/Version-1.2.0-informational?style=flat-square) ![AppVersion: v1.3.0](https://img.shields.io/badge/AppVersion-v1.3.0-informational?style=flat-square)

## Overview

This Helm Chart deploys the Datadog Private Action Runner inside a Kubernetes cluster. The Private Action Runner enables you to:

- Execute private actions from [Datadog Workflow Automation](https://docs.datadoghq.com/service_management/workflows/)
- Run [App Builder](https://docs.datadoghq.com/service_management/app_builder/) actions
- Interact with resources in your Kubernetes cluster
- Connect to internal services that aren't accessible from the public internet

## Prerequisites

Before installing the chart, ensure you have:

* Kubernetes cluster running
* `kubectl` CLI installed and configured to access your cluster
* `helm` CLI installed
* Appropriate permissions in your Kubernetes environment to create resources like Deployments, Services, and RBAC objects

## Installation

### Add the Datadog Helm Repository

```bash
helm repo add datadog https://helm.datadoghq.com
helm repo update
```

### Create a Private Action Runner in Datadog

1. Go to the [Private Action Runner tab](https://app.datadoghq.com/workflow/private-action-runners) in your Datadog account
2. Click "New Private Action Runner"
3. Configure your runner and select the list of actions you want to enable
4. Select "Kubernetes" as the deployment method
5. Note the config that gets printed in your terminal (URN, privateKey, baseUrl, actionsAllowlist, etc.)

### Install the Chart

Create a `values.yaml` file with your runner configuration (see the [examples/values.yaml](examples/values.yaml) for a complete example):

```yaml
runner:
  config:
    urn: "YOUR_RUNNER_URN"
    privateKey: "YOUR_RUNNER_PRIVATE_KEY"
```

Install the chart:

```bash
helm install <RELEASE_NAME> datadog/private-action-runner -f values.yaml
```

### Verify the Installation

Check that the runner pod is running:

```bash
kubectl get pods -l app.kubernetes.io/instance=<RELEASE-NAME>
```

## Upgrading

### Upgrading from 0.x to 1.0.0

> **Important:** Version 1.0.0 introduces breaking changes to the values.yaml structure. If you're upgrading from version 0.x, please follow the dedicated upgrade [UPGRADING.md](UPGRADING.md) guide.

### General Upgrade Process

To upgrade to the latest version:

```bash
helm repo update
helm upgrade <RELEASE_NAME> datadog/private-action-runner -f values.yaml
```

## Usage

### Using Connection Credentials

To use private actions that require credentials:

1. Configure [connection credentials](https://docs.datadoghq.com/service_management/workflows/private_actions/private_action_credentials) in your `values.yaml` file
2. Update your Helm release:
```bash
helm upgrade <RELEASE_NAME> datadog/private-action-runner -f values.yaml
```
3. Create the connection in [Datadog](https://app.datadoghq.com/actions/connections)

### Using Kubernetes Actions

To enable Kubernetes actions:

1. Go to the [Workflow connections page](https://app.datadoghq.com/workflow/connections)
2. Create a new connection, select your private action runner, and use **Service account authentication**
3. Enable the actions you want in your `values.yaml` file:

```yaml
runner:
  kubernetesActions:
    pods: ["get", "list"]
    deployments: ["get", "list", "create", "update"]
```

4. Update your Helm release
```bash
helm upgrade <RELEASE_NAME> datadog/private-action-runner -f values.yaml
```

## Going Further

* Learn more about [Kubernetes RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac)
* Deploy several runners with different permissions for different teams or environments
* Learn more about [Private actions](https://docs.datadoghq.com/actions/private_actions/)

## Advanced Configuration

### Using Kubernetes Secrets for Runner Identity

For enhanced security, you can store the runner's identity (URN and private key) in a Kubernetes secret instead of in the values.yaml file:

```bash
# Create a secret with runner's private key and urn
kubectl create secret generic runner-identity \
  --from-literal RUNNER_URN=YOUR_RUNNER_URN \
  --from-literal RUNNER_PRIVATE_KEY=YOUR_RUNNER_PRIVATE_KEY

# Alternatively, store only the private key in the secret
kubectl create secret generic <secret-name> \
  --from-literal RUNNER_PRIVATE_KEY=YOUR_RUNNER_PRIVATE_KEY
```

Then reference this secret in your values.yaml:

```yaml
runner:
  runnerIdentitySecret: "runner-identity"
  config:
    # When using runnerIdentitySecret, you can omit these values
    # urn: "YOUR_RUNNER_URN"  # Only needed if not in the secret
    # privateKey: "YOUR_RUNNER_PRIVATE_KEY"
```

### Using Kubernetes Secrets for Credentials

You can also store connection credentials in Kubernetes secrets:

```bash
# Create a secret with multiple credential files
kubectl create secret generic action-credentials \
  --from-literal jenkins_token.json='{"auth_type": "Token Auth", "credentials": [{"tokenName": "username", "tokenValue": "USERNAME"}, {"tokenName": "token", "tokenValue": "TOKEN"}, {"tokenName": "domain", "tokenValue": "DOMAIN" }]}' \
  --from-literal gitlab_token.json='{"auth_type": "Token Auth", "credentials": [{"tokenName": "baseURL", "tokenValue": "GITLAB_BASE_URL"}, {"tokenName": "gitlabApiToken", "tokenValue": "GITLAB_API_TOKEN"}]}'

# Or create separate secrets for different services
kubectl create secret generic jenkins-credentials \
  --from-literal jenkins_token.json='{"auth_type": "Token Auth", "credentials": [{"tokenName": "username", "tokenValue": "USERNAME"}, {"tokenName": "token", "tokenValue": "TOKEN"}, {"tokenName": "domain", "tokenValue": "DOMAIN" }]}'
```

Reference these secrets in your values.yaml:

```yaml
runner:
  credentialSecrets:
    # Mount all files from the secret at /etc/dd-action-runner/credentials/
    - secretName: action-credentials
      directoryName: ""
    # Mount files in a subdirectory at /etc/dd-action-runner/credentials/jenkins/
    - secretName: jenkins-credentials
      directoryName: "jenkins"
```

## Architecture

The Private Action Runner Helm chart deploys the following components:

- **Deployment**: Runs the Private Action Runner container
- **Service**: Exposes the runner's HTTP endpoint for health checks and App Builder mode
- **ServiceAccount**: Identity used by the runner to interact with the Kubernetes API
- **Role/ClusterRole**: Defines permissions for the runner to perform Kubernetes actions
- **RoleBinding/ClusterRoleBinding**: Associates the ServiceAccount with the Role/ClusterRole
- **Secret**: Stores the runner configuration and credentials

## Troubleshooting

1. Check if the pod is running:
   ```bash
   kubectl get pods -l app.kubernetes.io/instance=<RELEASE-NAME>
   ```

2. Check the pod logs for connection issues:
   ```bash
   kubectl logs -l app.kubernetes.io/instance=<RELEASE-NAME>
   ```

3. Verify that the URN and private key are correct in your values.yaml or secret

### Connection Credential Issues

If actions requiring credentials fail:

1. Verify that your credential files are properly formatted
2. Check that the credentials are mounted correctly in the pod:
   ```bash
   kubectl exec <pod-name> -- ls /etc/dd-action-runner/credentials/
   ## Depending on how you pass the credentials they might appear in a different directory
   kubectl exec <pod-name> -- ls /etc/dd-action-runner/
   ```

3. Check the pod logs for credential-related errors

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| $schema | string | `"./values.schema.json"` | Schema for the values file, enables support in Jetbrains IDEs. You should probably use https://raw.githubusercontent.com/DataDog/helm-charts/refs/heads/main/charts/private-action-runner/values.schema.json. |
| fullnameOverride | string | `""` | Override the full qualified app name |
| image | object | `{"repository":"gcr.io/datadoghq/private-action-runner","tag":"v1.3.0"}` | Current Datadog Private Action Runner image |
| nameOverride | string | `""` | Override name of app |
| runner.config | object | `{"actionsAllowlist":[],"ddBaseURL":"https://app.datadoghq.com","modes":["workflowAutomation","appBuilder"],"port":9016,"privateKey":"CHANGE_ME_PRIVATE_KEY_FROM_CONFIG","urn":"CHANGE_ME_URN_FROM_CONFIG"}` | Configuration for the Datadog Private Action Runner |
| runner.config.actionsAllowlist | list | `[]` | List of actions that the Datadog Private Action Runner is allowed to execute |
| runner.config.ddBaseURL | string | `"https://app.datadoghq.com"` | Base URL of the Datadog app |
| runner.config.modes | list | `["workflowAutomation","appBuilder"]` | Modes that the runner can run in |
| runner.config.port | int | `9016` | Port for HTTP server liveness checks and App Builder mode |
| runner.config.privateKey | string | `"CHANGE_ME_PRIVATE_KEY_FROM_CONFIG"` | The runner's privateKey from the enrollment page |
| runner.config.urn | string | `"CHANGE_ME_URN_FROM_CONFIG"` | The runner's URN from the enrollment page |
| runner.credentialFiles | list | `[]` | List of credential files to be used by the Datadog Private Action Runner |
| runner.credentialSecrets | list | `[]` | References to kubernetes secrets that contain credentials to be used by the Datadog Private Action Runner |
| runner.env | list | `[]` | Environment variables to be passed to the Datadog Private Action Runner |
| runner.kubernetesActions | object | `{"configMaps":[],"controllerRevisions":[],"cronJobs":[],"customObjects":[],"customResourceDefinitions":[],"daemonSets":[],"deployments":[],"endpoints":[],"events":[],"jobs":[],"limitRanges":[],"namespaces":[],"nodes":[],"persistentVolumeClaims":[],"persistentVolumes":[],"podTemplates":[],"pods":["get","list"],"replicaSets":[],"replicationControllers":[],"resourceQuotas":[],"serviceAccounts":[],"services":[],"statefulSets":[]}` | Add Kubernetes actions to the `config.actionsAllowlist` and corresponding permissions for the service account |
| runner.kubernetesActions.configMaps | list | `[]` | Actions related to configMaps (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runner.kubernetesActions.controllerRevisions | list | `[]` | Actions related to controllerRevisions (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runner.kubernetesActions.cronJobs | list | `[]` | Actions related to cronJobs (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runner.kubernetesActions.customObjects | list | `[]` | Actions related to customObjects (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple"). You also need to add appropriate `kubernetesPermissions`. |
| runner.kubernetesActions.customResourceDefinitions | list | `[]` | Actions related to customResourceDefinitions (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runner.kubernetesActions.daemonSets | list | `[]` | Actions related to daemonSets (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runner.kubernetesActions.deployments | list | `[]` | Actions related to deployments (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple", "restart", "rollback", "scale") |
| runner.kubernetesActions.endpoints | list | `[]` | Actions related to endpoints (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runner.kubernetesActions.events | list | `[]` | Actions related to events (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runner.kubernetesActions.jobs | list | `[]` | Actions related to jobs (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runner.kubernetesActions.limitRanges | list | `[]` | Actions related to limitRanges (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runner.kubernetesActions.namespaces | list | `[]` | Actions related to namespaces (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runner.kubernetesActions.nodes | list | `[]` | Actions related to nodes (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runner.kubernetesActions.persistentVolumeClaims | list | `[]` | Actions related to persistentVolumeClaims (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runner.kubernetesActions.persistentVolumes | list | `[]` | Actions related to persistentVolumes (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runner.kubernetesActions.podTemplates | list | `[]` | Actions related to podTemplates (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runner.kubernetesActions.pods | list | `["get","list"]` | Actions related to pods (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runner.kubernetesActions.replicaSets | list | `[]` | Actions related to replicaSets (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runner.kubernetesActions.replicationControllers | list | `[]` | Actions related to replicationControllers (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runner.kubernetesActions.resourceQuotas | list | `[]` | Actions related to resourceQuotas (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runner.kubernetesActions.serviceAccounts | list | `[]` | Actions related to serviceAccounts (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runner.kubernetesActions.services | list | `[]` | Actions related to services (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runner.kubernetesActions.statefulSets | list | `[]` | Actions related to statefulSets (options: "get", "list", "create", "update", "patch", "delete", "deleteMultiple") |
| runner.kubernetesPermissions | list | `[]` | Kubernetes permissions to provide in addition to the one that will be inferred from `kubernetesActions` (useful for customObjects) |
| runner.replicas | int | `1` | Number of pod instances for the Datadog Private Action Runner |
| runner.resources | object | `{"limits":{"cpu":"250m","memory":"1Gi"},"requests":{"cpu":"250m","memory":"1Gi"}}` | Resource requirements for the Datadog Private Action Runner container |
| runner.resources.limits | object | `{"cpu":"250m","memory":"1Gi"}` | Resource limits for the runner container |
| runner.resources.requests | object | `{"cpu":"250m","memory":"1Gi"}` | Resource requests for the runner container |
| runner.roleType | string | `"Role"` | Type of kubernetes role to create (either "Role" or "ClusterRole") |
| runner.runnerIdentitySecret | string | `""` | Reference to a kubernetes secrets that contains the runner identity |
