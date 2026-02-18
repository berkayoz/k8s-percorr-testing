package conformance

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var destDir string

var _ = BeforeSuite(func(ctx SpecContext) {
	By("verifying sonobuoy CLI")
	err := r.Cmd(ctx, "sonobuoy", "version")
	Expect(err).NotTo(HaveOccurred(), "sonobuoy CLI not found")

	destDir, err = getResultsDir()
	Expect(err).NotTo(HaveOccurred(), "failed to prepare results directory")
})

var _ = AfterSuite(func(ctx SpecContext) {
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

		tarball := strings.TrimSpace(string(tarballPath))
		Expect(tarball).NotTo(BeEmpty(), "sonobuoy retrieve returned empty path")

		By("parsing sonobuoy results")
		summary, err := sonobuoyResults(ctx, tarball)
		Expect(err).NotTo(HaveOccurred(), "sonobuoy results failed")

		output := string(summary)
		Expect(output).NotTo(ContainSubstring("Status: failed"),
			"conformance tests reported failures")
	})
})
