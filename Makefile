MODULE?=github.com/canonical/k8s-percorr-testing

.PHONY: test fmt

test:
	go test ./tests -v

fmt:
	gofmt -w .
