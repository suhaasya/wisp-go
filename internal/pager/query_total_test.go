package pager

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/driver/mock"
)

func TestUpdateQueryTotalAfterFetch(t *testing.T) {
	t.Parallel()
	r := &result{source: Source{Kind: SourceQuery}, total: Total{Kind: "estimate", Value: 0}}

	updateQueryTotalAfterFetch(r, 0, 1, 150)
	if r.total.Kind != "exact" || r.total.Value != 1 {
		t.Fatalf("single row total = %+v, want exact 1", r.total)
	}

	r.total = Total{Kind: "estimate", Value: 0}
	updateQueryTotalAfterFetch(r, 0, 150, 150)
	if r.total.Kind != "estimate" || r.total.Value != 151 {
		t.Fatalf("full page total = %+v, want estimate 151", r.total)
	}

	r.total = Total{Kind: "estimate", Value: 0}
	updateQueryTotalAfterFetch(r, 0, 0, 150)
	if r.total.Kind != "exact" || r.total.Value != 0 {
		t.Fatalf("empty total = %+v, want exact 0", r.total)
	}
}

func TestManagerQueryWindowTotalAfterFetch(t *testing.T) {
	t.Parallel()
	var d mock.Driver
	conn, err := d.Open(t.Context(), driver.ConnectParams{
		Extra: mock.Config{
			TotalRows: 5,
			Columns:   []driver.ColumnMeta{{Name: "id", Type: "int4"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	m := NewManager(conn, "postgres", func() time.Time { return time.Unix(0, 0) })
	ctx := t.Context()
	open, err := m.Open(ctx, Source{Kind: SourceQuery, SQL: "SELECT id FROM user WHERE id=1"})
	if err != nil {
		t.Fatal(err)
	}
	if open.Total.Value != 0 {
		t.Fatalf("open total = %d, want 0 before first window", open.Total.Value)
	}

	raw, err := m.WindowJSON(ctx, open.RID, 0, 150)
	if err != nil {
		t.Fatal(err)
	}
	var win struct {
		Total Total `json:"total"`
	}
	if err := json.Unmarshal(raw, &win); err != nil {
		t.Fatal(err)
	}
	if win.Total.Kind != "exact" || win.Total.Value != 5 {
		t.Fatalf("window total = %+v, want exact 5", win.Total)
	}
}
