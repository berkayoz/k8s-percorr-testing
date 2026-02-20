package conformance

import (
	"context"

	"github.com/canonical/k8s-percorr-testing/internal/run"
	. "github.com/onsi/ginkgo/v2"
)

var r = run.New(GinkgoWriter)

func sonobuoyRun(ctx context.Context) error {
	return r.Cmd(ctx, "sonobuoy", "run", "--mode=quick", "--wait")
}

func sonobuoyRetrieve(ctx context.Context, destDir string) ([]byte, error) {
	return r.CmdOutput(ctx, "sonobuoy", "retrieve", destDir)
}

func sonobuoyDumpResults(ctx context.Context, tarball string) ([]byte, error) {
	return r.CmdOutput(ctx, "sonobuoy", "results", tarball, "--mode", "dump", "--plugin", "e2e")
}

func sonobuoyDelete(ctx context.Context) error {
	return r.Cmd(ctx, "sonobuoy", "delete", "--wait")
}
