package conformance

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var destDir string

var _ = BeforeSuite(func(ctx SpecContext) {
	// Verify sonobuoy CLI is available.
	GinkgoWriter.Println("Verifying sonobuoy CLI...")
	err := r.Cmd(ctx, "sonobuoy", "version")
	Expect(err).NotTo(HaveOccurred(), "sonobuoy CLI not found")

	// Prepare results directory.
	destDir, err = getResultsDir()
	Expect(err).NotTo(HaveOccurred(), "failed to prepare results directory")
	GinkgoWriter.Printf("Results will be stored in: %s\n", destDir)
})

var _ = AfterSuite(func(ctx SpecContext) {
	GinkgoWriter.Println("Cleaning up sonobuoy resources...")
	sonobuoyDelete(ctx)
})

var _ = Describe("CNCF Conformance", Ordered, Serial, func() {
	It("should run sonobuoy conformance tests", func(ctx SpecContext) {
		GinkgoWriter.Println("Running sonobuoy certified-conformance mode...")
		err := sonobuoyRun(ctx)
		Expect(err).NotTo(HaveOccurred(), "sonobuoy run failed")
	})

	It("should retrieve and validate results", func(ctx SpecContext) {
		GinkgoWriter.Println("Retrieving sonobuoy results...")
		tarballPath, err := sonobuoyRetrieve(ctx, destDir)
		Expect(err).NotTo(HaveOccurred(), "sonobuoy retrieve failed")

		tarball := strings.TrimSpace(string(tarballPath))
		Expect(tarball).NotTo(BeEmpty(), "sonobuoy retrieve returned empty path")
		GinkgoWriter.Printf("Results tarball: %s\n", tarball)

		GinkgoWriter.Println("Parsing sonobuoy results...")
		summary, err := sonobuoyResults(ctx, tarball)
		Expect(err).NotTo(HaveOccurred(), "sonobuoy results failed")

		output := string(summary)
		GinkgoWriter.Printf("Results summary:\n%s\n", output)
		Expect(output).NotTo(ContainSubstring("Status: failed"),
			"conformance tests reported failures")
	})
})
