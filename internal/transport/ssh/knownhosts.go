package ssh

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// HostKeyCallback builds a known_hosts verifier that returns HostKeyNeeded or HostKeyMismatch.
func HostKeyCallback(knownHostsPath string) (gossh.HostKeyCallback, error) {
	if strings.TrimSpace(knownHostsPath) == "" {
		return nil, errors.New("ssh: known_hosts path required")
	}
	cb, err := knownhosts.New(knownHostsPath)
	if err != nil {
		if os.IsNotExist(err) {
			cb, err = knownhosts.New()
			if err != nil {
				return nil, err
			}
			return wrapCallback(cb, knownHostsPath), nil
		}
		return nil, err
	}
	return wrapCallback(cb, knownHostsPath), nil
}

func wrapCallback(inner gossh.HostKeyCallback, _ string) gossh.HostKeyCallback {
	return func(hostname string, remote net.Addr, key gossh.PublicKey) error {
		err := inner(hostname, remote, key)
		if err == nil {
			return nil
		}
		host, port := splitHostPort(hostname, remote)
		fp := fingerprintSHA256(key)
		var keyErr *knownhosts.KeyError
		if errors.As(err, &keyErr) {
			if len(keyErr.Want) > 0 {
				return &HostKeyMismatch{Host: host, Port: port, Fingerprint: fp}
			}
			need := &HostKeyNeeded{
				Host:        host,
				Port:        port,
				KeyType:     keyTypeName(key),
				Fingerprint: fp,
				Key:         key,
			}
			RememberPendingHostKey(fp, key)
			return need
		}
		if strings.Contains(strings.ToLower(err.Error()), "unknown") {
			need := &HostKeyNeeded{
				Host:        host,
				Port:        port,
				KeyType:     keyTypeName(key),
				Fingerprint: fp,
				Key:         key,
			}
			RememberPendingHostKey(fp, key)
			return need
		}
		return err
	}
}

func splitHostPort(hostname string, remote net.Addr) (string, int) {
	host := hostname
	port := 22
	if h, p, ok := parseHostPort(hostname); ok {
		host, port = h, p
	} else if remote != nil {
		if h, p, ok := parseHostPort(remote.String()); ok {
			host, port = h, p
		}
	}
	return host, port
}

func parseHostPort(s string) (string, int, bool) {
	if s == "" {
		return "", 0, false
	}
	// [host]:port or host:port
	if strings.HasPrefix(s, "[") {
		i := strings.Index(s, "]")
		if i < 0 {
			return "", 0, false
		}
		host := s[1:i]
		rest := strings.TrimPrefix(s[i+1:], ":")
		if rest == "" {
			return host, 22, true
		}
		var p int
		_, err := fmt.Sscanf(rest, "%d", &p)
		if err != nil {
			return host, 22, true
		}
		return host, p, true
	}
	host, rest, ok := strings.Cut(s, ":")
	if !ok {
		return s, 22, true
	}
	var p int
	if _, err := fmt.Sscanf(rest, "%d", &p); err != nil {
		return host, 22, true
	}
	return host, p, true
}

// AppendKnownHost adds a host key line to the user's known_hosts file.
func AppendKnownHost(path, host string, port int, key gossh.PublicKey) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	line := knownhosts.Line([]string{knownHostName(host, port)}, key)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600) //nolint:gosec // G304: ~/.ssh/known_hosts
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	if _, err := f.WriteString(line + "\n"); err != nil {
		return err
	}
	return nil
}

func knownHostName(host string, port int) string {
	if port == 22 {
		return host
	}
	if strings.Contains(host, ":") {
		return fmt.Sprintf("[%s]:%d", host, port)
	}
	return fmt.Sprintf("%s:%d", host, port)
}
