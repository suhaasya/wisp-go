package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/suhaasya/wisp-go/internal/store"
)

const (
	journalVersion  = 1
	journalMaxBytes = 2 * 1024 * 1024
	journalDebounce = time.Second
)

// JournalBinding is staged edits for one result binding.
type JournalBinding struct {
	TabID   string   `json:"tab_id,omitempty"`
	RID     string   `json:"rid,omitempty"`
	Schema  string   `json:"schema"`
	Table   string   `json:"table"`
	KeyCols []string `json:"key_cols"`
	Ops     []Change `json:"ops"`
}

// JournalDocument is the on-disk journal for one profile.
type JournalDocument struct {
	Version   int              `json:"version"`
	Workspace WorkspaceState   `json:"workspace"`
	Bindings  []JournalBinding `json:"bindings,omitempty"`
}

// JournalStore persists crash-safe workspace and change sets.
type JournalStore struct {
	paths     store.Paths
	workspace *WorkspaceStore
	mu        sync.Mutex
	timers    map[string]*time.Timer
	pending   map[string]struct{}
}

// NewJournalStore returns a debounced journal writer.
func NewJournalStore(paths store.Paths, workspace *WorkspaceStore) *JournalStore {
	return &JournalStore{
		paths:     paths,
		workspace: workspace,
		timers:    make(map[string]*time.Timer),
		pending:   make(map[string]struct{}),
	}
}

// ScheduleFlush debounces a journal write for profileID.
func (j *JournalStore) ScheduleFlush(profileID string, snapshot func() (JournalDocument, error)) {
	if profileID == "" {
		return
	}
	j.mu.Lock()
	defer j.mu.Unlock()
	if t, ok := j.timers[profileID]; ok {
		t.Stop()
	}
	j.pending[profileID] = struct{}{}
	j.timers[profileID] = time.AfterFunc(journalDebounce, func() {
		j.mu.Lock()
		delete(j.timers, profileID)
		delete(j.pending, profileID)
		j.mu.Unlock()
		doc, err := snapshot()
		if err != nil {
			return
		}
		_ = j.Save(profileID, doc)
	})
}

// Save writes journal atomically.
func (j *JournalStore) Save(profileID string, doc JournalDocument) error {
	if profileID == "" || strings.Contains(profileID, "/") {
		return errors.New("journal: profile id required")
	}
	if doc.Version == 0 {
		doc.Version = journalVersion
	}
	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	if len(data) > journalMaxBytes {
		return fmt.Errorf("journal: exceeds %d byte cap", journalMaxBytes)
	}
	j.mu.Lock()
	defer j.mu.Unlock()
	path := j.paths.JournalJSON(profileID)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Load reads a journal; missing file returns empty doc.
func (j *JournalStore) Load(profileID string) (JournalDocument, error) {
	path := j.paths.JournalJSON(profileID)
	data, err := os.ReadFile(filepath.Clean(path)) //nolint:gosec // G304: profile journal path
	if err != nil {
		if os.IsNotExist(err) {
			return JournalDocument{Version: journalVersion}, nil
		}
		return JournalDocument{}, err
	}
	var doc JournalDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return JournalDocument{}, fmt.Errorf("journal json: %w", err)
	}
	if doc.Version == 0 {
		doc.Version = journalVersion
	}
	return doc, nil
}

// Clear removes the journal file (clean shutdown).
func (j *JournalStore) Clear(profileID string) error {
	j.mu.Lock()
	if t, ok := j.timers[profileID]; ok {
		t.Stop()
		delete(j.timers, profileID)
	}
	delete(j.pending, profileID)
	j.mu.Unlock()
	path := j.paths.JournalJSON(profileID)
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// Restore applies journal workspace to disk and returns bindings to replay.
func (j *JournalStore) Restore(profileID string) (JournalDocument, error) {
	doc, err := j.Load(profileID)
	if err != nil {
		return JournalDocument{}, err
	}
	if len(doc.Workspace.Tabs) > 0 && j.workspace != nil {
		if err := j.workspace.Save(profileID, doc.Workspace); err != nil {
			return JournalDocument{}, err
		}
	}
	return doc, nil
}

// Snapshot builds a journal document from workspace and session bindings.
func SnapshotJournal(profileID string, ws WorkspaceState, sessions []*Session) JournalDocument {
	doc := JournalDocument{Version: journalVersion, Workspace: ws}
	for _, s := range sessions {
		if s == nil || s.profile.ID != profileID {
			continue
		}
		doc.Bindings = append(doc.Bindings, s.journalBindings()...)
	}
	return doc
}

func (s *Session) journalBindings() []JournalBinding {
	s.changeMu.RLock()
	defer s.changeMu.RUnlock()
	var out []JournalBinding
	for rid, rc := range s.resultChanges {
		if rc == nil || len(rc.journal) == 0 {
			continue
		}
		out = append(out, JournalBinding{
			TabID:   rc.tabID,
			RID:     rid,
			Schema:  rc.schema,
			Table:   rc.table,
			KeyCols: append([]string(nil), rc.keyCols...),
			Ops:     append([]Change(nil), rc.journal...),
		})
	}
	return out
}
