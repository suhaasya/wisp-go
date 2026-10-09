package session_test

import (
	"context"
	"errors"
	"fmt"
	"net"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/suhaasya/wisp-go/internal/session"
)

func TestMapConnectionErrorDNS(t *testing.T) {
	err := &net.DNSError{IsNotFound: true, Name: "db.example.com"}
	got := session.MapConnectionError(err)
	if got.Code != "dns_failure" {
		t.Fatalf("code = %q", got.Code)
	}
	if got.Message != "Could not resolve the database hostname." {
		t.Fatalf("message = %q", got.Message)
	}
}

func TestMapConnectionErrorRefused(t *testing.T) {
	err := &net.OpError{Op: "dial", Err: errors.New("connection refused")}
	got := session.MapConnectionError(err)
	if got.Code != "connection_refused" {
		t.Fatalf("code = %q", got.Code)
	}
	if got.Message != "The server refused the connection." {
		t.Fatalf("message = %q", got.Message)
	}
}

func TestMapConnectionErrorTimeout(t *testing.T) {
	err := context.DeadlineExceeded
	got := session.MapConnectionError(err)
	if got.Code != "timeout" {
		t.Fatalf("code = %q", got.Code)
	}
	if got.Message != "The connection timed out." {
		t.Fatalf("message = %q", got.Message)
	}
}

func TestMapConnectionErrorAuthFailed(t *testing.T) {
	pg := &pgconn.PgError{Code: "28P01", Message: "password authentication failed for user \"app\""}
	got := session.MapConnectionError(pg)
	if got.Code != "auth_failed" {
		t.Fatalf("code = %q", got.Code)
	}
	if got.Message != "Authentication failed. Check username and password." {
		t.Fatalf("message = %q", got.Message)
	}
}

func TestMapConnectionErrorUnknownDatabase(t *testing.T) {
	pg := &pgconn.PgError{Code: "3D000", Message: `database "missing" does not exist`}
	got := session.MapConnectionError(pg)
	if got.Code != "unknown_database" {
		t.Fatalf("code = %q", got.Code)
	}
	if got.Message != "The database name does not exist on this server." {
		t.Fatalf("message = %q", got.Message)
	}
}

func TestMapConnectionErrorTLSRequired(t *testing.T) {
	pg := &pgconn.PgError{Code: "28000", Message: "no pg_hba.conf entry ... SSL off"}
	got := session.MapConnectionError(pg)
	if got.Code != "tls_required" {
		t.Fatalf("code = %q", got.Code)
	}
}

func TestMapConnectionErrorSSH(t *testing.T) {
	err := fmt.Errorf("ssh: handshake failed: %w", errors.New("host key mismatch"))
	got := session.MapConnectionError(err)
	if got.Code != "ssh_failed" {
		t.Fatalf("code = %q", got.Code)
	}
	if got.Message != "SSH tunnel failed." {
		t.Fatalf("message = %q", got.Message)
	}
}

func TestMapConnectionErrorGeneric(t *testing.T) {
	got := session.MapConnectionError(errors.New("something unexpected"))
	if got.Code != "connection_failed" {
		t.Fatalf("code = %q", got.Code)
	}
}
