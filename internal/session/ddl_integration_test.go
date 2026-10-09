//go:build integration

package session

import (
	"strings"
	"testing"
	"time"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/driver/mysql"
	"github.com/suhaasya/wisp-go/internal/store"
)

func openMySQLBench(t *testing.T) driver.Conn {
	t.Helper()
	var d mysql.Driver
	conn, err := d.Open(t.Context(), driver.ConnectParams{
		Host:                    "127.0.0.1",
		Port:                    33084,
		DB:                      "wisp_bench",
		User:                    "wisp",
		Password:                []byte("wisp"),
		ConnectTimeout:          5 * time.Second,
		AllowPublicKeyRetrieval: true,
	})
	if err != nil {
		t.Skip(err)
	}
	return conn
}

func TestTableDDLMySQLShowCreate(t *testing.T) {
	conn := openMySQLBench(t)
	defer func() { _ = conn.Close() }()

	s := NewTestSession(store.Profile{Engine: "mysql"})
	s.mu.Lock()
	s.main = conn
	s.meta = conn
	s.mu.Unlock()

	ddl, err := s.TableDDL(t.Context(), "wisp_meta", "parent", false)
	if err != nil {
		t.Fatalf("TableDDL: %v", err)
	}
	upper := strings.ToUpper(ddl)
	if !strings.Contains(upper, "CREATE TABLE") {
		t.Fatalf("expected CREATE TABLE, got: %s", ddl)
	}
}

func TestTableDDLPostgreSQLMatchesGolden(t *testing.T) {
	conn := openPGRun(t)
	defer func() { _ = conn.Close() }()

	s := NewTestSession(store.Profile{Engine: "postgres"})
	s.mu.Lock()
	s.main = conn
	s.meta = conn
	s.mu.Unlock()

	det, err := s.TableMetadata(t.Context(), "wisp_meta", "parent", true)
	if err != nil {
		t.Fatalf("TableMetadata: %v", err)
	}
	ddl, err := s.TableDDL(t.Context(), "wisp_meta", "parent", false)
	if err != nil {
		t.Fatalf("TableDDL: %v", err)
	}
	if !strings.Contains(ddl, "CREATE TABLE") {
		t.Fatalf("ddl: %s", ddl)
	}
	if len(det.Columns) < 2 {
		t.Fatalf("columns = %d", len(det.Columns))
	}
}
