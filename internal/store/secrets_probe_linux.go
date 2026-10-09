//go:build linux

package store

import (
	"strings"

	ss "github.com/zalando/go-keyring/secret_service"
)

func isKeyringLocked(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "locked") ||
		strings.Contains(msg, "User cancelled") ||
		strings.Contains(msg, "was cancelled")
}

func probeKeyringPlatform() (KeyringProbeResult, error) {
	svc, err := ss.NewSecretService()
	if err != nil {
		return KeyringProbeUnavailable, nil
	}
	session, err := svc.OpenSession()
	if err != nil {
		if isKeyringLocked(err) {
			return KeyringProbeLocked, nil
		}
		return KeyringProbeUnavailable, nil
	}
	_ = svc.Close(session)
	return KeyringProbeAvailable, nil
}
