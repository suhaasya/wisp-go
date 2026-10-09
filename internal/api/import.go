package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/suhaasya/wisp-go/internal/session"
)

type importPreviewRequest struct {
	Path string `json:"path"`
}

type importStartRequest struct {
	Path        string                        `json:"path"`
	Kind        string                        `json:"kind"`
	Schema      string                        `json:"schema"`
	Table       string                        `json:"table"`
	Delimiter   string                        `json:"delimiter,omitempty"`
	HasHeader   bool                          `json:"has_header"`
	NullMarkers []string                      `json:"null_markers,omitempty"`
	BatchSize   int                           `json:"batch_size"`
	TxMode      string                        `json:"tx_mode,omitempty"`
	StopOnError bool                          `json:"stop_on_error"`
	Confirm     bool                          `json:"confirm"`
	Mapping     []session.ImportColumnMapping `json:"mapping"`
	FileColumns []string                      `json:"file_columns"`
}

func (h *SessionsHandler) postImportPreview(w http.ResponseWriter, r *http.Request, sid string) {
	s, ok := h.mgr.Get(sid)
	if !ok {
		writeError(w, http.StatusNotFound, "session_not_found", "Session not found.", "")
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "could not read body", "")
		return
	}
	var req importPreviewRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid JSON", err.Error())
		return
	}
	if h.fs == nil {
		writeError(w, http.StatusInternalServerError, "import_unavailable", "Import is not configured.", "")
		return
	}
	resolved, err := h.fs.ResolveHomePath(req.Path)
	if err != nil {
		writeError(w, http.StatusForbidden, "path_forbidden", "Path must stay under your home directory.", err.Error())
		return
	}
	s.TouchTabRef(time.Now())
	out, err := s.PreviewImport(r.Context(), session.ImportPreviewRequest{Path: resolved})
	if err != nil {
		writeError(w, http.StatusBadRequest, "import_preview_failed", "Could not preview import file.", err.Error())
		return
	}
	writeJSON(w, out)
}

func (h *SessionsHandler) postImportStart(w http.ResponseWriter, r *http.Request, sid string) {
	s, ok := h.mgr.Get(sid)
	if !ok {
		writeError(w, http.StatusNotFound, "session_not_found", "Session not found.", "")
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "could not read body", "")
		return
	}
	var req importStartRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid JSON", err.Error())
		return
	}
	if h.fs == nil {
		writeError(w, http.StatusInternalServerError, "import_unavailable", "Import is not configured.", "")
		return
	}
	resolved, err := h.fs.ResolveHomePath(req.Path)
	if err != nil {
		writeError(w, http.StatusForbidden, "path_forbidden", "Path must stay under your home directory.", err.Error())
		return
	}
	s.TouchTabRef(time.Now())
	iid, err := s.StartImport(r.Context(), session.ImportStartRequest{
		Path:        resolved,
		Kind:        req.Kind,
		Schema:      req.Schema,
		Table:       req.Table,
		Delimiter:   req.Delimiter,
		HasHeader:   req.HasHeader,
		NullMarkers: req.NullMarkers,
		BatchSize:   req.BatchSize,
		TxMode:      req.TxMode,
		StopOnError: req.StopOnError,
		Confirm:     req.Confirm,
		Mapping:     req.Mapping,
		FileColumns: req.FileColumns,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, "import_start_failed", "Could not start import.", err.Error())
		return
	}
	writeJSON(w, map[string]string{"import_id": iid})
}

func (h *SessionsHandler) postImportCancel(w http.ResponseWriter, r *http.Request, sid, iid string) {
	s, ok := h.mgr.Get(sid)
	if !ok {
		writeError(w, http.StatusNotFound, "session_not_found", "Session not found.", "")
		return
	}
	if err := s.CancelImport(iid); err != nil {
		writeError(w, http.StatusNotFound, "import_not_found", "Import not found.", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SessionsHandler) getImportErrors(w http.ResponseWriter, r *http.Request, sid, iid string) {
	s, ok := h.mgr.Get(sid)
	if !ok {
		writeError(w, http.StatusNotFound, "session_not_found", "Session not found.", "")
		return
	}
	csv, ok := s.ImportErrorsCSV(iid)
	if !ok {
		writeError(w, http.StatusNotFound, "import_not_found", "Import not found.", "")
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="import-errors.csv"`)
	_, _ = w.Write([]byte(csv)) //nolint:gosec // CSV error report generated server-side
	s.ReleaseImport(iid)
}

func (h *SessionsHandler) serveImport(w http.ResponseWriter, r *http.Request, sid, rest string) {
	rest = strings.TrimPrefix(rest, "import/")
	rest = strings.TrimSuffix(rest, "/")
	switch {
	case rest == "preview" && r.Method == http.MethodPost:
		h.postImportPreview(w, r, sid)
	case rest == "start" && r.Method == http.MethodPost:
		h.postImportStart(w, r, sid)
	case strings.HasSuffix(rest, "/cancel") && r.Method == http.MethodPost:
		iid := strings.TrimSuffix(rest, "/cancel")
		iid = strings.TrimSuffix(iid, "/")
		h.postImportCancel(w, r, sid, iid)
	case strings.HasSuffix(rest, "/errors") && r.Method == http.MethodGet:
		iid := strings.TrimSuffix(rest, "/errors")
		iid = strings.TrimSuffix(iid, "/")
		h.getImportErrors(w, r, sid, iid)
	default:
		http.Error(w, "not found", http.StatusNotFound)
	}
}
