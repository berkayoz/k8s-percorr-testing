package chaos

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"text/template"
	"time"

	"github.com/canonical/k8s-percorr-testing/internal/report"
)

// ExperimentResult holds the verdict and status of a single chaos experiment,
// extracted from a ChaosResult custom resource.
type ExperimentResult struct {
	Name                   string
	Verdict                string
	Phase                  string
	FailStep               string
	ProbeSuccessPercentage string
}

// chaosResultItem mirrors the JSON structure of a single ChaosResult resource.
type chaosResultItem struct {
	Status struct {
		ExperimentStatus struct {
			Verdict                string `json:"verdict"`
			Phase                  string `json:"phase"`
			FailStep               string `json:"failStep"`
			ProbeSuccessPercentage string `json:"probeSuccessPercentage"`
		} `json:"experimentStatus"`
	} `json:"status"`
}

// reportData is the top-level data passed to the Markdown template.
type reportData struct {
	Timestamp   string
	TotalCount  int
	PassedCount int
	FailedCount int
	PassRate    float64
	Results     []ExperimentResult
}

//go:embed report.md.tmpl
var markdownTemplate string

var tmpl = template.Must(template.New("chaos-report").Parse(markdownTemplate))

// ParseResult parses the JSON output of a single ChaosResult object
// (e.g. `kubectl get chaosresult <name> -o json`) and returns an ExperimentResult.
// The caller supplies the experiment name directly.
func ParseResult(name string, jsonData []byte) (ExperimentResult, error) {
	var item chaosResultItem
	if err := json.Unmarshal(jsonData, &item); err != nil {
		return ExperimentResult{}, fmt.Errorf("parsing chaos result JSON: %w", err)
	}

	return ExperimentResult{
		Name:                   name,
		Verdict:                item.Status.ExperimentStatus.Verdict,
		Phase:                  item.Status.ExperimentStatus.Phase,
		FailStep:               item.Status.ExperimentStatus.FailStep,
		ProbeSuccessPercentage: item.Status.ExperimentStatus.ProbeSuccessPercentage,
	}, nil
}

// Generate renders a Markdown chaos report and writes it to w.
func Generate(results []ExperimentResult, w io.Writer) error {
	data := buildReportData(results)
	return tmpl.Execute(w, data)
}

// GenerateToFile is a convenience wrapper that writes the report to outputPath.
func GenerateToFile(results []ExperimentResult, outputPath string) error {
	return report.WriteToFile(outputPath, func(w io.Writer) error {
		return Generate(results, w)
	})
}

// buildReportData computes summary statistics and enriches results with categories.
func buildReportData(results []ExperimentResult) *reportData {
	var passed, failed int
	for _, r := range results {
		if r.Verdict == "Pass" {
			passed++
		} else {
			failed++
		}
	}

	total := len(results)
	var passRate float64
	if total > 0 {
		passRate = float64(passed) / float64(total) * 100
	}

	return &reportData{
		Timestamp:   time.Now().UTC().Format("2006-01-02 15:04:05 UTC"),
		TotalCount:  total,
		PassedCount: passed,
		FailedCount: failed,
		PassRate:    passRate,
		Results:     results,
	}
}
