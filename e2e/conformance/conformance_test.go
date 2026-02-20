package conformance

import (
	"path/filepath"
	"strings"

	"github.com/canonical/k8s-percorr-testing/internal/report"
	conformancereport "github.com/canonical/k8s-percorr-testing/internal/report/conformance"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/vmware-tanzu/sonobuoy/pkg/client/results"
)

var (
	destDir string
	tarball string
)

var _ = BeforeSuite(func(ctx SpecContext) {
	By("verifying sonobuoy CLI")
	err := r.Cmd(ctx, "sonobuoy", "version")
	Expect(err).NotTo(HaveOccurred(), "sonobuoy CLI not found")

	destDir, err = report.Dir()
	Expect(err).NotTo(HaveOccurred(), "failed to prepare results directory")
})

var _ = AfterSuite(func(ctx SpecContext) {
	if tarball != "" {
		By("generating conformance report")
		dump, err := sonobuoyDumpResults(ctx, tarball)
		Expect(err).NotTo(HaveOccurred(), "sonobuoy dump results failed")

		result, err := conformancereport.ParseDumpResults(dump)
		Expect(err).NotTo(HaveOccurred(), "parsing dump results failed")

		err = conformancereport.GenerateToFile(result, filepath.Join(destDir, "conformance-report.md"))
		Expect(err).NotTo(HaveOccurred(), "generating conformance report failed")
	}

	By("cleaning up sonobuoy resources")
	err := sonobuoyDelete(ctx)
	Expect(err).NotTo(HaveOccurred())
})

var _ = Describe("CNCF Conformance", Ordered, Serial, func() {
	It("should run sonobuoy conformance tests", func(ctx SpecContext) {
		By("running sonobuoy certified-conformance mode")
		err := sonobuoyRun(ctx)
		Expect(err).NotTo(HaveOccurred(), "sonobuoy run failed")
	})

	It("should retrieve and validate results", func(ctx SpecContext) {
		By("retrieving sonobuoy results")
		tarballPath, err := sonobuoyRetrieve(ctx, destDir)
		Expect(err).NotTo(HaveOccurred(), "sonobuoy retrieve failed")

		tarball = strings.TrimSpace(string(tarballPath))
		Expect(tarball).NotTo(BeEmpty(), "sonobuoy retrieve returned empty path")

		By("parsing sonobuoy dump results")
		dump, err := sonobuoyDumpResults(ctx, tarball)
		Expect(err).NotTo(HaveOccurred(), "sonobuoy dump results failed")

		result, err := conformancereport.ParseDumpResults(dump)
		Expect(err).NotTo(HaveOccurred(), "parsing dump results failed")

		Expect(result.Status).NotTo(Equal(results.StatusFailed),
			"conformance tests reported failures")
	})
})
