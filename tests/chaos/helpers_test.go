package chaos

import (
	"context"

	"github.com/canonical/k8s-percorr-testing/pkg/run"
	. "github.com/onsi/ginkgo/v2"
)

var r = run.New(GinkgoWriter)

// Chaos / Litmus constants.
const (
	chaosExperimentsSubdir  = "manifests/experiments"
	chaosNginxManifest      = "manifests/nginx.yaml"
	chaosSuperuserManifest  = "manifests/superuser.yaml"
	chaosNamespace          = "litmus"
	chaosReleaseName        = "litmus"
	chaosHelmRepo           = "https://litmuschaos.github.io/litmus-helm/"
	chaosHelmChart          = "litmuschaos/litmus-core"
	chaosExperimentsChart   = "litmuschaos/kubernetes-chaos"
	chaosExperimentsRelease = "litmus-experiments"
)

// --- Helm helpers ---

func helmInstallLitmus(ctx context.Context) error {
	if err := r.Cmd(ctx, "helm", "repo", "add", "litmuschaos", chaosHelmRepo); err != nil {
		return err
	}
	// Install litmus-core operator
	if err := r.Cmd(ctx, "helm", "install", chaosReleaseName, chaosHelmChart,
		"--namespace", chaosNamespace,
		"--create-namespace",
		"--wait",
		"--timeout", "5m",
	); err != nil {
		return err
	}
	// Install kubernetes-chaos experiments
	return r.Cmd(ctx, "helm", "install", chaosExperimentsRelease, chaosExperimentsChart,
		"--namespace", chaosNamespace,
		"--wait",
		"--timeout", "5m",
	)
}

func helmUninstallLitmus(ctx context.Context) {
	// Uninstall experiments chart first
	if err := r.Cmd(ctx, "helm", "uninstall", chaosExperimentsRelease,
		"--namespace", chaosNamespace, "--wait", "--timeout", "2m"); err != nil {
		GinkgoWriter.Printf("WARNING: Failed to uninstall experiments chart: %v\n", err)
	}
	// Then uninstall operator
	if err := r.Cmd(ctx, "helm", "uninstall", chaosReleaseName,
		"--namespace", chaosNamespace, "--wait", "--timeout", "2m"); err != nil {
		GinkgoWriter.Printf("WARNING: Failed to uninstall Litmus: %v\n", err)
	}
}
