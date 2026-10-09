package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/suhaasya/wisp-go/internal/store"
)

// EventBroadcaster pushes SSE events to connected tabs.
type EventBroadcaster interface {
	Broadcast(typ string, data any)
}

// SettingsHandler serves GET/PUT /api/v1/settings.
type SettingsHandler struct {
	store     *store.SettingsStore
	broadcast EventBroadcaster
}

// NewSettingsHandler returns a settings API handler.
func NewSettingsHandler(st *store.SettingsStore, bc EventBroadcaster) *SettingsHandler {
	return &SettingsHandler{store: st, broadcast: bc}
}

// ServeHTTP implements http.Handler.
func (h *SettingsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.get(w, r)
	case http.MethodPut:
		h.put(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *SettingsHandler) get(w http.ResponseWriter, _ *http.Request) {
	doc, err := h.store.DocumentJSON()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "settings_read", "could not read settings", err.Error())
		return
	}
	writeJSON(w, doc)
}

func (h *SettingsHandler) put(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "could not read body", "")
		return
	}
	var doc map[string]any
	if err := json.Unmarshal(body, &doc); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "settings must be valid JSON", err.Error())
		return
	}
	preview, err := h.store.PreviewDocument(doc)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_settings", err.Error(), "")
		return
	}
	if err := validateSettings(preview); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_settings", err.Error(), "")
		return
	}
	if _, err := h.store.CommitDocument(doc); err != nil {
		writeError(w, http.StatusInternalServerError, "settings_write", "could not save settings", err.Error())
		return
	}
	out, err := h.store.DocumentJSON()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "settings_read", "could not read settings", err.Error())
		return
	}
	if h.broadcast != nil {
		h.broadcast.Broadcast("settings", out)
	}
	writeJSON(w, out)
}

func validateSettings(s store.Settings) error {
	switch s.Theme {
	case "light", "dark", "system":
	default:
		return errValidation("theme must be light, dark, or system")
	}
	switch s.Density {
	case "compact", "comfortable":
	default:
		return errValidation("density must be compact or comfortable")
	}
	if s.WindowRows < 50 || s.WindowRows > 500 {
		return errValidation("window_rows out of range")
	}
	if s.WindowsInMemory < 1 || s.WindowsInMemory > 10 {
		return errValidation("windows_in_memory out of range")
	}
	if s.ResultCapMB < 8 || s.ResultCapMB > 128 {
		return errValidation("result_cap_mb out of range")
	}
	if s.KeepAliveSec < 15 || s.KeepAliveSec > 600 {
		return errValidation("keep_alive_sec out of range")
	}
	if s.IdleExitMin < 1 || s.IdleExitMin > 120 {
		return errValidation("idle_exit_min out of range")
	}
	if s.GoMemLimitMB < 32 || s.GoMemLimitMB > 256 {
		return errValidation("go_mem_limit_mb out of range")
	}
	return nil
}

type validationError string

func (e validationError) Error() string { return string(e) }

func errValidation(msg string) error { return validationError(msg) }

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

// writeRawJSON writes server-generated JSON (e.g. pre-encoded pager windows).
func writeRawJSON(w http.ResponseWriter, raw []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// #nosec G705 — body is JSON from the pager encoder, not request input.
	_, _ = w.Write(raw)
}

func writeError(w http.ResponseWriter, status int, code, message, detail string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": code, "message": message, "detail": detail},
	})
}
