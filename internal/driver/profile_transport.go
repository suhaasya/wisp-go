package driver

import (
	"fmt"

	"github.com/suhaasya/wisp-go/internal/transport"
)

// ApplyProfileTransport maps a saved profile transport onto ConnectParams.
func ApplyProfileTransport(params *ConnectParams, engine, kind, host, path string, port int) error {
	if kind == "" {
		kind = "tcp"
	}
	params.TransportKind = kind
	switch kind {
	case "tcp":
		params.Host = host
		params.Port = port
		params.TransportPath = ""
		return nil
	case "socket":
		if path == "" {
			path = host
		}
		if path == "" {
			return fmt.Errorf("socket path required")
		}
		switch engine {
		case "postgres":
			dir, p, err := transport.ParsePostgresSocketPath(path)
			if err != nil {
				return err
			}
			params.Host = dir
			params.Port = p
			params.TransportPath = path
		case "mysql", "mariadb":
			sock, err := transport.ParseMySQLSocketPath(path)
			if err != nil {
				return err
			}
			params.TransportPath = sock
			params.Host = "localhost"
			params.Port = 3306
		default:
			return fmt.Errorf("socket transport not supported for %s", engine)
		}
		return nil
	case "pipe":
		if engine != "mysql" && engine != "mariadb" {
			return fmt.Errorf("named pipe only supported for MySQL/MariaDB")
		}
		raw := path
		if raw == "" {
			raw = host
		}
		name, err := transport.ParsePipeName(raw)
		if err != nil {
			return err
		}
		params.TransportPath = name
		params.Host = "localhost"
		params.Port = 3306
		return nil
	default:
		return fmt.Errorf("unknown transport kind %q", kind)
	}
}
