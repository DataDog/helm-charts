all: docs

docs:
	./.github/helm-docs.sh

.PHONY: test-e2e-ci
test-e2e-ci:
	bash ./.gitlab/run_e2e_tests.sh -p ci


.PHONY: test-e2e
test-e2e:
	bash ./.gitlab/run_e2e_tests.sh
