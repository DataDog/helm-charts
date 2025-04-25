# Operator EKS Add-on

This is a wrapper chart for installing EKS add-on. Charts required for the add-on are added as a dependency to this chart. Chart itself doesn't contain any templates or configurable properties.

## Version Mapping
| `operator-addon-chart` | `datadog-operator` | `datadog-crds` | Operator | Agent | Cluster Agent |
| :-: | :-: | :-: | :-: | :-: | :-: |
| < 0.1.6 | 1.0.5 | 1.0.1 | 1.0.3 | 7.43.1 | 7.43.1 | 
| 0.1.6 | 1.4.1 | 1.3.0 | 1.3.0 | 7.47.1 | 7.47.1 |
| 0.1.7 | 1.5.1 | 1.4.0 | 1.4.0 | 7.50.3 | 7.50.3 |
| 0.1.9 | 1.8.1 | 1.7.0 | 1.7.0 | 7.54.0 | 7.54.0 |
| 0.1.10 | 2.5.1 | 2.3.0 | 1.11.1 | 7.60.0 | 7.60.0 |
| 0.1.12 | 2.7.0 | 2.4.1 | 1.12.1 | 7.62.2 | 7.62.2 |
| 0.1.13 | 2.9.1 | 2.7.0 | 1.14.0 | 7.64.3 | 7.64.3 |

* 0.1.8 failed validation and didn't go through.
* 0.1.11 failed validation and didn't go through.

## Pushing Add-on Chart
Below steps have been validated using `Helm v3.12.0`.

* Merge `main` branch into `operator-eks-addon` branch: this will require to resolve conflicts manually.
* Update:
    * `charts/operator-eks-addon/CHANGELOG.md`: provide an entry for the new version.
    * `charts/operator-eks-addon/Chart.yaml`: update the version and the `datadog-operator` version dependency
    * `charts/operator-eks-addon/README.md`: add a row to the version mapping
* Build dependencies - this step is necessary to update the dependent charts under `charts/operator-eks-addon/charts` and `charts/datadog-operator/charts`. Latter is necessary to include correct version of `datadog-crds` chart in the `datadog-operator`.
```sh
helm dependency update ./charts/datadog-operator
helm dependency build ./charts/datadog-operator
helm dependency update ./charts/operator-eks-addon
helm dependency build ./charts/operator-eks-addon
```

* Package chart - this step creates a chart archive e.g. `operator-eks-addon-0.1.0.tgz`
```sh
helm package ./charts/operator-eks-addon
```

* Validate the artifact
```sh
# Unpack the chart and go the the chart folder
tar -xzf operator-eks-addon-0.1.3.tgz -C /tmp/
cd /tmp/operator-eks-addon

# Review chart version and dependency version are correct
Ensure in all the manifests/templates the absence of unsupported Helm objects.

# Render chart in a file
helm template datadog-operator . -n datadog-agent > operator-addon.yaml

# Render chart with a sample override
helm template datadog-operator . --set datadog-operator.image.tag=1.2.0 > operator-addon.yaml
```
Make sure the rendered manifest contains the CRD, Operator tag override works and uses correct EKS repo. 

Install it on a Kind cluster after replacing registry `709825985650.dkr.ecr.us-east-1.amazonaws.com` with a public one.

```sh
kubectl apply -f operator-addon.yaml
```
Confirm Operator is deploy and pods reach running state. Afterwards, create secret and apply default `DatadogAgent` manifest and make sure agents reach running state and metric shows up in the app.


* Push chart to EKS repo - first we need to authenticate Helm with the repo, then we push the chart archive to the Marketplace repository. This will upload the chart at `datadog/helm-charts/operator-eks-addon` and tag it with version `0.1.0`. See [ECR documentation][eks-helm-push] for more details.
```sh
aws ecr get-login-password --region us-east-1 | helm registry login --username AWS --password-stdin 709825985650.dkr.ecr.us-east-1.amazonaws.com
helm push operator-eks-addon-0.1.0.tgz oci://709825985650.dkr.ecr.us-east-1.amazonaws.com/datadog/helm-charts
```

* Validation - with AWS CLI we can list images in a give reposotiry.
```sh
aws ecr describe-images --registry-id 709825985650 --region us-east-1  --repository-name datadog/helm-charts/operator-eks-addon
{
    "imageDetails": [
        {
            "registryId": "709825985650",
            "repositoryName": "datadog/helm-charts/operator-eks-addon",
            "imageDigest": "sha256:d6e54cbe69bfb962f0a4e16c9b29a7572f6aaf479de347f91bea8331a1a867f9",
            "imageTags": [
                "0.1.0"
            ],
            "imageSizeInBytes": 63269,
            "imagePushedAt": 1690215560.0,
            "imageManifestMediaType": "application/vnd.oci.image.manifest.v1+json",
            "artifactMediaType": "application/vnd.cncf.helm.config.v1+json"
        }
    ]
}
```

## Pushing Container Images
Images required during add-on instasllation must be available through the EKS marketplace repository. Each image can be copied simply by `crane copy`. Make sure all referenced tags are uploaded to the respective repository.
```sh
aws ecr get-login-password --region us-east-1|crane auth login --username AWS --password-stdin 709825985650.dkr.ecr.us-east-1.amazonaws.com

❯ crane copy gcr.io/datadoghq/operator:1.0.3 709825985650.dkr.ecr.us-east-1.amazonaws.com/datadog/operator:1.0.3
```

To validate, describe the repository
```sh
aws ecr describe-images --registry-id 709825985650 --region us-east-1  --repository-name datadog/operator
..
        {
            "registryId": "709825985650",
            "repositoryName": "datadog/operator",
            "imageDigest": "sha256:e7ad530ca73db7324186249239dec25556b4d60d85fa9ba0374dd2d0468795b3",
            "imageTags": [
                "1.0.3"
            ],
..
```

[eks-helm-push]: https://docs.aws.amazon.com/AmazonECR/latest/userguide/push-oci-artifact.html
