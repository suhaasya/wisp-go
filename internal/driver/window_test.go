package driver_test

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/suhaasya/wisp-go/internal/driver"
)

func TestWindowMarshalJSONGolden(t *testing.T) {
	w := driver.NewWindow([]driver.ColumnMeta{
		{Name: "id", Type: "int8"},
		{Name: "status", Type: "order_status"},
	}, 2)
	colID := w.Column(0)
	colID.SetInt(0, 1248920)
	colID.SetInt(1, 1248919)
	colStatus := w.Column(1)
	colStatus.SetText(0, "paid")
	colStatus.SetText(1, "shipped")
	colStatus.SetNull(1)

	got, err := json.Marshal(w)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("testdata/window_golden.json")
	if err != nil {
		t.Fatal(err)
	}
	want = bytes.TrimSpace(want)
	if !bytes.Equal(got, want) {
		t.Fatalf("JSON mismatch\ngot:  %s\nwant: %s", got, want)
	}
}

func TestWindowMarshalJSONTruncatedCell(t *testing.T) {
	w := driver.NewWindow([]driver.ColumnMeta{{Name: "body", Type: "text"}}, 1)
	long := make([]byte, 5000)
	for i := range long {
		long[i] = 'x'
	}
	w.Column(0).SetText(0, string(long))

	got, err := json.Marshal(w)
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Truncated [][]int `json:"truncated"`
		Data      [][]any `json:"data"`
	}
	if err := json.Unmarshal(got, &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.Truncated) != 1 || len(doc.Truncated[0]) != 1 || doc.Truncated[0][0] != 0 {
		t.Fatalf("truncated = %v", doc.Truncated)
	}
	s, ok := doc.Data[0][0].(string)
	if !ok || len(s) != 4096 {
		t.Fatalf("data len = %d want 4096", len(s))
	}
}

func TestWindowMarshalJSONZeroRows(t *testing.T) {
	w := driver.NewWindow([]driver.ColumnMeta{
		{Name: "id", Type: "int8"},
		{Name: "note", Type: "text"},
	}, 0)
	if _, err := w.MarshalJSON(); err != nil {
		t.Fatal(err)
	}
}

func TestWindowMarshalJSONTextNullAfterFilledRows(t *testing.T) {
	const rows = 40
	w := driver.NewWindow([]driver.ColumnMeta{{Name: "note", Type: "text"}}, rows)
	b := w.Column(0)
	for r := 0; r < rows-1; r++ {
		b.SetText(r, "x")
	}
	b.SetNull(rows - 1)
	w.Rows = rows
	if _, err := w.MarshalJSON(); err != nil {
		t.Fatal(err)
	}
}

func BenchmarkWindowMarshal150x10x20(b *testing.B) {
	const rows, cols = 150, 10
	meta := make([]driver.ColumnMeta, cols)
	for i := range meta {
		meta[i] = driver.ColumnMeta{Name: "c", Type: "text"}
	}
	cell := "01234567890123456789"
	w := driver.NewWindow(meta, rows)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		driver.ResetWindow(w, meta, rows)
		for c := 0; c < cols; c++ {
			bld := w.Column(c)
			for r := 0; r < rows; r++ {
				bld.SetText(r, cell)
			}
		}
		if _, err := w.MarshalJSON(); err != nil {
			b.Fatal(err)
		}
	}
}

func TestBenchmarkWindowAllocsBudget(t *testing.T) {
	const maxAllocs = 30
	result := testing.Benchmark(BenchmarkWindowMarshal150x10x20)
	if result.AllocsPerOp() > maxAllocs {
		t.Fatalf("allocs/op = %v want ≤ %d", result.AllocsPerOp(), maxAllocs)
	}
	const rows, cols, strLen = 150, 10, 20
	meta := make([]driver.ColumnMeta, cols)
	for i := range meta {
		meta[i] = driver.ColumnMeta{Name: "c", Type: "text"}
	}
	cell := "01234567890123456789"
	w := driver.NewWindow(meta, rows)
	driver.ResetWindow(w, meta, rows)
	for c := 0; c < cols; c++ {
		bld := w.Column(c)
		for r := 0; r < rows; r++ {
			bld.SetText(r, cell)
		}
	}
	raw := float64(rows * cols * strLen)
	if float64(w.ArenaLen()) > 1.5*raw {
		t.Fatalf("arena bytes %d want ≤ %.0f (raw %v)", w.ArenaLen(), 1.5*raw, raw)
	}
	if _, err := w.MarshalJSON(); err != nil {
		t.Fatal(err)
	}
}
