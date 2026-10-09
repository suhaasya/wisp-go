package ssh

import (
	"fmt"

	gossh "golang.org/x/crypto/ssh"
)

// HostKeyNeeded is returned when the server host key is not in known_hosts.
type HostKeyNeeded struct {
	Host        string
	Port        int
	KeyType     string
	Fingerprint string
	Key         gossh.PublicKey
}

func (e *HostKeyNeeded) Error() string {
	return fmt.Sprintf("ssh: unknown host key for %s:%d (%s %s)", e.Host, e.Port, e.KeyType, e.Fingerprint)
}

// HostKeyMismatch is returned when the server key does not match known_hosts.
type HostKeyMismatch struct {
	Host        string
	Port        int
	Fingerprint string
}

func (e *HostKeyMismatch) Error() string {
	return fmt.Sprintf("ssh: host key mismatch for %s:%d (%s)", e.Host, e.Port, e.Fingerprint)
}
