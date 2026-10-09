package mover_test

import (
	"path/filepath"
	"testing"

	"github.com/suhaasya/wisp-go/internal/mover"
)

func TestPreviewCSVCommaQuotedFixture(t *testing.T) {
	path := filepath.Join("testdata", "csv", "comma_quoted.csv")
	got, err := mover.PreviewCSV(t.Context(), path, 100)
	if err != nil {
		t.Fatal(err)
	}
	if got.Encoding != "utf-8" {
		t.Fatalf("encoding=%q", got.Encoding)
	}
	if got.Delimiter != "," {
		t.Fatalf("delimiter=%q", got.Delimiter)
	}
	if len(got.Header) != 2 || got.Header[0] != "name" || got.Header[1] != "note" {
		t.Fatalf("header=%v", got.Header)
	}
	if len(got.Rows) != 2 {
		t.Fatalf("rows=%d want 2", len(got.Rows))
	}
	if got.Rows[0][0] != "a" || got.Rows[0][1] != "comma,inside" {
		t.Fatalf("row0=%v", got.Rows[0])
	}
	if got.Rows[1][1] != "line1\nline2" {
		t.Fatalf("row1 note=%q", got.Rows[1][1])
	}
}

func TestPreviewCSVTabDelimiterFixture(t *testing.T) {
	path := filepath.Join("testdata", "csv", "tab.tsv")
	got, err := mover.PreviewCSV(t.Context(), path, 100)
	if err != nil {
		t.Fatal(err)
	}
	if got.Delimiter != "\t" {
		t.Fatalf("delimiter=%q", got.Delimiter)
	}
	if len(got.Rows) != 2 {
		t.Fatalf("rows=%d", len(got.Rows))
	}
}
