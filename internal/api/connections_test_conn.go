package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/suhaasya/wisp-go/internal/session"
	"github.com/suhaasya/wisp-go/internal/store"
)

const connectionTestPath = "/api/v1/connections/test"

// ConnectionTestHandler serves POST /connections/test.
type ConnectionTestHandler struct {
	secrets store.Secrets
}

// NewConnectionTestHandler returns a test-connection handler.
func NewConnectionTestHandler(secrets store.Secrets) *ConnectionTestHandler {
	return &ConnectionTestHandler{secrets: secrets}
}

type connectionTestRequest struct {
	ProfileID string  `json:"profile_id"`
	Password  *string `json:"password"`
	store.Profile
}

type connectionTestOK struct {
	OK        bool   `json:"ok"`
	LatencyMS int64  `json:"latency_ms"`
	Server    string `json:"server"`
	TLS       string `json:"tls,omitempty"`
	ViaSSH    bool   `json:"via_ssh"`
}

// ServeHTTP implements http.Handler.
func (h *ConnectionTestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || r.URL.Path != connectionTestPath {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "could not read body", "")
		return
	}
	var req connectionTestRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid JSON", err.Error())
		return
	}
	pw, err := h.resolvePassword(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "secrets", err.Error(), "")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	start := time.Now()
	conn, ping, info, err := session.TestConnect(ctx, req.Profile, pw)
	if err != nil {
		ce := session.MapConnectionError(err)
		writeError(w, http.StatusBadGateway, ce.Code, ce.Message, ce.Detail)
		return
	}
	defer func() { _ = conn.Close() }()
	latency := ping.Milliseconds()
	if latency < 1 {
		latency = time.Since(start).Milliseconds()
	}
	if latency < 1 {
		latency = 1
	}
	label := strings.TrimSpace(info.Engine + " " + info.Version)
	if label == "" {
		label = info.Version
	}
	writeJSON(w, connectionTestOK{
		OK:        true,
		LatencyMS: latency,
		Server:    label,
	})
}

func (h *ConnectionTestHandler) resolvePassword(req connectionTestRequest) ([]byte, error) {
	if req.Password != nil {
		return []byte(*req.Password), nil
	}
	if req.ProfileID != "" && h.secrets != nil {
		b, ok, err := h.secrets.Get(req.ProfileID, store.SecretKindDBPassword)
		if err != nil {
			return nil, err
		}
		if ok {
			return b, nil
		}
	}
	return nil, nil
}
