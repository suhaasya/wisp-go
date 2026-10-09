package server

import (
	"context"
	"net"
	"net/http"
	"time"
)

const (
	readHeaderTimeout = 5 * time.Second
	idleTimeout       = 120 * time.Second
	maxHeaderBytes    = 64 << 10
)

// ListenLoopback binds only to 127.0.0.1 (port 0 = random).
func ListenLoopback(port int) (net.Listener, error) {
	addr := "127.0.0.1:" + itoa(port)
	lc := net.ListenConfig{}
	return lc.Listen(context.Background(), "tcp", addr)
}

// NewHTTPServer returns an http.Server configured for the local Wisp listener.
func NewHTTPServer(handler http.Handler) *http.Server {
	return &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		IdleTimeout:       idleTimeout,
		MaxHeaderBytes:    maxHeaderBytes,
	}
}
