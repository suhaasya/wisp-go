package server_test

import (
	"os"
	"testing"
	"time"

	"github.com/suhaasya/wisp-go/internal/server"
	"github.com/suhaasya/wisp-go/internal/store"
)

func TestIdleLifecycle_shutsDownAfterIdleExit(t *testing.T) {
	dir := t.TempDir()
	p := store.PathsFromDirs(dir, dir)
	if err := p.EnsureRuntime(); err != nil {
		t.Fatal(err)
	}
	instancePath := p.InstanceJSON()
	if err := server.WriteInstanceJSON(instancePath, server.InstanceRecord{PID: 1, Port: 1}); err != nil {
		t.Fatal(err)
	}

	idle := 30 * time.Millisecond
	lc := server.NewIdleLifecycle(idle, func() {
		_ = server.RemoveInstanceJSON(instancePath)
	})

	lc.NotifySSEConnected()
	lc.NotifySSEDisconnected()

	select {
	case <-lc.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for idle shutdown")
	}
	if _, err := os.Stat(instancePath); !os.IsNotExist(err) {
		t.Fatalf("instance.json still present: %v", err)
	}
}
