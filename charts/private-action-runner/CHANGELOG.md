# Datadog changelog

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
