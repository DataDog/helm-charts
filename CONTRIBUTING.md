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

We have two major groups of tests
* Unit tests - these are lightweight tests utilizing Helm to verify:
  * Error-free rendering of the templates.
  * Correctness of specific values in the rendered manifests.
  * Rendered manifests against baselines saved in the repo.
* Integration tests - these test run against cluster in the local Kubernetes context or Kind cluster in the CI.
  * Tests install one or multiple charts and assert that certain resources reach expected state.

### Prerequisites

Tests have been validated using:
* Go v1.20
* Helm v3.10.1

They may work with older versions, though.

### Running the Tests
Go sources are located under the `test` directory. The repository uses [Go workspace][go-ws], so tests can be run from the repository root using `make`.

#### Unit Tests
To run unit tests, run `make unit-test`.

For changes which require baseline file update run `make update-test-baselines`. This will update all baseline files which should be included in the PR and pushed upstream.

#### Integration Tests
Integration tests run against locally configured context. We use [Terratest][terratest] for interacting with Helm and Kubectl.

Each test creates unique namespace and subsequent resources are created in this namespace. Clean-up upon test completion is best effort, so it's recommended to run test against disposable cluster. **Make sure your don't accidentally run the test against production cluster.**

**Prerequisites**
* Kubeconfig context targeting test cluster. Local and CI tests have been tested using Kind cluster.
* Environment Variables:
  * `APP_KEY`
  * `API_KEY`
  * `K8S_VERSION` e.g. "v1.24"

To run tests, run `make integration-tests`. You can run tests from IDE too (tested with VScode) as long as the environment variables are configured properly.

## How to update a README file

In each chart, the `README.md` file is generated from the corresponding `README.md.gotmpl` and `values.yaml` files. Instead of modifying the `README.md` file directly:
1. Update either the `README.md.gotmpl` or `values.yaml` file.
1. Run `.github/helm-docs.sh` to update the README.


[go-ws]:https://go.dev/ref/mod#workspaces
[terratest]:https://github.com/gruntwork-io/terratest