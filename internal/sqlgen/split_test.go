package sqlgen_test

import (
	"testing"

	"github.com/suhaasya/wisp-go/internal/sqlgen"
)

func TestSplitStatements_semicolonInSingleQuote(t *testing.T) {
	sql := "SELECT ';'; SELECT 2"
	got := sqlgen.SplitStatements(sql)
	want := []sqlgen.StmtRange{{Start: 0, End: 10}, {Start: 12, End: 20}}
	assertRanges(t, sql, got, want)
}

func TestSplitStatements_lineComment(t *testing.T) {
	sql := "SELECT 1; -- ; drop\nSELECT 2"
	got := sqlgen.SplitStatements(sql)
	want := []sqlgen.StmtRange{{Start: 0, End: 8}, {Start: 20, End: 28}}
	assertRanges(t, sql, got, want)
}

func TestSplitStatements_dollarQuote(t *testing.T) {
	sql := "SELECT $tag$; inside $tag$; SELECT 2"
	got := sqlgen.SplitStatements(sql)
	want := []sqlgen.StmtRange{{Start: 0, End: 26}, {Start: 28, End: 36}}
	assertRanges(t, sql, got, want)
}

func TestSplitStatements_delimiter(t *testing.T) {
	sql := "DELIMITER //\nCREATE PROCEDURE p() BEGIN SELECT 1;//\nDELIMITER ;\nSELECT 2;"
	got := sqlgen.SplitStatements(sql)
	if len(got) != 2 {
		t.Fatalf("len=%d want 2: %+v", len(got), got)
	}
	if sql[got[0].Start:got[0].End] != "CREATE PROCEDURE p() BEGIN SELECT 1;" {
		t.Fatalf("first=%q", sql[got[0].Start:got[0].End])
	}
	if sql[got[1].Start:got[1].End] != "SELECT 2" {
		t.Fatalf("second=%q", sql[got[1].Start:got[1].End])
	}
}

func TestStatementAtOffset_picksContaining(t *testing.T) {
	sql := "SELECT 1;\n\nSELECT 2"
	r, ok := sqlgen.StatementAtOffset(sql, 12)
	if !ok {
		t.Fatal("ok=false")
	}
	if sql[r.Start:r.End] != "SELECT 2" {
		t.Fatalf("got %q", sql[r.Start:r.End])
	}
}

func TestSplitStatements_backtickIdentifier(t *testing.T) {
	sql := "SELECT `a;b`; SELECT 2"
	want := []sqlgen.StmtRange{{Start: 0, End: 12}, {Start: 14, End: 22}}
	assertRanges(t, sql, sqlgen.SplitStatements(sql), want)
}

func TestSplitStatements_twoSimple(t *testing.T) {
	sql := "SELECT 1; SELECT 2"
	got := sqlgen.SplitStatements(sql)
	want := []sqlgen.StmtRange{{Start: 0, End: 8}, {Start: 10, End: 18}}
	if len(got) != len(want) {
		t.Fatalf("len=%d want %d", len(got), len(want))
	}
	assertRanges(t, sql, got, want)
}

func assertRanges(t *testing.T, sql string, got, want []sqlgen.StmtRange) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len=%d want %d got=%+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("[%d]=%+v want %+v", i, got[i], want[i])
		}
		if sql[got[i].Start:got[i].End] != sql[want[i].Start:want[i].End] {
			t.Fatalf("slice[%d]=%q want %q", i, sql[got[i].Start:got[i].End], sql[want[i].Start:want[i].End])
		}
	}
}
