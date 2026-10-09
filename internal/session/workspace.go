package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/suhaasya/wisp-go/internal/sqlgen"
	"github.com/suhaasya/wisp-go/internal/store"
)

const workspaceVersion = 1

const (
	TabKindTableData      = "table_data"
	TabKindTableStructure = "table_structure"
	TabKindQuery          = "query"
	TabKindServerTools    = "server_tools"
)

// WorkspaceTabState is serializable tab state kept in Go (not live DOM).
type WorkspaceTabState struct {
	ID        string                `json:"id"`
	Kind      string                `json:"kind"`
	Schema    string                `json:"schema,omitempty"`
	Table     string                `json:"table,omitempty"`
	Pinned    bool                  `json:"pinned,omitempty"`
	QueryText string                `json:"query_text,omitempty"`
	ScrollTop float64               `json:"scroll_top,omitempty"`
	Selection *GridSelectionState   `json:"selection,omitempty"`
	Browse    *WorkspaceBrowseState `json:"browse,omitempty"`
}

// WorkspaceBrowseState mirrors applied table filter/sort from the page.
type WorkspaceBrowseState struct {
	RawMode  bool               `json:"raw_mode,omitempty"`
	RawWhere string             `json:"raw_where,omitempty"`
	Filter   sqlgen.FilterGroup `json:"filter,omitempty"`
	Sort     []sqlgen.SortKey   `json:"sort,omitempty"`
}

// GridSelectionState holds selected cell/row for restore.
type GridSelectionState struct {
	AnchorRow int `json:"anchor_row"`
	AnchorCol int `json:"anchor_col"`
}

// WorkspaceState is the open workspace for one profile.
type WorkspaceState struct {
	Version     int                 `json:"version"`
	ActiveTabID string              `json:"active_tab_id,omitempty"`
	Tabs        []WorkspaceTabState `json:"tabs"`
}

// WorkspaceStore loads and saves workspace JSON per profile.
type WorkspaceStore struct {
	paths    store.Paths
	profiles *store.ProfileStore
	mu       sync.Mutex
}

// NewWorkspaceStore returns a store backed by profile-scoped JSON files.
func NewWorkspaceStore(paths store.Paths, profiles *store.ProfileStore) *WorkspaceStore {
	return &WorkspaceStore{paths: paths, profiles: profiles}
}

// Load reads workspace state for profileID.
func (w *WorkspaceStore) Load(profileID string) (WorkspaceState, error) {
	if err := w.requireProfile(profileID); err != nil {
		return WorkspaceState{}, err
	}
	path := w.paths.WorkspaceJSON(profileID)
	data, err := os.ReadFile(filepath.Clean(path)) //nolint:gosec // profileID validated; path under DataDir/workspace
	if err != nil {
		if os.IsNotExist(err) {
			return WorkspaceState{Version: workspaceVersion, Tabs: nil}, nil
		}
		return WorkspaceState{}, err
	}
	var st WorkspaceState
	if err := json.Unmarshal(data, &st); err != nil {
		return WorkspaceState{}, fmt.Errorf("workspace json: %w", err)
	}
	if st.Version == 0 {
		st.Version = workspaceVersion
	}
	return st, nil
}

// Save writes workspace state atomically.
func (w *WorkspaceStore) Save(profileID string, st WorkspaceState) error {
	if err := w.requireProfile(profileID); err != nil {
		return err
	}
	if st.Version == 0 {
		st.Version = workspaceVersion
	}
	data, err := json.Marshal(st)
	if err != nil {
		return err
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	path := w.paths.WorkspaceJSON(profileID)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func (w *WorkspaceStore) requireProfile(profileID string) error {
	if profileID == "" || strings.Contains(profileID, "/") || strings.Contains(profileID, "..") {
		return errors.New("workspace: profile id required")
	}
	if w.profiles == nil {
		return errors.New("workspace: no profile store")
	}
	if _, ok := w.profiles.Get(profileID); !ok {
		return ErrProfileNotFound
	}
	return nil
}

// ErrProfileNotFound is returned when the profile id is unknown.
var ErrProfileNotFound = errors.New("workspace: profile not found")
