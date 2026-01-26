# Default variables
SHELL = /usr/bin/env bash -o pipefail
GOTESTSUM_FORMAT?=standard-verbose

# E2E environment variables
E2E_CONFIG_PARAMS?=
E2E_KEY_PAIR_NAME=ci.helm-charts
DD_TEAM?=container-ecosystems
DD_TAGS?=
E2E_BUILD_TAGS?="e2e e2e_autopilot e2e_autopilot_systemprobe e2e_autopilot_csi"

## Local profile
E2E_PROFILE?=local
export E2E_KEY_PAIR_NAME?=${USER}
export E2E_API_KEY?=
export E2E_APP_KEY?=
export PULUMI_CONFIG_PASSPHRASE?=

## CI profile
CI_ENV_NAMES?=aws/agent-qa

ifdef ${CI_PIPELINE_ID}
override E2E_PROFILE=ci
endif
ifdef ${CI_PROJECT_ID}
override E2E_PROFILE=ci
endif

ifeq ($(E2E_PROFILE), ci)
export E2E_KEY_PAIR_NAME
export CI_ENV_NAMES
export DD_TEAM
export DD_TAGS
endif

.PHONY: all
all: fmt vet test test-e2e

.PHONY: test
test: fmt vet unit-test integration-test

.PHONY: fmt
fmt:
	go fmt -C test ./...

.PHONY: vet
vet:
	go vet -C test ./...

.PHONY: unit-test
unit-test: unit-test-datadog unit-test-operator unit-test-private-action-runner unit-test-datadog-csi-driver

.PHONY: unit-test-datadog
unit-test-datadog:
	helm dependency update ./charts/datadog 2>/dev/null
	go test -C test ./datadog -count=1

.PHONY: unit-test-operator
unit-test-operator:
	helm dependency update ./charts/datadog-operator 2>/dev/null
	go test -C test ./datadog-operator -count=1

.PHONY: unit-test-datadog-csi-driver
unit-test-datadog-csi-driver:
	helm dependency update ./charts/datadog-csi-driver 2>/dev/null
	go test -C test ./datadog-csi-driver -count=1

.PHONY: unit-test-private-action-runner
unit-test-private-action-runner:
	go test -C test ./private-action-runner -count=1

.PHONY: update-test-baselines
update-test-baselines: update-test-baselines-datadog-agent update-test-baselines-operator update-test-baselines-private-action-runner update-test-baselines-datadog-csi-driver

.PHONY: update-test-baselines-private-action-runner
update-test-baselines-private-action-runner:
	go test -C test ./private-action-runner -count=1 -args -updateBaselines=true

.PHONY: update-test-baselines-operator
update-test-baselines-operator:
	helm dependency update ./charts/datadog-operator 2>/dev/null
	go test -C test ./datadog-operator -count=1 -args -updateBaselines=true

.PHONY: update-test-baselines-datadog-agent
update-test-baselines-datadog-agent:
	helm dependency update ./charts/datadog 2>/dev/null
	go test -C test ./datadog -count=1 -args -updateBaselines=true

.PHONY: update-test-baselines-datadog-csi-driver
update-test-baselines-datadog-csi-driver:
	helm dependency update ./charts/datadog-csi-driver 2>/dev/null
	go test -C test ./datadog-csi-driver -count=1 -args -updateBaselines=true

.PHONY: integration-test
integration-test:
	go test -C test/integ --tags=integration -count=1 -v

# Running E2E tests locally:
## Must be connected to appgate
## E2E make target commands must be prepended with `aws-vault exec sso-agent-sandbox-account-admin --`

# aws-vault exec sso-agent-sandbox-account-admin -- make test-e2e
.PHONY: test-e2e
test-e2e: fmt vet e2e-test

# aws-vault exec sso-agent-sandbox-account-admin -- make e2e-test
.PHONY: e2e-test
e2e-test:
	E2E_CONFIG_PARAMS=$(E2E_CONFIG_PARAMS) E2E_PROFILE=$(E2E_PROFILE) E2E_AGENT_VERSION=$(E2E_AGENT_VERSION) go test -C test/e2e ./... --tags=$(E2E_BUILD_TAGS) -v -vet=off -timeout 1h -count=1
