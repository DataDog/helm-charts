# Default variables
SHELL = /usr/bin/env bash -o pipefail
GOTESTSUM_FORMAT?=standard-verbose

# E2E environment variables
E2E_CONFIG_PARAMS?=
DD_TEAM?=container-ecosystems
DD_TAGS?=

## Local profile
E2E_PROFILE?=local
export AWS_KEYPAIR_NAME?=${USER}
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
export CI_ENV_NAMES
export DD_TEAM
export DD_TAGS
endif

.PHONY: all
all: fmt vet test test-e2e

.PHONY: test
test: fmt vet unit-test

.PHONY: fmt
fmt:
	go fmt -C test ./...

.PHONY: vet
vet:
	go vet -C test ./...

.PHONY: unit-test
unit-test:
	go test -C test ./... -count=1 -skip=E2E

.PHONY: update-test-baselines
update-test-baselines:
	go test -C test ./... -count=1 -skip=E2E -args -updateBaselines=true

.PHONY: test-e2e
test-e2e: fmt vet e2e-test

.PHONY: e2e-test
e2e-test:
	E2E_CONFIG_PARAMS=$(E2E_CONFIG_PARAMS) E2E_PROFILE=$(E2E_PROFILE) gotestsum --packages=./test/... --format standard-verbose --format-hide-empty-pkg -- -run=E2E -vet=off -timeout 1h -count=1

.PHONY: e2e-test-preserve-stacks
e2e-test-preserve-stacks:
	E2E_CONFIG_PARAMS=$(E2E_CONFIG_PARAMS) E2E_PROFILE=$(E2E_PROFILE) gotestsum --packages=./test/... --format standard-verbose --format-hide-empty-pkg -- -run=E2E -vet=off -timeout 1h -count=1 -args -preserveStacks=true

.PHONY: e2e-test-cleanup-stacks
e2e-test-cleanup-stacks:
	E2E_CONFIG_PARAMS=$(E2E_CONFIG_PARAMS) E2E_PROFILE=$(E2E_PROFILE) gotestsum --packages=./test/... --format standard-verbose --format-hide-empty-pkg -- -run=E2E -vet=off -timeout 1h -count=1 -args -destroyStacks=true
