package kubeburner

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/canonical/k8s-percorr-testing/internal/report"
	kbreport "github.com/canonical/k8s-percorr-testing/internal/report/kubeburner"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	metricsDir string
	reportsDir string
)

var _ = BeforeSuite(func() {
	var err error
	reportsDir, err = report.Dir()
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	if metricsDir != "" {
		reportPath := filepath.Join(reportsDir, "api-intensive-report.md")
		err := kbreport.GenerateToFile(metricsDir, apiIntensiveConfig, reportPath)
		Expect(err).NotTo(HaveOccurred())

		fmt.Fprintf(GinkgoWriter, "Report written to %s\n", reportPath)
	}
})

var _ = Describe("API Intensive", func() {
	It("should complete the kube-burner api-intensive workload", func(ctx SpecContext) {
		workDir, err := filepath.Abs(manifestsSubdir)
		Expect(err).NotTo(HaveOccurred())

		DeferCleanup(func(ctx SpecContext) {
			err := destroyKubeBurner(ctx, workDir, apiIntensiveConfig)
			Expect(err).NotTo(HaveOccurred())
		})

		metricsDir, err = os.MkdirTemp("", "kubeburner-metrics-*")
		Expect(err).NotTo(HaveOccurred())

		err = runKubeBurner(ctx, workDir, apiIntensiveConfig, metricsDir)
		Expect(err).NotTo(HaveOccurred())
	})
})
