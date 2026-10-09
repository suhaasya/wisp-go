package mover_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/driver/mock"
	"github.com/suhaasya/wisp-go/internal/mover"
)

func TestStreamQueryMockRowCount(t *testing.T) {
	const total = 2500
	cfg := mock.Config{
		TotalRows: total,
		Columns: []driver.ColumnMeta{
			{Name: "id", Type: "int8"},
			{Name: "note", Type: "text"},
		},
	}
	conn, err := mock.Driver{}.Open(context.Background(), driver.ConnectParams{Extra: cfg})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()

	var buf bytes.Buffer
	n, err := mover.StreamQuery(context.Background(), conn, "SELECT 1", nil, mover.FormatCSV, mover.Options{IncludeHeader: true}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if n != total {
		t.Fatalf("rows=%d want %d", n, total)
	}
	lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
	if len(lines) != total+1 {
		t.Fatalf("csv lines=%d want %d", len(lines), total+1)
	}
}
