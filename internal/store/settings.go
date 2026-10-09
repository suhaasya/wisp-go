package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// SettingsVersion is the on-disk schema version.
const SettingsVersion = 1

// Settings are user preferences persisted in config/settings.json.
type Settings struct {
	Version                   int               `json:"version"`
	Theme                     string            `json:"theme"`
	Density                   string            `json:"density"`
	WindowRows                int               `json:"window_rows"`
	WindowsInMemory           int               `json:"windows_in_memory"`
	ResultCapMB               int               `json:"result_cap_mb"`
	KeepAliveSec              int               `json:"keep_alive_sec"`
	IdleExitMin               int               `json:"idle_exit_min"`
	GoMemLimitMB              int               `json:"go_mem_limit_mb"`
	ConfirmWritesOnProduction bool              `json:"confirm_writes_on_production"`
	Keybindings               map[string]string `json:"keybindings"`
	Telemetry                 bool              `json:"telemetry"`
}

var knownSettingKeys = map[string]struct{}{
	"version": {}, "theme": {}, "density": {}, "window_rows": {}, "windows_in_memory": {},
	"result_cap_mb": {}, "keep_alive_sec": {}, "idle_exit_min": {}, "go_mem_limit_mb": {},
	"confirm_writes_on_production": {}, "keybindings": {}, "telemetry": {},
}

// DefaultSettings returns factory defaults from the product spec.
func DefaultSettings() Settings {
	return Settings{
		Version:                   SettingsVersion,
		Theme:                     "system",
		Density:                   "compact",
		WindowRows:                150,
		WindowsInMemory:           3,
		ResultCapMB:               32,
		KeepAliveSec:              60,
		IdleExitMin:               10,
		GoMemLimitMB:              64,
		ConfirmWritesOnProduction: true,
		Keybindings:               map[string]string{},
		Telemetry:                 false,
	}
}

// SettingsStore loads and saves settings atomically.
type SettingsStore struct {
	path  string
	mu    sync.RWMutex
	cur   Settings
	extra map[string]json.RawMessage
}

// NewSettingsStore returns a store for the given JSON file path.
func NewSettingsStore(path string) *SettingsStore {
	return &SettingsStore{
		path:  path,
		cur:   DefaultSettings(),
		extra: map[string]json.RawMessage{},
	}
}

// Current returns the in-memory settings (last good).
func (st *SettingsStore) Current() Settings {
	st.mu.RLock()
	defer st.mu.RUnlock()
	return st.cur
}

// Load reads settings from disk or defaults when missing.
func (st *SettingsStore) Load() (Settings, error) {
	st.mu.Lock()
	defer st.mu.Unlock()

	data, err := os.ReadFile(st.path)
	if err != nil {
		if os.IsNotExist(err) {
			st.cur = DefaultSettings()
			st.extra = map[string]json.RawMessage{}
			return st.cur, nil
		}
		return Settings{}, err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return st.cur, fmt.Errorf("settings json: %w", err)
	}
	extra := map[string]json.RawMessage{}
	for k, v := range raw {
		if _, known := knownSettingKeys[k]; !known {
			extra[k] = v
		}
	}
	knownBytes, err := json.Marshal(rawFromKnown(raw))
	if err != nil {
		return st.cur, err
	}
	s := DefaultSettings()
	if err := json.Unmarshal(knownBytes, &s); err != nil {
		return st.cur, fmt.Errorf("settings json: %w", err)
	}
	applyMissingDefaults(&s)
	if s.Keybindings == nil {
		s.Keybindings = map[string]string{}
	}
	st.cur = s
	st.extra = extra
	return st.cur, nil
}

func rawFromKnown(raw map[string]json.RawMessage) map[string]json.RawMessage {
	out := map[string]json.RawMessage{}
	for k := range knownSettingKeys {
		if v, ok := raw[k]; ok {
			out[k] = v
		}
	}
	return out
}

func applyMissingDefaults(s *Settings) {
	d := DefaultSettings()
	if s.Version == 0 {
		s.Version = d.Version
	}
	if s.Theme == "" {
		s.Theme = d.Theme
	}
	if s.Density == "" {
		s.Density = d.Density
	}
	if s.WindowRows == 0 {
		s.WindowRows = d.WindowRows
	}
	if s.WindowsInMemory == 0 {
		s.WindowsInMemory = d.WindowsInMemory
	}
	if s.ResultCapMB == 0 {
		s.ResultCapMB = d.ResultCapMB
	}
	if s.KeepAliveSec == 0 {
		s.KeepAliveSec = d.KeepAliveSec
	}
	if s.IdleExitMin == 0 {
		s.IdleExitMin = d.IdleExitMin
	}
	if s.GoMemLimitMB == 0 {
		s.GoMemLimitMB = d.GoMemLimitMB
	}
}

// Save atomically writes settings (0600).
func (st *SettingsStore) Save(s Settings) error {
	st.mu.Lock()
	defer st.mu.Unlock()
	st.cur = s
	return st.saveLocked(s)
}

// DocumentJSON returns settings plus preserved unknown keys for API responses.
func (st *SettingsStore) DocumentJSON() (map[string]any, error) {
	st.mu.RLock()
	defer st.mu.RUnlock()
	var out map[string]any
	b, err := json.Marshal(st.cur)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	for k, v := range st.extra {
		var anyVal any
		if err := json.Unmarshal(v, &anyVal); err != nil {
			return nil, err
		}
		out[k] = anyVal
	}
	return out, nil
}

// PreviewDocument merges doc over current settings without mutating the store.
func (st *SettingsStore) PreviewDocument(doc map[string]any) (Settings, error) {
	st.mu.RLock()
	defer st.mu.RUnlock()
	return mergeDocument(st.cur, doc)
}

// CommitDocument merges doc, updates in-memory state, and saves to disk.
func (st *SettingsStore) CommitDocument(doc map[string]any) (Settings, error) {
	st.mu.Lock()
	defer st.mu.Unlock()
	s, extra, err := mergeDocumentLocked(st.cur, st.extra, doc)
	if err != nil {
		return Settings{}, err
	}
	st.cur = s
	st.extra = extra
	if err := st.saveLocked(s); err != nil {
		return Settings{}, err
	}
	return s, nil
}

func mergeDocument(cur Settings, doc map[string]any) (Settings, error) {
	s, _, err := mergeDocumentLocked(cur, nil, doc)
	return s, err
}

func mergeDocumentLocked(cur Settings, extra map[string]json.RawMessage, doc map[string]any) (Settings, map[string]json.RawMessage, error) {
	raw, err := json.Marshal(doc)
	if err != nil {
		return Settings{}, nil, err
	}
	var asMap map[string]json.RawMessage
	if err := json.Unmarshal(raw, &asMap); err != nil {
		return Settings{}, nil, err
	}
	knownBytes, err := json.Marshal(rawFromKnown(asMap))
	if err != nil {
		return Settings{}, nil, err
	}
	s := cur
	if err := json.Unmarshal(knownBytes, &s); err != nil {
		return Settings{}, nil, errors.New("invalid settings fields")
	}
	applyMissingDefaults(&s)
	if s.Keybindings == nil {
		s.Keybindings = map[string]string{}
	}
	outExtra := map[string]json.RawMessage{}
	for k, v := range extra {
		outExtra[k] = v
	}
	for k, v := range asMap {
		if _, known := knownSettingKeys[k]; !known {
			outExtra[k] = v
		}
	}
	return s, outExtra, nil
}

func (st *SettingsStore) saveLocked(s Settings) error {
	if s.Keybindings == nil {
		s.Keybindings = map[string]string{}
	}
	out := map[string]json.RawMessage{}
	known, err := json.Marshal(s)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(known, &out); err != nil {
		return err
	}
	for k, v := range st.extra {
		out[k] = v
	}
	data, err := json.Marshal(out)
	if err != nil {
		return err
	}
	dir := filepath.Dir(st.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	tmp := st.path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600) //nolint:gosec // G304: path is SettingsJSON under user config dir
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, st.path); err != nil {
		return err
	}
	return os.Chmod(st.path, 0o600)
}
