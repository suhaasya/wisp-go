package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const snippetsVersion = 1

// Snippet is a saved SQL fragment.
type Snippet struct {
	ID   string   `json:"id"`
	Name string   `json:"name"`
	Tags []string `json:"tags,omitempty"`
	SQL  string   `json:"sql"`
}

// SnippetsDocument is snippets.json on disk.
type SnippetsDocument struct {
	Version  int       `json:"version"`
	Snippets []Snippet `json:"snippets"`
}

// SnippetsStore persists snippets.json.
type SnippetsStore struct {
	paths Paths
	mu    sync.RWMutex
}

// NewSnippetsStore returns a snippets store.
func NewSnippetsStore(paths Paths) *SnippetsStore {
	return &SnippetsStore{paths: paths}
}

// Load reads snippets from disk.
func (s *SnippetsStore) Load() (SnippetsDocument, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	path := s.paths.SnippetsJSON()
	data, err := os.ReadFile(path) //nolint:gosec // G304: data dir
	if err != nil {
		if os.IsNotExist(err) {
			return SnippetsDocument{Version: snippetsVersion, Snippets: nil}, nil
		}
		return SnippetsDocument{}, err
	}
	var doc SnippetsDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return SnippetsDocument{}, err
	}
	if doc.Version == 0 {
		doc.Version = snippetsVersion
	}
	return doc, nil
}

// Save writes snippets atomically.
func (s *SnippetsStore) Save(doc SnippetsDocument) error {
	if doc.Version == 0 {
		doc.Version = snippetsVersion
	}
	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	if len(data) > 512*1024 {
		return errors.New("snippets: document too large")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	path := s.paths.SnippetsJSON()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Upsert adds or replaces a snippet by id.
func (s *SnippetsStore) Upsert(sn Snippet) error {
	sn.Name = strings.TrimSpace(sn.Name)
	sn.SQL = strings.TrimSpace(sn.SQL)
	if sn.Name == "" || sn.SQL == "" {
		return errors.New("snippets: name and sql required")
	}
	if sn.ID == "" {
		sn.ID = NewUUIDv7()
	}
	doc, err := s.Load()
	if err != nil {
		return err
	}
	replaced := false
	for i := range doc.Snippets {
		if doc.Snippets[i].ID == sn.ID {
			doc.Snippets[i] = sn
			replaced = true
			break
		}
	}
	if !replaced {
		doc.Snippets = append(doc.Snippets, sn)
	}
	return s.Save(doc)
}
