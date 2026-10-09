package transport

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const detectProbeTimeout = 200 * time.Millisecond

// LocalSuggestion is one detected local database instance for the UI.
type LocalSuggestion struct {
	ID       string `json:"id"`
	Engine   string `json:"engine"`
	Label    string `json:"label"`
	User     string `json:"user"`
	Database string `json:"database"`
	Kind     string `json:"kind"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Path     string `json:"path"`
}

// ParsePostgresSocketPath returns the pgx socket directory and port from a directory or .s.PGSQL.* file path.
func ParsePostgresSocketPath(path string) (dir string, port int, err error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", 0, errors.New("empty socket path")
	}
	path = filepath.Clean(path)
	base := filepath.Base(path)
	if strings.HasPrefix(base, ".s.PGSQL.") {
		suffix := strings.TrimPrefix(base, ".s.PGSQL.")
		p, convErr := strconv.Atoi(suffix)
		if convErr != nil || p < 1 || p > 65535 {
			return "", 0, fmt.Errorf("invalid postgres socket file name: %s", base)
		}
		return filepath.Dir(path), p, nil
	}
	return path, 5432, nil
}

// ParseMySQLSocketPath returns the absolute unix socket path.
func ParseMySQLSocketPath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", errors.New("empty socket path")
	}
	if !filepath.IsAbs(path) {
		return "", errors.New("mysql socket path must be absolute")
	}
	return filepath.Clean(path), nil
}

// ParsePipeName normalizes a Windows named pipe to a bare name (e.g. MySQL).
func ParsePipeName(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", errors.New("empty pipe name")
	}
	s = strings.TrimPrefix(s, `\\.\pipe\`)
	s = strings.TrimPrefix(s, `\.\pipe\`)
	s = strings.TrimPrefix(s, `\\.\Pipe\`)
	if s == "" {
		return "", errors.New("empty pipe name")
	}
	return s, nil
}

// KnownPostgresSocketDirs returns platform-specific directories that may contain PostgreSQL sockets.
func KnownPostgresSocketDirs() []string {
	var out []string
	out = append(out, "/tmp", "/var/run/postgresql", "/run/postgresql")
	if runtime.GOOS == "darwin" {
		out = append(out,
			"/opt/homebrew/var/postgresql@16",
			"/opt/homebrew/var/postgres",
			"/usr/local/var/postgresql@16",
			"/usr/local/var/postgres",
		)
		if home, err := os.UserHomeDir(); err == nil {
			out = append(out, filepath.Join(home, "Library", "Application Support", "Postgres", "var-16"))
		}
	}
	return dedupePaths(out)
}

// KnownMySQLSocketPaths returns common unix socket file paths.
func KnownMySQLSocketPaths() []string {
	var out []string
	if runtime.GOOS == "darwin" {
		out = append(out,
			"/tmp/mysql.sock",
			"/var/run/mysqld/mysqld.sock",
			"/opt/homebrew/var/mysql/mysql.sock",
			"/usr/local/var/mysql/mysql.sock",
		)
	}
	if runtime.GOOS == "linux" {
		out = append(out,
			"/var/run/mysqld/mysqld.sock",
			"/run/mysqld/mysqld.sock",
			"/var/lib/mysql/mysql.sock",
		)
	}
	return dedupePaths(out)
}

func dedupePaths(in []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, p := range in {
		p = filepath.Clean(p)
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out
}

// ProbeTCP reports whether host:port accepts a TCP connection within timeout.
func ProbeTCP(ctx context.Context, host string, port int, timeout time.Duration) bool {
	if timeout <= 0 {
		timeout = detectProbeTimeout
	}
	dctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	d := net.Dialer{}
	conn, err := d.DialContext(dctx, "tcp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func socketExists(path string) bool {
	st, err := os.Stat(path)
	if err != nil {
		return false
	}
	if st.Mode()&os.ModeSocket != 0 {
		return true
	}
	// Directory containing .s.PGSQL.* or mysql.sock by name
	if st.IsDir() {
		return postgresSocketInDir(path) || filepath.Base(path) == "mysql.sock"
	}
	return false
}

func postgresSocketInDir(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".s.PGSQL.") {
			return true
		}
	}
	return false
}

func firstPostgresSocketFile(dir string) (fullPath string, port int, ok bool) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", 0, false
	}
	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), ".s.PGSQL.") {
			continue
		}
		full := filepath.Join(dir, e.Name())
		dir2, port2, err := ParsePostgresSocketPath(full)
		if err != nil {
			continue
		}
		return filepath.Join(dir2, e.Name()), port2, true
	}
	return "", 0, false
}

// OSUsername returns the current OS user for peer-auth prefill.
func OSUsername() string {
	if u := os.Getenv("USER"); u != "" {
		return u
	}
	if u := os.Getenv("USERNAME"); u != "" {
		return u
	}
	return "postgres"
}

// DetectLocal scans known socket locations and localhost ports.
func DetectLocal(ctx context.Context) []LocalSuggestion {
	var out []LocalSuggestion
	user := OSUsername()

	for _, dir := range KnownPostgresSocketDirs() {
		if ctx.Err() != nil {
			return out
		}
		if st, err := os.Stat(dir); err == nil && st.IsDir() && postgresSocketInDir(dir) {
			sockPath, port, ok := firstPostgresSocketFile(dir)
			if !ok {
				continue
			}
			out = append(out, LocalSuggestion{
				ID:       "pg-socket-" + dir,
				Engine:   "postgres",
				Label:    "PostgreSQL (local socket)",
				User:     user,
				Database: "postgres",
				Kind:     "socket",
				Host:     dir,
				Port:     port,
				Path:     sockPath,
			})
		}
	}

	for _, sock := range KnownMySQLSocketPaths() {
		if ctx.Err() != nil {
			return out
		}
		if socketExists(sock) {
			out = append(out, LocalSuggestion{
				ID:       "mysql-socket-" + sock,
				Engine:   "mysql",
				Label:    "MySQL (local socket)",
				User:     "root",
				Database: "mysql",
				Kind:     "socket",
				Host:     "",
				Port:     3306,
				Path:     sock,
			})
		}
	}

	tcpChecks := []struct {
		engine string
		port   int
		label  string
		db     string
		user   string
	}{
		{"postgres", 5432, "PostgreSQL (localhost)", "postgres", user},
		{"mysql", 3306, "MySQL (localhost)", "mysql", "root"},
		{"mariadb", 3307, "MariaDB (localhost)", "mysql", "root"},
	}
	for _, c := range tcpChecks {
		if ProbeTCP(ctx, "127.0.0.1", c.port, detectProbeTimeout) {
			out = append(out, LocalSuggestion{
				ID:       fmt.Sprintf("%s-tcp-127.0.0.1-%d", c.engine, c.port),
				Engine:   c.engine,
				Label:    c.label,
				User:     c.user,
				Database: c.db,
				Kind:     "tcp",
				Host:     "127.0.0.1",
				Port:     c.port,
				Path:     "",
			})
		}
	}
	return out
}
