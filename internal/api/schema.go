package api

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/suhaasya/wisp-go/internal/driver"
)

type schemasResponse struct {
	Databases []string `json:"databases,omitempty"`
	Schemas   []string `json:"schemas,omitempty"`
}

type objectsResponse struct {
	Names []string `json:"names"`
}

func (h *SessionsHandler) serveSchema(w http.ResponseWriter, r *http.Request, sid, rest string) {
	refresh := queryRefresh(r)
	switch {
	case rest == "complete" && r.Method == http.MethodGet:
		h.getCompletionSchema(w, r, sid)
	case rest == "schemas" && r.Method == http.MethodGet:
		h.getSchemas(w, r, sid, refresh)
	case rest == "objects" && r.Method == http.MethodGet:
		h.getObjects(w, r, sid, refresh)
	case strings.HasPrefix(rest, "tables/") && r.Method == http.MethodGet:
		h.getTableDetails(w, r, sid, strings.TrimPrefix(rest, "tables/"), refresh)
	default:
		http.Error(w, "not found", http.StatusNotFound)
	}
}

func queryRefresh(r *http.Request) bool {
	v := strings.EqualFold(r.URL.Query().Get("refresh"), "true")
	if v {
		return true
	}
	v = strings.EqualFold(r.URL.Query().Get("refresh"), "1")
	return v
}

func (h *SessionsHandler) getSchemas(w http.ResponseWriter, r *http.Request, sid string, refresh bool) {
	s, ok := h.mgr.Get(sid)
	if !ok {
		writeError(w, http.StatusNotFound, "session_not_found", "Session not found.", "")
		return
	}
	if refresh {
		s.InvalidateMetadata()
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	dbs, schemas, err := s.SchemaSnapshot(ctx)
	if err != nil {
		writeError(w, http.StatusBadGateway, "schema_list_failed", "Could not list schemas.", err.Error())
		return
	}
	s.TouchTabRef(time.Now())
	writeJSON(w, schemasResponse{Databases: dbs, Schemas: schemas})
}

func (h *SessionsHandler) getObjects(w http.ResponseWriter, r *http.Request, sid string, refresh bool) {
	s, ok := h.mgr.Get(sid)
	if !ok {
		writeError(w, http.StatusNotFound, "session_not_found", "Session not found.", "")
		return
	}
	schema := r.URL.Query().Get("schema")
	kindStr := r.URL.Query().Get("kind")
	if schema == "" || kindStr == "" {
		writeError(w, http.StatusBadRequest, "invalid_query", "schema and kind are required", "")
		return
	}
	kind := driver.ObjectKind(kindStr)
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	names, err := s.ListObjects(ctx, schema, kind, refresh)
	if err != nil {
		writeError(w, http.StatusBadGateway, "object_list_failed", "Could not list objects.", err.Error())
		return
	}
	s.TouchTabRef(time.Now())
	if names == nil {
		names = []string{}
	}
	writeJSON(w, objectsResponse{Names: names})
}

func (h *SessionsHandler) getTableDetails(w http.ResponseWriter, r *http.Request, sid, path string, refresh bool) {
	s, ok := h.mgr.Get(sid)
	if !ok {
		writeError(w, http.StatusNotFound, "session_not_found", "Session not found.", "")
		return
	}
	schema, table, ok := splitSchemaTablePath(path)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_path", "expected /tables/{schema}/{table}", "")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	det, err := s.TableMetadata(ctx, schema, table, refresh)
	if err != nil {
		writeError(w, http.StatusBadGateway, "table_details_failed", "Could not load table details.", err.Error())
		return
	}
	s.TouchTabRef(time.Now())
	writeJSON(w, det)
}

func (h *SessionsHandler) getCompletionSchema(w http.ResponseWriter, r *http.Request, sid string) {
	s, ok := h.mgr.Get(sid)
	if !ok {
		writeError(w, http.StatusNotFound, "session_not_found", "Session not found.", "")
		return
	}
	schema := strings.TrimSpace(r.URL.Query().Get("schema"))
	if schema == "" {
		writeError(w, http.StatusBadRequest, "invalid_query", "schema is required", "")
		return
	}
	var extra []string
	if raw := r.URL.Query().Get("tables"); raw != "" {
		for _, part := range strings.Split(raw, ",") {
			if t := strings.TrimSpace(part); t != "" {
				extra = append(extra, t)
			}
		}
	}
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	out, err := s.CompletionSchema(ctx, schema, extra)
	if err != nil {
		writeError(w, http.StatusBadGateway, "complete_failed", "Could not build completion schema.", err.Error())
		return
	}
	s.TouchTabRef(time.Now())
	writeJSON(w, out)
}

func splitSchemaTablePath(path string) (schema, table string, ok bool) {
	path = strings.Trim(path, "/")
	if path == "" {
		return "", "", false
	}
	i := strings.Index(path, "/")
	if i <= 0 || i >= len(path)-1 {
		return "", "", false
	}
	schema, err := url.PathUnescape(path[:i])
	if err != nil {
		return "", "", false
	}
	table, err = url.PathUnescape(path[i+1:])
	if err != nil || table == "" {
		return "", "", false
	}
	return schema, table, true
}
