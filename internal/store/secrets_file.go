package store

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"
)

const (
	secretsFileMagic    = "WISPSEC1"
	secretsSaltLen      = 16
	secretsKeyLen       = 32
	secretsArgonTime    = 3
	secretsArgonMemory  = 64 * 1024 // KiB → 64 MiB
	secretsArgonThreads = 1
	secretsVerifyPlain  = "wisp-secret-store-v1"
)

// NewFileSecrets returns an encrypted file-backed Secrets store (locked until Unlock).
func NewFileSecrets(path string) *FileSecrets {
	return &FileSecrets{path: path}
}

// FileSecrets stores secrets in secrets.bin under the data directory.
type FileSecrets struct {
	path string

	mu       sync.Mutex
	unlocked bool
	key      []byte
	salt     []byte
	records  map[string][]byte
}

// Unlock derives the master key and loads secrets from disk.
func (f *FileSecrets) Unlock(masterPassword []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	pw := append([]byte(nil), masterPassword...)
	defer zeroBytes(pw)

	key, salt, records, err := loadFileSecrets(f.path, pw)
	debug.FreeOSMemory()
	if err != nil {
		return err
	}
	f.zeroLocked()
	f.key = key
	f.salt = salt
	f.records = records
	f.unlocked = true
	return nil
}

// Locked reports whether Unlock has succeeded.
func (f *FileSecrets) Locked() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return !f.unlocked
}

// Close clears in-memory key material.
func (f *FileSecrets) Close() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.zeroLocked()
}

func (f *FileSecrets) zeroLocked() {
	if f.key != nil {
		zeroBytes(f.key)
		f.key = nil
	}
	for k, v := range f.records {
		zeroBytes(v)
		delete(f.records, k)
	}
	f.records = nil
	f.salt = nil
	f.unlocked = false
}

func (f *FileSecrets) Get(profileID string, kind SecretKind) ([]byte, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if !f.unlocked {
		return nil, false, nil
	}
	key := SecretAccountKey(profileID, kind)
	v, ok := f.records[key]
	if !ok {
		return nil, false, nil
	}
	out := make([]byte, len(v))
	copy(out, v)
	return out, true, nil
}

func (f *FileSecrets) Set(profileID string, kind SecretKind, value []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if !f.unlocked {
		return ErrSecretsLocked
	}
	acc := SecretAccountKey(profileID, kind)
	valCopy := make([]byte, len(value))
	copy(valCopy, value)
	if f.records == nil {
		f.records = make(map[string][]byte)
	}
	if old, ok := f.records[acc]; ok {
		zeroBytes(old)
	}
	f.records[acc] = valCopy
	return f.persistLocked()
}

func (f *FileSecrets) DeleteAll(profileID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if !f.unlocked {
		return ErrSecretsLocked
	}
	for _, kind := range allSecretKinds {
		acc := SecretAccountKey(profileID, kind)
		if old, ok := f.records[acc]; ok {
			zeroBytes(old)
			delete(f.records, acc)
		}
	}
	return f.persistLocked()
}

func (f *FileSecrets) persistLocked() error {
	return writeSecretsFile(f.path, f.salt, f.key, f.records)
}

func loadFileSecrets(path string, masterPassword []byte) ([]byte, []byte, map[string][]byte, error) {
	data, err := os.ReadFile(path) //nolint:gosec // G304: path is SecretsBin under user data dir
	if err != nil {
		if os.IsNotExist(err) {
			salt := make([]byte, secretsSaltLen)
			if _, err := io.ReadFull(rand.Reader, salt); err != nil {
				return nil, nil, nil, err
			}
			key := deriveSecretsKey(masterPassword, salt)
			records := make(map[string][]byte)
			if err := writeSecretsFile(path, salt, key, records); err != nil {
				zeroBytes(key)
				return nil, nil, nil, err
			}
			return key, salt, records, nil
		}
		return nil, nil, nil, err
	}
	salt, records, err := parseSecretsFile(data, masterPassword)
	if err != nil {
		return nil, nil, nil, err
	}
	key := deriveSecretsKey(masterPassword, salt)
	return key, salt, records, nil
}

func deriveSecretsKey(password, salt []byte) []byte {
	return argon2.IDKey(password, salt, secretsArgonTime, secretsArgonMemory, secretsArgonThreads, secretsKeyLen)
}

func writeSecretsFile(path string, salt, key []byte, records map[string][]byte) error {
	if len(salt) != secretsSaltLen {
		return errors.New("invalid secrets salt")
	}
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return err
	}
	var body []byte
	verifyNonce, err := randomNonce(aead.NonceSize())
	if err != nil {
		return err
	}
	verifyCT := aead.Seal(nil, verifyNonce, []byte(secretsVerifyPlain), nil)
	body = appendRecord(body, verifyNonce, verifyCT)

	for acc, plain := range records {
		nonce, err := randomNonce(aead.NonceSize())
		if err != nil {
			return err
		}
		ct := aead.Seal(nil, nonce, plain, []byte(acc))
		body = appendString(body, acc)
		body = appendRecord(body, nonce, ct)
	}

	out := make([]byte, 0, len(secretsFileMagic)+secretsSaltLen+len(body))
	out = append(out, []byte(secretsFileMagic)...)
	out = append(out, salt...)
	out = append(out, body...)

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, out, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func parseSecretsFile(data []byte, masterPassword []byte) ([]byte, map[string][]byte, error) {
	if len(data) < len(secretsFileMagic)+secretsSaltLen {
		return nil, nil, UnlockError{Message: "secrets file corrupt"}
	}
	if string(data[:len(secretsFileMagic)]) != secretsFileMagic {
		return nil, nil, UnlockError{Message: "secrets file corrupt"}
	}
	salt := data[len(secretsFileMagic) : len(secretsFileMagic)+secretsSaltLen]
	key := deriveSecretsKey(masterPassword, salt)
	defer zeroBytes(key)

	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, nil, err
	}
	rest := data[len(secretsFileMagic)+secretsSaltLen:]
	rest, verifyNonce, verifyCT, err := readRecord(rest)
	if err != nil {
		return nil, nil, UnlockError{Message: "secrets file corrupt"}
	}
	plain, err := aead.Open(nil, verifyNonce, verifyCT, nil)
	if err != nil || string(plain) != secretsVerifyPlain {
		return nil, nil, UnlockError{Message: "wrong master password"}
	}

	records := make(map[string][]byte)
	for len(rest) > 0 {
		var acc string
		rest, acc, err = readString(rest)
		if err != nil {
			return nil, nil, UnlockError{Message: "secrets file corrupt"}
		}
		var nonce, ct []byte
		rest, nonce, ct, err = readRecord(rest)
		if err != nil {
			return nil, nil, UnlockError{Message: "secrets file corrupt"}
		}
		val, err := aead.Open(nil, nonce, ct, []byte(acc))
		if err != nil {
			return nil, nil, UnlockError{Message: "secrets file corrupt"}
		}
		valCopy := make([]byte, len(val))
		copy(valCopy, val)
		records[acc] = valCopy
	}
	return salt, records, nil
}

func appendString(b []byte, s string) []byte {
	if len(s) > 0xffff {
		panic("secrets record key too long")
	}
	u := make([]byte, 2)
	binary.BigEndian.PutUint16(u, uint16(len(s))) //nolint:gosec // G115: bounded above
	b = append(b, u...)
	return append(b, s...)
}

func readString(b []byte) ([]byte, string, error) {
	if len(b) < 2 {
		return b, "", errors.New("truncated")
	}
	n := int(binary.BigEndian.Uint16(b[:2]))
	b = b[2:]
	if len(b) < n {
		return b, "", errors.New("truncated")
	}
	return b[n:], string(b[:n]), nil
}

func appendRecord(b, nonce, ct []byte) []byte {
	if len(nonce) > 0xffff || len(ct) > 0xffffffff {
		panic("secrets record too large")
	}
	u := make([]byte, 2)
	binary.BigEndian.PutUint16(u, uint16(len(nonce))) //nolint:gosec // G115: bounded above
	b = append(b, u...)
	b = append(b, nonce...)
	v := make([]byte, 4)
	binary.BigEndian.PutUint32(v, uint32(len(ct))) //nolint:gosec // G115: bounded above
	b = append(b, v...)
	return append(b, ct...)
}

func readRecord(b []byte) ([]byte, []byte, []byte, error) {
	if len(b) < 2 {
		return b, nil, nil, errors.New("truncated")
	}
	nn := int(binary.BigEndian.Uint16(b[:2]))
	b = b[2:]
	if len(b) < nn {
		return b, nil, nil, errors.New("truncated")
	}
	nonce := b[:nn]
	b = b[nn:]
	if len(b) < 4 {
		return b, nil, nil, errors.New("truncated")
	}
	nc := int(binary.BigEndian.Uint32(b[:4]))
	b = b[4:]
	if len(b) < nc {
		return b, nil, nil, errors.New("truncated")
	}
	return b[nc:], nonce, b[:nc], nil
}

func randomNonce(n int) ([]byte, error) {
	nonce := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return nonce, nil
}

func zeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
