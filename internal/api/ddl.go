package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/suhaasya/wisp-go/internal/session"
)

func (h *SessionsHandler) serveDDL(w http.ResponseWriter, r *http.Request, sid, sub string) {
	switch sub {
	case "preview":
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.postDDLPreview(w, r, sid)
	case "execute":
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.postDDLExecute(w, r, sid)
	default:
		http.Error(w, "not found", http.StatusNotFound)
	}
}

func (h *SessionsHandler) postDDLPreview(w http.ResponseWriter, r *http.Request, sid string) {
	s, ok := h.mgr.Get(sid)
	if !ok {
		writeError(w, http.StatusNotFound, "session_not_found", "Session not found.", "")
		return
	}
	req, err := decodeDDLRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", err.Error(), "")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	out, err := s.PreviewDDL(ctx, req)
	if err != nil {
		writeChangeError(w, err)
		return
	}
	s.TouchTabRef(time.Now())
	writeJSON(w, out)
}

type ddlExecuteBody struct {
	session.DDLPreviewRequest
	Confirm       bool   `json:"confirm"`
	ObjectConfirm string `json:"object_confirm"`
}

func (h *SessionsHandler) postDDLExecute(w http.ResponseWriter, r *http.Request, sid string) {
	s, ok := h.mgr.Get(sid)
	if !ok {
		writeError(w, http.StatusNotFound, "session_not_found", "Session not found.", "")
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "Could not read body.", "")
		return
	}
	var req ddlExecuteBody
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON.", err.Error())
		return
	}
	if req.Action == "" {
		writeError(w, http.StatusBadRequest, "invalid_body", "action required", "")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()
	err = s.ExecuteDDL(ctx, req.DDLPreviewRequest, session.DDLExecuteOptions{
		Confirm: req.Confirm, ObjectConfirm: req.ObjectConfirm,
	})
	if err != nil {
		writeChangeError(w, err)
		return
	}
	s.TouchTabRef(time.Now())
	writeJSON(w, map[string]bool{"ok": true})
}

func decodeDDLRequest(r *http.Request) (session.DDLPreviewRequest, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return session.DDLPreviewRequest{}, err
	}
	var raw struct {
		Action string `json:"action"`
		session.DDLPreviewRequest
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return session.DDLPreviewRequest{}, err
	}
	req := raw.DDLPreviewRequest
	if raw.Action != "" {
		req.Action = session.DDLAction(raw.Action)
	}
	if req.Action == "" {
		return session.DDLPreviewRequest{}, errDDLActionRequired
	}
	return req, nil
}

var errDDLActionRequired = ddlErr("action required")

type ddlErr string

func (e ddlErr) Error() string { return string(e) }
