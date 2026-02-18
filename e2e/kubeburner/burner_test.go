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

func runKubeBurner(ctx context.Context, workDir, configFile string) error {
	return r.CmdWithDir(ctx, workDir, "kube-burner", "init", "--skip-log-file", "-c", configFile)
}
