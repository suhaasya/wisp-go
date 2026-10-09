package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/suhaasya/wisp-go/internal/session"
)

func (h *SessionsHandler) serveServerTools(w http.ResponseWriter, r *http.Request, sid, rest string) {
	switch {
	case rest == "processes" && r.Method == http.MethodPost:
		h.postServerToolsProcesses(w, r, sid)
	case strings.HasPrefix(rest, "processes/") && r.Method == http.MethodPost:
		parts := strings.Split(strings.Trim(strings.TrimPrefix(rest, "processes/"), "/"), "/")
		if len(parts) != 2 || parts[1] != "cancel" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		h.postServerToolsProcessAction(w, r, sid, parts[0])
	case rest == "variables" && r.Method == http.MethodGet:
		h.getServerToolsVariables(w, r, sid)
	case rest == "users" && r.Method == http.MethodGet:
		h.getServerToolsUsers(w, r, sid)
	default:
		http.Error(w, "not found", http.StatusNotFound)
	}
}

func (h *SessionsHandler) postServerToolsProcesses(w http.ResponseWriter, r *http.Request, sid string) {
	s, ok := h.mgr.Get(sid)
	if !ok {
		writeError(w, http.StatusNotFound, "session_not_found", "Session not found.", "")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	out, err := s.OpenServerToolsProcesses(ctx)
	if err != nil {
		if session.IsWriteBlockedReadOnly(err) {
			writeError(w, http.StatusForbidden, "read_only", "Session is read-only.", "")
			return
		}
		writeError(w, http.StatusBadGateway, "process_list_failed", "Could not load process list.", err.Error())
		return
	}
	s.TouchTabRef(time.Now())
	writeJSON(w, out)
}

type serverToolsProcessActionRequest struct {
	Terminate bool `json:"terminate"`
}

func (h *SessionsHandler) postServerToolsProcessAction(w http.ResponseWriter, r *http.Request, sid, pidStr string) {
	s, ok := h.mgr.Get(sid)
	if !ok {
		writeError(w, http.StatusNotFound, "session_not_found", "Session not found.", "")
		return
	}
	pid, err := strconv.ParseInt(pidStr, 10, 64)
	if err != nil || pid <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_pid", "Invalid process id.", "")
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "could not read body", "")
		return
	}
	var req serverToolsProcessActionRequest
	if len(body) > 0 {
		if err := json.Unmarshal(body, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json", "invalid JSON", err.Error())
			return
		}
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	if err := s.CancelServerProcess(ctx, pid, req.Terminate); err != nil {
		if session.IsWriteBlockedReadOnly(err) {
			writeError(w, http.StatusForbidden, "read_only", "Session is read-only.", "")
			return
		}
		writeError(w, http.StatusBadGateway, "process_cancel_failed", "Could not cancel process.", err.Error())
		return
	}
	s.TouchTabRef(time.Now())
	w.WriteHeader(http.StatusNoContent)
}

func (h *SessionsHandler) getServerToolsVariables(w http.ResponseWriter, r *http.Request, sid string) {
	s, ok := h.mgr.Get(sid)
	if !ok {
		writeError(w, http.StatusNotFound, "session_not_found", "Session not found.", "")
		return
	}
	search := strings.TrimSpace(r.URL.Query().Get("q"))
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	vars, err := s.ServerToolsVariables(ctx, search)
	if err != nil {
		writeError(w, http.StatusBadGateway, "variables_failed", "Could not load variables.", err.Error())
		return
	}
	s.TouchTabRef(time.Now())
	writeJSON(w, map[string]any{"variables": vars})
}

func (h *SessionsHandler) getServerToolsUsers(w http.ResponseWriter, r *http.Request, sid string) {
	s, ok := h.mgr.Get(sid)
	if !ok {
		writeError(w, http.StatusNotFound, "session_not_found", "Session not found.", "")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	out, err := s.ServerToolsUsers(ctx)
	if err != nil {
		writeError(w, http.StatusBadGateway, "users_failed", "Could not load users.", err.Error())
		return
	}
	s.TouchTabRef(time.Now())
	writeJSON(w, out)
}
