package api

import "testing"

func TestSplitSchemaTablePath(t *testing.T) {
	schema, table, ok := splitSchemaTablePath("wisp_meta/parent")
	if !ok || schema != "wisp_meta" || table != "parent" {
		t.Fatalf("got %q %q ok=%v", schema, table, ok)
	}
	schema, table, ok = splitSchemaTablePath("wisp%20meta/parent%20v")
	if !ok || schema != "wisp meta" || table != "parent v" {
		t.Fatalf("escaped: got %q %q ok=%v", schema, table, ok)
	}
	if _, _, ok := splitSchemaTablePath("onlyone"); ok {
		t.Fatal("expected fail for single segment")
	}
}
