package ssh

import (
	"sync"

	gossh "golang.org/x/crypto/ssh"
)

var pendingKeys sync.Map

// RememberPendingHostKey stores a host key until the user trusts it.
func RememberPendingHostKey(fingerprint string, key gossh.PublicKey) {
	if fingerprint == "" || key == nil {
		return
	}
	pendingKeys.Store(fingerprint, key)
}

// TakePendingHostKey removes and returns a pending host key.
func TakePendingHostKey(fingerprint string) (gossh.PublicKey, bool) {
	v, ok := pendingKeys.LoadAndDelete(fingerprint)
	if !ok {
		return nil, false
	}
	key, ok := v.(gossh.PublicKey)
	return key, ok
}
