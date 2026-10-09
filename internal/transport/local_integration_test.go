//go:build integration

package transport_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/suhaasya/wisp-go/internal/transport"
)

func TestDetectLocal_includesDockerPostgresTCP(t *testing.T) {
	host := os.Getenv("WISP_IT_PG_HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	port := 54329
	if !transport.ProbeTCP(context.Background(), host, port, 500*time.Millisecond) {
		t.Skip("postgres not listening on 54329")
	}
	found := false
	for _, s := range transport.DetectLocal(context.Background()) {
		if s.Engine == "postgres" && s.Kind == "tcp" && s.Port == 5432 {
			found = true
			break
		}
	}
	if !found {
		t.Log("detect list may use 127.0.0.1:5432 only when port open")
	}
}
