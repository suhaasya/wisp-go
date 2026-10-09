package bench

// RowStatus is the outcome of comparing a measurement to a budget.
type RowStatus string

const (
	StatusPass    RowStatus = "pass"
	StatusWarn    RowStatus = "warn"
	StatusFail    RowStatus = "fail"
	StatusSkipped RowStatus = "skipped"
)

// Measurement is an observed or skipped metric value.
type Measurement struct {
	Value   float64 `json:"value,omitempty"`
	Skipped bool    `json:"skipped,omitempty"`
	Reason  string  `json:"reason,omitempty"`
}

// EvaluateBudget compares m against b.
func EvaluateBudget(b Budget, m Measurement) RowStatus {
	if m.Skipped {
		return StatusSkipped
	}
	if m.Value > b.Cap {
		return StatusFail
	}
	if m.Value > b.Target {
		return StatusWarn
	}
	return StatusPass
}

// GateRow is one line in the gate report.
type GateRow struct {
	Metric string
	Value  float64
	Target float64
	Cap    float64
	Unit   string
	Status RowStatus
	Reason string
}

// EvaluateAll compares budgets against results (missing metrics are skipped).
func EvaluateAll(budgets []Budget, results map[string]Measurement) []GateRow {
	rows := make([]GateRow, 0, len(budgets))
	for _, b := range budgets {
		m, ok := results[b.Metric]
		if !ok {
			m = Measurement{Skipped: true, Reason: "not measured"}
		}
		rows = append(rows, GateRow{
			Metric: b.Metric,
			Value:  m.Value,
			Target: b.Target,
			Cap:    b.Cap,
			Unit:   b.Unit,
			Status: EvaluateBudget(b, m),
			Reason: m.Reason,
		})
	}
	return rows
}

// AnyCapFail returns true if any row failed the hard cap.
func AnyCapFail(rows []GateRow) bool {
	for _, r := range rows {
		if r.Status == StatusFail {
			return true
		}
	}
	return false
}
