package kubeburner

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("API Intensive", func() {
	It("should complete the kube-burner api-intensive workload", func(ctx SpecContext) {
		workDir, err := filepath.Abs(manifestsSubdir)
		Expect(err).NotTo(HaveOccurred())

		err = runKubeBurner(ctx, workDir, apiIntensiveConfig)
		Expect(err).NotTo(HaveOccurred())
	})
})
