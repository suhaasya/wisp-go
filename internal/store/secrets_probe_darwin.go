//go:build darwin

package store

import "os/exec"

func probeKeyringPlatform() (KeyringProbeResult, error) {
	if _, err := exec.LookPath("/usr/bin/security"); err != nil {
		return KeyringProbeUnavailable, nil
	}
	return KeyringProbeAvailable, nil
}
