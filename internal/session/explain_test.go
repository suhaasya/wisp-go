package session

import (
	"testing"
)

const samplePGExplain = `[{"Plan":{"Node Type":"Seq Scan","Relation Name":"users","Alias":"","Startup Cost":0.00,"Total Cost":15.30,"Plan Rows":930,"Actual Total Time":0.015,"Actual Rows":1,"Plans":[]},"Planning Time":0.1,"Execution Time":0.02}]`

func TestParsePostgresExplainJSON(t *testing.T) {
	root, err := parsePostgresExplainJSON(samplePGExplain)
	if err != nil {
		t.Fatal(err)
	}
	if root.Node != "Seq Scan" {
		t.Fatalf("node=%q", root.Node)
	}
	if root.Relation != "users" {
		t.Fatalf("relation=%q", root.Relation)
	}
	if root.Cost != "0.00..15.30" {
		t.Fatalf("cost=%q", root.Cost)
	}
	if root.Time == "" {
		t.Fatal("expected time from analyze")
	}
}
