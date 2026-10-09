package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const fsListPrefix = "/api/v1/fs/list"

// FSHandler serves home-scoped directory listings.
type FSHandler struct {
	home string
}

// NewFSHandler returns a file picker API rooted at home.
func NewFSHandler(home string) *FSHandler {
	home = filepath.Clean(home)
	if r, err := filepath.EvalSymlinks(home); err == nil {
		home = r
	}
	return &FSHandler{home: home}
}

type fsEntry struct {
	Name string `json:"name"`
	Dir  bool   `json:"dir"`
}

type fsListResponse struct {
	Path    string    `json:"path"`
	Entries []fsEntry `json:"entries"`
}

// ServeHTTP implements http.Handler.
func (h *FSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet || r.URL.Path != fsListPrefix {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	raw := r.URL.Query().Get("path")
	if raw == "" {
		raw = "~"
	}
	resolved, err := h.resolveHomePath(raw)
	if err != nil {
		writeError(w, http.StatusForbidden, "path_forbidden", "Path must stay under your home directory.", err.Error())
		return
	}
	entries, err := os.ReadDir(resolved)
	if err != nil {
		writeError(w, http.StatusBadRequest, "list_failed", "Could not read directory.", err.Error())
		return
	}
	out := fsListResponse{Path: displayPath(h.home, resolved), Entries: make([]fsEntry, 0, len(entries))}
	for _, e := range entries {
		if e.Name() == "." || e.Name() == ".." {
			continue
		}
		out.Entries = append(out.Entries, fsEntry{Name: e.Name(), Dir: e.IsDir()})
	}
	writeJSON(w, out)
}

// ResolveHomePath maps ~ paths to an absolute path under the user home directory.
func (h *FSHandler) ResolveHomePath(raw string) (string, error) {
	return h.resolveHomePath(raw)
}

func (h *FSHandler) resolveHomePath(raw string) (string, error) {
	p := raw
	if strings.HasPrefix(p, "~/") {
		p = filepath.Join(h.home, strings.TrimPrefix(p, "~/"))
	} else if p == "~" {
		p = h.home
	} else if filepath.IsAbs(p) {
		p = filepath.Clean(p)
	} else {
		p = filepath.Join(h.home, filepath.Clean(p))
	}
	p = filepath.Clean(p)
	real := p
	if r, err := filepath.EvalSymlinks(p); err == nil {
		real = r
	}
	if !pathWithinHome(h.home, real) {
		return "", errPathOutsideHome
	}
	return real, nil
}

var errPathOutsideHome = os.ErrPermission

func pathWithinHome(home, path string) bool {
	home = filepath.Clean(home)
	path = filepath.Clean(path)
	if path == home {
		return true
	}
	rel, err := filepath.Rel(home, path)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func displayPath(home, abs string) string {
	if abs == home {
		return "~"
	}
	if strings.HasPrefix(abs, home+string(filepath.Separator)) {
		return "~/" + strings.TrimPrefix(abs, home+string(filepath.Separator))
	}
	return abs
}
