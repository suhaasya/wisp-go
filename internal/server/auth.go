package server

import (
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

const (
	cookieName       = "wisp_sid"
	csrfHeader       = "X-Wisp-CSRF"
	launchTokenTTL   = 60 * time.Second
	launchTokenBytes = 32
)

// Session holds per-tab session state after launch token exchange.
type Session struct {
	CSRF    string
	Created time.Time
}

type launchEntry struct {
	created time.Time
	used    bool
}

// Guard enforces local-only access, launch tokens, sessions, and CSRF.
type Guard struct {
	port int
	dev  bool

	mu           sync.Mutex
	sessions     map[string]Session
	launchTokens map[string]launchEntry
	clock        func() time.Time
}

// NewGuard returns a Guard using real time.
func NewGuard(port int) *Guard {
	return NewGuardWithClock(port, time.Now)
}

// NewGuardWithClock returns a Guard with an injectable clock (tests).
func NewGuardWithClock(port int, clock func() time.Time) *Guard {
	return &Guard{
		port:         port,
		sessions:     make(map[string]Session),
		launchTokens: make(map[string]launchEntry),
		clock:        clock,
	}
}

// SetPort updates the expected loopback port (after listen).
func (g *Guard) SetPort(port int) {
	g.port = port
}

// SetDev enables development CSP exceptions (Vite HMR).
func (g *Guard) SetDev(dev bool) {
	g.dev = dev
}

// MintLaunchURL creates a one-time launch token URL for the launcher or handoff.
func (g *Guard) MintLaunchURL() string {
	raw, err := randomBytes(launchTokenBytes)
	if err != nil {
		return "http://127.0.0.1:" + itoa(g.port) + "/"
	}
	tok := base64.RawURLEncoding.EncodeToString(raw)
	g.mu.Lock()
	g.launchTokens[tok] = launchEntry{created: g.clock(), used: false}
	g.mu.Unlock()
	return "http://127.0.0.1:" + itoa(g.port) + "/auth/exchange?t=" + tok
}

// CSRF returns the CSRF token for the session cookie on r, or "".
func (g *Guard) CSRF(r *http.Request) string {
	sid := sessionIDFromRequest(r)
	if sid == "" {
		return ""
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	s, ok := g.sessions[sid]
	if !ok {
		return ""
	}
	return s.CSRF
}

// RegisterRoutes mounts bootstrap handlers on mux (inside Guard.Wrap).
func (g *Guard) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/auth/exchange", g.handleExchange)
	mux.HandleFunc("/api/v1/hello", g.handleHello)
}

func (g *Guard) handleExchange(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	tok := r.URL.Query().Get("t")
	if tok == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if !g.redeemLaunchToken(tok) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	sid, err := randomHex(16)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	csrf, err := randomHex(16)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	g.mu.Lock()
	g.sessions[sid] = Session{CSRF: csrf, Created: g.clock()}
	g.mu.Unlock()

	http.SetCookie(w, &http.Cookie{ //nolint:gosec // G124: loopback-only; Secure omitted intentionally
		Name:     cookieName,
		Value:    sid,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (g *Guard) redeemLaunchToken(tok string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	var matchedKey string
	var entry launchEntry
	for k, e := range g.launchTokens {
		if len(k) != len(tok) {
			continue
		}
		if subtle.ConstantTimeCompare([]byte(k), []byte(tok)) == 1 {
			matchedKey = k
			entry = e
			break
		}
	}
	if matchedKey == "" {
		pad := make([]byte, len(tok))
		_ = subtle.ConstantTimeCompare([]byte(tok), pad)
		return false
	}
	if entry.used || g.clock().Sub(entry.created) > launchTokenTTL {
		return false
	}
	entry.used = true
	g.launchTokens[matchedKey] = entry
	return true
}

func (g *Guard) handleHello(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	csrf := g.CSRF(r)
	if csrf == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"csrf": csrf})
}

func (g *Guard) sessionValid(r *http.Request) bool {
	sid := sessionIDFromRequest(r)
	if sid == "" {
		return false
	}
	g.mu.Lock()
	_, ok := g.sessions[sid]
	g.mu.Unlock()
	return ok
}

func sessionIDFromRequest(r *http.Request) string {
	c, err := r.Cookie(cookieName)
	if err != nil {
		return ""
	}
	return c.Value
}

func itoa(n int) string {
	return strconvItoa(n)
}
