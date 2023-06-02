.PHONY: all
all: fmt vet test

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

.PHONY: integration-tests
integration-tests:
	go test -C test/integ --tags=integration -count=1 -v

