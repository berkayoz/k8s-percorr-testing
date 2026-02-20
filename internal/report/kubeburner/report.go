package kubeburner

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/canonical/k8s-percorr-testing/internal/report"
)

// JobConfig contains the job configuration nested under jobConfig in the JSON.
type JobConfig struct {
	Name       string  `json:"name"`
	JobType    string  `json:"jobType"`
	Iterations int     `json:"jobIterations"`
	QPS        float64 `json:"qps"`
	Burst      int     `json:"burst"`
}

// JobSummary maps the JSON produced by kube-burner's local indexer for each
// job execution (found in jobSummary.json).
type JobSummary struct {
	JobConfig   JobConfig `json:"jobConfig"`
	Passed      bool      `json:"passed"`
	Elapsed     float64   `json:"elapsedTime"`
	AchievedQPS float64   `json:"achievedQps"`
}

// PodLatencyQuantiles maps a single quantile row from
// podLatencyQuantilesMeasurement-<job>.json.
type PodLatencyQuantiles struct {
	QuantileName string  `json:"quantileName"`
	P50          int64   `json:"P50"`
	P95          int64   `json:"P95"`
	P99          int64   `json:"P99"`
	Max          int64   `json:"max"`
	Avg          float64 `json:"avg"`
	JobName      string  `json:"jobName"`
}

// jobLatency groups quantile rows for a single job.
type jobLatency struct {
	JobName   string
	Quantiles []PodLatencyQuantiles
}

// reportData is the top-level data passed to the Markdown template.
type reportData struct {
	Timestamp  string
	ConfigFile string
	Jobs       []JobSummary
	Latencies  []jobLatency
}

//go:embed report.md.tmpl
var markdownTemplate string

var tmpl = template.Must(template.New("report").Parse(markdownTemplate))

// Generate reads kube-burner local-indexer output from metricsDir, renders a
// Markdown report and writes it to w.
func Generate(metricsDir, configFile string, w io.Writer) error {
	data, err := collect(metricsDir)
	if err != nil {
		return err
	}
	data.ConfigFile = configFile
	data.Timestamp = time.Now().UTC().Format("2006-01-02 15:04:05 UTC")
	return tmpl.Execute(w, data)
}

// GenerateToFile is a convenience wrapper that writes the report to outputPath.
func GenerateToFile(metricsDir, configFile, outputPath string) error {
	return report.WriteToFile(outputPath, func(w io.Writer) error {
		return Generate(metricsDir, configFile, w)
	})
}

// collect reads jobSummary.json and any podLatencyQuantilesMeasurement-*.json
// files from metricsDir.
func collect(metricsDir string) (*reportData, error) {
	data := &reportData{}

	// Job summaries.
	summaryPath := filepath.Join(metricsDir, "jobSummary.json")
	raw, err := os.ReadFile(summaryPath)
	if err != nil {
		return nil, fmt.Errorf("reading job summary: %w", err)
	}
	if err := json.Unmarshal(raw, &data.Jobs); err != nil {
		return nil, fmt.Errorf("parsing job summary: %w", err)
	}

	// Pod latency quantiles — one file per job that produces them.
	matches, _ := filepath.Glob(filepath.Join(metricsDir, "podLatencyQuantilesMeasurement-*.json"))
	for _, m := range matches {
		raw, err := os.ReadFile(m)
		if err != nil {
			continue // silently skip unreadable files
		}
		var quantiles []PodLatencyQuantiles
		if err := json.Unmarshal(raw, &quantiles); err != nil {
			continue
		}
		if len(quantiles) == 0 {
			continue
		}
		// Derive job name from filename:
		// podLatencyQuantilesMeasurement-<jobName>.json
		base := filepath.Base(m)
		jobName := strings.TrimPrefix(base, "podLatencyQuantilesMeasurement-")
		jobName = strings.TrimSuffix(jobName, ".json")
		data.Latencies = append(data.Latencies, jobLatency{
			JobName:   jobName,
			Quantiles: quantiles,
		})
	}

	return data, nil
}
