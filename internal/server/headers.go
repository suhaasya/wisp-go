package server

import "net/http"

const baseCSP = "default-src 'self'; script-src 'self'; img-src 'self' data:; connect-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'self'"

func applySecurityHeaders(w http.ResponseWriter, path string) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")
	if len(path) >= 5 && path[:5] == "/api/" {
		w.Header().Set("Cache-Control", "no-store")
	}
	// Never send CORS headers.
	w.Header().Del("Access-Control-Allow-Origin")
	w.Header().Del("Access-Control-Allow-Credentials")
}

func setCSPDefault(w http.ResponseWriter) {
	w.Header().Set("Content-Security-Policy", baseCSP+"; style-src 'self'")
}

func setCSPWithNonce(w http.ResponseWriter, nonce string) {
	w.Header().Set("Content-Security-Policy", baseCSP+"; style-src 'self' 'nonce-"+nonce+"'")
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
		setCSPDefault(w.ResponseWriter)
	}
}
