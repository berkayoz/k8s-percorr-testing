package conformance

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"strings"
	"text/template"
	"time"

	"github.com/canonical/k8s-percorr-testing/internal/report"
	"github.com/vmware-tanzu/sonobuoy/pkg/client/results"
	"gopkg.in/yaml.v3"
)

// reportData is the top-level data passed to the Markdown template.
type reportData struct {
	Timestamp   string
	TotalCount  int
	PassedCount int
	FailedCount int
	Results     []results.Item
	Failures    []results.Item
}

//go:embed report.md.tmpl
var markdownTemplate string

var tmpl = template.Must(template.New("conformance-report").Parse(markdownTemplate))

// suiteLifecyclePrefixes are Ginkgo suite setup/teardown node name prefixes
// that are not actual test cases and should be excluded from the report.
var suiteLifecyclePrefixes = []string{
	"[BeforeSuite]",
	"[AfterSuite]",
	"[SynchronizedBeforeSuite]",
	"[SynchronizedAfterSuite]",
	"[ReportBeforeSuite]",
	"[ReportAfterSuite]",
}

// ParseDumpResults parses the YAML output of
// `sonobuoy results --mode dump <tarball>` and returns the first document
// as a results.Item tree.
func ParseDumpResults(data []byte) (*results.Item, error) {
	var item results.Item
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&item); err != nil {
		return nil, fmt.Errorf("decoding dump results: %w", err)
	}
	return &item, nil
}

// Generate renders a Markdown conformance report and writes it to w.
func Generate(item *results.Item, w io.Writer) error {
	data := buildReportData(item)
	return tmpl.Execute(w, data)
}

// GenerateToFile is a convenience wrapper that writes the report to outputPath.
func GenerateToFile(item *results.Item, outputPath string) error {
	return report.WriteToFile(outputPath, func(w io.Writer) error {
		return Generate(item, w)
	})
}

// buildReportData computes summary statistics and filters for failures.
// Only passed and failed test cases are counted; skipped tests and Ginkgo
// suite lifecycle nodes (BeforeSuite, AfterSuite, etc.) are excluded.
func buildReportData(item *results.Item) *reportData {
	var passed, failed int
	var all, failures []results.Item

	item.Walk(func(i *results.Item) error {
		if !i.IsLeaf() {
			return nil
		}
		if isSuiteLifecycleNode(i.Name) {
			return nil
		}
		switch i.Status {
		case results.StatusPassed:
			passed++
			all = append(all, *i)
		case results.StatusFailed:
			failed++
			all = append(all, *i)
			failures = append(failures, *i)
		}
		return nil
	})

	return &reportData{
		Timestamp:   time.Now().UTC().Format("2006-01-02 15:04:05 UTC"),
		TotalCount:  passed + failed,
		PassedCount: passed,
		FailedCount: failed,
		Results:     all,
		Failures:    failures,
	}
}

func isSuiteLifecycleNode(name string) bool {
	for _, prefix := range suiteLifecyclePrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}
