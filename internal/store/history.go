package store

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	HistoryRotateSizeBytes = 5 * 1024 * 1024
	historyMaxMatches      = 200
)

// HistoryEntry is one executed statement record (no bound parameter values).
type HistoryEntry struct {
	TS     int64  `json:"ts"`
	SQL    string `json:"sql"`
	MS     int64  `json:"ms,omitempty"`
	Rows   int64  `json:"rows,omitempty"`
	Status string `json:"status"`
}

// HistoryStore appends JSONL query history per profile.
type HistoryStore struct {
	paths Paths
	mu    sync.Mutex
}

// NewHistoryStore returns a history store under paths.DataDir.
func NewHistoryStore(paths Paths) *HistoryStore {
	return &HistoryStore{paths: paths}
}

// Append records one history line when enabled for the profile.
func (h *HistoryStore) Append(profile Profile, entry HistoryEntry) error {
	if profile.DisableQueryHistory {
		return nil
	}
	if profile.ID == "" || strings.Contains(profile.ID, "/") || strings.Contains(profile.ID, "..") {
		return errors.New("history: profile id required")
	}
	entry.SQL = strings.TrimSpace(entry.SQL)
	if entry.SQL == "" {
		return errors.New("history: sql required")
	}
	if entry.Status == "" {
		entry.Status = "ok"
	}
	line, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	path := h.paths.HistoryJSONL(profile.ID)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600) //nolint:gosec // profile-scoped data path
	if err != nil {
		return err
	}
	if _, err := f.Write(append(line, '\n')); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return h.rotateIfNeededLocked(profile.ID, path)
}

// RotateIfNeeded compacts history when the file exceeds the rotate threshold.
func (h *HistoryStore) RotateIfNeeded(profile Profile) error {
	if profile.ID == "" {
		return errors.New("history: profile id required")
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.rotateIfNeededLocked(profile.ID, h.paths.HistoryJSONL(profile.ID))
}

func (h *HistoryStore) rotateIfNeededLocked(profileID, path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if info.Size() < HistoryRotateSizeBytes {
		return nil
	}
	data, err := os.ReadFile(path) //nolint:gosec // G304: profile history path
	if err != nil {
		return err
	}
	lines := splitJSONL(data)
	if len(lines) <= 1 {
		return nil
	}
	keepFrom := len(lines) / 2
	kept := lines[keepFrom:]
	out := bytes.Join(kept, nil)
	if len(kept) > 0 && !bytes.HasSuffix(out, []byte("\n")) {
		out = append(out, '\n')
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, out, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func splitJSONL(data []byte) [][]byte {
	var lines [][]byte
	for _, part := range bytes.Split(data, []byte("\n")) {
		if len(bytes.TrimSpace(part)) == 0 {
			continue
		}
		lines = append(lines, append(part, '\n'))
	}
	return lines
}

// Search writes up to 200 matching entries as NDJSON, newest first.
func (h *HistoryStore) Search(profile Profile, q string, w *bytes.Buffer) error {
	if profile.ID == "" {
		return errors.New("history: profile id required")
	}
	path := h.paths.HistoryJSONL(profile.ID)
	data, err := os.ReadFile(path) //nolint:gosec // G304: profile history path
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	q = strings.ToLower(strings.TrimSpace(q))
	lines := splitJSONL(data)
	matches := 0
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		if q != "" && !bytes.Contains(bytes.ToLower(line), []byte(q)) {
			continue
		}
		w.Write(line)
		matches++
		if matches >= historyMaxMatches {
			break
		}
	}
	return nil
}

// SearchReader streams matches to w (for HTTP).
func (h *HistoryStore) SearchReader(profile Profile, q string, w interface{ Write([]byte) (int, error) }) error {
	var buf bytes.Buffer
	if err := h.Search(profile, q, &buf); err != nil {
		return err
	}
	if buf.Len() == 0 {
		return nil
	}
	_, err := w.Write(buf.Bytes())
	return err
}

// CountLines returns non-empty JSONL lines (tests).
func (h *HistoryStore) CountLines(profileID string) (int, error) {
	data, err := os.ReadFile(h.paths.HistoryJSONL(profileID)) //nolint:gosec // G304: test helper
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	return len(splitJSONL(data)), nil
}

// ValidateHistoryEntry rejects empty sql after sanitize.
func ValidateHistoryEntry(e HistoryEntry) error {
	if strings.TrimSpace(e.SQL) == "" {
		return fmt.Errorf("history: empty sql")
	}
	return nil
}

// ScanHistoryFile reads entries in reverse with optional filter (tests).
func ScanHistoryFile(path string, q string, max int, fn func(HistoryEntry) bool) error {
	f, err := os.Open(path) //nolint:gosec // G304: tests
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	if err := sc.Err(); err != nil {
		return err
	}
	q = strings.ToLower(q)
	n := 0
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(line), q) {
			continue
		}
		var e HistoryEntry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			continue
		}
		if !fn(e) {
			break
		}
		n++
		if n >= max {
			break
		}
	}
	return nil
}
