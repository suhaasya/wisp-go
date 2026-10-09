package update

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReplaceExecutable_roundTrip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	exe := filepath.Join(dir, "wisp")
	if err := os.WriteFile(exe, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	newBin := []byte("#!/bin/sh\necho new\n")
	if err := ReplaceExecutable(exe, newBin); err != nil {
		t.Fatalf("ReplaceExecutable: %v", err)
	}
	got, err := os.ReadFile(exe)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(newBin) {
		t.Fatalf("content = %q", got)
	}
}
