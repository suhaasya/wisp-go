package transport_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/suhaasya/wisp-go/internal/transport"
)

func certDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../../ci/certs"))
}

func readCert(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(certDir(t), name))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func startTLSServer(t *testing.T, certFile, keyFile string) string {
	t.Helper()
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		t.Fatal(err)
	}
	ln, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{Certificates: []tls.Certificate{cert}})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			if tc, ok := c.(*tls.Conn); ok {
				_ = tc.HandshakeContext(context.Background())
			}
			_ = c.Close()
		}
	}()
	return ln.Addr().String()
}

func dialTLS(t *testing.T, addr string, cfg *tls.Config) error {
	t.Helper()
	ctx := context.Background()
	if cfg == nil {
		var d net.Dialer
		conn, err := d.DialContext(ctx, "tcp", addr)
		if err != nil {
			return err
		}
		return conn.Close()
	}
	conn, err := (&tls.Dialer{Config: cfg}).DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	return conn.Close()
}

func TestBuildTLSConfigVerifyFullOK(t *testing.T) {
	dir := certDir(t)
	addr := startTLSServer(t, filepath.Join(dir, "server.pem"), filepath.Join(dir, "server.key"))
	host, _, _ := net.SplitHostPort(addr)

	cfg, err := transport.BuildTLSConfig(transport.TLSParams{
		Mode:       "verify-full",
		ServerName: "db.test",
		CASource:   "file",
		CAFile:     filepath.Join(dir, "ca.pem"),
	})
	if err != nil {
		t.Fatal(err)
	}
	cfg.ServerName = "db.test"
	if err := dialTLS(t, addr, cfg); err != nil {
		t.Fatalf("dial: %v", err)
	}
	_ = host
}

func TestBuildTLSConfigVerifyFullHostnameMismatch(t *testing.T) {
	dir := certDir(t)
	addr := startTLSServer(t, filepath.Join(dir, "wrong-host.pem"), filepath.Join(dir, "wrong-host.key"))

	cfg, err := transport.BuildTLSConfig(transport.TLSParams{
		Mode:       "verify-full",
		ServerName: "db.test",
		CASource:   "file",
		CAFile:     filepath.Join(dir, "ca.pem"),
	})
	if err != nil {
		t.Fatal(err)
	}
	err = dialTLS(t, addr, cfg)
	if err == nil {
		t.Fatal("expected hostname verification failure")
	}
}

func TestBuildTLSConfigVerifyCAAllowsNameMismatch(t *testing.T) {
	dir := certDir(t)
	addr := startTLSServer(t, filepath.Join(dir, "wrong-host.pem"), filepath.Join(dir, "wrong-host.key"))

	cfg, err := transport.BuildTLSConfig(transport.TLSParams{
		Mode:       "verify-ca",
		ServerName: "db.test",
		CASource:   "file",
		CAFile:     filepath.Join(dir, "ca.pem"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := dialTLS(t, addr, cfg); err != nil {
		t.Fatalf("verify-ca should trust CA-signed cert: %v", err)
	}
}

func TestBuildTLSConfigRequireEncryptsWithoutCA(t *testing.T) {
	dir := certDir(t)
	addr := startTLSServer(t, filepath.Join(dir, "server.pem"), filepath.Join(dir, "server.key"))

	cfg, err := transport.BuildTLSConfig(transport.TLSParams{
		Mode:       "require",
		ServerName: "db.test",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := dialTLS(t, addr, cfg); err != nil {
		t.Fatalf("require should accept server cert: %v", err)
	}
}

func TestBuildTLSConfigDisableNil(t *testing.T) {
	cfg, err := transport.BuildTLSConfig(transport.TLSParams{Mode: "disable"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg != nil {
		t.Fatalf("disable: want nil config, got %+v", cfg)
	}
}

func TestBuildTLSConfigClientCert(t *testing.T) {
	dir := certDir(t)
	certPEM := readCert(t, "server.pem")
	keyPEM, err := os.ReadFile(filepath.Join(dir, "server.key"))
	if err != nil {
		t.Fatal(err)
	}
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatal(err)
	}
	ca := x509.NewCertPool()
	if !ca.AppendCertsFromPEM(readCert(t, "ca.pem")) {
		t.Fatal("ca pem")
	}
	ln, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    ca,
		MinVersion:   tls.VersionTLS12,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			if tc, ok := c.(*tls.Conn); ok {
				_ = tc.HandshakeContext(context.Background())
			}
			_ = c.Close()
		}
	}()
	addr := ln.Addr().String()

	cfg, err := transport.BuildTLSConfig(transport.TLSParams{
		Mode:       "verify-full",
		ServerName: "db.test",
		CASource:   "file",
		CAFile:     filepath.Join(dir, "ca.pem"),
		ClientCert: filepath.Join(dir, "client.pem"),
		ClientKey:  filepath.Join(dir, "client.key"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := dialTLS(t, addr, cfg); err != nil {
		t.Fatalf("mutual TLS: %v", err)
	}
}
