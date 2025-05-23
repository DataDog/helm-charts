# Datadog changelog

## 1.1.2

* Add customizable resource limits and requests for the private action runner container

## 1.1.1

* Bump runner version to `v1.3.0`

## 1.1.0

* Add the `$schema` key to the `values.yaml` file to enable schema validation in IDEs.

## 1.0.3

* Allow a `global` object in values so this chart can be used in a subchart.

## 1.0.2

* Update private action runner version to `v1.2.0`
  * Bugfix: `HTTP_PROXY`, `HTTPS_PROXY` and `NO_PROXY` are now honored for all http requests from the runner
  * Feat: more flexible credentials loading.

## 1.0.1

* Improve Readme

## 1.0.0

* BREAKING CHANGES: Updates the chart for simplification and better following of Helm best practices. See [UPGRADING.md](UPGRADING.md) for more details.

## 0.20.1

* Various cleanup for the chart.

## 0.20.0

* Add the ability to specify kubernetes secrets to store credential files.

## 0.19.0

* Use a role instead of a cluster role for the runner's service account by default.

## 0.18.0

* Add the ability to specify a kubernetes secret to store the runner's identity.

## 0.17.2

* Update postgresql credentials file example

## 0.17.1

* Update private action image version to `v1.1.1`

## 0.17.0

* Update private action image version to `v1.0.0`

## 0.16.0

* Add support for passing environment variables to the Datadog Private Action Runner container.

## 0.15.8

* Update private action image version to `v0.1.14-beta`

## 0.15.7

* Update private action image version to `v0.1.12-beta`

## 0.15.6

* Update private action image version to `v0.1.11-beta`

## 0.15.5

* Add gitlab credentials file example

## 0.15.4

* Update private action image version to `v0.1.10-beta`

## 0.15.3

* Update private action image version to `v0.1.9-beta`

## 0.15.2

* Update private action image version to `v0.1.8-beta`

## 0.15.1

* Update private action image version to `v0.1.6-beta`

## 0.15.0

* Update private action image version to `v0.1.5-beta`

## 0.14.3

* Add GitLab private actions and fix image repository link.

## 0.14.2

* Update private action image version to `v0.1.3-beta`

## 0.14.1

* Update private action image version to `v0.1.2-beta`

## 0.14.0

* Add support for `kubernetesActions`.

## 0.13.0

* Update private action image version to `v0.1.1-beta`

## 0.12.0

* Introduced `credentialFiles` key in `values.yaml` for secret management. Deprecated the `connectionCredentials` key
* Fixed issue where specifying connection secrets under `connectionCredentials` can result in the Helm chart generating malformed JSON

## 0.11.0

* Added top level `port` configuration option, superseding `appBuilder.port`. Update the private action image to the beta image, `v0.1.0-beta`.

### 0.10.0

* Update private action image version to `v0.0.1-alpha31`.

### 0.9.1

* Added ability to configure connection credentials in `config.yaml`.

### 0.9.0

* Update private action image version to `v0.0.1-alpha29`.

### 0.8.1

* Minor tweaks to YAML formatting in the runner configuration

### 0.8.0

* Send MANAGED_BY environment variable to container. Update private action image version to `v0.0.1-alpha28`.

### 0.7.0

* Simplify README instructions to reflect the new Kubernetes UI. Split image value to be consistent with other charts. Fix bug requiring port for Workflow mode.

### 0.6.0

* Update private action image version to `v0.0.1-alpha27`.

### 0.5.0

* Update private action image version to `v0.0.1-alpha26`.

### 0.4.0

* Revert private action image version to `v0.0.1-alpha24`, apply patch to fix labels in `deployments.yaml`, and add newlines to end of all yaml files.

### 0.3.0

* Update private action image version to `v0.0.1-alpha25`.

### 0.2.0

* Update private action image version to `v0.0.1-alpha24` and add port to example config.

### 0.1.0

* Initial version
