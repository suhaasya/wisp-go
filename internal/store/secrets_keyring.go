package store

import (
	"errors"

	"github.com/zalando/go-keyring"
)

// KeyringSecrets stores secrets in the OS keychain.
type KeyringSecrets struct{}

// NewKeyringSecrets returns a keyring-backed Secrets store.
func NewKeyringSecrets() *KeyringSecrets {
	return &KeyringSecrets{}
}

func (k *KeyringSecrets) Get(profileID string, kind SecretKind) ([]byte, bool, error) {
	user := SecretAccountKey(profileID, kind)
	pass, err := keyring.Get(secretServiceName, user)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return []byte(pass), true, nil
}

func (k *KeyringSecrets) Set(profileID string, kind SecretKind, value []byte) error {
	user := SecretAccountKey(profileID, kind)
	return keyring.Set(secretServiceName, user, string(value))
}

func (k *KeyringSecrets) DeleteAll(profileID string) error {
	var firstErr error
	for _, kind := range allSecretKinds {
		user := SecretAccountKey(profileID, kind)
		if err := keyring.Delete(secretServiceName, user); err != nil {
			if errors.Is(err, keyring.ErrNotFound) {
				continue
			}
			firstErr = err
		}
	}
	return firstErr
}

// KeyringProbeResult is the outcome of checking OS keychain availability.
type KeyringProbeResult int

const (
	KeyringProbeAvailable KeyringProbeResult = iota
	KeyringProbeUnavailable
	KeyringProbeLocked
)

// KeyringProbe checks whether the OS keychain can be used.
type KeyringProbe interface {
	Probe() (KeyringProbeResult, error)
}

type defaultKeyringProbe struct{}

func (defaultKeyringProbe) Probe() (KeyringProbeResult, error) {
	return ProbeKeyring()
}

// DefaultKeyringProbe is used by OpenSecrets.
var DefaultKeyringProbe KeyringProbe = defaultKeyringProbe{}

type fakeKeyringProbe struct {
	result KeyringProbeResult
}

// FakeKeyringProbe returns a probe that always reports result (tests).
func FakeKeyringProbe(result KeyringProbeResult) KeyringProbe {
	return fakeKeyringProbe{result: result}
}

func (f fakeKeyringProbe) Probe() (KeyringProbeResult, error) {
	return f.result, nil
}

// ProbeKeyring checks whether the OS keychain can be used without storing probe secrets.
func ProbeKeyring() (KeyringProbeResult, error) {
	return probeKeyringPlatform()
}
