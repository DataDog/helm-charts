# Upgrade from version 0.x to version 1.x

Version 1.0.0 introduces changes to simplify the chart and better align with Helm best practices. The most significant change is the restructuring of the values.yaml file.

## Breaking Changes in Values.yaml Structure

### 1. Runners Array to Single Runner Object

In version 0.x, the chart used a top-level `runners` array where each runner had its own configuration:

```yaml
# Old structure (0.x)
runners:
  - name: "custom-runner"
    config:
      ddBaseURL: "https://app.datadoghq.com"
      # other config options...
    roleType: "Role"
    # other runner options...
```

In version 1.0.0, this has been simplified to a single `runner` object:

```yaml
# New structure (1.0.0)
runner:
  roleType: "Role"
  config:
    ddBaseURL: "https://app.datadoghq.com"
    # other config options...
  # other runner options...
```

### 2. Credential Files and Secrets Moved Under Runner

In version 0.x, credential files and secrets were defined at the top level:

```yaml
# Old structure (0.x)
runners:
  - name: "custom-runner"
    # runner configuration...

credentialFiles:
  - fileName: "http_basic_creds.json"
    data: |
      # credential data...

credentialSecrets:
  - secretName: "my-secret"
    directoryName: "my-directory"
```

In version 1.0.0, these have been moved under the `runner` object:

```yaml
# New structure (1.0.0)
runner:
  # runner configuration...
  credentialFiles:
    - fileName: "http_basic_creds.json"
      data: |
        # credential data...

  credentialSecrets:
    - secretName: "my-secret"
      directoryName: "my-directory"
```

## Migration Guide

To migrate your values.yaml file from version 0.x to 1.0.0:

1. If you have multiple runners defined in the `runners` array, you'll have to to create one helm release per runner.

2. Move the configuration from `runners[0]` to the new `runner` object, removing the `name` field.

3. Move any top-level `credentialFiles` and `credentialSecrets` under the `runner` object.

Example migration:

```yaml
# Old structure (0.x)
runners:
  - name: "custom-runner"
    config:
      ddBaseURL: "https://app.datadoghq.com"
      urn: "my-urn"
      privateKey: "my-private-key"
    roleType: "Role"
    kubernetesActions:
      pods: ["get", "list"]

credentialFiles:
  - fileName: "creds.json"
    data: |
      { "credentials": [] }

# New structure (1.0.0)
runner:
  config:
    ddBaseURL: "https://app.datadoghq.com"
    urn: "my-urn"
    privateKey: "my-private-key"
  roleType: "Role"
  kubernetesActions:
    pods: ["get", "list"]
  credentialFiles:
    - fileName: "creds.json"
      data: |
        { "credentials": [] }
```

## Run the upgrade

```bash
helm upgrade <release-name> datadog/private-action-runner -f your-migrated-values.yaml -n <namespace>
```

## Use older version of the chart

If you need to use the older version of the chart, you can specify the version in your Helm command:

```bash
helm install <release-name> datadog/private-action-runner --version 0.20.1 -f your-values.yaml -n <namespace>
```