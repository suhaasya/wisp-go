package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/session"
)

const sessionsPrefix = "/api/v1/sessions"

// SessionsHandler serves database session lifecycle APIs.
type SessionsHandler struct {
	mgr *session.Manager
	fs  *FSHandler
}

// NewSessionsHandler returns the sessions handler.
func NewSessionsHandler(mgr *session.Manager, fs *FSHandler) *SessionsHandler {
	return &SessionsHandler{mgr: mgr, fs: fs}
}

type createSessionRequest struct {
	ProfileID string `json:"profile_id"`
}

type createSessionResponse struct {
	SID        string         `json:"sid"`
	ServerInfo sessionInfoDTO `json:"server_info"`
}

type sessionInfoDTO struct {
	State    string `json:"state"`
	Engine   string `json:"engine"`
	Version  string `json:"version"`
	Database string `json:"database"`
	User     string `json:"user"`
	TLS      string `json:"tls,omitempty"`
	ViaSSH   bool   `json:"via_ssh,omitempty"`
	ReadOnly bool   `json:"read_only"`
	SafeMode bool   `json:"safe_mode"`
	Env      string `json:"env"`
}

// ServeHTTP implements http.Handler.
func (h *SessionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, sessionsPrefix)
	path = strings.TrimPrefix(path, "/")
	switch {
	case path == "" && r.Method == http.MethodPost:
		h.create(w, r)
	case strings.HasSuffix(path, "/info") && r.Method == http.MethodGet:
		id := strings.TrimSuffix(path, "/info")
		id = strings.TrimSuffix(id, "/")
		h.info(w, r, id)
	case path != "" && r.Method == http.MethodDelete:
		h.delete(w, r, path)
	default:
		if i := strings.Index(path, "/"); i > 0 {
			sid, rest := path[:i], path[i+1:]
			switch rest {
			case "run":
				if r.Method == http.MethodPost {
					h.postRun(w, r, sid)
					return
				}
			case "explain":
				if r.Method == http.MethodPost {
					h.postExplain(w, r, sid)
					return
				}
			case "export":
				if r.Method == http.MethodGet {
					h.getExport(w, r, sid)
					return
				}
			case "import/preview", "import/start":
				h.serveImport(w, r, sid, rest)
				return
			case "schemas", "objects", "complete":
				h.serveSchema(w, r, sid, rest)
				return
			default:
				if strings.HasPrefix(rest, "ddl/") {
					h.serveDDL(w, r, sid, strings.TrimPrefix(rest, "ddl/"))
					return
				}
				if strings.HasPrefix(rest, "tables/") {
					h.serveSchema(w, r, sid, rest)
					return
				}
				if rest == "results" || strings.HasPrefix(rest, "results/") {
					h.serveResults(w, r, sid, rest)
					return
				}
				if strings.HasPrefix(rest, "cancel/") && r.Method == http.MethodPost {
					qid := strings.TrimPrefix(rest, "cancel/")
					qid = strings.TrimSuffix(qid, "/")
					h.postCancelRun(w, r, sid, qid)
					return
				}
				if strings.HasPrefix(rest, "import/") {
					h.serveImport(w, r, sid, rest)
					return
				}
				if strings.HasPrefix(rest, "export/") && strings.HasSuffix(rest, "/cancel") && r.Method == http.MethodPost {
					eid := strings.TrimPrefix(rest, "export/")
					eid = strings.TrimSuffix(eid, "/cancel")
					eid = strings.TrimSuffix(eid, "/")
					h.postExportCancel(w, r, sid, eid)
					return
				}
			}
		}
		http.Error(w, "not found", http.StatusNotFound)
	}
}

func (h *SessionsHandler) create(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "could not read body", "")
		return
	}
	var req createSessionRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid JSON", err.Error())
		return
	}
	if req.ProfileID == "" {
		writeError(w, http.StatusBadRequest, "invalid_body", "profile_id required", "")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	s, info, err := h.mgr.Create(ctx, req.ProfileID)
	if err != nil {
		writeError(w, http.StatusBadGateway, "session_open_failed", "Could not open a database session.", err.Error())
		return
	}
	writeJSON(w, createSessionResponse{
		SID:        s.ID(),
		ServerInfo: infoFromSession(s, info),
	})
}

func (h *SessionsHandler) info(w http.ResponseWriter, r *http.Request, id string) {
	s, ok := h.mgr.Get(id)
	if !ok {
		writeError(w, http.StatusNotFound, "session_not_found", "Session not found.", "")
		return
	}
	s.TouchTabRef(time.Now())
	writeJSON(w, infoFromSession(s, s.ServerInfo()))
}

func (h *SessionsHandler) delete(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.mgr.Delete(id); err != nil {
		writeError(w, http.StatusNotFound, "session_not_found", "Session not found.", "")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func infoFromSession(s *session.Session, info driver.ServerInfo) sessionInfoDTO {
	p := s.Profile()
	return sessionInfoDTO{
		State:    string(s.State()),
		Engine:   info.Engine,
		Version:  info.Version,
		Database: info.Database,
		User:     info.User,
		TLS:      info.TLS,
		ViaSSH:   info.ViaSSH,
		ReadOnly: p.ReadOnly,
		SafeMode: p.SafeMode,
		Env:      p.Env,
	}
}
