package transport

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"strings"
)

// TLSParams holds TLS options for a TCP database connection.
type TLSParams struct {
	Mode       string // disable, prefer, require, verify-ca, verify-full
	ServerName string
	CASource   string // system or file
	CAFile     string
	ClientCert string
	ClientKey  string
}

// NegotiatedTLS describes the TLS session after connect.
type NegotiatedTLS struct {
	Version string
	Cipher  string
}

// BuildTLSConfig returns a client tls.Config for the given mode, or nil for disable.
func BuildTLSConfig(p TLSParams) (*tls.Config, error) {
	mode := strings.ToLower(strings.TrimSpace(p.Mode))
	if mode == "" || mode == "disable" {
		return nil, nil
	}
	if mode == "prefer" {
		// Prefer tries TLS first at the driver layer; config matches require when TLS is attempted.
		mode = "require"
	}
	cfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
		ServerName: strings.TrimSpace(p.ServerName),
	}
	if p.ClientCert != "" && p.ClientKey != "" {
		cert, err := tls.LoadX509KeyPair(p.ClientCert, p.ClientKey)
		if err != nil {
			return nil, fmt.Errorf("tls: client certificate: %w", err)
		}
		cfg.Certificates = []tls.Certificate{cert}
	}
	switch mode {
	case "require":
		cfg.InsecureSkipVerify = true
		cfg.VerifyConnection = func(tls.ConnectionState) error { return nil }
	case "verify-ca", "verify-full":
		roots, err := loadRoots(p)
		if err != nil {
			return nil, err
		}
		cfg.RootCAs = roots
		if mode == "verify-ca" {
			// Default verification runs before VerifyConnection and includes hostname/IP
			// checks. InsecureSkipVerify defers to verifyCAOnly for chain validation only.
			cfg.InsecureSkipVerify = true
			cfg.VerifyConnection = verifyCAOnly(roots)
		}
	default:
		return nil, fmt.Errorf("tls: unknown mode %q", p.Mode)
	}
	return cfg, nil
}

func loadRoots(p TLSParams) (*x509.CertPool, error) {
	src := strings.ToLower(strings.TrimSpace(p.CASource))
	if src == "" || src == "system" {
		pool, err := x509.SystemCertPool()
		if err != nil {
			pool = x509.NewCertPool()
		}
		if p.CAFile != "" {
			if err := appendCAPEM(pool, p.CAFile); err != nil {
				return nil, err
			}
		}
		return pool, nil
	}
	if src == "file" {
		if strings.TrimSpace(p.CAFile) == "" {
			return nil, errors.New("tls: CA file path required")
		}
		pool := x509.NewCertPool()
		if err := appendCAPEM(pool, p.CAFile); err != nil {
			return nil, err
		}
		return pool, nil
	}
	return nil, fmt.Errorf("tls: unknown CA source %q", p.CASource)
}

func appendCAPEM(pool *x509.CertPool, path string) error {
	// #nosec G304 — path comes from the user's profile (file picker), not from request input.
	pem, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("tls: read CA %q: %w", path, err)
	}
	if !pool.AppendCertsFromPEM(pem) {
		return fmt.Errorf("tls: no certificates in %q", path)
	}
	return nil
}

func verifyCAOnly(roots *x509.CertPool) func(tls.ConnectionState) error {
	return func(cs tls.ConnectionState) error {
		if len(cs.PeerCertificates) == 0 {
			return errors.New("tls: server did not send a certificate")
		}
		opts := x509.VerifyOptions{Roots: roots}
		if _, err := cs.PeerCertificates[0].Verify(opts); err != nil {
			return err
		}
		return nil
	}
}

// DescribeNegotiated formats TLS version and cipher for the test line.
func DescribeNegotiated(cs *tls.ConnectionState) NegotiatedTLS {
	if cs == nil || !cs.HandshakeComplete {
		return NegotiatedTLS{}
	}
	ver := tlsVersionName(cs.Version)
	cipher := tls.CipherSuiteName(cs.CipherSuite)
	if ver == "" {
		return NegotiatedTLS{}
	}
	if cipher == "" {
		cipher = "unknown cipher"
	}
	return NegotiatedTLS{Version: ver, Cipher: cipher}
}

func tlsVersionName(v uint16) string {
	switch v {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return ""
	}
}

// FormatNegotiated returns a single string for API responses.
func FormatNegotiated(n NegotiatedTLS) string {
	if n.Version == "" {
		return ""
	}
	if n.Cipher == "" {
		return n.Version
	}
	return n.Version + ", " + n.Cipher
}
