//go:build integration

package postgres_test

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/driver/postgres"
	"github.com/suhaasya/wisp-go/internal/session"
	"github.com/suhaasya/wisp-go/internal/store"
)

func certDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../../../ci/certs"))
}

func tlsPGParams(t *testing.T, mode, host string) driver.ConnectParams {
	t.Helper()
	dir := certDir(t)
	p := pgParams(t, envIntOr("WISP_PG_TLS_PORT", 54330))
	p.Host = host
	p.TLS = driver.TLSParams{
		Mode:       mode,
		ServerName: host,
		CASource:   "file",
		CAFile:     filepath.Join(dir, "ca.pem"),
	}
	return p
}

func openTLSConn(t *testing.T, mode, host string) driver.Conn {
	t.Helper()
	d := postgres.Driver{}
	conn, err := d.Open(t.Context(), tlsPGParams(t, mode, host))
	if err != nil {
		t.Skipf("postgres-tls not available (start ci/docker-compose postgres-tls): %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func TestPostgresTLSRequire(t *testing.T) {
	conn := openTLSConn(t, "require", "127.0.0.1")
	info, err := conn.ServerInfo(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if info.TLS == "" {
		t.Fatal("expected TLS details on ServerInfo")
	}
	if !strings.Contains(info.TLS, "TLS") {
		t.Fatalf("tls line: %q", info.TLS)
	}
}

func TestPostgresTLSVerifyFullHostnameFail(t *testing.T) {
	d := postgres.Driver{}
	p := tlsPGParams(t, "verify-full", "127.0.0.1")
	p.TLS.ServerName = "db.test"
	_, err := d.Open(t.Context(), p)
	if err == nil {
		t.Fatal("expected verify-full hostname failure")
	}
	if strings.Contains(err.Error(), "connection refused") {
		t.Skipf("postgres-tls not available: %v", err)
	}
	ce := session.MapConnectionError(err)
	if ce.Code != "tls_hostname" {
		t.Fatalf("code=%q want tls_hostname message=%q", ce.Code, ce.Message)
	}
}

func TestPostgresTLSDisableRejected(t *testing.T) {
	d := postgres.Driver{}
	p := tlsPGParams(t, "disable", "127.0.0.1")
	p.TLS = driver.TLSParams{Mode: "disable", ServerName: "127.0.0.1"}
	_, err := d.Open(t.Context(), p)
	if err == nil {
		t.Fatal("expected TLS-only server to reject plain connection")
	}
}

func TestPostgresTLSVerifyFullOK(t *testing.T) {
	_ = openTLSConn(t, "verify-full", "localhost")
}

func TestPostgresTLSProfileViaSession(t *testing.T) {
	dir := certDir(t)
	prof := store.Profile{
		Engine:   "postgres",
		Database: envOr("WISP_PG_DB", "wisp_bench"),
		User:     envOr("WISP_PG_USER", "wisp"),
		Transport: store.Transport{
			Kind: "tcp",
			Host: "127.0.0.1",
			Port: envIntOr("WISP_PG_TLS_PORT", 54330),
		},
		SSL: store.SSLConfig{
			Mode:     "verify-full",
			CASource: "file",
			CAFile:   filepath.Join(dir, "ca.pem"),
		},
	}
	conn, _, info, err := session.TestConnect(t.Context(), prof, []byte(envOr("WISP_PG_PASSWORD", "wisp")), session.SSHConnectSecrets{})
	if err != nil {
		t.Skipf("postgres-tls not available: %v", err)
	}
	defer conn.Close()
	if info.TLS == "" {
		t.Fatal("missing negotiated TLS on test connect")
	}
}
