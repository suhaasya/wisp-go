//go:build integration

package mysql_test

import (
	"context"
	"os"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/suhaasya/wisp-go/internal/bench"
	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/driver/mysql"
)

func myParams(t *testing.T, port int, user, password string, allowPubKey bool) driver.ConnectParams {
	t.Helper()
	host := envOr("WISP_MYSQL_HOST", "127.0.0.1")
	if port == 0 {
		port = envIntOr("WISP_MYSQL_PORT", 33084)
	}
	if user == "" {
		user = envOr("WISP_MYSQL_USER", "wisp")
	}
	if password == "" {
		password = envOr("WISP_MYSQL_PASSWORD", "wisp")
	}
	return driver.ConnectParams{
		Host:                    host,
		Port:                    port,
		DB:                      envOr("WISP_MYSQL_DB", "wisp_bench"),
		User:                    user,
		Password:                []byte(password),
		ConnectTimeout:          5 * time.Second,
		AllowPublicKeyRetrieval: allowPubKey,
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
	return openConnAs(t, port, "", "", true)
}

func openConnAs(t *testing.T, port int, user, password string, allowPubKey bool) driver.Conn {
	t.Helper()
	var d mysql.Driver
	conn, err := d.Open(t.Context(), myParams(t, port, user, password, allowPubKey))
	if err != nil {
		t.Skipf("mysql not available: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func TestConnectAndPing(t *testing.T) {
	for name, port := range mysqlServicePorts() {
		t.Run(name, func(t *testing.T) {
			if name == "mysql57" && runtime.GOARCH == "arm64" {
				t.Skip("mysql:5.7 image is amd64-only")
			}
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

func mysqlServicePorts() map[string]int {
	return map[string]int{
		"mysql57":    33057,
		"mysql80":    33080,
		"mysql84":    33084,
		"mariadb106": 33106,
		"mariadb114": 33114,
	}
}

func TestServerInfoMySQL(t *testing.T) {
	conn := openConn(t, 33084)
	info, err := conn.ServerInfo(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if info.Engine != "mysql" {
		t.Fatalf("engine = %q", info.Engine)
	}
	if info.Version == "" || info.Database == "" || info.User == "" {
		t.Fatalf("info = %+v", info)
	}
}

func TestServerInfoMariaDB(t *testing.T) {
	conn := openConn(t, 33114)
	info, err := conn.ServerInfo(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if info.Engine != "mariadb" {
		t.Fatalf("engine = %q", info.Engine)
	}
}

func TestQueryTypesRoundTrip(t *testing.T) {
	conn := openConn(t, 0)
	rows, err := conn.Query(t.Context(), `SELECT CAST(7 AS SIGNED), CAST(42 AS SIGNED), 3.5, CAST(1 AS UNSIGNED), 'hello', NULL`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	w, err := rows.NextWindow(10)
	if err != nil || w == nil || w.Rows != 1 {
		t.Fatalf("window = %v err=%v", w, err)
	}
	if got, ok := w.CellInt(0, 0); !ok || got != 7 {
		t.Fatalf("int = %d ok=%v", got, ok)
	}
	if got, ok := w.CellInt(1, 0); !ok || got != 42 {
		t.Fatalf("int = %d", got)
	}
	if got, ok := w.CellFloat(2, 0); !ok || got != 3.5 {
		t.Fatalf("float = %v", got)
	}
	if got, ok := w.CellInt(3, 0); !ok || got != 1 {
		t.Fatalf("flag = %d ok=%v", got, ok)
	}
	if got, ok := w.CellText(4, 0); !ok || got != "hello" {
		t.Fatalf("text = %q", got)
	}
	if _, ok := w.CellText(5, 0); ok {
		t.Fatal("expected null")
	}
}

func TestCancelSleep(t *testing.T) {
	conn := openConn(t, 0)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		rows, err := conn.Query(ctx, "SELECT SLEEP(30)")
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
		if err == nil {
			t.Fatal("expected error after cancel")
		}
		if time.Since(start) > 500*time.Millisecond {
			t.Fatalf("cancel took %v", time.Since(start))
		}
	case <-time.After(2 * time.Second):
		t.Fatal("query did not stop")
	}
}

func TestStreamOneMillionRows(t *testing.T) {
	conn := openConn(t, 0)
	q := `WITH RECURSIVE seq AS (
		SELECT 1 AS n
		UNION ALL
		SELECT n + 1 FROM seq WHERE n < 1000000
	) SELECT n FROM seq`
	rows, err := conn.Query(t.Context(), q)
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
				t.Fatalf("row %d missing", r)
			}
			if v <= last {
				t.Fatalf("order %d after %d", v, last)
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
	q := `WITH RECURSIVE seq AS (
		SELECT 1 AS n
		UNION ALL
		SELECT n + 1 FROM seq WHERE n < 1000000
	) SELECT n FROM seq`
	rows, err := conn.Query(t.Context(), q)
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
	if rss > 80 {
		t.Fatalf("RSS %.1f MB exceeds cap 80", rss)
	}
}

func TestAuthCachingSHA2(t *testing.T) {
	conn := openConnAs(t, 33084, "wisp", "wisp", true)
	if _, err := conn.Ping(t.Context()); err != nil {
		t.Fatal(err)
	}
}

func TestAuthNativePassword(t *testing.T) {
	conn := openConnAs(t, 33084, "wisp_native", "wisp", false)
	if _, err := conn.Ping(t.Context()); err != nil {
		t.Fatal(err)
	}
}
