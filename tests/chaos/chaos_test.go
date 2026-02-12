package chaos

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/canonical/k8s-percorr-testing/pkg/k8sutil"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	chaosTestsDir string
	nginxManifest string
	clientset     *kubernetes.Clientset
)

var _ = BeforeSuite(func(ctx SpecContext) {
	var err error
	_, clientset, err = k8sutil.NewDefaultClientset()
	Expect(err).NotTo(HaveOccurred())

	chaosTestsDir, err = filepath.Abs(chaosExperimentsSubdir)
	Expect(err).NotTo(HaveOccurred())

	nginxManifest, err = filepath.Abs(chaosNginxManifest)
	Expect(err).NotTo(HaveOccurred())

	superuserManifest, err := filepath.Abs(chaosSuperuserManifest)
	Expect(err).NotTo(HaveOccurred())

	// Install Litmus via Helm
	GinkgoWriter.Println("Installing Litmus operator via Helm...")
	err = helmInstallLitmus(ctx)
	Expect(err).NotTo(HaveOccurred())

	// Deploy superuser
	GinkgoWriter.Println("Deploying superuser...")
	err = r.Cmd(ctx, "kubectl", "apply", "-f", superuserManifest)
	Expect(err).NotTo(HaveOccurred())

	// Deploy nginx target application
	GinkgoWriter.Println("Deploying nginx target application...")
	err = r.Cmd(ctx, "kubectl", "apply", "-f", nginxManifest)
	Expect(err).NotTo(HaveOccurred())

	// Wait for nginx deployment rollout
	GinkgoWriter.Println("Waiting for nginx deployment rollout...")
	err = r.Cmd(ctx, "kubectl", "rollout", "status",
		"deployment/nginx-deployment", "-n", chaosNamespace, "--timeout=2m")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func(ctx SpecContext) {
	// Delete nginx target
	if err := r.Cmd(ctx, "kubectl", "delete", "-f", nginxManifest, "--ignore-not-found"); err != nil {
		GinkgoWriter.Printf("WARNING: Failed to delete nginx: %v\n", err)
	}

	// Uninstall Litmus (both charts)
	GinkgoWriter.Println("Uninstalling Litmus...")
	helmUninstallLitmus(ctx)

	// Delete namespace
	if clientset != nil {
		err := clientset.CoreV1().Namespaces().Delete(ctx, chaosNamespace, metav1.DeleteOptions{})
		if err != nil {
			GinkgoWriter.Printf("WARNING: Failed to delete namespace %s: %v\n", chaosNamespace, err)
		}
	}
})

var _ = Describe("Litmus Chaos", Ordered, Serial, func() {
	DescribeTable("should pass", func(ctx SpecContext, experiment string) {
		experimentFile := filepath.Join(chaosTestsDir, fmt.Sprintf("%s.yaml", experiment))
		chaosResultName := fmt.Sprintf("%s-%s", experiment, experiment)

		DeferCleanup(func(ctx SpecContext) {
			GinkgoWriter.Printf("[%s] Cleaning up...\n", experiment)
			r.Cmd(ctx, "kubectl", "delete", "-f", experimentFile, "--ignore-not-found", "--wait")
			r.Cmd(ctx, "kubectl", "delete", "chaosresults", chaosResultName,
				"-n", chaosNamespace, "--ignore-not-found", "--wait")
		})

		GinkgoWriter.Printf("[%s] Applying %s.yaml...\n", experiment, experiment)
		err := r.Cmd(ctx, "kubectl", "apply", "-f", experimentFile)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func(g Gomega) string {
			data, err := r.CmdOutput(ctx, "kubectl", "get", "chaosresult", chaosResultName,
				"-n", chaosNamespace, "-o", "jsonpath={.status.experimentStatus.verdict}")
			g.Expect(err).NotTo(HaveOccurred())
			return string(data)
		}).WithContext(ctx).WithTimeout(10 * time.Minute).WithPolling(10 * time.Second).Should(Equal("Pass"))
	},
		Entry(nil, "container-kill"),
		Entry(nil, "disk-fill"),
		Entry(nil, "docker-service-kill"),
		Entry(nil, "kubelet-service-kill"),
		Entry(nil, "node-cpu-hog"),
		Entry(nil, "node-io-stress"),
		Entry(nil, "node-memory-hog"),
		Entry(nil, "node-taint"),
		Entry(nil, "pod-autoscaler"),
		Entry(nil, "pod-cpu-hog"),
		Entry(nil, "pod-cpu-hog-exec"),
		Entry(nil, "pod-delete"),
		Entry(nil, "pod-dns-error"),
		Entry(nil, "pod-dns-spoof"),
		Entry(nil, "pod-http-latency"),
		Entry(nil, "pod-http-modify-body"),
		Entry(nil, "pod-http-modify-header"),
		Entry(nil, "pod-http-reset-peer"),
		Entry(nil, "pod-http-status-code"),
		Entry(nil, "pod-io-stress"),
		Entry(nil, "pod-memory-hog"),
		Entry(nil, "pod-memory-hog-exec"),
		Entry(nil, "pod-network-corruption"),
		Entry(nil, "pod-network-duplication"),
		Entry(nil, "pod-network-latency"),
		Entry(nil, "pod-network-loss"),
		Entry(nil, "pod-network-partition"),
	)
})
