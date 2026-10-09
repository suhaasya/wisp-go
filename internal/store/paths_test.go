package store_test

import (
	"path/filepath"
	"testing"

	"github.com/suhaasya/wisp-go/internal/store"
)

func TestPaths_layout(t *testing.T) {
	t.Parallel()
	p := store.PathsFromDirs("/cfg", "/cache")
	if p.ConfigDir != filepath.Join("/cfg", "wisp") {
		t.Fatalf("ConfigDir = %q", p.ConfigDir)
	}
	if p.DataDir != filepath.Join("/cfg", "wisp", "data") {
		t.Fatalf("DataDir = %q", p.DataDir)
	}
	if p.CacheDir != filepath.Join("/cache", "wisp") {
		t.Fatalf("CacheDir = %q", p.CacheDir)
	}
	if p.RuntimeDir != filepath.Join("/cache", "wisp", "runtime") {
		t.Fatalf("RuntimeDir = %q", p.RuntimeDir)
	}
	if p.InstanceJSON() != filepath.Join(p.RuntimeDir, "instance.json") {
		t.Fatalf("InstanceJSON = %q", p.InstanceJSON())
	}
	if p.InstanceLock() != filepath.Join(p.RuntimeDir, "instance.lock") {
		t.Fatalf("InstanceLock = %q", p.InstanceLock())
	}
	if p.HandoffSecret() != filepath.Join(p.RuntimeDir, "handoff.secret") {
		t.Fatalf("HandoffSecret = %q", p.HandoffSecret())
	}
}
