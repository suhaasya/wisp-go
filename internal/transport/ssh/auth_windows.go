//go:build windows

package ssh

import (
	"net"
)

func platformAgentConn(path string) (net.Conn, error) {
	conn, err := net.Dial("unix", path)
	if err != nil {
		return net.Dial("npipe", path)
	}
	return conn, nil
}
