package api

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

// LaunchURLMinter returns a fresh one-time login URL for the launcher.
type LaunchURLMinter func() string

// HandoffHandler serves POST /internal/handoff for second-instance launch.
type HandoffHandler struct {
	secretPath string
	mint       LaunchURLMinter
}

// NewHandoffHandler validates the handoff secret file and mints login URLs.
func NewHandoffHandler(secretPath string, mint LaunchURLMinter) *HandoffHandler {
	return &HandoffHandler{secretPath: secretPath, mint: mint}
}

// ServeHTTP implements http.Handler.
func (h *HandoffHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	got, ok := bearerToken(r.Header.Get("Authorization"))
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	want, err := os.ReadFile(h.secretPath)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if subtle.ConstantTimeCompare([]byte(got), want) != 1 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	url := h.mint()
	_ = json.NewEncoder(w).Encode(map[string]string{"url": url})
}

func bearerToken(h string) (string, bool) {
	const prefix = "Bearer "
	if !strings.HasPrefix(h, prefix) {
		return "", false
	}
	t := strings.TrimSpace(h[len(prefix):])
	if t == "" {
		return "", false
	}
	return t, true
}
