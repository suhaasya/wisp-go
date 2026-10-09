package server

import "net/http"

func applySecurityHeaders(w http.ResponseWriter, path string) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")
	if len(path) >= 5 && path[:5] == "/api/" {
		w.Header().Set("Cache-Control", "no-store")
	}
	w.Header().Del("Access-Control-Allow-Origin")
	w.Header().Del("Access-Control-Allow-Credentials")
}

func cspBase(dev bool) string {
	connect := "connect-src 'self'"
	style := ""
	if dev {
		// Vite injects Tailwind and HMR as inline <style> tags when Go proxies the dev server.
		connect = "connect-src 'self' ws://127.0.0.1:5173 ws://localhost:5173"
		style = "style-src 'self' 'unsafe-inline'; "
	}
	return "default-src 'self'; script-src 'self'; " + style + "img-src 'self' data:; " + connect + "; frame-ancestors 'none'; base-uri 'none'; form-action 'self'"
}

func setCSPDefault(w http.ResponseWriter, dev bool) {
	w.Header().Set("Content-Security-Policy", cspBase(dev))
}

func setCSPWithNonce(w http.ResponseWriter, nonce string, dev bool) {
	w.Header().Set("Content-Security-Policy", cspBase(dev)+"; style-src 'self' 'nonce-"+nonce+"'")
}

type responseWriter struct {
	http.ResponseWriter
	guard *Guard
	path  string
}

func (w *responseWriter) WriteHeader(code int) {
	w.apply()
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.apply()
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) apply() {
	if w.Header().Get("X-Content-Type-Options") != "" {
		return
	}
	applySecurityHeaders(w.ResponseWriter, w.path)
	if w.Header().Get("Content-Security-Policy") == "" {
		setCSPDefault(w.ResponseWriter, w.guard != nil && w.guard.dev)
	}
}
