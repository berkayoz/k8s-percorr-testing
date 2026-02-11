MODULE?=github.com/canonical/k8s-percorr-testing

.PHONY: test test-no-bg fmt

test:
	go test ./tests -v -timeout 30m

test-no-bg:
	go test ./tests -v -timeout 30m -- --bg-load=false

fmt:
	gofmt -w .
