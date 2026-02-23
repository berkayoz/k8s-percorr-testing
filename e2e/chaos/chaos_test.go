package chaos

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/canonical/k8s-percorr-testing/internal/k8sutil"
	"github.com/canonical/k8s-percorr-testing/internal/report"
	chaosreport "github.com/canonical/k8s-percorr-testing/internal/report/chaos"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	chaosTestsDir          string
	nginxManifest          string
	serviceSymlinksManifest string
	clientset              *kubernetes.Clientset
	reportsDir             string
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

	reportsDir, err = report.Dir()
	Expect(err).NotTo(HaveOccurred())

	By("installing Litmus operator via Helm")
	err = helmInstallLitmus(ctx)
	Expect(err).NotTo(HaveOccurred())

	By("deploying superuser")
	err = r.Cmd(ctx, "kubectl", "apply", "-f", superuserManifest)
	Expect(err).NotTo(HaveOccurred())

	serviceSymlinksManifest, err = filepath.Abs(chaosServiceSymlinksManifest)
	Expect(err).NotTo(HaveOccurred())

	By("deploying service-symlinks DaemonSet")
	err = r.Cmd(ctx, "kubectl", "apply", "-f", serviceSymlinksManifest)
	Expect(err).NotTo(HaveOccurred())

	By("waiting for service-symlinks rollout")
	err = r.Cmd(ctx, "kubectl", "rollout", "status",
		"daemonset/service-symlinks", "-n", chaosNamespace, "--timeout=2m")
	Expect(err).NotTo(HaveOccurred())

	By("deploying nginx target application")
	err = r.Cmd(ctx, "kubectl", "apply", "-f", nginxManifest)
	Expect(err).NotTo(HaveOccurred())

	By("waiting for nginx deployment rollout")
	err = r.Cmd(ctx, "kubectl", "rollout", "status",
		"deployment/nginx-deployment", "-n", chaosNamespace, "--timeout=2m")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func(ctx SpecContext) {
	By("generating chaos report")
	var results []chaosreport.ExperimentResult
	for _, exp := range experiments {
		chaosResultName := fmt.Sprintf("%s-%s", exp, exp)
		data, err := r.CmdOutput(ctx, "kubectl", "get", "chaosresult", chaosResultName,
			"-n", chaosNamespace, "-o", "json")
		if err != nil {
			// Skip experiments whose ChaosResult doesn't exist.
			continue
		}

		result, err := chaosreport.ParseResult(exp, data)
		Expect(err).NotTo(HaveOccurred())
		results = append(results, result)
	}

	err := chaosreport.GenerateToFile(results,
		filepath.Join(reportsDir, "chaos-report.md"))
	Expect(err).NotTo(HaveOccurred())

	By("cleaning up chaos results")
	err = r.Cmd(ctx, "kubectl", "delete", "chaosresults", "--all",
		"-n", chaosNamespace)
	Expect(err).NotTo(HaveOccurred())

	err = r.Cmd(ctx, "kubectl", "delete", "-f", nginxManifest, "--ignore-not-found")
	Expect(err).NotTo(HaveOccurred())

	By("cleaning up service-symlinks DaemonSet")
	err = r.Cmd(ctx, "kubectl", "delete", "-f", serviceSymlinksManifest, "--ignore-not-found")
	Expect(err).NotTo(HaveOccurred())

	By("uninstalling Litmus")
	err = helmUninstallLitmus(ctx)
	Expect(err).NotTo(HaveOccurred())

	if clientset != nil {
		err = clientset.CoreV1().Namespaces().Delete(ctx, chaosNamespace, metav1.DeleteOptions{})
		Expect(err).NotTo(HaveOccurred())
	}
})

// experimentEntries builds []TableEntry from the shared experiments list.
func experimentEntries() []TableEntry {
	entries := make([]TableEntry, 0, len(experiments))
	for _, e := range experiments {
		entries = append(entries, Entry(nil, e))
	}
	return entries
}

var _ = Describe("Litmus Chaos", Ordered, Serial, func() {
	DescribeTable("should pass", func(ctx SpecContext, experiment string) {
		experimentFile := filepath.Join(chaosTestsDir, fmt.Sprintf("%s.yaml", experiment))
		chaosResultName := fmt.Sprintf("%s-%s", experiment, experiment)

		DeferCleanup(func(ctx SpecContext) {
			By("cleaning up")
			r.Cmd(ctx, "kubectl", "delete", "-f", experimentFile, "--ignore-not-found", "--wait")
		})

		By("applying experiment")
		err := r.Cmd(ctx, "kubectl", "apply", "-f", experimentFile)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func(g Gomega) string {
			data, err := r.CmdOutput(ctx, "kubectl", "get", "chaosresult", chaosResultName,
				"-n", chaosNamespace, "-o", "jsonpath={.status.experimentStatus.verdict}")
			g.Expect(err).NotTo(HaveOccurred())
			return string(data)
		}).WithContext(ctx).WithTimeout(10 * time.Minute).WithPolling(10 * time.Second).Should(Equal("Pass"))
	},
		experimentEntries(),
	)

})
