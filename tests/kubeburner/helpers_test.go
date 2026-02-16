package kubeburner

import (
	"context"

	"github.com/canonical/k8s-percorr-testing/pkg/run"
	. "github.com/onsi/ginkgo/v2"
)

var r = run.New(GinkgoWriter)

// Manifest constants.
const (
	manifestsSubdir    = "manifests"
	apiIntensiveConfig = "api-intensive.yml"
)

func runKubeBurner(ctx context.Context, workDir, configFile string) error {
	return r.CmdWithDir(ctx, workDir, "kube-burner", "init", "--skip-log-file", "-c", configFile)
}
