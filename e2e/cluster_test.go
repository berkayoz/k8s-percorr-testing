package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Cluster", func() {
	It("should have at least one node or namespace", func(ctx SpecContext) {
		// Check namespaces as a light-weight verification
		nss, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(len(nss.Items)).To(BeNumerically(">", 0))
	})
})
