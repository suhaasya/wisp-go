package server_test

import (
	"sync/atomic"
	"testing"

	"github.com/suhaasya/wisp-go/internal/server"
)

func TestEventHub_clientCountAndHooks(t *testing.T) {
	var connected, disconnected int32
	server.SSEActivityHooks.OnConnected = func() { atomic.AddInt32(&connected, 1) }
	server.SSEActivityHooks.OnDisconnected = func() { atomic.AddInt32(&disconnected, 1) }

	hub := server.NewEventHub(server.EventHubConfig{Dev: false})
	if hub.ClientCount() != 0 {
		t.Fatalf("initial count = %d", hub.ClientCount())
	}
	hub.AddClient()
	hub.AddClient()
	if hub.ClientCount() != 2 {
		t.Fatalf("count = %d, want 2", hub.ClientCount())
	}
	if atomic.LoadInt32(&connected) != 1 {
		t.Fatalf("OnConnected calls = %d, want 1", connected)
	}
	hub.RemoveClient()
	if hub.ClientCount() != 1 {
		t.Fatalf("count = %d, want 1", hub.ClientCount())
	}
	if atomic.LoadInt32(&disconnected) != 0 {
		t.Fatalf("OnDisconnected early = %d", disconnected)
	}
	hub.RemoveClient()
	if hub.ClientCount() != 0 {
		t.Fatalf("count = %d, want 0", hub.ClientCount())
	}
	if atomic.LoadInt32(&disconnected) != 1 {
		t.Fatalf("OnDisconnected calls = %d, want 1", disconnected)
	}
}
