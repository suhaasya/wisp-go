package store_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/suhaasya/wisp-go/internal/store"
)

func TestSettingsStore_missingFileReturnsDefaults(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	st := store.NewSettingsStore(path)
	got, err := st.Load()
	if err != nil {
		t.Fatal(err)
	}
	want := store.DefaultSettings()
	if got.Theme != want.Theme || got.WindowRows != want.WindowRows {
		t.Fatalf("got %+v, want defaults theme=%s rows=%d", got, want.Theme, want.WindowRows)
	}
}

func TestSettingsStore_roundTripAndMode0600(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("Unix file mode bits are not enforced on Windows temp files")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	st := store.NewSettingsStore(path)
	s := store.DefaultSettings()
	s.Theme = "dark"
	s.WindowRows = 200
	if err := st.Save(s); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %o, want 0600", info.Mode().Perm())
	}
	loaded, err := st.Load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Theme != "dark" || loaded.WindowRows != 200 {
		t.Fatalf("loaded %+v", loaded)
	}
}

func TestSettingsStore_preservesUnknownKeys(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	raw := `{"version":1,"theme":"light","future_flag":true,"nested":{"a":1}}`
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	st := store.NewSettingsStore(path)
	if _, err := st.Load(); err != nil {
		t.Fatal(err)
	}
	if err := st.Save(st.Current()); err != nil {
		t.Fatal(err)
	}
	out, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatal(err)
	}
	if _, ok := m["future_flag"]; !ok {
		t.Fatalf("future_flag missing: %s", out)
	}
	if _, ok := m["nested"]; !ok {
		t.Fatalf("nested missing: %s", out)
	}
}

func TestSettingsStore_corruptFileKeepsLastGood(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	st := store.NewSettingsStore(path)
	s := store.DefaultSettings()
	s.Theme = "dark"
	if err := st.Save(s); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := st.Load()
	if err == nil {
		t.Fatal("expected load error for corrupt file")
	}
	if st.Current().Theme != "dark" {
		t.Fatalf("last good theme = %q, want dark", st.Current().Theme)
	}
}
