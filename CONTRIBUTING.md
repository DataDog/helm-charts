# Contributing

All contributions improving our Helm charts are welcome. If you'd like to contribute a bug fix or a feature, you can directly open a pull request with your changes.

We aim to follow high quality standards, thus your PR must follow some rules:

- Make sure any new parameter is documented
- Make sure the chart version has been bumped in the corresponding chart's `Chart.yaml`.
- Make sure to describe your change in the corresponding chart's `CHANGELOG.md`.
- Make sure any new feature is tested by modifying or adding a file in `ci/`
- Make sure your changes are compatible (or protected) with older Kubernetes version (CI will validate this down to 1.14)
- Make sure you updated documentation (after bumping `Chart.yaml`) by running `.github/helm-docs.sh`

Our team will then happily review and merge contributions!

## Go Tests

Go tests ensure quality and correctness of our Helm charts. These tests are intended to validate charts and catch any potential issues early in the development process.

These tests run as part of the CI workflow. They can be used locally, during development as well.

We have three major groups of tests
* Unit tests - these are lightweight tests utilizing Helm to verify:
  * Error-free rendering of the templates.
  * Correctness of specific values in the rendered manifests.
  * Rendered manifests against baselines saved in the repo.
* Integration tests - these test run against cluster in the local Kubernetes context or Kind cluster in the CI.
  * Tests install one or multiple charts and assert that certain resources reach expected state.
* End-to-End test - these tests target cloud infrastructure deployed by [Pulumi][pulumi].

### Prerequisites

Tests have been validated using:
* Go v1.20
* Helm v3.10.1

They may work with older versions, though.

### Running the Tests
Go sources are located under the `test` directory.

#### Unit Tests
To run unit tests.

```shell
 make unit-test
 ```

For changes which require baseline file update run `make update-test-baselines`. This will update all baseline files which should be included in the PR and pushed upstream.

#### Integration Tests
Integration tests run against locally configured context. We use [Terratest][terratest] for interacting with Helm and Kubectl.

Each test creates a unique namespace and subsequent resources are created in this namespace. Clean-up upon test completion is best effort, so it's recommended to run test against disposable cluster. **Make sure you don't accidentally run the test against a production cluster.**

**Prerequisites**
* Kubeconfig context targeting test cluster. Local and CI tests have been tested using Kind cluster.
* Environment Variables:
  * `APP_KEY`
  * `API_KEY`
  * `K8S_VERSION` e.g. "v1.24"

Use below `make` targets to run integration tests or integration and unit tests together respectively.

```shell
make integration-test
make test
```
You can run tests from IDE too (tested with VScode) as long as the environment variables are configured properly.

#### End-to-End Tests
The helm-charts end-to-end (E2E) tests run on [Pulumi][pulumi]-deployed test infrastructures, defined as "stacks". The test infrastructures are deployed using the [`test-infra-definitions`][test-infra-repo] and [`datadog-agent`][agent-e2e-source] E2E frameworks.

**Prerequisites**
Internal Datadog users may run E2E locally with the following prerequisites:

* Access to the AWS `agent-sandbox` account
* AWS keypair with your public ssh key created in the `agent-sandbox` account
* Completed steps 1-4 of the `test-infra-definitions` [Quick start guide][test-infra-quickstart]
* Environment Variables:
  * AWS_KEYPAIR_NAME
  * E2E_API_KEY
  * E2E_APP_KEY
  * PULUMI_CONFIG_PASSPHRASE

To run E2E tests locally, run `aws-vault exec sso-agent-sandbox-account-admin -- make test-e2e`. This creates the E2E infrastructure stacks, runs tests in the infrastructure, and performs stack cleanup upon test completion.

```shell
aws-vault exec sso-agent-sandbox-account-admin -- make test-e2e
```

To keep an E2E Pulumi stack running upon test completion, run `make e2e-test-preserve-stacks`. This is useful for developing tests on Pulumi infrastructures that have a long startup time (such as AWS EKS).

```shell
aws-vault exec sso-agent-sandbox-account-admin -- make e2e-test-preserve-stacks
```

To clean up existing stacks, run:

```shell
aws-vault exec sso-agent-sandbox-account-admin -- make e2e-test-cleanup-stacks
```
 
## How to update a README file

In each chart, the `README.md` file is generated from the corresponding `README.md.gotmpl` and `values.yaml` files. Instead of modifying the `README.md` file directly:
1. Update either the `README.md.gotmpl` or `values.yaml` file.
1. Run `.github/helm-docs.sh` to update the README.


[go-ws]:https://go.dev/ref/mod#workspaces
[terratest]:https://github.com/gruntwork-io/terratest
[pulumi]:https://www.pulumi.com/
[test-infra-repo]:https://github.com/DataDog/test-infra-definitions
[agent-e2e-source]:https://github.com/DataDog/datadog-agent/tree/main/test/new-e2e
[test-infra-quickstart]:https://github.com/DataDog/test-infra-definitions#quick-start-guide