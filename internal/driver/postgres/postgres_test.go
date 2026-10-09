//go:build integration

package postgres_test

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/suhaasya/wisp-go/internal/bench"
	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/driver/postgres"
)

func pgParams(t *testing.T, port int) driver.ConnectParams {
	t.Helper()
	host := envOr("WISP_PG_HOST", "127.0.0.1")
	if port == 0 {
		port = envIntOr("WISP_PG_PORT", 54317)
	}
	return driver.ConnectParams{
		Host:           host,
		Port:           port,
		DB:             envOr("WISP_PG_DB", "wisp_bench"),
		User:           envOr("WISP_PG_USER", "wisp"),
		Password:       []byte(envOr("WISP_PG_PASSWORD", "wisp")),
		ConnectTimeout: 5 * time.Second,
	}
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envIntOr(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil {
			return n
		}
	}
	return def
}

func openConn(t *testing.T, port int) driver.Conn {
	t.Helper()
	var d postgres.Driver
	conn, err := d.Open(t.Context(), pgParams(t, port))
	if err != nil {
		t.Skipf("postgres not available: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func TestConnectAndPing(t *testing.T) {
	ports := pgVersionPorts()
	for name, port := range ports {
		t.Run(name, func(t *testing.T) {
			conn := openConn(t, port)
			d, err := conn.Ping(t.Context())
			if err != nil {
				t.Fatalf("Ping: %v", err)
			}
			if d <= 0 {
				t.Fatalf("latency = %v", d)
			}
		})
	}
}

func pgVersionPorts() map[string]int {
	return map[string]int{
		"pg12": 54312,
		"pg13": 54313,
		"pg14": 54314,
		"pg15": 54315,
		"pg16": 54316,
		"pg17": 54317,
	}
}

func TestServerInfo(t *testing.T) {
	conn := openConn(t, 0)
	info, err := conn.ServerInfo(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if info.Engine != "postgres" {
		t.Fatalf("engine = %q", info.Engine)
	}
	if info.Version == "" {
		t.Fatal("empty version")
	}
	if info.Database != envOr("WISP_PG_DB", "wisp_bench") {
		t.Fatalf("database = %q", info.Database)
	}
	if info.User != envOr("WISP_PG_USER", "wisp") {
		t.Fatalf("user = %q", info.User)
	}
}

func TestQueryTypesRoundTrip(t *testing.T) {
	conn := openConn(t, 0)
	rows, err := conn.Query(t.Context(), `SELECT 7::int2, 42::int8, 3.5::float8, true, 'hello'::text, NULL::text`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	w, err := rows.NextWindow(10)
	if err != nil {
		t.Fatal(err)
	}
	if w == nil || w.Rows != 1 {
		t.Fatalf("window rows = %v", w)
	}
	if got, ok := w.CellInt(0, 0); !ok || got != 7 {
		t.Fatalf("int2 = %d ok=%v", got, ok)
	}
	if got, ok := w.CellInt(1, 0); !ok || got != 42 {
		t.Fatalf("int8 = %d", got)
	}
	if got, ok := w.CellFloat(2, 0); !ok || got != 3.5 {
		t.Fatalf("float8 = %v", got)
	}
	if got, ok := w.CellBool(3, 0); !ok || !got {
		t.Fatalf("bool = %v ok=%v", got, ok)
	}
	if got, ok := w.CellText(4, 0); !ok || got != "hello" {
		t.Fatalf("text = %q", got)
	}
	if _, ok := w.CellText(5, 0); ok {
		t.Fatal("null column should be missing")
	}
}

func TestCancelPgSleep(t *testing.T) {
	conn := openConn(t, 0)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		rows, err := conn.Query(ctx, "SELECT pg_sleep(30)")
		if err != nil {
			done <- err
			return
		}
		_ = rows.Close()
		done <- nil
	}()
	time.Sleep(50 * time.Millisecond)
	start := time.Now()
	cancel()
	select {
	case err := <-done:
		elapsed := time.Since(start)
		if err == nil {
			t.Fatal("expected cancel error")
		}
		if elapsed > 500*time.Millisecond {
			t.Fatalf("cancel took %v want ≤ 500ms", elapsed)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("query did not finish after cancel")
	}
}

func TestStreamOneMillionRows(t *testing.T) {
	conn := openConn(t, 0)
	rows, err := conn.Query(t.Context(), "SELECT i FROM generate_series(1, 1000000) AS g(i)")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	var count int64
	var last int64
	for {
		w, err := rows.NextWindow(150)
		if err != nil {
			t.Fatal(err)
		}
		if w == nil {
			break
		}
		for r := 0; r < w.Rows; r++ {
			v, ok := w.CellInt(0, r)
			if !ok {
				t.Fatalf("row %d: missing value", r)
			}
			if v <= last {
				t.Fatalf("order: %d after %d", v, last)
			}
			last = v
			count++
		}
	}
	if count != 1_000_000 {
		t.Fatalf("count = %d", count)
	}
}

func TestStreamOneMillionRSSBudget(t *testing.T) {
	conn := openConn(t, 0)
	rows, err := conn.Query(t.Context(), "SELECT i FROM generate_series(1, 1000000) AS g(i)")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	for {
		w, err := rows.NextWindow(150)
		if err != nil {
			t.Fatal(err)
		}
		if w == nil {
			break
		}
	}
	rss, err := bench.RSSMegabytes()
	if err != nil {
		t.Skip("RSS unavailable:", err)
	}
	const capMB = 80
	if rss > capMB {
		t.Fatalf("RSS %.1f MB exceeds scroll cap %.0f MB", rss, capMB)
	}
}
