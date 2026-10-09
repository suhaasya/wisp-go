package server_test

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/suhaasya/wisp-go/internal/server"
)

func startSecurityServer(t *testing.T, now func() time.Time) (*httptest.Server, *server.Guard) {
	t.Helper()
	g := server.NewGuardWithClock(0, now)
	mux := http.NewServeMux()
	g.RegisterRoutes(mux, "test", nil, nil, nil, nil, nil, nil, func() (map[string]any, error) { return map[string]any{}, nil })
	mux.HandleFunc("/api/v1/mutate", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	ts := httptest.NewUnstartedServer(g.Wrap(mux))
	ln, err := server.ListenLoopback(0)
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	g.SetPort(port)
	ts.Listener = ln
	ts.Start()
	t.Cleanup(ts.Close)
	return ts, g
}

func hostFor(ts *httptest.Server) string {
	return ts.Listener.Addr().String()
}

func exchangeSession(t *testing.T, ts *httptest.Server, g *server.Guard) *http.Cookie {
	t.Helper()
	url := g.MintLaunchURL()
	tok := strings.Split(url, "t=")[1]
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, ts.URL+"/auth/exchange?t="+tok, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Host = hostFor(ts)
	client := &http.Client{CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("exchange status = %d, want 303", resp.StatusCode)
	}
	for _, c := range resp.Cookies() {
		if c.Name == "wisp_sid" {
			return c
		}
	}
	t.Fatal("no wisp_sid cookie")
	return nil
}

func TestSecurity_rejections(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	ts, g := startSecurityServer(t, func() time.Time { return now })

	sid := exchangeSession(t, ts, g)

	tests := []struct {
		name       string
		method     string
		path       string
		host       string
		cookie     *http.Cookie
		origin     string
		csrf       string
		wantStatus int
	}{
		{name: "wrong_host", method: http.MethodGet, path: "/api/v1/hello", host: "evil.local:" + hostPort(ts), wantStatus: 421},
		{name: "missing_cookie", method: http.MethodGet, path: "/api/v1/hello", host: hostFor(ts), wantStatus: 401},
		{name: "wrong_origin_post", method: http.MethodPost, path: "/api/v1/mutate", host: hostFor(ts), origin: "https://evil.example", csrf: "x", wantStatus: 403, cookie: sid},
		{name: "missing_csrf", method: http.MethodPost, path: "/api/v1/mutate", host: hostFor(ts), origin: "http://127.0.0.1:" + hostPort(ts), wantStatus: 403, cookie: sid},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequestWithContext(t.Context(), tc.method, ts.URL+tc.path, http.NoBody)
			if err != nil {
				t.Fatal(err)
			}
			host := tc.host
			if host == "" {
				host = hostFor(ts)
			}
			req.Host = host
			if tc.cookie != nil {
				req.AddCookie(tc.cookie)
			}
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}
			if tc.csrf != "" {
				req.Header.Set("X-Wisp-CSRF", tc.csrf)
			}
			assertStatus(t, req, tc.wantStatus)
		})
	}

	t.Run("reused_launch_token", func(t *testing.T) {
		url := g.MintLaunchURL()
		tok := strings.Split(url, "t=")[1]
		req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, ts.URL+"/auth/exchange?t="+tok, nil)
		req.Host = hostFor(ts)
		client := &http.Client{CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		}}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
		req2, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, ts.URL+"/auth/exchange?t="+tok, nil)
		req2.Host = hostFor(ts)
		resp2, err := client.Do(req2)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = resp2.Body.Close() }()
		if resp2.StatusCode != http.StatusUnauthorized {
			t.Fatalf("reuse status = %d, want 401", resp2.StatusCode)
		}
	})

	t.Run("expired_launch_token", func(t *testing.T) {
		tick := []time.Time{now}
		ts2, g2 := startSecurityServer(t, func() time.Time { return tick[0] })
		url := g2.MintLaunchURL()
		tok := strings.Split(url, "t=")[1]
		tick[0] = now.Add(61 * time.Second)
		req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, ts2.URL+"/auth/exchange?t="+tok, nil)
		req.Host = hostFor(ts2)
		resp := doRequest(t, req)
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expired status = %d, want 401", resp.StatusCode)
		}
	})
}

func TestSecurity_helloAndCSRF(t *testing.T) {
	ts, g := startSecurityServer(t, time.Now)
	sid := exchangeSession(t, ts, g)

	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, ts.URL+"/api/v1/hello", nil)
	req.Host = hostFor(ts)
	req.AddCookie(sid)
	resp := doRequest(t, req)
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("hello status = %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), `"csrf"`) {
		t.Fatalf("hello body = %s", body)
	}
	assertNoCORS(t, resp)
}

func TestSecurity_crossOriginCannotMutate(t *testing.T) {
	ts, g := startSecurityServer(t, time.Now)
	sid := exchangeSession(t, ts, g)

	req, _ := http.NewRequestWithContext(t.Context(), http.MethodPost, ts.URL+"/api/v1/mutate", http.NoBody)
	req.Host = hostFor(ts)
	req.AddCookie(sid)
	req.Header.Set("Origin", "http://evil.local:"+hostPort(ts))
	resp := doRequest(t, req)
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", resp.StatusCode)
	}
	assertNoCORS(t, resp)
}

func TestSecurity_validMutation(t *testing.T) {
	ts, g := startSecurityServer(t, time.Now)
	sid := exchangeSession(t, ts, g)
	csrf := fetchCSRF(t, ts, sid, g)

	req, _ := http.NewRequestWithContext(t.Context(), http.MethodPost, ts.URL+"/api/v1/mutate", http.NoBody)
	req.Host = hostFor(ts)
	req.AddCookie(sid)
	req.Header.Set("Origin", "http://127.0.0.1:"+hostPort(ts))
	req.Header.Set("X-Wisp-CSRF", csrf)
	resp := doRequest(t, req)
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
}

func TestSecurity_headers(t *testing.T) {
	ts, g := startSecurityServer(t, time.Now)
	sid := exchangeSession(t, ts, g)
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, ts.URL+"/api/v1/hello", nil)
	req.Host = hostFor(ts)
	req.AddCookie(sid)
	resp := doRequest(t, req)
	defer func() { _ = resp.Body.Close() }()
	if resp.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("missing nosniff")
	}
	if resp.Header.Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("unexpected CORS header")
	}
	if cc := resp.Header.Get("Cache-Control"); cc != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", cc)
	}
}

func fetchCSRF(t *testing.T, ts *httptest.Server, sid *http.Cookie, g *server.Guard) string {
	t.Helper()
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, ts.URL+"/api/v1/hello", nil)
	req.Host = hostFor(ts)
	req.AddCookie(sid)
	resp := doRequest(t, req)
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	const prefix = `"csrf":"`
	i := strings.Index(string(body), prefix)
	if i < 0 {
		t.Fatalf("csrf not in %s", body)
	}
	rest := string(body)[i+len(prefix):]
	end := strings.IndexByte(rest, '"')
	return rest[:end]
}

func hostPort(ts *httptest.Server) string {
	_, port, _ := net.SplitHostPort(ts.Listener.Addr().String())
	return port
}

func assertStatus(t *testing.T, req *http.Request, want int) {
	t.Helper()
	resp := doRequest(t, req)
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != want {
		t.Fatalf("status = %d, want %d", resp.StatusCode, want)
	}
	assertNoCORS(t, resp)
}

func doRequest(t *testing.T, req *http.Request) *http.Response {
	t.Helper()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func assertNoCORS(t *testing.T, resp *http.Response) {
	t.Helper()
	if v := resp.Header.Get("Access-Control-Allow-Origin"); v != "" {
		t.Fatalf("Access-Control-Allow-Origin = %q", v)
	}
}
