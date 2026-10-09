//go:build !windows

package transport

import (
	"context"
	"errors"
	"net"
)

// DialPipe is only supported on Windows.
func DialPipe(ctx context.Context, pipeName string) (net.Conn, error) {
	_ = ctx
	_ = pipeName
	return nil, errors.New("named pipes require Windows")
}
