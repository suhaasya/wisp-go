//go:build !windows

package driver

import (
	"context"
	"errors"
	"net"
)

// DialMySQLPipe connects to a Windows named pipe (unsupported on this OS).
func DialMySQLPipe(ctx context.Context, pipeName string) (net.Conn, error) {
	_ = ctx
	_ = pipeName
	return nil, errors.New("named pipes require Windows")
}
