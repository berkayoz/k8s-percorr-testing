# k8s-percorr-testing
PerCorr testing suite for Canonical Kubernetes

Repository to run Ginkgo-based tests against a Kubernetes cluster.

Quickstart

- Install Go (1.25+).
- Ensure `KUBECONFIG` env var points to a cluster (or run in-cluster).
- Run the tests:

```bash
make test
```

Tests live under the `tests/` package and use `github.com/onsi/ginkgo/v2`.
