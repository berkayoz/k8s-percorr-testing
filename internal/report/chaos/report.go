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
	Description            string
	URL                    string
	Verdict                string
	Phase                  string
	FailStep               string
	ProbeSuccessPercentage string
}

type experimentInfo struct {
	Description string
	URL         string
}

var experimentMeta = map[string]experimentInfo{
	"container-kill":          {"Kills the application container and validates recovery", "https://litmuschaos.github.io/litmus/experiments/categories/pods/container-kill/"},
	"disk-fill":               {"Fills the ephemeral storage of a pod to test disk pressure handling", "https://litmuschaos.github.io/litmus/experiments/categories/pods/disk-fill/"},
	"docker-service-kill":     {"Kills the container runtime service on a node to simulate runtime failure", "https://litmuschaos.github.io/litmus/experiments/categories/nodes/docker-service-kill/"},
	"kubelet-service-kill":    {"Kills the kubelet service on a node to test node-level recovery", "https://litmuschaos.github.io/litmus/experiments/categories/nodes/kubelet-service-kill/"},
	"node-cpu-hog":            {"Stresses CPU on a Kubernetes node to test resource contention", "https://litmuschaos.github.io/litmus/experiments/categories/nodes/node-cpu-hog/"},
	"node-io-stress":          {"Stresses I/O on a Kubernetes node to test disk performance", "https://litmuschaos.github.io/litmus/experiments/categories/nodes/node-io-stress/"},
	"node-memory-hog":         {"Stresses memory on a Kubernetes node to test OOM behaviour", "https://litmuschaos.github.io/litmus/experiments/categories/nodes/node-memory-hog/"},
	"node-taint":              {"Taints a node to evict pods and test rescheduling", "https://litmuschaos.github.io/litmus/experiments/categories/nodes/node-taint/"},
	"pod-autoscaler":          {"Scales replicas to test horizontal pod autoscaling behaviour", "https://litmuschaos.github.io/litmus/experiments/categories/pods/pod-autoscaler/"},
	"pod-cpu-hog":             {"Stresses CPU of a pod using system resources", "https://litmuschaos.github.io/litmus/experiments/categories/pods/pod-cpu-hog/"},
	"pod-cpu-hog-exec":        {"Stresses CPU of a pod using exec into the container", "https://litmuschaos.github.io/litmus/experiments/categories/pods/pod-cpu-hog-exec/"},
	"pod-delete":              {"Deletes a pod to test application availability and recovery", "https://litmuschaos.github.io/litmus/experiments/categories/pods/pod-delete/"},
	"pod-dns-error":           {"Injects DNS errors to test application DNS failure handling", "https://litmuschaos.github.io/litmus/experiments/categories/pods/pod-dns-error/"},
	"pod-dns-spoof":           {"Spoofs DNS responses to redirect traffic and test resilience", "https://litmuschaos.github.io/litmus/experiments/categories/pods/pod-dns-spoof/"},
	"pod-http-latency":        {"Injects latency into HTTP responses of a pod", "https://litmuschaos.github.io/litmus/experiments/categories/pods/pod-http-latency/"},
	"pod-http-modify-body":    {"Modifies HTTP response body of a pod to test error handling", "https://litmuschaos.github.io/litmus/experiments/categories/pods/pod-http-modify-body/"},
	"pod-http-modify-header":  {"Modifies HTTP response headers of a pod to test resilience", "https://litmuschaos.github.io/litmus/experiments/categories/pods/pod-http-modify-header/"},
	"pod-http-reset-peer":     {"Resets HTTP peer connections to test connection handling", "https://litmuschaos.github.io/litmus/experiments/categories/pods/pod-http-reset-peer/"},
	"pod-http-status-code":    {"Modifies HTTP response status code of a pod", "https://litmuschaos.github.io/litmus/experiments/categories/pods/pod-http-status-code/"},
	"pod-io-stress":           {"Stresses I/O of a pod to test disk performance under load", "https://litmuschaos.github.io/litmus/experiments/categories/pods/pod-io-stress/"},
	"pod-memory-hog":          {"Stresses memory of a pod using system resources", "https://litmuschaos.github.io/litmus/experiments/categories/pods/pod-memory-hog/"},
	"pod-memory-hog-exec":     {"Stresses memory of a pod using exec into the container", "https://litmuschaos.github.io/litmus/experiments/categories/pods/pod-memory-hog-exec/"},
	"pod-network-corruption":  {"Injects network packet corruption on a pod", "https://litmuschaos.github.io/litmus/experiments/categories/pods/pod-network-corruption/"},
	"pod-network-duplication": {"Injects network packet duplication on a pod", "https://litmuschaos.github.io/litmus/experiments/categories/pods/pod-network-duplication/"},
	"pod-network-latency":     {"Injects network latency on a pod to test slow-network resilience", "https://litmuschaos.github.io/litmus/experiments/categories/pods/pod-network-latency/"},
	"pod-network-loss":        {"Injects network packet loss on a pod to test reliability", "https://litmuschaos.github.io/litmus/experiments/categories/pods/pod-network-loss/"},
	"pod-network-partition":   {"Partitions a pod from the network to test isolation handling", "https://litmuschaos.github.io/litmus/experiments/categories/pods/pod-network-partition/"},
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

// buildReportData computes summary statistics and enriches results with metadata.
func buildReportData(results []ExperimentResult) *reportData {
	var passed, failed int
	enriched := make([]ExperimentResult, len(results))
	for i, r := range results {
		if r.Verdict == "Pass" {
			passed++
		} else {
			failed++
		}
		enriched[i] = r
		if meta, ok := experimentMeta[r.Name]; ok {
			enriched[i].Description = meta.Description
			enriched[i].URL = meta.URL
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
		Results:     enriched,
	}
}
