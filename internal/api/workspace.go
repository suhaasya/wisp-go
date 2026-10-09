package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/suhaasya/wisp-go/internal/session"
)

const workspacePrefix = "/api/v1/workspace/"

// WorkspaceHandler serves GET/PUT workspace state per profile.
type WorkspaceHandler struct {
	store *session.WorkspaceStore
}

// NewWorkspaceHandler returns the workspace persistence handler.
func NewWorkspaceHandler(store *session.WorkspaceStore) *WorkspaceHandler {
	return &WorkspaceHandler{store: store}
}

// ServeHTTP implements http.Handler.
func (h *WorkspaceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, workspacePrefix) {
		http.NotFound(w, r)
		return
	}
	profileID := strings.TrimPrefix(r.URL.Path, workspacePrefix)
	profileID = strings.Trim(profileID, "/")
	if profileID == "" || strings.Contains(profileID, "/") {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		h.get(w, r, profileID)
	case http.MethodPut:
		h.put(w, r, profileID)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *WorkspaceHandler) get(w http.ResponseWriter, _ *http.Request, profileID string) {
	st, err := h.store.Load(profileID)
	if err != nil {
		if errors.Is(err, session.ErrProfileNotFound) {
			writeError(w, http.StatusNotFound, "profile_not_found", "Profile not found.", "")
			return
		}
		writeError(w, http.StatusBadGateway, "workspace_load_failed", "Could not load workspace.", err.Error())
		return
	}
	writeJSON(w, st)
}

func (h *WorkspaceHandler) put(w http.ResponseWriter, r *http.Request, profileID string) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "could not read body", "")
		return
	}
	var st session.WorkspaceState
	if err := json.Unmarshal(body, &st); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid JSON", err.Error())
		return
	}
	if err := h.store.Save(profileID, st); err != nil {
		if errors.Is(err, session.ErrProfileNotFound) {
			writeError(w, http.StatusNotFound, "profile_not_found", "Profile not found.", "")
			return
		}
		writeError(w, http.StatusBadGateway, "workspace_save_failed", "Could not save workspace.", err.Error())
		return
	}
	writeJSON(w, st)
}
