package transport_test

import (
	"testing"

	"github.com/suhaasya/wisp-go/internal/transport"
)

func TestParseWindowsPipeName(t *testing.T) {
	t.Parallel()
	name, err := transport.ParsePipeName(`\\.\pipe\MySQL`)
	if err != nil {
		t.Fatal(err)
	}
	if name != "MySQL" {
		t.Fatalf("name = %q", name)
	}
	name2, err := transport.ParsePipeName("MySQL")
	if err != nil {
		t.Fatal(err)
	}
	if name2 != "MySQL" {
		t.Fatalf("name2 = %q", name2)
	}
}
