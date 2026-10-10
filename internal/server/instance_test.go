package server_test

import (
	"os"
	"runtime"
	"testing"

	"github.com/suhaasya/wisp-go/internal/server"
	"github.com/suhaasya/wisp-go/internal/store"
)

func TestInstanceLock_exclusive(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := store.PathsFromDirs(dir, dir)
	if err := p.EnsureRuntime(); err != nil {
		t.Fatal(err)
	}
	lockPath := p.InstanceLock()

	first, err := server.AcquireInstanceLock(lockPath)
	if err != nil {
		t.Fatalf("first AcquireInstanceLock: %v", err)
	}
	t.Cleanup(func() { _ = first.Release() })

	_, err = server.AcquireInstanceLock(lockPath)
	if err != server.ErrInstanceAlreadyRunning {
		t.Fatalf("second lock err = %v, want ErrInstanceAlreadyRunning", err)
	}
}

func TestWriteInstanceJSON_mode(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("Unix file mode bits are not enforced on Windows temp files")
	}
	dir := t.TempDir()
	path := dir + "/instance.json"
	rec := server.InstanceRecord{PID: 42, Port: 8080}
	if err := server.WriteInstanceJSON(path, rec); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("instance.json mode = %o, want 0600", info.Mode().Perm())
	}
}
