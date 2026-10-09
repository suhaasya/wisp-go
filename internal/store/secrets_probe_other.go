//go:build !darwin && !linux && !windows

package store

func probeKeyringPlatform() (KeyringProbeResult, error) {
	return KeyringProbeUnavailable, nil
}
