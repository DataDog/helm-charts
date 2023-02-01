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

## How to update README.md files content

In each chart, the `README.md` file is generated from the corresponding `README.md.gotmpl` and `values.yaml` files.
When the command `.github/helm-docs.sh`, the content of the README.md is updated.

So if the contribution requires an update in the README.md, the modification should be done either in the `README.md.gotmpl` or `values.yaml` files.
