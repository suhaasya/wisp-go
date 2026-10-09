package session

import (
	"context"
	"crypto/x509"
	"errors"
	"net"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5/pgconn"
	sshtransport "github.com/suhaasya/wisp-go/internal/transport/ssh"
)

// ConnectionError is a user-facing connection failure with technical detail.
type ConnectionError struct {
	Code    string
	Message string
	Detail  string
}

// MapConnectionError maps driver/network errors to plain language.
func MapConnectionError(err error) ConnectionError {
	if err == nil {
		return ConnectionError{Code: "unknown", Message: "Connection failed.", Detail: ""}
	}
	var dns *net.DNSError
	if errors.As(err, &dns) {
		return ConnectionError{
			Code:    "dns_failure",
			Message: "Could not resolve the database hostname.",
			Detail:  dns.Error(),
		}
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return ConnectionError{
			Code:    "timeout",
			Message: "The connection timed out.",
			Detail:  err.Error(),
		}
	}
	var op *net.OpError
	if errors.As(err, &op) {
		if strings.Contains(strings.ToLower(op.Err.Error()), "connection refused") {
			return ConnectionError{
				Code:    "connection_refused",
				Message: "The server refused the connection.",
				Detail:  op.Error(),
			}
		}
		if strings.Contains(strings.ToLower(op.Err.Error()), "i/o timeout") ||
			strings.Contains(strings.ToLower(op.Err.Error()), "timeout") {
			return ConnectionError{
				Code:    "timeout",
				Message: "The connection timed out.",
				Detail:  op.Error(),
			}
		}
	}
	var pg *pgconn.PgError
	if errors.As(err, &pg) {
		switch pg.Code {
		case "28P01", "28000":
			if strings.Contains(strings.ToLower(pg.Message), "ssl") {
				return ConnectionError{
					Code:    "tls_required",
					Message: "The server requires a secure (TLS) connection.",
					Detail:  pg.Message,
				}
			}
			return ConnectionError{
				Code:    "auth_failed",
				Message: "Authentication failed. Check username and password.",
				Detail:  pg.Message,
			}
		case "3D000":
			return ConnectionError{
				Code:    "unknown_database",
				Message: "The database name does not exist on this server.",
				Detail:  pg.Message,
			}
		}
	}
	var my *mysql.MySQLError
	if errors.As(err, &my) {
		switch my.Number {
		case 1045:
			return ConnectionError{
				Code:    "auth_failed",
				Message: "Authentication failed. Check username and password.",
				Detail:  my.Message,
			}
		case 1049:
			return ConnectionError{
				Code:    "unknown_database",
				Message: "The database name does not exist on this server.",
				Detail:  my.Message,
			}
		case 3159:
			return ConnectionError{
				Code:    "tls_required",
				Message: "The server requires a secure (TLS) connection.",
				Detail:  my.Message,
			}
		}
	}
	low := strings.ToLower(err.Error())
	var mismatch *sshtransport.HostKeyMismatch
	if errors.As(err, &mismatch) {
		return ConnectionError{
			Code:    "ssh_host_key_mismatch",
			Message: "The SSH host key changed. Connection blocked for safety.",
			Detail:  err.Error(),
		}
	}
	if strings.Contains(low, "ssh") {
		return ConnectionError{
			Code:    "ssh_failed",
			Message: "SSH tunnel failed.",
			Detail:  err.Error(),
		}
	}
	var hostErr x509.HostnameError
	if errors.As(err, &hostErr) {
		return ConnectionError{
			Code:    "tls_hostname",
			Message: "The server certificate does not match the database hostname.",
			Detail:  err.Error(),
		}
	}
	if strings.Contains(low, "wanted to match") || strings.Contains(low, "not valid for any names") {
		return ConnectionError{
			Code:    "tls_hostname",
			Message: "The server certificate does not match the database hostname.",
			Detail:  err.Error(),
		}
	}
	if strings.Contains(low, "tls") || strings.Contains(low, "ssl") || strings.Contains(low, "certificate") {
		return ConnectionError{
			Code:    "tls_mismatch",
			Message: "TLS settings do not match what the server expects.",
			Detail:  err.Error(),
		}
	}
	return ConnectionError{
		Code:    "connection_failed",
		Message: "Could not connect to the database.",
		Detail:  err.Error(),
	}
}
