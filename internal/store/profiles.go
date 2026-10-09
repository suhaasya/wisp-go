package store

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const profilesVersion = 1

// Profile is a saved connection definition without secrets.
type Profile struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Folder       string          `json:"folder"`
	Engine       string          `json:"engine"`
	Transport    Transport       `json:"transport"`
	User         string          `json:"user"`
	Database     string          `json:"database"`
	PasswordMode string          `json:"password_mode"`
	SSL          SSLConfig       `json:"ssl"`
	SSH          *SSHConfig      `json:"ssh,omitempty"`
	Env          string          `json:"env"`
	ReadOnly     bool            `json:"read_only"`
	SafeMode     bool            `json:"safe_mode"`
	Options      ConnOptions     `json:"options"`
	Advanced     ProfileAdvanced `json:"advanced"`
	LastUsed     *time.Time      `json:"last_used,omitempty"`
}

// ProfileAdvanced holds session defaults configured on connect.
type ProfileAdvanced struct {
	DefaultSchema       string `json:"default_schema,omitempty"`
	TimeZone            string `json:"time_zone,omitempty"`
	StatementTimeoutSec int    `json:"statement_timeout_sec,omitempty"`
	KeepAliveSec        int    `json:"keep_alive_sec,omitempty"`
	RunOnConnect        string `json:"run_on_connect,omitempty"`
}

// Transport describes how to reach the database server.
type Transport struct {
	Kind string `json:"kind"`
	Host string `json:"host"`
	Port int    `json:"port"`
	Path string `json:"path"`
}

// SSLConfig holds TLS options (no secrets in profile JSON).
type SSLConfig struct {
	Mode string `json:"mode"`
}

// SSHConfig holds SSH tunnel options (no secrets in profile JSON).
type SSHConfig struct {
	Host    string `json:"host"`
	Port    int    `json:"port"`
	User    string `json:"user"`
	KeyPath string `json:"key_path"`
}

// ConnOptions holds driver-specific connection flags.
type ConnOptions struct {
	AllowPublicKeyRetrieval bool `json:"allow_public_key_retrieval,omitempty"`
}

// ConnectionsDocument is the on-disk connections file.
type ConnectionsDocument struct {
	Version     int       `json:"version"`
	Folders     []string  `json:"folders"`
	Connections []Profile `json:"connections"`
	Warning     string    `json:"warning,omitempty"`
}

// ProfileStore persists connection profiles in connections.json.
type ProfileStore struct {
	path string
	mu   sync.RWMutex
	doc  ConnectionsDocument
}

// NewProfileStore returns a store for the given JSON path.
func NewProfileStore(path string) *ProfileStore {
	return &ProfileStore{
		path: path,
		doc: ConnectionsDocument{
			Version:     profilesVersion,
			Folders:     []string{},
			Connections: []Profile{},
		},
	}
}

// Load reads profiles from disk.
func (st *ProfileStore) Load() (ConnectionsDocument, error) {
	st.mu.Lock()
	defer st.mu.Unlock()
	data, err := os.ReadFile(st.path) //nolint:gosec // G304: user config path
	if err != nil {
		if os.IsNotExist(err) {
			st.doc = emptyConnectionsDoc()
			return st.doc, nil
		}
		return ConnectionsDocument{}, err
	}
	var doc ConnectionsDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		_ = backupCorruptFile(st.path, data)
		st.doc = emptyConnectionsDoc()
		st.doc.Warning = "connections file was corrupt and was reset"
		return st.doc, nil
	}
	if doc.Version == 0 {
		doc.Version = profilesVersion
	}
	if doc.Folders == nil {
		doc.Folders = []string{}
	}
	if doc.Connections == nil {
		doc.Connections = []Profile{}
	}
	st.doc = doc
	return st.doc, nil
}

func emptyConnectionsDoc() ConnectionsDocument {
	return ConnectionsDocument{
		Version:     profilesVersion,
		Folders:     []string{},
		Connections: []Profile{},
	}
}

func backupCorruptFile(path string, data []byte) error {
	ts := time.Now().UTC().Format("20060102T150405Z")
	bak := path + ".bak-" + ts
	if err := os.WriteFile(bak, data, 0o600); err != nil { //nolint:gosec // G703: bak path derived from user config file
		return err
	}
	_ = os.Remove(path)
	return nil
}

// Document returns the current in-memory document (including load warning).
func (st *ProfileStore) Document() ConnectionsDocument {
	st.mu.RLock()
	defer st.mu.RUnlock()
	return st.doc
}

// Create adds a profile with a new UUIDv7 id.
func (st *ProfileStore) Create(p Profile) (Profile, error) {
	if err := ValidateProfile(p); err != nil {
		return Profile{}, err
	}
	st.mu.Lock()
	defer st.mu.Unlock()
	p.ID = NewUUIDv7()
	st.doc.Connections = append(st.doc.Connections, p)
	st.ensureFolder(p.Folder)
	if err := st.saveLocked(); err != nil {
		return Profile{}, err
	}
	return p, nil
}

func (st *ProfileStore) ensureFolder(name string) {
	if name == "" {
		return
	}
	for _, f := range st.doc.Folders {
		if f == name {
			return
		}
	}
	st.doc.Folders = append(st.doc.Folders, name)
}

// Update replaces a profile by id.
func (st *ProfileStore) Update(p Profile) error {
	if p.ID == "" {
		return errors.New("profile id required")
	}
	if err := ValidateProfile(p); err != nil {
		return err
	}
	st.mu.Lock()
	defer st.mu.Unlock()
	for i, c := range st.doc.Connections {
		if c.ID == p.ID {
			st.doc.Connections[i] = p
			st.ensureFolder(p.Folder)
			return st.saveLocked()
		}
	}
	return errors.New("profile not found")
}

// Delete removes a profile by id.
func (st *ProfileStore) Delete(id string) error {
	st.mu.Lock()
	defer st.mu.Unlock()
	for i, c := range st.doc.Connections {
		if c.ID == id {
			st.doc.Connections = append(st.doc.Connections[:i], st.doc.Connections[i+1:]...)
			return st.saveLocked()
		}
	}
	return errors.New("profile not found")
}

// Duplicate clones a profile with a new id and "(copy)" name suffix.
func (st *ProfileStore) Duplicate(id string) (Profile, error) {
	st.mu.Lock()
	defer st.mu.Unlock()
	for _, c := range st.doc.Connections {
		if c.ID == id {
			dup := c
			dup.ID = NewUUIDv7()
			dup.Name = c.Name + " (copy)"
			st.doc.Connections = append(st.doc.Connections, dup)
			if err := st.saveLocked(); err != nil {
				return Profile{}, err
			}
			return dup, nil
		}
	}
	return Profile{}, errors.New("profile not found")
}

// Get returns one profile by id.
func (st *ProfileStore) Get(id string) (Profile, bool) {
	st.mu.RLock()
	defer st.mu.RUnlock()
	for _, c := range st.doc.Connections {
		if c.ID == id {
			return c, true
		}
	}
	return Profile{}, false
}

func (st *ProfileStore) saveLocked() error {
	st.doc.Version = profilesVersion
	data, err := json.Marshal(st.doc)
	if err != nil {
		return err
	}
	dir := filepath.Dir(st.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	tmp := st.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmp, st.path); err != nil {
		return err
	}
	return os.Chmod(st.path, 0o600)
}

// ValidateProfile checks engine, env, and related fields.
func ValidateProfile(p Profile) error {
	switch p.Engine {
	case "postgres", "mysql", "mariadb":
	default:
		return fmt.Errorf("engine must be postgres, mysql, or mariadb")
	}
	switch p.Env {
	case "local", "development", "staging", "production":
	default:
		return fmt.Errorf("env must be local, development, staging, or production")
	}
	switch p.PasswordMode {
	case "keychain", "ask", "none":
	default:
		return fmt.Errorf("password_mode must be keychain, ask, or none")
	}
	switch p.Transport.Kind {
	case "tcp", "socket", "pipe":
	default:
		return fmt.Errorf("transport.kind must be tcp, socket, or pipe")
	}
	if p.Name == "" {
		return errors.New("name required")
	}
	return nil
}

// NewUUIDv7 returns a time-ordered UUID version 7 string.
func NewUUIDv7() string {
	var b [16]byte
	now := time.Now().UnixMilli()
	b[0] = byte(now >> 40) //nolint:gosec // G115: UUIDv7 timestamp bytes
	b[1] = byte(now >> 32) //nolint:gosec // G115: UUIDv7 timestamp bytes
	b[2] = byte(now >> 24) //nolint:gosec // G115: UUIDv7 timestamp bytes
	b[3] = byte(now >> 16) //nolint:gosec // G115: UUIDv7 timestamp bytes
	b[4] = byte(now >> 8)  //nolint:gosec // G115: UUIDv7 timestamp bytes
	b[5] = byte(now)       //nolint:gosec // G115: UUIDv7 timestamp bytes
	_, _ = rand.Read(b[6:])
	b[6] = (b[6] & 0x0f) | 0x70
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uint32(b[0])<<24|uint32(b[1])<<16|uint32(b[2])<<8|uint32(b[3]),
		uint16(b[4])<<8|uint16(b[5]),
		uint16(b[6])<<8|uint16(b[7]),
		uint16(b[8])<<8|uint16(b[9]),
		uint64(b[10])<<40|uint64(b[11])<<32|uint64(b[12])<<24|uint64(b[13])<<16|uint64(b[14])<<8|uint64(b[15]),
	)
}
