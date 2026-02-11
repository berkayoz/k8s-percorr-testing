package tests

import (
	"fmt"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
)

const (
	bgReleaseName = "bg-load"
	bgNamespace   = "bg-load"
	bgChartSubdir = "manifests/k8s-bg-load"
)

func helmInstallBgLoad(chartDir string) error {
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
	cmd := exec.Command("helm", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("helm install failed: %w\nOutput: %s", err, string(output))
	}
	GinkgoWriter.Printf("Background load deployed:\n%s\n", string(output))
	return nil
}

func helmUninstallBgLoad() error {
	cmd := exec.Command("helm", "uninstall", bgReleaseName,
		"--namespace", bgNamespace,
		"--wait",
		"--timeout", "2m",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("helm uninstall failed: %w\nOutput: %s", err, string(output))
	}
	GinkgoWriter.Printf("Background load removed:\n%s\n", string(output))
	return nil
}
