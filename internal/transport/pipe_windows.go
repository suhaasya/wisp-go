//go:build windows

package transport

import (
	"context"
	"net"

	winio "github.com/Microsoft/go-winio"
)

// DialPipe connects to a Windows named pipe for MySQL.
func DialPipe(ctx context.Context, pipeName string) (net.Conn, error) {
	name, err := ParsePipeName(pipeName)
	if err != nil {
		return nil, err
	}
	path := `\\.\pipe\` + name
	return winio.DialPipeContext(ctx, path)
}
