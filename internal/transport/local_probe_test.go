package transport_test

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/suhaasya/wisp-go/internal/transport"
)

func TestProbeTCP_openPort(t *testing.T) {
	t.Parallel()
	lc := net.ListenConfig{}
	ln, err := lc.Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = ln.Close() }()
	_, portStr, _ := net.SplitHostPort(ln.Addr().String())
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatal(err)
	}
	if !transport.ProbeTCP(context.Background(), "127.0.0.1", port, 500*time.Millisecond) {
		t.Fatal("expected open port")
	}
}
