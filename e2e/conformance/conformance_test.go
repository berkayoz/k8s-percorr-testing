package conformance

import (
	"os"
	"path/filepath"
	"strings"

	conformancereport "github.com/canonical/k8s-percorr-testing/internal/report/conformance"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
	"github.com/vmware-tanzu/sonobuoy/pkg/client/results"
)

var tarballDir string

var _ = BeforeSuite(func(ctx SpecContext) {
	By("verifying sonobuoy CLI")
	err := r.Cmd(ctx, "sonobuoy", "version")
	Expect(err).NotTo(HaveOccurred(), "sonobuoy CLI not found")

	tarballDir, err = os.MkdirTemp("", "sonobuoy-*")
	Expect(err).NotTo(HaveOccurred(), "failed to create temp directory for sonobuoy tarball")
})

var _ = AfterSuite(func(ctx SpecContext) {
	By("cleaning up sonobuoy resources")
	err := sonobuoyDelete(ctx)
	Expect(err).NotTo(HaveOccurred())
})

var _ = ReportAfterSuite("conformance markdown report", func(report types.Report) {
	var item *results.Item
	for _, spec := range report.SpecReports {
		for _, entry := range spec.ReportEntries {
			if entry.Name == "conformance-result" {
				v := entry.GetRawValue().(results.Item)
				item = &v
			}
		}
	}
	if item == nil {
		return
	}

	// Ginkgo resolves --output-dir into the JSONReport path before the
	// test binary runs, so filepath.Dir gives us the same output directory
	// used for JSON/JUnit reports. When no --json-report flag is set the
	// field is empty and filepath.Dir returns ".", writing to cwd.
	_, rc := GinkgoConfiguration()
	reportPath := filepath.Join(filepath.Dir(rc.JSONReport), "conformance-report.md")
	if err := conformancereport.GenerateToFile(item, reportPath); err != nil {
		GinkgoWriter.Printf("Failed to generate conformance report: %v\n", err)
		return
	}

	GinkgoWriter.Printf("Conformance report written to %s\n", reportPath)
})

var _ = Describe("CNCF Conformance", Ordered, Serial, func() {
	It("should run sonobuoy conformance tests", func(ctx SpecContext) {
		By("running sonobuoy certified-conformance mode")
		err := sonobuoyRun(ctx)
		Expect(err).NotTo(HaveOccurred(), "sonobuoy run failed")
	})

	It("should retrieve and validate results", func(ctx SpecContext) {
		By("retrieving sonobuoy results")
		tarballPath, err := sonobuoyRetrieve(ctx, tarballDir)
		Expect(err).NotTo(HaveOccurred(), "sonobuoy retrieve failed")

		tarball := strings.TrimSpace(string(tarballPath))
		Expect(tarball).NotTo(BeEmpty(), "sonobuoy retrieve returned empty path")

		By("parsing sonobuoy dump results")
		dump, err := sonobuoyDumpResults(ctx, tarball)
		Expect(err).NotTo(HaveOccurred(), "sonobuoy dump results failed")

		result, err := conformancereport.ParseDumpResults(dump)
		Expect(err).NotTo(HaveOccurred(), "parsing dump results failed")

		Expect(result.Status).NotTo(Equal(results.StatusFailed),
			"conformance tests reported failures")

		AddReportEntry("conformance-result", *result)
	})
})
