// Package ssh implements SSH bastion tunnels and jump hosts for Wisp.
package ssh

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"
)

// ConfigHost holds parsed OpenSSH config fields for one Host block.
type ConfigHost struct {
	Patterns     []string
	HostName     string
	User         string
	Port         int
	IdentityFile string
	ProxyJump    string
}

// ParseConfig reads ~/.ssh/config-style files (no Match, Include, or ProxyCommand).
func ParseConfig(r *bufio.Reader) ([]ConfigHost, error) {
	var blocks []ConfigHost
	var cur *ConfigHost
	flush := func() {
		if cur != nil && len(cur.Patterns) > 0 {
			blocks = append(blocks, *cur)
		}
		cur = nil
	}
	for {
		line, err := r.ReadString('\n')
		if line != "" {
			line = strings.TrimSpace(stripComment(line))
			if line == "" {
				continue
			}
			lower := strings.ToLower(line)
			if strings.HasPrefix(lower, "host ") {
				flush()
				pat := strings.Fields(line)[1:]
				cur = &ConfigHost{Patterns: pat, Port: 22}
				continue
			}
			if cur == nil {
				continue
			}
			parts := strings.Fields(line)
			if len(parts) < 2 {
				continue
			}
			key := strings.ToLower(parts[0])
			val := strings.TrimSpace(strings.Join(parts[1:], " "))
			switch key {
			case "hostname":
				cur.HostName = val
			case "user":
				cur.User = val
			case "port":
				if p, err := strconv.Atoi(val); err == nil {
					cur.Port = p
				}
			case "identityfile":
				cur.IdentityFile = unquote(val)
			case "proxyjump":
				cur.ProxyJump = val
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
	}
	flush()
	return blocks, nil
}

// ParseConfigFile loads path (missing file → empty config).
func ParseConfigFile(path string) ([]ConfigHost, error) {
	f, err := os.Open(path) //nolint:gosec // G304: ~/.ssh/config path
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return ParseConfig(bufio.NewReader(f))
}

// ResolveHost finds the first matching Host block for alias.
func ResolveHost(blocks []ConfigHost, alias string) (ConfigHost, bool) {
	for _, b := range blocks {
		for _, p := range b.Patterns {
			if matchHostPattern(p, alias) {
				out := b
				if out.HostName == "" {
					out.HostName = alias
				}
				if out.Port == 0 {
					out.Port = 22
				}
				return out, true
			}
		}
	}
	return ConfigHost{}, false
}

func matchHostPattern(pat, alias string) bool {
	if pat == "*" {
		return true
	}
	return strings.EqualFold(pat, alias)
}

func stripComment(line string) string {
	if i := strings.Index(line, "#"); i >= 0 {
		return line[:i]
	}
	return line
}

func unquote(s string) string {
	if len(s) >= 2 && ((s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'')) {
		return s[1 : len(s)-1]
	}
	return s
}
