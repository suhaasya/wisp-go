package store

import (
	"errors"
	"fmt"
)

const secretServiceName = "dev.wisp"

// SecretKind identifies which secret slot on a profile is used.
type SecretKind string

const (
	SecretKindDBPassword    SecretKind = "db_password"
	SecretKindSSHPassword   SecretKind = "ssh_password"
	SecretKindSSHPassphrase SecretKind = "ssh_passphrase"
)

var allSecretKinds = []SecretKind{
	SecretKindDBPassword,
	SecretKindSSHPassword,
	SecretKindSSHPassphrase,
}

// Secrets stores profile secrets outside config files.
type Secrets interface {
	Get(profileID string, kind SecretKind) ([]byte, bool, error)
	Set(profileID string, kind SecretKind, value []byte) error
	DeleteAll(profileID string) error
}

// SecretAccountKey is the keyring user name for a profile secret.
func SecretAccountKey(profileID string, kind SecretKind) string {
	return profileID + "/" + string(kind)
}

// ErrSecretsLocked is returned when the file backend needs unlock.
var ErrSecretsLocked = errors.New("secrets locked")

// ErrSecretServiceLocked is returned when the OS keychain exists but is locked.
var ErrSecretServiceLocked = errors.New("secret service locked")

// UnlockError means the master password was wrong or the file is corrupt.
type UnlockError struct {
	Message string
}

func (e UnlockError) Error() string {
	if e.Message == "" {
		return "secrets unlock failed"
	}
	return e.Message
}

// IsUnlockError reports unlock failures (wrong master password).
func IsUnlockError(err error) bool {
	var u UnlockError
	return errors.As(err, &u)
}

// IsSecretServiceLocked reports when the OS keychain is locked.
func IsSecretServiceLocked(err error) bool {
	return errors.Is(err, ErrSecretServiceLocked)
}

// SecretsBackend identifies which store implementation is active.
type SecretsBackend string

const (
	SecretsBackendKeyring SecretsBackend = "keyring"
	SecretsBackendFile    SecretsBackend = "file"
)

// SecretsStore is the active secrets backend for the running instance.
type SecretsStore struct {
	Backend SecretsBackend
	Secrets
	file *FileSecrets
}

// OpenSecretsStore probes the keychain and returns the store for the process.
func OpenSecretsStore(p Paths, probe KeyringProbe) (*SecretsStore, error) {
	sec, backend, err := OpenSecrets(p, probe)
	if err != nil {
		return nil, err
	}
	st := &SecretsStore{Backend: backend, Secrets: sec}
	if backend == SecretsBackendFile {
		st.file = sec.(*FileSecrets)
	}
	return st, nil
}

// MasterLocked reports whether the file backend awaits unlock.
func (s *SecretsStore) MasterLocked() bool {
	return s.file != nil && s.file.Locked()
}

// UnlockMaster unlocks the encrypted file backend.
func (s *SecretsStore) UnlockMaster(masterPassword []byte) error {
	if s.file == nil {
		return nil
	}
	return s.file.Unlock(masterPassword)
}

// OpenSecrets selects keyring or encrypted file storage from a startup probe.
func OpenSecrets(p Paths, probe KeyringProbe) (Secrets, SecretsBackend, error) {
	if probe == nil {
		probe = DefaultKeyringProbe
	}
	res, err := probe.Probe()
	if err != nil {
		return nil, "", err
	}
	switch res {
	case KeyringProbeAvailable:
		return NewKeyringSecrets(), SecretsBackendKeyring, nil
	case KeyringProbeUnavailable:
		return NewFileSecrets(p.SecretsBin()), SecretsBackendFile, nil
	case KeyringProbeLocked:
		return nil, "", ErrSecretServiceLocked
	default:
		return nil, "", fmt.Errorf("unknown keyring probe result")
	}
}
