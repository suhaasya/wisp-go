package bench

import (
	"encoding/json"
	"os"
)

// ResultsFile is the JSON format written by size/bench/tabbench.
type ResultsFile struct {
	Metrics map[string]Measurement `json:"metrics"`
}

// LoadResults reads a results JSON file.
func LoadResults(path string) (ResultsFile, error) {
	data, err := os.ReadFile(path) //nolint:gosec // G304: bench results paths from CI/Makefile
	if err != nil {
		return ResultsFile{}, err
	}
	var f ResultsFile
	if err := json.Unmarshal(data, &f); err != nil {
		return ResultsFile{}, err
	}
	if f.Metrics == nil {
		f.Metrics = map[string]Measurement{}
	}
	return f, nil
}

// MergeResults merges multiple result files (later files override).
func MergeResults(files ...ResultsFile) map[string]Measurement {
	out := map[string]Measurement{}
	for _, f := range files {
		for k, v := range f.Metrics {
			out[k] = v
		}
	}
	return out
}

// AppendResults merges metrics into an existing results file.
func AppendResults(path string, metrics map[string]Measurement) error {
	existing, err := LoadResults(path)
	if err != nil {
		existing = ResultsFile{Metrics: map[string]Measurement{}}
	}
	for k, v := range metrics {
		existing.Metrics[k] = v
	}
	return WriteResults(path, existing.Metrics)
}

// WriteResults writes metrics to path.
func WriteResults(path string, metrics map[string]Measurement) error {
	if err := os.MkdirAll(dirOf(path), 0o750); err != nil {
		return err
	}
	data, err := json.MarshalIndent(ResultsFile{Metrics: metrics}, "", "  ")
	if err != nil {
		return err
	}
	//nolint:gosec // G306: bench artifacts are local CI output with 0600
	return os.WriteFile(path, append(data, '\n'), 0o600)
}

func dirOf(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[:i]
		}
	}
	return "."
}
