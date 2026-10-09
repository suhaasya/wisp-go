package server

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestEventHub_streamsHeartbeatAndMemory(t *testing.T) {
	hub := NewEventHub(EventHubConfig{
		HeartbeatEvery: 20 * time.Millisecond,
		MemoryEvery:    20 * time.Millisecond,
		RSSMegabytes:   func() (float64, error) { return 18.5, nil },
	})
	mux := http.NewServeMux()
	mux.Handle("/api/v1/events", hub)
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	ctx, cancel := context.WithTimeout(t.Context(), 500*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL+"/api/v1/events", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "text/event-stream") {
		t.Fatalf("content-type = %q", ct)
	}

	sc := bufio.NewScanner(resp.Body)
	var sawHeartbeat, sawMemory bool
	for sc.Scan() {
		line := sc.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		var ev SSEEvent
		if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &ev); err != nil {
			t.Fatal(err)
		}
		switch ev.Type {
		case "heartbeat":
			sawHeartbeat = true
		case "memory":
			sawMemory = true
			m, ok := ev.Data.(map[string]any)
			if !ok {
				t.Fatalf("memory data type %T", ev.Data)
			}
			if m["go_mb"] != 18.5 {
				t.Fatalf("go_mb = %v", m["go_mb"])
			}
		}
		if sawHeartbeat && sawMemory {
			break
		}
	}
	if !sawHeartbeat || !sawMemory {
		t.Fatalf("heartbeat=%v memory=%v", sawHeartbeat, sawMemory)
	}
}
