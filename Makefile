# default variables
SHELL = /usr/bin/env bash -o pipefail
GOTESTSUM_FORMAT?=pkgname-and-test-fails

## Local profile
export AWS_KEYPAIR_NAME?=${USER}
export E2E_API_KEY?=
export E2E_APP_KEY?=
export PULUMI_CONFIG_PASSPHRASE?=

## CI profile
E2E_PROFILE?=local
E2E_CONFIG_PARAMS?=
CI_ENV_NAMES?=aws/agent-qa
DD_TEAM?=container-ecosystems
DD_TAGS?=

ifdef ${CI_PIPELINE_ID}
override E2E_PROFILE=ci
endif
ifdef ${CI_PROJECT_ID}
override E2E_PROFILE=ci
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
test-e2e: e2e-test
# TODO add `fmt` and `vet` targets when test-infra-definitions gitlab runner image bumps GO version to 1.20

.PHONY: e2e-test
e2e-test:
ifeq ($(E2E_PROFILE), ci)
	E2E_PROFILE=$(E2E_PROFILE)
	E2E_CONFIG_PARAMS=$(E2E_CONFIG_PARAMS)
	CI_ENV_NAMES=$(CI_ENV_NAMES)
	DD_TEAM=$(DD_TEAM)
	DD_TAGS=$(DD_TAGS)
endif
	GOTESTSUM_FORMAT=$(GOTESTSUM_FORMAT) gotestsum --packages=./test/... --format-hide-empty-pkg -- -run=E2E -vet=off -timeout 1h -count=1

.PHONY: e2e-test-preserve-stacks
e2e-test-preserve-stacks:
ifeq ($(E2E_PROFILE), ci)
	E2E_PROFILE=$(E2E_PROFILE)
	E2E_CONFIG_PARAMS=$(E2E_CONFIG_PARAMS)
	CI_ENV_NAMES=$(CI_ENV_NAMES)
	DD_TEAM=$(DD_TEAM)
	DD_TAGS=$(DD_TAGS)
endif
	GOTESTSUM_FORMAT=$(GOTESTSUM_FORMAT) gotestsum --packages=./test/... --format-hide-empty-pkg -- -run=E2E -vet=off -timeout 1h -count=1 -args -preserveStacks=true
