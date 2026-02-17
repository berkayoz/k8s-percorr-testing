package conformance

import (
	"context"
	"flag"
	"os"

	"github.com/canonical/k8s-percorr-testing/pkg/run"
	. "github.com/onsi/ginkgo/v2"
)

var r = run.New(GinkgoWriter)

var resultsDir string

func init() {
	flag.StringVar(&resultsDir, "results-dir", "", "Directory to store sonobuoy results (defaults to a temp dir)")
}

func sonobuoyRun(ctx context.Context) error {
	return r.Cmd(ctx, "sonobuoy", "run", "--mode=certified-conformance", "--wait")
}

func sonobuoyRetrieve(ctx context.Context, destDir string) ([]byte, error) {
	return r.CmdOutput(ctx, "sonobuoy", "retrieve", destDir)
}

func sonobuoyResults(ctx context.Context, tarball string) ([]byte, error) {
	return r.CmdOutput(ctx, "sonobuoy", "results", tarball)
}

func sonobuoyDelete(ctx context.Context) {
	if err := r.Cmd(ctx, "sonobuoy", "delete", "--wait"); err != nil {
		GinkgoWriter.Printf("WARNING: Failed to delete sonobuoy resources: %v\n", err)
	}
}

func getResultsDir() (string, error) {
	if resultsDir != "" {
		return resultsDir, nil
	}
	return os.MkdirTemp("", "sonobuoy-results-*")
}
