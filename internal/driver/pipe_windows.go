//go:build windows

package driver

import (
	"context"
	"fmt"
	"net"
	"strings"

	winio "github.com/Microsoft/go-winio"
)

// DialMySQLPipe connects to a Windows named pipe for MySQL.
func DialMySQLPipe(ctx context.Context, pipeName string) (net.Conn, error) {
	name := strings.TrimSpace(pipeName)
	name = strings.TrimPrefix(name, `\\.\pipe\`)
	if name == "" {
		return nil, fmt.Errorf("empty pipe name")
	}
	return winio.DialPipeContext(ctx, `\\.\pipe\`+name)
}
