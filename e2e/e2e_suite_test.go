package e2e

import (
	"path/filepath"
	"testing"

	"github.com/canonical/k8s-percorr-testing/internal/k8sutil"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var cfg *rest.Config
var clientset *kubernetes.Clientset

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2E Suite")
}

var _ = BeforeSuite(func(ctx SpecContext) {
	var err error
	cfg, clientset, err = k8sutil.NewDefaultClientset()
	Expect(err).NotTo(HaveOccurred())

	if bgLoad {
		chartPath, err := filepath.Abs(bgChartSubdir)
		Expect(err).NotTo(HaveOccurred())

		By("deploying background load")
		err = helmInstallBgLoad(ctx, chartPath)
		Expect(err).NotTo(HaveOccurred())
	} else {
		By("background load disabled")
	}
})

var _ = AfterSuite(func(ctx SpecContext) {
	if bgLoad {
		By("cleaning up background load")
		err := r.Cmd(ctx, "helm", "uninstall", bgReleaseName,
			"--namespace", bgNamespace, "--wait", "--timeout", "2m")
		Expect(err).NotTo(HaveOccurred())
		if clientset != nil {
			err = clientset.CoreV1().Namespaces().Delete(ctx, bgNamespace, metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
		}
	}
})
