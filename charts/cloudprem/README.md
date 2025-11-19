# CloudPrem

![Version: 0.1.12](https://img.shields.io/badge/Version-0.1.12-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v0.1.15](https://img.shields.io/badge/AppVersion-v0.1.15-informational?style=flat-square)

## Using the Datadog Helm repository

Add and update the Datadog Helm repository to your Helm repositories:

```sh
helm repo add datadog https://helm.datadoghq.com
helm repo update
```

## Prerequisites

- AWS account
- Kubernetes `1.25+` ([EKS](https://aws.amazon.com/eks/) preferred)
- [AWS Load Balancer Controller](https://kubernetes-sigs.github.io/aws-load-balancer-controller)
- PostgreSQL database ([RDS](https://aws.amazon.com/rds/) preferred)
- S3 bucket

## Quick start

### Creating the Kubernetes namespace

```sh
kubectl create namespace <namespace name>
```

### Storing the PostgreSQL database connection string as a Kubernetes secret

```sh
kubectl create secret generic <secret name> --from-literal QW_METASTORE_URI=postgres://<username>:<password>@<endpoint>:<port>/<database> -n <namespace name>
```

### Customizing the Helm chart

Create a `datadog-values.yaml` file to override the default values with your custom configuration. This is where you define environment-specific settings such as the image tag, AWS account ID, service account, ingress setup, resource requests and limits, and more.
Any parameters not explicitly overridden in `datadog-values.yaml` will fall back to the defaults defined in the chart’s `values.yaml`. Here is an example of a `datadog-values.yaml` file with such overrides:

```yaml
datadog:
  # The Datadog [site](https://docs.datadoghq.com/getting_started/site/) to connect to. Defaults to `datadoghq.com`.
  site: datadoghq.eu
  # The name of the existing Secret containing the Datadog API key. The secret key name must be `api-key`.
  apiKeyExistingSecret: datadog-api-key

cloudprem:
  index:
    # The retention period for the index specified as a human-readable duration such as `30d`, `6m`, or `1y`. Defaults to 30 days.
    retention: 90d

aws:
  accountId: "123456789012"
  # AWS partition, set to "aws" by default, but should be set to "aws-cn" for China regions
  partition: aws-cn

# Environment variables
# Any environment variables defined here will be available to all pods in the deployment
environment:
  AWS_REGION: cn-north-1

# Service account configuration
# If `serviceAccount.create` is set to `true`, a service account will be created with the specified name.
# The service account will be annotated with the IAM role ARN if `aws.partition`, `aws.accountId`, and serviceAccount.eksRoleName` are set.
# Additional annotations can be added using serviceAccount.extraAnnotations.
serviceAccount:
  create: true
  name: cloudprem
  # The name of the IAM role to use for the service account. If set, the following annotations will be added to the service account:
  # - eks.amazonaws.com/role-arn: arn:<aws.partition>:iam::<aws.accountId>:role/<serviceAccount.eksRoleName>
  # - eks.amazonaws.com/sts-regional-endpoints: "true"
  eksRoleName: cloudprem
  extraAnnotations: {}

# CloudPrem node configuration
config:
  # The root URI where index data will be stored. This should be an S3 path.
  # All indexes created in CloudPrem will be stored under this location.
  default_index_root_uri: s3://<bucket name>/indexes

# Ingress configuration
ingress:
  # The internal ingress is used by Datadog agents and other collectors running outside
  # the Kubernetes cluster to send their logs to CloudPrem.
  internal:
    enabled: true
    name: cloudprem-internal
    host: cloudprem.acme.internal
    # Additional annotations can be added to customize the ALB behavior.
    extraAnnotations:
      alb.ingress.kubernetes.io/load-balancer-name: cloudprem-internal

# Metastore configuration
# The metastore is responsible for storing and managing index metadata.
# It requires a PostgreSQL database connection string to be provided via a Kubernetes secret.
# The secret should contain a key named `QW_METASTORE_URI` with a value in the format:
# postgresql://<username>:<password>@<host>:<port>/<database>
#
# The metastore connection string is mounted into the pods using extraEnvFrom to reference the secret.
metastore:
  extraEnvFrom:
    - secretRef:
        name: cloudprem-metastore-uri

# Indexer configuration
# The indexer is responsible for processing and indexing incoming data it receives data from various sources (e.g., Datadog agents, log collectors)
# and transforms it into searchable files called "splits" stored in S3.
#
# The indexer is horizontally scalable - you can increase `replicaCount` or enable autoscaling to increase the cluster's indexing throughput.
# Resource requests and limits should be tuned based on your indexing workload.
#
# The default values are suitable for moderate indexing loads of up to 20MB/s per indexer pod.
indexer:
  replicaCount: 2

  resources:
    requests:
      cpu: "4"
      memory: "8Gi"
    limits:
      cpu: "4"
      memory: "8Gi"

  autoscaling:
    enabled: false

# Searcher configuration
# The searcher is responsible for executing search queries against the indexed data stored in S3.
# It handles search requests from Datadog's query service and returns matching results.
#
# The searcher is horizontally scalable - you can increase `replicaCount` or enable autoscaling to handle more concurrent search requests.
# Resource requirements for searchers are highly workload-dependent and should be determined empirically.
# Key factors that impact searcher performance include:
# - Query complexity (e.g., number of terms, use of wildcards or regex)
# - Query concurrency (number of simultaneous search requests)
# - Amount of data scanned per query
# - Data access patterns (cache hit rates)
#
# Memory is particularly important for searchers as they cache frequently accessed index data in memory.
# Monitor searcher metrics and adjust resources based on observed performance and workload characteristics.
searcher:
  replicaCount: 2

  resources:
    requests:
      cpu: "4"
      memory: "16Gi"
    limits:
      cpu: "4"
      memory: "16Gi"

  autoscaling:
    enabled: false

# Additional ConfigMaps
# Create custom ConfigMaps for application configuration, scripts, or other data
extraConfigMaps:
  - name: custom-app-config
    labels:
      component: application
    annotations:
      description: "Custom application configuration"
    data:
      app.properties: |
        database.pool.size=10
        logging.level=INFO
      init-script.sh: |
        #!/bin/bash
        echo "Initializing application..."
```

### Installing or upgrading the Helm chart

```sh
helm upgrade --install <release name> datadog/cloudprem \
  -n <namespace name> \
  -f datadog-values.yaml
```

### Uninstalling the Helm chart
To uninstall the deployment:

```sh
helm uninstall <release name>
```

This command removes all the Kubernetes resources associated with the chart and deletes the release.

## Helm Chart values (non-exhaustive)

| Key | Type | Default | Description
| :--------------- |:---------------:| -----:|--- |
|aws.accountId | string | null | AWS account ID used for the EKS role ARN service account annotation|
|aws.partition | string | aws | AWS partition used for the EKS role ARN service account annotation|
|config.* | dict | config defaults | Config used by the CloudPrem prods|
|environment | dict | {} | Key-value environment variables passed to CloudPrem pods|
|environmentFrom | list | [] | List of sources to populate environment variables (e.g., Secrets or ConfigMaps)|
|extraConfigMaps | list | [] | Additional ConfigMaps to create with custom data and configuration|
|image.pullPolicy | string | IfNotPresent | Image pull policy for CloudPrem containers|
|image.repository | string | public.ecr.aws/datadog/cloudprem | Repository of the CloudPrem image|
|image.tag | string | devel | Tag of the CloudPrem image to deploy|
|ingress.internal.enabled | bool | false | Whether to enable the internal ingress|
|ingress.internal.host | string | null | Hostname for internal ingress access|
|ingress.internal.name | string | null | Name of the internal ingress resource|
|ingress.internal.extraAnnotations | dict | {} | Annotations to add to the internal ingress resource|
|ingress.public.enabled | bool | false | Whether to enable the public ingress|
|ingress.public.extraAnnotations | dict | {} | Annotations to add to the public ingress resource|
|ingress.public.host | string | null | Hostname for public ingress access|
|ingress.public.name | string | null | Name of the public ingress resource|
|serviceAccount.create | bool | true | Whether to create a new Kubernetes service account|
|serviceAccount.eksRoleName | string | null | IAM role name to associate with the service account|
|serviceAccount.extraAnnotations | dict | {} | Extra annotations to add to the service account|
|serviceAccount.name | string | null | Name of the service account used by the CloudPrem pods|
