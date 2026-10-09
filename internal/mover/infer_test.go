package mover_test

import (
	"testing"

	"github.com/suhaasya/wisp-go/internal/mover"
)

func TestInferColumnTypesLiterals(t *testing.T) {
	header := []string{"id", "flag", "note"}
	rows := [][]string{
		{"1", "true", "a"},
		{"2", "false", "b"},
	}
	got := mover.InferColumnTypes(header, rows)
	if len(got) != 3 {
		t.Fatalf("len=%d", len(got))
	}
	if got[0].Type != "bigint" {
		t.Fatalf("id type=%q", got[0].Type)
	}
	if got[1].Type != "boolean" {
		t.Fatalf("flag type=%q", got[1].Type)
	}
	if got[2].Type != "varchar(255)" {
		t.Fatalf("note type=%q", got[2].Type)
	}
}
