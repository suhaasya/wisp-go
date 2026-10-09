package bench_test

import (
	"testing"

	"github.com/suhaasya/wisp-go/internal/bench"
)

func TestEvaluateBudget_capFail(t *testing.T) {
	t.Parallel()
	b := bench.Budget{Metric: "idle_rss_mb", Target: 20, Cap: 30, Unit: "MB"}
	got := bench.EvaluateBudget(b, bench.Measurement{Value: 31})
	if got != bench.StatusFail {
		t.Fatalf("status = %v, want fail", got)
	}
}

func TestEvaluateBudget_thirtyMegabyteSliceFailsIdleCap(t *testing.T) {
	t.Parallel()
	// Simulates a 30 MB global allocation pushing idle RSS over the 30 MB deal breaker.
	b := bench.Budget{Metric: "idle_rss_mb", Target: 20, Cap: 30, Unit: "MB"}
	got := bench.EvaluateBudget(b, bench.Measurement{Value: 30.01})
	if got != bench.StatusFail {
		t.Fatalf("status = %v, want fail for RSS over cap", got)
	}
}

func TestEvaluateBudget_targetWarn(t *testing.T) {
	t.Parallel()
	b := bench.Budget{Metric: "binary_mb", Target: 15, Cap: 25, Unit: "MB"}
	got := bench.EvaluateBudget(b, bench.Measurement{Value: 18})
	if got != bench.StatusWarn {
		t.Fatalf("status = %v, want warn", got)
	}
}

func TestEvaluateBudget_underTargetPass(t *testing.T) {
	t.Parallel()
	b := bench.Budget{Metric: "js_initial_kb_gz", Target: 60, Cap: 100, Unit: "KB"}
	got := bench.EvaluateBudget(b, bench.Measurement{Value: 50})
	if got != bench.StatusPass {
		t.Fatalf("status = %v, want pass", got)
	}
}

func TestEvaluateBudget_skippedNotPass(t *testing.T) {
	t.Parallel()
	b := bench.Budget{Metric: "scroll_1m_rss_mb", Target: 50, Cap: 80, Unit: "MB"}
	got := bench.EvaluateBudget(b, bench.Measurement{Skipped: true, Reason: "pager not wired"})
	if got != bench.StatusSkipped {
		t.Fatalf("status = %v, want skipped", got)
	}
}

func TestEvaluateBudget_jsInitialCap100(t *testing.T) {
	t.Parallel()
	for _, b := range bench.TabBudgets {
		if b.Metric == "js_initial_kb_gz" {
			got := bench.EvaluateBudget(b, bench.Measurement{Value: 150})
			if got != bench.StatusFail {
				t.Fatalf("150 KB initial JS: status = %v, want fail (cap 100)", got)
			}
			return
		}
	}
	t.Fatal("js_initial_kb_gz not in TabBudgets")
}

func TestAnyCapFail_detectsFailure(t *testing.T) {
	t.Parallel()
	rows := []bench.GateRow{{Status: bench.StatusPass}, {Status: bench.StatusFail}}
	if !bench.AnyCapFail(rows) {
		t.Fatal("expected cap fail")
	}
}
