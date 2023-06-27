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

### Prerequisites

Tests have been validated using:
* Go v1.20
* Helm v3.10.1

They may work with older versions, though.

### Running the Tests
Go sources are located under the `test` directory.

#### Unit Tests

```shell
 make test
 ```

#### End-to-End Tests
The helm-charts end-to-end (E2E) tests run on [Pulumi](https://www.pulumi.com/)-deployed test infrastructures, defined as "stacks". The test infrastructures are deployed using the [`test-infra-definitions`](https://github.com/DataDog/test-infra-definitions) and [`datadog-agent`](https://github.com/DataDog/datadog-agent/tree/main/test/new-e2e) E2E frameworks.

**Prerequisites**
Internal Datadog users may run E2E locally with the following prerequisites:

* Access to the AWS `agent-sandbox` account
* AWS keypair with your public ssh key created in the `agent-sandbox` account
* Completed steps 1-4 of the `test-infra-definitions` [Quick start guide](https://github.com/DataDog/test-infra-definitions#quick-start-guide)
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
