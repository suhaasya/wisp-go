package server

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

// ErrInstanceAlreadyRunning is returned when instance.lock is held by another process.
var ErrInstanceAlreadyRunning = errors.New("wisp instance already running")

// InstanceRecord is persisted to instance.json.
type InstanceRecord struct {
	PID     int       `json:"pid"`
	Port    int       `json:"port"`
	Started time.Time `json:"started"`
}

// InstanceLock is an exclusive runtime lock.
type InstanceLock struct {
	release func() error
}

// Release drops the instance lock.
func (l *InstanceLock) Release() error {
	if l == nil || l.release == nil {
		return nil
	}
	return l.release()
}

// WriteInstanceJSON writes record with mode 0600.
func WriteInstanceJSON(path string, rec InstanceRecord) error {
	data, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return err
	}
	return os.Chmod(path, 0o600)
}

// ReadInstanceJSON loads instance metadata.
func ReadInstanceJSON(path string) (InstanceRecord, error) {
	data, err := os.ReadFile(path) //nolint:gosec // G304: runtime metadata path under user cache
	if err != nil {
		return InstanceRecord{}, err
	}
	var rec InstanceRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return InstanceRecord{}, err
	}
	return rec, nil
}

// RemoveInstanceJSON deletes instance.json if present.
func RemoveInstanceJSON(path string) error {
	err := os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

// EnsureHandoffSecret creates a 32-byte secret file with mode 0600 if missing.
func EnsureHandoffSecret(path string) ([]byte, error) {
	data, err := os.ReadFile(path) //nolint:gosec // G304: runtime secret path under user cache
	if err == nil {
		return data, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	raw, err := randomBytes(32)
	if err != nil {
		return nil, err
	}
	secret := []byte(hex.EncodeToString(raw))
	if err := os.WriteFile(path, secret, 0o600); err != nil {
		return nil, err
	}
	return secret, nil
}

// LoadHandoffSecret reads the handoff secret file.
func LoadHandoffSecret(path string) ([]byte, error) {
	data, err := os.ReadFile(path) //nolint:gosec // G304: runtime secret path under user cache
	if err != nil {
		return nil, fmt.Errorf("handoff secret: %w", err)
	}
	return data, nil
}
