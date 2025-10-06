# List of things to change in the next breaking changes release


## useSeparateSecretForCredentials

`runner.useSeparateSecretForCredentials` should be true by default, we want the files layout to match the default values suggested in the UI.
So runner config should be in `/etc/dd-action-runner/config/config.yaml` and credentials in `/etc/dd-action-runner/config/credentials/` subdirectory.