package kubeburner

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCollect(t *testing.T) {
	data, err := collect("testdata")
	if err != nil {
		t.Fatalf("collect: %v", err)
	}

	if got := len(data.Jobs); got != 3 {
		t.Fatalf("expected 3 job summaries, got %d", got)
	}
	if data.Jobs[0].JobConfig.Name != "api-intensive" {
		t.Errorf("expected first job name api-intensive, got %s", data.Jobs[0].JobConfig.Name)
	}
	if !data.Jobs[0].Passed {
		t.Error("expected first job to have passed")
	}

	if got := len(data.Latencies); got != 1 {
		t.Fatalf("expected 1 latency group, got %d", got)
	}
	if data.Latencies[0].JobName != "api-intensive" {
		t.Errorf("expected latency job name api-intensive, got %s", data.Latencies[0].JobName)
	}
	if got := len(data.Latencies[0].Quantiles); got != 4 {
		t.Fatalf("expected 4 quantile rows, got %d", got)
	}
}

func TestGenerate(t *testing.T) {
	var buf bytes.Buffer
	if err := Generate("testdata", "api-intensive.yaml", &buf); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	out := buf.String()

	mustContain := []string{
		"# kube-burner Performance Report",
		"**Config**: api-intensive.yaml",
		"## Job Summary",
		"| api-intensive |",
		"| api-intensive-patch |",
		"| PASS |",
		"## Pod Latency (Quantiles)",
		"### api-intensive",
		"| PodScheduled |",
		"| ContainersReady |",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("report missing expected string: %q", s)
		}
	}
}

func TestGenerateToFile(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "subdir", "report.md")

	if err := GenerateToFile("testdata", "api-intensive.yaml", outPath); err != nil {
		t.Fatalf("GenerateToFile: %v", err)
	}

	raw, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("reading output file: %v", err)
	}
	if !strings.Contains(string(raw), "# kube-burner Performance Report") {
		t.Error("output file missing report header")
	}
}

func TestCollectMissingSummary(t *testing.T) {
	dir := t.TempDir()
	_, err := collect(dir)
	if err == nil {
		t.Fatal("expected error for missing jobSummary.json")
	}
}
