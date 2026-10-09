package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// EventHub streams SSE events to connected tabs and tracks client count for idle exit.
type EventHub struct {
	dev bool

	mu      sync.Mutex
	clients int
	subs    map[chan []byte]struct{}

	heartbeatEvery time.Duration
	memoryEvery    time.Duration
	rssFn          func() (float64, error)
	now            func() time.Time
}

// EventHubConfig configures tick intervals and RSS sampling.
type EventHubConfig struct {
	Dev            bool
	HeartbeatEvery time.Duration
	MemoryEvery    time.Duration
	RSSMegabytes   func() (float64, error)
	Clock          func() time.Time
}

// NewEventHub returns a hub wired to SSEActivityHooks for idle lifecycle.
func NewEventHub(cfg EventHubConfig) *EventHub {
	hb := cfg.HeartbeatEvery
	if hb == 0 {
		hb = 15 * time.Second
	}
	mem := cfg.MemoryEvery
	if mem == 0 {
		mem = 5 * time.Second
	}
	rss := cfg.RSSMegabytes
	if rss == nil {
		rss = func() (float64, error) { return 0, nil }
	}
	now := cfg.Clock
	if now == nil {
		now = time.Now
	}
	return &EventHub{
		dev:            cfg.Dev,
		subs:           map[chan []byte]struct{}{},
		heartbeatEvery: hb,
		memoryEvery:    mem,
		rssFn:          rss,
		now:            now,
	}
}

// ClientCount is the number of active SSE connections.
func (h *EventHub) ClientCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.clients
}

func (h *EventHub) registerSub(ch chan []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.subs == nil {
		h.subs = map[chan []byte]struct{}{}
	}
	h.subs[ch] = struct{}{}
}

func (h *EventHub) unregisterSub(ch chan []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.subs, ch)
}

// Broadcast sends an SSE payload to all connected tabs.
func (h *EventHub) Broadcast(typ string, data any) {
	payload, err := json.Marshal(SSEEvent{Type: typ, Data: data})
	if err != nil {
		return
	}
	line := []byte(fmt.Sprintf("data: %s\n\n", payload))
	h.mu.Lock()
	subs := make([]chan []byte, 0, len(h.subs))
	for ch := range h.subs {
		subs = append(subs, ch)
	}
	h.mu.Unlock()
	for _, ch := range subs {
		select {
		case ch <- line:
		default:
		}
	}
}

// AddClient registers a new SSE client.
func (h *EventHub) AddClient() {
	h.mu.Lock()
	wasZero := h.clients == 0
	h.clients++
	h.mu.Unlock()
	if wasZero && SSEActivityHooks.OnConnected != nil {
		SSEActivityHooks.OnConnected()
	}
}

// RemoveClient unregisters an SSE client.
func (h *EventHub) RemoveClient() {
	h.mu.Lock()
	if h.clients > 0 {
		h.clients--
	}
	nowZero := h.clients == 0
	h.mu.Unlock()
	if nowZero && SSEActivityHooks.OnDisconnected != nil {
		SSEActivityHooks.OnDisconnected()
	}
}

// SSEEvent is the JSON payload on the events stream.
type SSEEvent struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

// ServeHTTP implements GET /api/v1/events.
func (h *EventHub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Connection", "keep-alive")

	h.AddClient()
	defer h.RemoveClient()

	events := make(chan []byte, 8)
	h.registerSub(events)
	defer h.unregisterSub(events)

	ctx := r.Context()
	heartbeat := time.NewTicker(h.heartbeatEvery)
	defer heartbeat.Stop()
	memory := time.NewTicker(h.memoryEvery)
	defer memory.Stop()

	writeLine := func(line []byte) {
		_, _ = w.Write(line)
		flusher.Flush()
	}

	h.writeEvent(w, flusher, "heartbeat", map[string]any{"ts": h.now().UTC().Format(time.RFC3339)})

	for {
		select {
		case <-ctx.Done():
			return
		case line := <-events:
			writeLine(line)
		case <-heartbeat.C:
			h.writeEvent(w, flusher, "heartbeat", map[string]any{"ts": h.now().UTC().Format(time.RFC3339)})
		case <-memory.C:
			mb, err := h.rssFn()
			if err != nil {
				mb = 0
			}
			h.writeEvent(w, flusher, "memory", map[string]any{"go_mb": mb})
		}
	}
}

func (h *EventHub) writeEvent(w http.ResponseWriter, flusher http.Flusher, typ string, data any) {
	payload, err := json.Marshal(SSEEvent{Type: typ, Data: data})
	if err != nil {
		return
	}
	_, _ = fmt.Fprintf(w, "data: %s\n\n", payload)
	flusher.Flush()
}

// BroadcastShutdown notifies all connected clients (called during server shutdown).
func (h *EventHub) BroadcastShutdown() {
	h.Broadcast("shutdown", map[string]any{})
}
