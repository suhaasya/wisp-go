package mover_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/suhaasya/wisp-go/internal/mover"
)

func TestFormatCSVFieldLiterals(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"plain", "plain"},
		{"a,b", `"a,b"`},
		{`he said "hi"`, `"he said ""hi"""`},
		{"line1\nline2", "\"line1\nline2\""},
		{"", ""},
	}
	for _, tc := range tests {
		if got := mover.FormatCSVField(tc.in); got != tc.want {
			t.Fatalf("FormatCSVField(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestCSVWriterRowGoldenLine(t *testing.T) {
	var buf bytes.Buffer
	w := mover.NewCSVWriter(&buf)
	if err := w.WriteRow([]string{"a,b", `x"y`, "ok"}); err != nil {
		t.Fatal(err)
	}
	want := `"a,b","x""y",ok` + "\n"
	if buf.String() != want {
		t.Fatalf("got %q want %q", buf.String(), want)
	}
}

func TestCSVWriterRowJoinsWithComma(t *testing.T) {
	var buf bytes.Buffer
	w := mover.NewCSVWriter(&buf)
	_ = w.WriteRow([]string{"1", "2", "3"})
	if !strings.HasPrefix(buf.String(), "1,2,3\n") {
		t.Fatalf("got %q", buf.String())
	}
}
