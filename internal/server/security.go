package server

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

func (g *Guard) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &responseWriter{ResponseWriter: w, guard: g, path: r.URL.Path}
		if !g.validHost(r) {
			http.Error(rw, "misdirected request", http.StatusMisdirectedRequest)
			return
		}
		if g.needsAuth(r) && !g.sessionValid(r) {
			http.Error(rw, "unauthorized", http.StatusUnauthorized)
			return
		}
		if g.needsCSRF(r) && !g.checkCSRF(r) {
			http.Error(rw, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(rw, r)
	})
}

func (g *Guard) validHost(r *http.Request) bool {
	host := r.Host
	if i := strings.IndexByte(host, ':'); i < 0 {
		return false
	}
	portPart := host[strings.LastIndexByte(host, ':')+1:]
	if portPart != itoa(g.port) {
		return false
	}
	name := host[:strings.LastIndexByte(host, ':')]
	return name == "127.0.0.1" || name == "localhost"
}

func (g *Guard) needsAuth(r *http.Request) bool {
	p := r.URL.Path
	if p == "/auth/exchange" {
		return false
	}
	if len(p) >= 8 && p[:8] == "/assets/" {
		return false
	}
	return true
}

func (g *Guard) needsCSRF(r *http.Request) bool {
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		return false
	default:
		return true
	}
}

func (g *Guard) checkCSRF(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if !g.allowedOrigin(origin) {
		return false
	}
	want := g.CSRF(r)
	if want == "" {
		return false
	}
	got := r.Header.Get(csrfHeader)
	if got == "" {
		return false
	}
	return subtleStringEqual(got, want)
}

func (g *Guard) allowedOrigin(origin string) bool {
	if origin == "http://127.0.0.1:"+itoa(g.port) {
		return true
	}
	if origin == "http://localhost:"+itoa(g.port) {
		return true
	}
	return false
}

func subtleStringEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
