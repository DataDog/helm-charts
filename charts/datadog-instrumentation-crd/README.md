# Datadog Instrumentation CRD

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![AppVersion: 1](https://img.shields.io/badge/AppVersion-1-informational?style=flat-square)

Installs the DatadogInstrumentation custom resource definition used by the Datadog Cluster Agent.

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| annotations | object | `{}` | Annotations to add to the DatadogInstrumentation CRD. |
| keepCrds | bool | `false` | Instruct Helm to keep the DatadogInstrumentation CRD when the chart is uninstalled. |
