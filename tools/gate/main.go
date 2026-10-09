// Command gate compares bench results against PRD budgets.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/suhaasya/wisp-go/internal/bench"
)

func main() {
	paths := []string{
		"bench-results/size.json",
		"bench-results/go.json",
		"bench-results/tab.json",
	}
	if len(os.Args) > 1 {
		paths = os.Args[1:]
	}

	merged := map[string]bench.Measurement{}
	for _, p := range paths {
		f, err := bench.LoadResults(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "gate: skip %s: %v\n", p, err)
			continue
		}
		for k, v := range f.Metrics {
			merged[k] = v
		}
	}

	var rows []bench.GateRow
	rows = append(rows, bench.EvaluateAll(bench.GoBudgets, merged)...)
	rows = append(rows, bench.EvaluateAll(bench.TabBudgets, merged)...)

	printTable(rows)
	_ = os.MkdirAll("bench-results", 0o750)
	if err := os.WriteFile("bench-results/gate-comment.md", []byte(markdownTable(rows)), 0o600); err != nil { //nolint:gosec // G306: CI artifact
		fmt.Fprintf(os.Stderr, "gate: comment: %v\n", err)
	}

	if bench.AnyCapFail(rows) {
		os.Exit(1)
	}
}

func markdownTable(rows []bench.GateRow) string {
	var b strings.Builder
	b.WriteString("## Wisp budget gate\n\n")
	b.WriteString("| Metric | Value | Target | Cap | Status | Δ vs main |\n")
	b.WriteString("| --- | --- | --- | --- | --- | --- |\n")
	for _, r := range rows {
		val := "—"
		if r.Status != bench.StatusSkipped {
			val = fmt.Sprintf("%.2f %s", r.Value, r.Unit)
		}
		fmt.Fprintf(&b, "| %s | %s | %.0f | %.0f | %s | — |\n", r.Metric, val, r.Target, r.Cap, r.Status)
	}
	return b.String()
}

func printTable(rows []bench.GateRow) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "METRIC\tVALUE\tTARGET\tCAP\tSTATUS")
	for _, r := range rows {
		val := "-"
		if r.Status != bench.StatusSkipped {
			val = fmt.Sprintf("%.2f", r.Value)
		}
		_, _ = fmt.Fprintf(w, "%s\t%s %s\t%.0f\t%.0f\t%s\n", r.Metric, val, r.Unit, r.Target, r.Cap, r.Status)
	}
	_ = w.Flush()

	if os.Getenv("GATE_JSON") == "1" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(rows)
	}
}
