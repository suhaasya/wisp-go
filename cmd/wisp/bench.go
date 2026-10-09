package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/suhaasya/wisp-go/internal/bench"
	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/driver/mock"
	"github.com/suhaasya/wisp-go/internal/pager"
	"github.com/suhaasya/wisp-go/internal/server"
)

func benchScenarioFromArgs(args []string) (string, bool) {
	for i, a := range args {
		if a == "--bench-scenario" && i+1 < len(args) {
			return args[i+1], true
		}
	}
	return "", false
}

func runBenchScenario(name string) error {
	out := os.Getenv("WISP_BENCH_OUTPUT")
	if out == "" {
		out = "bench-results/go.json"
	}

	if name == "all" {
		return runAllBenchScenarios(out)
	}

	metrics, err := metricsForScenario(name)
	if err != nil {
		return err
	}
	if err := bench.AppendResults(out, metrics); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "wisp bench: updated %s (scenario %s)\n", out, name)
	return nil
}

func runAllBenchScenarios(out string) error {
	names := []string{"idle", "one_conn_500_tables", "ten_connections", "scroll_1m_rows", "start"}
	if err := os.Remove(out); err != nil && !os.IsNotExist(err) { //nolint:gosec // G703: WISP_BENCH_OUTPUT from env
		return err
	}
	for _, n := range names {
		if err := runBenchScenarioSingle(out, n); err != nil {
			return err
		}
	}
	fmt.Fprintf(os.Stderr, "wisp bench: wrote %s (all scenarios)\n", out)
	return nil
}

func runBenchScenarioSingle(out, name string) error {
	metrics, err := metricsForScenario(name)
	if err != nil {
		return err
	}
	return bench.AppendResults(out, metrics)
}

func metricsForScenario(name string) (map[string]bench.Measurement, error) {
	metrics := map[string]bench.Measurement{}
	switch name {
	case "idle":
		metrics["idle_rss_mb"] = skipped("SSE client attach not implemented (LW-009)")
	case "one_conn_500_tables":
		metrics["one_conn_500_rss_mb"] = skipped("database drivers and schema bench not implemented (M1)")
	case "ten_connections":
		metrics["per_conn_rss_mb"] = skipped("ten connection scenario not implemented (M1)")
	case "scroll_1m_rows":
		m, err := scroll1mMockPager()
		if err != nil {
			return nil, err
		}
		metrics["scroll_1m_rss_mb"] = m
	case "start":
		start := time.Now()
		ln, err := server.ListenLoopback(0)
		if err != nil {
			return nil, err
		}
		_ = ln.Close()
		metrics["start_ms"] = bench.Measurement{Value: float64(time.Since(start).Milliseconds())}
		if rss, err := bench.RSSMegabytes(); err == nil {
			metrics["rss_sample_mb"] = bench.Measurement{Value: rss}
		}
	default:
		return nil, fmt.Errorf("unknown bench scenario %q", name)
	}
	return metrics, nil
}

func skipped(reason string) bench.Measurement {
	return bench.Measurement{Skipped: true, Reason: reason}
}

func scroll1mMockPager() (bench.Measurement, error) {
	const total = 1_000_000
	var d mock.Driver
	conn, err := d.Open(context.Background(), driver.ConnectParams{
		Extra: mock.Config{
			TotalRows: total,
			Columns: []driver.ColumnMeta{
				{Name: "id", Type: "int8"},
				{Name: "name", Type: "text"},
			},
		},
	})
	if err != nil {
		return bench.Measurement{}, err
	}
	defer func() { _ = conn.Close() }()
	mgr := pager.NewManager(conn, "mock", time.Now)
	mgr.SetMockTotal(total)
	ctx := context.Background()
	open, err := mgr.Open(ctx, pager.Source{
		Kind:       pager.SourceTable,
		Schema:     "public",
		Table:      "big",
		KeyColumns: []string{"id"},
	})
	if err != nil {
		return bench.Measurement{}, err
	}
	if err := pager.ScrollEndWindows(ctx, mgr, open.RID, total, pager.DefaultWindowRows); err != nil {
		return bench.Measurement{}, err
	}
	rss, err := bench.RSSMegabytes()
	if err != nil {
		return bench.Measurement{Skipped: true, Reason: err.Error()}, nil
	}
	return bench.Measurement{Value: rss}, nil
}
