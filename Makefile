MODULE?=github.com/canonical/k8s-percorr-testing

.PHONY: test test-no-bg fmt

test:
	ginkgo -v --timeout 30m ./tests

test-no-bg:
	ginkgo -v --timeout 30m ./tests -- --bg-load=false

fmt:
	gofmt -w .
