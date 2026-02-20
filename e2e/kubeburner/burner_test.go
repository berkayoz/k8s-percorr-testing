package kubeburner

import (
	"context"

	"github.com/canonical/k8s-percorr-testing/internal/run"
	. "github.com/onsi/ginkgo/v2"
)

var r = run.New(GinkgoWriter)

// Manifest constants.
const (
	manifestsSubdir    = "testdata"
	apiIntensiveConfig = "api-intensive.yaml"
)

func runKubeBurner(ctx context.Context, workDir, configFile, metricsDir string) error {
	return r.CmdWithDir(ctx, workDir, "kube-burner", "init", "--skip-log-file",
		"-c", configFile, "--set", "metricsEndpoints.0.indexer.metricsDirectory="+metricsDir)
}

func destroyKubeBurner(ctx context.Context, workDir, configFile string) error {
	return r.CmdWithDir(ctx, workDir, "kube-burner", "destroy", "--skip-log-file", "-c", configFile)
}

