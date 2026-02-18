MODULE?=github.com/canonical/k8s-percorr-testing

.PHONY: test test-no-bg test-kube-burner test-chaos test-conformance test-all fmt

test:
	ginkgo -v --timeout 30m ./e2e

test-no-bg:
	ginkgo -v --timeout 30m ./e2e -- --bg-load=false

test-kube-burner:
	ginkgo -v --timeout 30m ./e2e/kubeburner

test-chaos:
	ginkgo -v --timeout 6h ./e2e/chaos

test-conformance:
	ginkgo -v --timeout 3h ./e2e/conformance

test-all:
	ginkgo -v --timeout 9h ./e2e ./e2e/kubeburner ./e2e/chaos ./e2e/conformance

fmt:
	gofmt -w .
