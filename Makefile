MODULE?=github.com/canonical/k8s-percorr-testing

.PHONY: test test-no-bg test-kube-burner test-chaos test-conformance test-all fmt

test:
	ginkgo -v --timeout 30m ./tests

test-no-bg:
	ginkgo -v --timeout 30m ./tests -- --bg-load=false

test-kube-burner:
	ginkgo -v --timeout 30m ./tests/kubeburner

test-chaos:
	ginkgo -v --timeout 6h ./tests/chaos

test-conformance:
	ginkgo -v --timeout 3h ./tests/conformance

test-all:
	ginkgo -v --timeout 9h ./tests ./tests/kubeburner ./tests/chaos ./tests/conformance

fmt:
	gofmt -w .
