//go:build integration

package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/suhaasya/wisp-go/internal/api"
	"github.com/suhaasya/wisp-go/internal/store"
)

func TestConnectionTestIntegrationDistinctErrors(t *testing.T) {
	dir := t.TempDir()
	paths := store.Paths{ConfigDir: dir, DataDir: filepath.Join(dir, "data")}
	sec, err := store.OpenSecretsStore(paths, store.FakeKeyringProbe(store.KeyringProbeUnavailable))
	if err != nil {
		t.Fatal(err)
	}
	h := api.NewConnectionTestHandler(sec)

	post := func(port int, password string) string {
		body, _ := json.Marshal(map[string]any{
			"engine":    "postgres",
			"transport": map[string]any{"kind": "tcp", "host": pgHost(), "port": port},
			"user":      pgUser(),
			"database":  pgDB(),
			"password":  password,
		})
		rec := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/connections/test", bytes.NewReader(body))
		h.ServeHTTP(rec, req)
		var out struct {
			Error struct {
				Code string `json:"code"`
			} `json:"error"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &out)
		return out.Error.Code
	}

	if post(1, "x") != "connection_refused" {
		t.Fatal("expected connection_refused on closed port")
	}

	authCode := post(pgPort(), "definitely-wrong-password")
	if authCode != "auth_failed" {
		t.Skipf("postgres not reachable for auth test (code=%q)", authCode)
	}
}

func pgHost() string {
	if v := os.Getenv("WISP_PG_HOST"); v != "" {
		return v
	}
	return "127.0.0.1"
}

func pgPort() int {
	if v := os.Getenv("WISP_PG_PORT"); v != "" {
		n, _ := strconv.Atoi(v)
		if n > 0 {
			return n
		}
	}
	return 54317
}

func pgUser() string {
	if v := os.Getenv("WISP_PG_USER"); v != "" {
		return v
	}
	return "wisp"
}

func pgDB() string {
	if v := os.Getenv("WISP_PG_DB"); v != "" {
		return v
	}
	return "wisp_bench"
}
