#### What this PR does / why we need it:

#### Which issue this PR fixes
*(optional, in `fixes #<issue number>(, fixes #<issue_number>, ...)` format, will close that issue when PR gets merged)*
  - fixes #

#### Special notes for your reviewer:

#### Checklist
[Place an '[x]' (no spaces) in all applicable fields. Please remove unrelated fields.]
- [ ] All commits are signed (see: [signing commits][1])
- [ ] Chart Version semver bump label has been added (use `<chartName>/minor-version`, `<chartName>/patch-version`, or `<chartName>/no-version-bump`)
- [ ] For `datadog` or `datadog-operator` chart or value changes, update the test baselines (run: `make update-test-baselines`)

GitHub CI takes care of the below, but are still required:
- [ ] Documentation has been updated with helm-docs (run: `.github/helm-docs.sh`)
- [ ] `CHANGELOG.md` has been updated 
- [ ] Variables are documented in the `README.md`

[1]: https://docs.github.com/en/authentication/managing-commit-signature-verification/signing-commits