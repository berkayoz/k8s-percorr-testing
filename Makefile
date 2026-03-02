MODULE?=github.com/canonical/k8s-percorr-testing

BG_CPU ?= 1
BG_MEMORY ?= 2Gi
BG_RPS ?= 100
BG_PAYLOAD_SIZE ?= 125000

.PHONY: test-kube-burner test-chaos test-conformance test-all fmt deploy-bg-load remove-bg-load

test-kube-burner:
	ginkgo -v --timeout 30m \
		--json-report=kubeburner.json --junit-report=kubeburner-junit.xml \
		--output-dir=reports \
		./e2e/kubeburner

test-chaos:
	ginkgo -v --timeout 6h \
		--json-report=chaos.json --junit-report=chaos-junit.xml \
		--output-dir=reports \
		./e2e/chaos

test-conformance:
	ginkgo -v --timeout 3h \
		--json-report=conformance.json --junit-report=conformance-junit.xml \
		--output-dir=reports \
		./e2e/conformance

test-all:
	ginkgo -v --timeout 9h \
		--json-report=all.json --junit-report=all-junit.xml \
		--output-dir=reports \
		./e2e/kubeburner ./e2e/chaos ./e2e/conformance

fmt:
	gofmt -w .

deploy-bg-load:
	helm upgrade --install bg-load e2e/testdata/k8s-bg-load \
		--namespace bg-load --create-namespace --wait --timeout 5m \
		--set compute.cpu=$(BG_CPU) \
		--set compute.memory=$(BG_MEMORY) \
		--set network.rps=$(BG_RPS) \
		--set network.payloadSize=$(BG_PAYLOAD_SIZE)

remove-bg-load:
	helm uninstall bg-load --namespace bg-load --wait --timeout 2m
	kubectl delete namespace bg-load --ignore-not-found
