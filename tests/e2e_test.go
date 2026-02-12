package tests

import (
	"path/filepath"

	"github.com/canonical/k8s-percorr-testing/pkg/k8sutil"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var cfg *rest.Config
var clientset *kubernetes.Clientset

var _ = BeforeSuite(func(ctx SpecContext) {
	var err error
	cfg, clientset, err = k8sutil.NewDefaultClientset()
	Expect(err).NotTo(HaveOccurred())

	if bgLoad {
		chartPath, err := filepath.Abs(bgChartSubdir)
		Expect(err).NotTo(HaveOccurred())

		GinkgoWriter.Printf("Deploying background load (cpu=%s, memory=%s, rps=%d, payloadSize=%d)\n",
			bgCPU, bgMemory, bgRPS, bgPayloadSize)

		err = helmInstallBgLoad(ctx, chartPath)
		Expect(err).NotTo(HaveOccurred())
	} else {
		GinkgoWriter.Println("Background load disabled (--bg-load=false)")
	}
})

var _ = AfterSuite(func(ctx SpecContext) {
	if bgLoad {
		GinkgoWriter.Println("Cleaning up background load...")
		if err := r.Cmd(ctx, "helm", "uninstall", bgReleaseName,
			"--namespace", bgNamespace, "--wait", "--timeout", "2m"); err != nil {
			GinkgoWriter.Printf("WARNING: Failed to uninstall background load: %v\n", err)
		}
		if clientset != nil {
			err := clientset.CoreV1().Namespaces().Delete(ctx, bgNamespace, metav1.DeleteOptions{})
			if err != nil {
				GinkgoWriter.Printf("WARNING: Failed to delete namespace %s: %v\n", bgNamespace, err)
			}
		}
	}
})

var _ = Describe("Cluster", func() {
	It("should have at least one node or namespace", func(ctx SpecContext) {
		// Check namespaces as a light-weight verification
		nss, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(len(nss.Items)).To(BeNumerically(">", 0))
	})
})

