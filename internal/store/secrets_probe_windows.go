//go:build windows

package store

func probeKeyringPlatform() (KeyringProbeResult, error) {
	return KeyringProbeAvailable, nil
}
