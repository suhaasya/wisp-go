package api

import (
	"net/http"
	"strings"

	"github.com/suhaasya/wisp-go/internal/store"
)

// HistoryHandler serves query history search.
type HistoryHandler struct {
	history  *store.HistoryStore
	profiles *store.ProfileStore
}

// NewHistoryHandler returns the history API handler.
func NewHistoryHandler(history *store.HistoryStore, profiles *store.ProfileStore) *HistoryHandler {
	return &HistoryHandler{history: history, profiles: profiles}
}

// ServeHTTP implements GET /api/v1/history.
func (h *HistoryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	profileID := strings.TrimSpace(r.URL.Query().Get("profile"))
	q := r.URL.Query().Get("q")
	if profileID == "" {
		writeError(w, http.StatusBadRequest, "invalid_query", "profile required", "")
		return
	}
	prof, ok := h.profiles.Get(profileID)
	if !ok {
		writeError(w, http.StatusNotFound, "profile_not_found", "Profile not found.", "")
		return
	}
	w.Header().Set("Content-Type", "application/x-ndjson")
	if err := h.history.SearchReader(prof, q, w); err != nil {
		writeError(w, http.StatusBadGateway, "history_failed", "Could not read history.", err.Error())
		return
	}
}
