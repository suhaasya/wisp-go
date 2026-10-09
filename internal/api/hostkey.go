package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/suhaasya/wisp-go/internal/session"
)

const hostKeyAcceptPath = "/api/v1/hostkeys/accept"

// HostKeyAcceptHandler appends a trusted host key to known_hosts.
type HostKeyAcceptHandler struct{}

// NewHostKeyAcceptHandler returns the accept handler.
func NewHostKeyAcceptHandler() *HostKeyAcceptHandler {
	return &HostKeyAcceptHandler{}
}

type hostKeyAcceptRequest struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Fingerprint string `json:"fingerprint"`
}

// ServeHTTP implements http.Handler.
func (h *HostKeyAcceptHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || r.URL.Path != hostKeyAcceptPath {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "could not read body", "")
		return
	}
	var req hostKeyAcceptRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid JSON", err.Error())
		return
	}
	if req.Host == "" || req.Fingerprint == "" {
		writeError(w, http.StatusBadRequest, "invalid_body", "host and fingerprint required", "")
		return
	}
	if err := session.AcceptHostKey(req.Host, req.Port, req.Fingerprint); err != nil {
		writeError(w, http.StatusBadRequest, "host_key_expired", "Host key prompt expired. Test the connection again.", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeHostKeyNeeded(w http.ResponseWriter, prompt *session.HostKeyPrompt) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadGateway)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]any{
			"code":    "host_key_needed",
			"message": "The SSH host key is not trusted yet.",
			"host_key": map[string]any{
				"host":        prompt.Host,
				"port":        prompt.Port,
				"key_type":    prompt.KeyType,
				"fingerprint": prompt.Fingerprint,
			},
		},
	})
}
