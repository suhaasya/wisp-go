//go:build !windows

package ssh

import (
	"context"
	"net"
)

func platformAgentConn(path string) (net.Conn, error) {
	var d net.Dialer
	return d.DialContext(context.Background(), "unix", path) //nolint:gosec // G704: SSH_AUTH_SOCK path from environment
}
