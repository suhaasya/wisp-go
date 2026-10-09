package server

import (
	"bytes"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"
	"time"
)

const viteDevOrigin = "http://127.0.0.1:5173"

// Static serves embedded frontend assets or proxies to Vite in dev.
type Static struct {
	FS   fs.FS
	Dev  bool
	Vite *url.URL
}

// NewStatic returns a Static handler for production or dev.
func NewStatic(assetFS fs.FS, dev bool) (*Static, error) {
	u, err := url.Parse(viteDevOrigin)
	if err != nil {
		return nil, err
	}
	return &Static{FS: assetFS, Dev: dev, Vite: u}, nil
}

// Register mounts static routes on mux (inside Guard.Wrap).
func (s *Static) Register(mux *http.ServeMux) {
	if s.Dev {
		proxy := httputil.NewSingleHostReverseProxy(s.Vite)
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			slog.Debug("vite proxy", "path", r.URL.Path, "err", err)
			if s.serveDevFallback(w, r) {
				return
			}
			http.Error(w, "Dev UI unavailable. Run make dev (Vite on :5173) or make web.", http.StatusBadGateway)
		}
		mux.Handle("/", proxy)
		return
	}
	mux.Handle("/assets/", http.HandlerFunc(s.serveAsset))
	mux.HandleFunc("/", s.serveIndex)
}

func (s *Static) serveAsset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	name := strings.TrimPrefix(r.URL.Path, "/")
	name = path.Clean(name)
	if !strings.HasPrefix(name, "assets/") {
		http.NotFound(w, r)
		return
	}
	if _, err := fs.Stat(s.FS, name); err != nil {
		http.NotFound(w, r)
		return
	}
	data, err := fs.ReadFile(s.FS, name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	http.ServeContent(w, r, path.Base(name), modTime(s.FS, name), bytes.NewReader(data))
}

func (s *Static) serveIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	data, err := fs.ReadFile(s.FS, "index.html")
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	nonce, err := randomHex(8)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	html := injectNonce(string(data), nonce)
	w.Header().Set("Cache-Control", "no-store")
	setCSPWithNonce(w, nonce, false)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if r.Method == http.MethodHead {
		w.WriteHeader(http.StatusOK)
		return
	}
	_, _ = w.Write([]byte(html))
}

func modTime(fsys fs.FS, name string) (t time.Time) {
	info, err := fs.Stat(fsys, name)
	if err != nil {
		return t
	}
	return info.ModTime()
}

// serveDevFallback serves the embedded production build when Vite is not reachable.
func (s *Static) serveDevFallback(w http.ResponseWriter, r *http.Request) bool {
	if s.FS == nil {
		return false
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		return false
	}
	if r.URL.Path == "/" || r.URL.Path == "" {
		s.serveIndex(w, r)
		return true
	}
	if strings.HasPrefix(r.URL.Path, "/assets/") {
		name := strings.TrimPrefix(r.URL.Path, "/")
		name = path.Clean(name)
		if _, err := fs.Stat(s.FS, name); err != nil {
			return false
		}
		s.serveAsset(w, r)
		return true
	}
	return false
}

func injectNonce(html, nonce string) string {
	meta := `<meta name="csp-nonce" content="` + nonce + `">`
	if strings.Contains(html, `name="csp-nonce"`) {
		return html
	}
	if i := strings.Index(html, "<head>"); i >= 0 {
		return html[:i+len("<head>")] + meta + html[i+len("<head>"):]
	}
	return meta + html
}
