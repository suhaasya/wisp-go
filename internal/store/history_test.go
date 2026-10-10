package store_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/suhaasya/wisp-go/internal/store"
)

func historySearchTimeLimit() time.Duration {
	if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
		return 250 * time.Millisecond
	}
	return 100 * time.Millisecond
}

func testPaths(t *testing.T) store.Paths {
	t.Helper()
	root := t.TempDir()
	return store.PathsFromDirs(root, root)
}

func TestHistoryAppendAndSearch(t *testing.T) {
	paths := testPaths(t)
	prof := store.Profile{ID: "p1", Name: "local", Engine: "postgres"}
	h := store.NewHistoryStore(paths)
	entry := store.HistoryEntry{
		TS:     time.Unix(100, 0).UnixMilli(),
		SQL:    "SELECT * FROM orders",
		MS:     12,
		Rows:   50,
		Status: "ok",
	}
	if err := h.Append(prof, entry); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := h.Search(prof, "orders", &buf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "orders") {
		t.Fatalf("search=%q", buf.String())
	}
}

func TestHistoryRotationKeepsNewestHalf(t *testing.T) {
	paths := testPaths(t)
	prof := store.Profile{ID: "p1", Name: "local", Engine: "postgres"}
	h := store.NewHistoryStore(paths)
	path := paths.HistoryJSONL("p1")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	pad := strings.Repeat("x", store.HistoryRotateSizeBytes/100+1024)
	var b strings.Builder
	for i := 0; i < 100; i++ {
		b.WriteString(`{"ts":1,"sql":"`)
		b.WriteString(pad)
		b.WriteString(`old"}`)
		b.WriteByte('\n')
	}
	for i := 0; i < 100; i++ {
		b.WriteString(`{"ts":2,"sql":"new"}`)
		b.WriteByte('\n')
	}
	data := b.String()
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := h.RotateIfNeeded(prof); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n")
	if len(lines) != 100 {
		t.Fatalf("lines=%d want 100 after rotate", len(lines))
	}
	for _, line := range lines {
		if !strings.Contains(line, "new") {
			t.Fatalf("kept old line: %s", line[:min(40, len(line))])
		}
	}
}

func TestHistoryDisabledWritesNothing(t *testing.T) {
	paths := testPaths(t)
	prof := store.Profile{ID: "p1", DisableQueryHistory: true}
	h := store.NewHistoryStore(paths)
	if err := h.Append(prof, store.HistoryEntry{TS: 1, SQL: "SELECT 1", Status: "ok"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(paths.HistoryJSONL("p1")); !os.IsNotExist(err) {
		t.Fatalf("stat err=%v want not exist", err)
	}
}

func TestHistorySearch50kUnder100ms(t *testing.T) {
	if testing.Short() {
		t.Skip("short")
	}
	paths := testPaths(t)
	prof := store.Profile{ID: "p1"}
	h := store.NewHistoryStore(paths)
	path := paths.HistoryJSONL("p1")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	needle := "UNIQUE_NEEDLE_034"
	for i := 0; i < 50_000; i++ {
		sql := "SELECT 1"
		if i == 25_000 {
			sql = "SELECT " + needle
		}
		line := `{"ts":1,"sql":"` + sql + `","status":"ok"}` + "\n"
		if _, err := f.WriteString(line); err != nil {
			t.Fatal(err)
		}
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	var buf bytes.Buffer
	if err := h.Search(prof, needle, &buf); err != nil {
		t.Fatal(err)
	}
	limit := historySearchTimeLimit()
	if elapsed := time.Since(start); elapsed > limit {
		t.Fatalf("search took %v want <= %v", elapsed, limit)
	}
	if !strings.Contains(buf.String(), needle) {
		t.Fatal("needle not found")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
