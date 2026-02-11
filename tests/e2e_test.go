package tests

import (
	"context"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var cfg *rest.Config
var clientset *kubernetes.Clientset

var _ = BeforeSuite(func() {
	var err error
	cfg, err = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	Expect(err).NotTo(HaveOccurred())

	clientset, err = kubernetes.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())

	if bgLoad {
		chartPath, err := filepath.Abs(bgChartSubdir)
		Expect(err).NotTo(HaveOccurred())

		GinkgoWriter.Printf("Deploying background load (cpu=%s, memory=%s, rps=%d, payloadSize=%d)\n",
			bgCPU, bgMemory, bgRPS, bgPayloadSize)

		err = helmInstallBgLoad(chartPath)
		Expect(err).NotTo(HaveOccurred())
	} else {
		GinkgoWriter.Println("Background load disabled (--bg-load=false)")
	}
})

var _ = AfterSuite(func() {
	if bgLoad {
		GinkgoWriter.Println("Cleaning up background load...")
		if err := helmUninstallBgLoad(); err != nil {
			GinkgoWriter.Printf("WARNING: Failed to uninstall background load: %v\n", err)
		}
		if clientset != nil {
			err := clientset.CoreV1().Namespaces().Delete(context.TODO(), bgNamespace, metav1.DeleteOptions{})
			if err != nil {
				GinkgoWriter.Printf("WARNING: Failed to delete namespace %s: %v\n", bgNamespace, err)
			}
		}
	}
})

var _ = Describe("Cluster", func() {
	It("should have at least one node or namespace", func() {
		// Check namespaces as a light-weight verification
		nss, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(len(nss.Items)).To(BeNumerically(">", 0))
	})
})

var _ = Describe("API Intensive", func() {
	It("should complete the kube-burner api-intensive workload", func() {
		workDir, err := filepath.Abs(apiIntensiveSubdir)
		Expect(err).NotTo(HaveOccurred())

		err = runKubeBurner(workDir, apiIntensiveConfig)
		Expect(err).NotTo(HaveOccurred())
	})
})
