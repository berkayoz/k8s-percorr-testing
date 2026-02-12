package tests

import (
	"context"
	"flag"
	"fmt"

	"github.com/canonical/k8s-percorr-testing/pkg/run"
	. "github.com/onsi/ginkgo/v2"
)

var r = run.New(GinkgoWriter)

// CLI flags.
var (
	bgLoad        bool
	bgCPU         string
	bgMemory      string
	bgRPS         int
	bgPayloadSize int
)

func init() {
	flag.BoolVar(&bgLoad, "bg-load", true, "Enable background load deployment")
	flag.StringVar(&bgCPU, "bg-cpu", "1", "CPU cores for compute stressor")
	flag.StringVar(&bgMemory, "bg-memory", "2Gi", "Memory for compute stressor")
	flag.IntVar(&bgRPS, "bg-rps", 100, "Requests per second for network load")
	flag.IntVar(&bgPayloadSize, "bg-payload-size", 125000, "Payload size in bytes for network load")
}

// Background load constants.
const (
	bgReleaseName = "bg-load"
	bgNamespace   = "bg-load"
	bgChartSubdir = "manifests/k8s-bg-load"
)

// --- Helm helpers ---

func helmInstallBgLoad(ctx context.Context, chartDir string) error {
	return r.Cmd(ctx, "helm",
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
	)
}

