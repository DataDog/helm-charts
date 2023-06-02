.PHONY: all
all: fmt vet test

.PHONY: test-e2e-ci
test-e2e-ci:
	bash ./.gitlab/run_e2e_tests.sh -p ci

.PHONY: test-e2e
test-e2e:
	bash ./.gitlab/run_e2e_tests.sh

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
	go test -C test ./... -count=1

.PHONY: update-test-baselines
update-test-baselines:
	go test -C test ./... -count=1 -args -updateBaselines=true
