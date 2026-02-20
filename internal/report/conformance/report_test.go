package conformance

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vmware-tanzu/sonobuoy/pkg/client/results"
)

func TestParseDumpResults(t *testing.T) {
	raw, err := os.ReadFile("testdata/dump-results.yaml")
	if err != nil {
		t.Fatalf("reading testdata: %v", err)
	}

	item, err := ParseDumpResults(raw)
	if err != nil {
		t.Fatalf("ParseDumpResults: %v", err)
	}

	if item.Status != results.StatusPassed {
		t.Errorf("unexpected top-level status: %q, want %q", item.Status, results.StatusPassed)
	}

	var leaves int
	item.Walk(func(i *results.Item) error {
		if i.IsLeaf() {
			leaves++
		}
		return nil
	})

	if leaves != 5 {
		t.Fatalf("got %d leaf items, want 5", leaves)
	}
}

func TestParseDumpResultsInvalidYAML(t *testing.T) {
	_, err := ParseDumpResults([]byte(":\n\t:bad"))
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestGenerate(t *testing.T) {
	item := &results.Item{
		Name:   "e2e",
		Status: results.StatusPassed,
		Items: []results.Item{
			{Name: "test-a", Status: results.StatusPassed},
			{Name: "test-b", Status: results.StatusPassed},
			{Name: "test-c", Status: results.StatusFailed},
			{Name: "test-d", Status: results.StatusSkipped},
		},
	}

	var buf bytes.Buffer
	if err := Generate(item, &buf); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	out := buf.String()

	mustContain := []string{
		"# CNCF Conformance Report",
		"| Total | 4 |",
		"| Passed | 2 |",
		"| Failed | 1 |",
		"| Skipped | 1 |",
		"## Failures",
		"| test-c |",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("report missing expected string: %q", s)
		}
	}

	// Only failures should appear in the table, not passed/skipped tests.
	if strings.Contains(out, "| test-a |") {
		t.Error("report should not list passed test in failures table")
	}
	if strings.Contains(out, "| test-d |") {
		t.Error("report should not list skipped test in failures table")
	}
}

func TestGenerateNoFailures(t *testing.T) {
	item := &results.Item{
		Name:   "e2e",
		Status: results.StatusPassed,
		Items: []results.Item{
			{Name: "test-a", Status: results.StatusPassed},
			{Name: "test-b", Status: results.StatusPassed},
		},
	}

	var buf bytes.Buffer
	if err := Generate(item, &buf); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	out := buf.String()

	if strings.Contains(out, "## Failures") {
		t.Error("report should not contain Failures section when all tests pass")
	}
	if !strings.Contains(out, "| Failed | 0 |") {
		t.Error("report should show Failed count as 0")
	}
}

func TestGenerateToFile(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "subdir", "conformance-report.md")

	item := &results.Item{
		Name:   "e2e",
		Status: results.StatusPassed,
		Items: []results.Item{
			{Name: "test-a", Status: results.StatusPassed},
			{Name: "test-b", Status: results.StatusFailed},
		},
	}

	if err := GenerateToFile(item, outPath); err != nil {
		t.Fatalf("GenerateToFile: %v", err)
	}

	raw, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("reading output file: %v", err)
	}
	if !strings.Contains(string(raw), "# CNCF Conformance Report") {
		t.Error("output file missing report header")
	}
}
