package session

import (
	"errors"
	"os"
	"path/filepath"

	sshtransport "github.com/suhaasya/wisp-go/internal/transport/ssh"
)

// HostKeyPrompt is returned to the page when SSH host key trust is required.
type HostKeyPrompt struct {
	Host        string
	Port        int
	KeyType     string
	Fingerprint string
}

// HostKeyPromptFrom reports whether err is an SSH host key trust prompt.
func HostKeyPromptFrom(err error) (*HostKeyPrompt, bool) {
	var need *sshtransport.HostKeyNeeded
	if errors.As(err, &need) {
		return &HostKeyPrompt{
			Host:        need.Host,
			Port:        need.Port,
			KeyType:     need.KeyType,
			Fingerprint: need.Fingerprint,
		}, true
	}
	return nil, false
}

// AcceptHostKey appends a previously seen host key to known_hosts.
func AcceptHostKey(host string, port int, fingerprint string) error {
	key, ok := sshtransport.TakePendingHostKey(fingerprint)
	if !ok {
		return errHostKeyExpired
	}
	if port == 0 {
		port = 22
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	path := filepath.Join(home, ".ssh", "known_hosts")
	return sshtransport.AppendKnownHost(path, host, port, key)
}

var errHostKeyExpired = &hostKeyExpiredError{}

type hostKeyExpiredError struct{}

func (e *hostKeyExpiredError) Error() string {
	return "host key prompt expired"
}
