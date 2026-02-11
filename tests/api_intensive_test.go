package tests

import (
	"context"
	"fmt"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
)

const (
	apiIntensiveSubdir = "../api-intensive"
	apiIntensiveConfig = "api-intensive.yml"
)

func runKubeBurner(ctx context.Context, workDir, configFile string) error {
	args := []string{
		"init",
		"-c", configFile,
	}
	cmd := exec.CommandContext(ctx, "kube-burner", args...)
	cmd.Dir = workDir
	cmd.Stdout = GinkgoWriter
	cmd.Stderr = GinkgoWriter
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("kube-burner failed: %w", err)
	}
	return nil
}
