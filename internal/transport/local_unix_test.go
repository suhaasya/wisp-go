//go:build !windows

package transport_test

import (
	"testing"

	"github.com/suhaasya/wisp-go/internal/transport"
)

func TestParsePostgresSocketPath_filePath(t *testing.T) {
	t.Parallel()
	dir, port, err := transport.ParsePostgresSocketPath("/var/run/postgresql/.s.PGSQL.5432")
	if err != nil {
		t.Fatal(err)
	}
	if dir != "/var/run/postgresql" {
		t.Fatalf("dir = %q", dir)
	}
	if port != 5432 {
		t.Fatalf("port = %d want 5432", port)
	}
}

func TestParsePostgresSocketPath_directory(t *testing.T) {
	t.Parallel()
	dir, port, err := transport.ParsePostgresSocketPath("/var/run/postgresql")
	if err != nil {
		t.Fatal(err)
	}
	if dir != "/var/run/postgresql" {
		t.Fatalf("dir = %q", dir)
	}
	if port != 5432 {
		t.Fatalf("port = %d want default 5432", port)
	}
}

func TestParseMySQLSocketPath(t *testing.T) {
	t.Parallel()
	path, err := transport.ParseMySQLSocketPath("/tmp/mysql.sock")
	if err != nil {
		t.Fatal(err)
	}
	if path != "/tmp/mysql.sock" {
		t.Fatalf("path = %q", path)
	}
}
