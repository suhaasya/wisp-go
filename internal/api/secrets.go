package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/suhaasya/wisp-go/internal/store"
)

// SecretsHandler serves secrets unlock and status (never returns stored secrets).
type SecretsHandler struct {
	store *store.SecretsStore
}

// NewSecretsHandler returns a secrets API handler.
func NewSecretsHandler(st *store.SecretsStore) *SecretsHandler {
	return &SecretsHandler{store: st}
}

// ServeHTTP implements http.Handler.
func (h *SecretsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/api/v1/secrets/unlock" && r.Method == http.MethodPost:
		h.postUnlock(w, r)
	case r.URL.Path == "/api/v1/secrets/status" && r.Method == http.MethodGet:
		h.getStatus(w, r)
	default:
		http.Error(w, "not found", http.StatusNotFound)
	}
}

func (h *SecretsHandler) getStatus(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, map[string]any{
		"backend": string(h.store.Backend),
		"locked":  h.store.MasterLocked(),
	})
}

func (h *SecretsHandler) postUnlock(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 4096))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "could not read body", "")
		return
	}
	var req struct {
		MasterPassword string `json:"master_password"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "unlock body must be valid JSON", err.Error())
		return
	}
	if req.MasterPassword == "" {
		writeError(w, http.StatusBadRequest, "invalid_body", "master_password required", "")
		return
	}
	if h.store.Backend != store.SecretsBackendFile {
		writeJSON(w, map[string]any{"unlocked": true})
		return
	}
	pw := []byte(req.MasterPassword)
	err = h.store.UnlockMaster(pw)
	for i := range pw {
		pw[i] = 0
	}
	if err != nil {
		if store.IsUnlockError(err) {
			writeError(w, http.StatusUnauthorized, "unlock_failed", "wrong master password", "")
			return
		}
		writeError(w, http.StatusInternalServerError, "unlock_failed", "could not unlock secrets", err.Error())
		return
	}
	writeJSON(w, map[string]any{"unlocked": true})
}
