MODULE?=github.com/canonical/k8s-percorr-testing

.PHONY: test test-no-bg test-kube-burner test-chaos test-all fmt

test:
	ginkgo -v --timeout 30m ./tests

test-no-bg:
	ginkgo -v --timeout 30m ./tests -- --bg-load=false

test-kube-burner:
	ginkgo -v --timeout 30m ./tests/kubeburner

test-chaos:
	ginkgo -v --timeout 6h ./tests/chaos

test-all:
	ginkgo -v --timeout 6h ./tests ./tests/kubeburner ./tests/chaos

fmt:
	gofmt -w .
