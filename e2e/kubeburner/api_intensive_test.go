package kubeburner

import (
	"os"
	"path/filepath"

	kbreport "github.com/canonical/k8s-percorr-testing/internal/report/kubeburner"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
)

var _ = ReportAfterSuite("kube-burner markdown report", func(report types.Report) {
	var data *kbreport.ReportData
	for _, spec := range report.SpecReports {
		for _, entry := range spec.ReportEntries {
			if entry.Name == "kubeburner-metrics" {
				v := entry.GetRawValue().(kbreport.ReportData)
				data = &v
			}
		}
	}
	if data == nil {
		return
	}

	// Ginkgo resolves --output-dir into the JSONReport path before the
	// test binary runs, so filepath.Dir gives us the same output directory
	// used for JSON/JUnit reports. When no --json-report flag is set the
	// field is empty and filepath.Dir returns ".", writing to cwd.
	_, rc := GinkgoConfiguration()
	reportPath := filepath.Join(filepath.Dir(rc.JSONReport), "api-intensive-report.md")
	if err := kbreport.GenerateToFile(data, apiIntensiveConfig, reportPath); err != nil {
		GinkgoWriter.Printf("Failed to generate kube-burner report: %v\n", err)
		return
	}
	GinkgoWriter.Printf("Report written to %s\n", reportPath)
})

var _ = Describe("API Intensive", func() {
	It("should complete the kube-burner api-intensive workload", func(ctx SpecContext) {
		workDir, err := filepath.Abs(manifestsSubdir)
		Expect(err).NotTo(HaveOccurred())

		DeferCleanup(func(ctx SpecContext) {
			err := destroyKubeBurner(ctx, workDir, apiIntensiveConfig)
			Expect(err).NotTo(HaveOccurred())
		})

		metricsDir, err := os.MkdirTemp("", "kubeburner-metrics-*")
		Expect(err).NotTo(HaveOccurred())

		err = runKubeBurner(ctx, workDir, apiIntensiveConfig, metricsDir)
		Expect(err).NotTo(HaveOccurred())

		data, err := kbreport.Collect(metricsDir)
		Expect(err).NotTo(HaveOccurred())
		AddReportEntry("kubeburner-metrics", *data)
	})
})
