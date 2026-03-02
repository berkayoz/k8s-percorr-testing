package chaos

import (
	"fmt"
	"path/filepath"
	"time"

	chaosreport "github.com/canonical/k8s-percorr-testing/internal/report/chaos"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
)

var (
	chaosTestsDir           string
	nginxManifest           string
	serviceSymlinksManifest string
)

var _ = BeforeSuite(func(ctx SpecContext) {
	var err error

	chaosTestsDir, err = filepath.Abs(chaosExperimentsSubdir)
	Expect(err).NotTo(HaveOccurred())

	nginxManifest, err = filepath.Abs(chaosNginxManifest)
	Expect(err).NotTo(HaveOccurred())

	superuserManifest, err := filepath.Abs(chaosSuperuserManifest)
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
	By("cleaning up chaos results")
	err := r.Cmd(ctx, "kubectl", "delete", "chaosresults", "--all",
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

	err = r.Cmd(ctx, "kubectl", "delete", "namespace", chaosNamespace, "--ignore-not-found")
	Expect(err).NotTo(HaveOccurred())
})

var _ = ReportAfterSuite("chaos markdown report", func(report types.Report) {
	var results []chaosreport.ExperimentResult
	for _, spec := range report.SpecReports {
		for _, entry := range spec.ReportEntries {
			if entry.Name == "chaos-result" {
				result, ok := entry.GetRawValue().(chaosreport.ExperimentResult)
				if !ok {
					continue
				}
				result.Duration = spec.RunTime
				results = append(results, result)
			}
		}
	}

	if len(results) == 0 {
		GinkgoWriter.Printf("No chaos results collected, skipping report generation\n")
		return
	}

	// Ginkgo resolves --output-dir into the JSONReport path before the
	// test binary runs, so filepath.Dir gives us the same output directory
	// used for JSON/JUnit reports. When no --json-report flag is set the
	// field is empty and filepath.Dir returns ".", writing to cwd.
	_, rc := GinkgoConfiguration()
	reportPath := filepath.Join(filepath.Dir(rc.JSONReport), "chaos-report.md")
	if err := chaosreport.GenerateToFile(results, reportPath); err != nil {
		GinkgoWriter.Printf("Failed to generate chaos report: %v\n", err)
		return
	}
	GinkgoWriter.Printf("Chaos report written to %s\n", reportPath)
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

		// Runs LAST (registered first, LIFO order).
		DeferCleanup(func(ctx SpecContext) {
			By("cleaning up")
			r.Cmd(ctx, "kubectl", "delete", "-f", experimentFile, "--ignore-not-found", "--wait")
		})

		// Runs FIRST (registered second, LIFO order) -- collect ChaosResult before cleanup deletes the manifest.
		DeferCleanup(func(ctx SpecContext) {
			data, err := r.CmdOutput(ctx, "kubectl", "get", "chaosresult", chaosResultName,
				"-n", chaosNamespace, "-o", "json")
			if err != nil {
				GinkgoWriter.Printf("Could not collect ChaosResult %s: %v\n", chaosResultName, err)
				return
			}
			result, err := chaosreport.ParseResult(experiment, data)
			if err != nil {
				GinkgoWriter.Printf("Could not parse ChaosResult %s: %v\n", chaosResultName, err)
				return
			}
			AddReportEntry("chaos-result", result)
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
