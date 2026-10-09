package server_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/suhaasya/wisp-go/internal/server"
	"github.com/suhaasya/wisp-go/internal/store"
)

func testRuntimePaths(t *testing.T) server.RuntimePaths {
	t.Helper()
	dir := t.TempDir()
	p := store.PathsFromDirs(dir, dir)
	if err := p.EnsureRuntime(); err != nil {
		t.Fatal(err)
	}
	return server.RuntimePathsFromStore(p.RuntimeDir, p.InstanceJSON(), p.InstanceLock(), p.HandoffSecret())
}

func TestIntegration_startHandoffStop(t *testing.T) {
	paths := testRuntimePaths(t)

	var opened []string
	launcher := server.Launcher{
		Platform: server.PlatformDarwin,
		Run: func(_ context.Context, name string, args ...string) error {
			if len(args) > 0 {
				opened = append(opened, args[len(args)-1])
			}
			return nil
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := server.StartConfig{
		Paths:     paths,
		NoBrowser: true,
		Launcher:  launcher,
		IdleExit:  time.Hour,
		LogLevel:  "debug",
	}

	start := time.Now()
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.RunOrHandoff(ctx, cfg)
	}()

	var rec server.InstanceRecord
	for time.Since(start) < 2*time.Second {
		var err error
		rec, err = server.ReadInstanceJSON(paths.InstanceJSON)
		if err == nil && rec.Port > 0 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if rec.Port == 0 {
		t.Fatal("server did not write instance.json")
	}
	if elapsed := time.Since(start); elapsed > 100*time.Millisecond {
		t.Fatalf("start to listening = %v, want ≤ 100ms", elapsed)
	}

	cfg2 := server.StartConfig{Paths: paths, Launcher: launcher}
	if err := server.RunSecondInstance(context.Background(), paths, cfg2); err != nil {
		t.Fatalf("RunSecondInstance: %v", err)
	}
	if len(opened) != 1 {
		t.Fatalf("opened URLs = %v, want one handoff open", opened)
	}

	cancel()
	select {
	case <-errCh:
	case <-time.After(3 * time.Second):
		t.Fatal("RunOrHandoff did not exit after cancel")
	}

	if _, err := os.Stat(paths.InstanceJSON); !os.IsNotExist(err) {
		t.Fatalf("instance.json still present after shutdown: %v", err)
	}
}
