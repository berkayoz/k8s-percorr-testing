package tests

import (
	"context"
	"fmt"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
)

const (
	bgReleaseName = "bg-load"
	bgNamespace   = "bg-load"
	bgChartSubdir = "manifests/k8s-bg-load"
)

func helmInstallBgLoad(ctx context.Context, chartDir string) error {
	args := []string{
		"upgrade", "--install",
		bgReleaseName,
		chartDir,
		"--namespace", bgNamespace,
		"--create-namespace",
		"--wait",
		"--timeout", "5m",
		"--set", fmt.Sprintf("compute.cpu=%s", bgCPU),
		"--set", fmt.Sprintf("compute.memory=%s", bgMemory),
		"--set", fmt.Sprintf("network.rps=%d", bgRPS),
		"--set", fmt.Sprintf("network.payloadSize=%d", bgPayloadSize),
	}
	cmd := exec.CommandContext(ctx, "helm", args...)
	cmd.Stdout = GinkgoWriter
	cmd.Stderr = GinkgoWriter
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("helm install failed: %w", err)
	}
	return nil
}

func helmUninstallBgLoad(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "helm", "uninstall", bgReleaseName,
		"--namespace", bgNamespace,
		"--wait",
		"--timeout", "2m",
	)
	cmd.Stdout = GinkgoWriter
	cmd.Stderr = GinkgoWriter
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("helm uninstall failed: %w", err)
	}
	return nil
}
