package chaos

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseResult(t *testing.T) {
	raw, err := os.ReadFile("testdata/chaosresult.json")
	if err != nil {
		t.Fatalf("reading testdata: %v", err)
	}

	result, err := ParseResult("pod-network-loss", raw)
	if err != nil {
		t.Fatalf("ParseResult: %v", err)
	}

	if result.Name != "pod-network-loss" {
		t.Errorf("Name = %q, want %q", result.Name, "pod-network-loss")
	}
	if result.Verdict != "Fail" {
		t.Errorf("Verdict = %q, want %q", result.Verdict, "Fail")
	}
	if result.Phase != "Completed" {
		t.Errorf("Phase = %q, want %q", result.Phase, "Completed")
	}
	if result.FailStep != "ChaosInject" {
		t.Errorf("FailStep = %q, want %q", result.FailStep, "ChaosInject")
	}
	if result.ProbeSuccessPercentage != "0" {
		t.Errorf("ProbeSuccessPercentage = %q, want %q", result.ProbeSuccessPercentage, "0")
	}
}

func TestGenerate(t *testing.T) {
	results := []ExperimentResult{
		{Name: "container-kill", Verdict: "Pass", Phase: "Completed", ProbeSuccessPercentage: "100"},
		{Name: "disk-fill", Verdict: "Pass", Phase: "Completed", ProbeSuccessPercentage: "100"},
		{Name: "pod-network-loss", Verdict: "Fail", Phase: "Completed", FailStep: "ChaosInject", ProbeSuccessPercentage: "0"},
	}

	var buf bytes.Buffer
	if err := Generate(results, &buf); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	out := buf.String()

	mustContain := []string{
		"# Chaos Testing Report",
		"**Total Experiments**: 3",
		"| Total | 3 |",
		"| Passed | 2 |",
		"| Failed | 1 |",
		"66.7%",
		"## Experiment Results",
		"| container-kill | Pass |",
		"| disk-fill | Pass |",
		"| pod-network-loss | Fail |",
		"| ChaosInject |",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("report missing expected string: %q", s)
		}
	}
}

func TestGenerateToFile(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "subdir", "chaos-report.md")

	results := []ExperimentResult{
		{Name: "container-kill", Verdict: "Pass", Phase: "Completed", ProbeSuccessPercentage: "100"},
	}

	if err := GenerateToFile(results, outPath); err != nil {
		t.Fatalf("GenerateToFile: %v", err)
	}

	raw, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("reading output file: %v", err)
	}
	if !strings.Contains(string(raw), "# Chaos Testing Report") {
		t.Error("output file missing report header")
	}
}

func TestParseResultInvalidJSON(t *testing.T) {
	_, err := ParseResult("test", []byte("not json"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
