package session

import (
	"errors"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/pager"
)

var errChangeNotEditable = errors.New("session: table has no primary key")

type resultChanges struct {
	cs       *ChangeSet
	keyCols  []string
	schema   string
	table    string
	editable bool
}

func (s *Session) registerResultChanges(rid string, keyCols []string, schema, table string) {
	s.changeMu.Lock()
	defer s.changeMu.Unlock()
	if s.resultChanges == nil {
		s.resultChanges = make(map[string]*resultChanges)
	}
	s.resultChanges[rid] = &resultChanges{
		cs:       NewChangeSet(keyCols),
		keyCols:  append([]string(nil), keyCols...),
		schema:   schema,
		table:    table,
		editable: len(keyCols) > 0,
	}
}

func (s *Session) dropResultChanges(rid string) {
	s.changeMu.Lock()
	delete(s.resultChanges, rid)
	s.changeMu.Unlock()
}

func (s *Session) resultChangesFor(rid string) (*resultChanges, bool) {
	s.changeMu.RLock()
	rc, ok := s.resultChanges[rid]
	s.changeMu.RUnlock()
	return rc, ok
}

func (s *Session) wirePagerOverlay() {
	p := s.ensurePager()
	if p == nil {
		return
	}
	p.SetOverlayProvider(s.resultOverlay)
}

func (s *Session) resultOverlay(rid string, w *driver.Window, offset int) pager.Overlay {
	rc, ok := s.resultChangesFor(rid)
	if !ok || rc == nil || rc.cs == nil || w == nil {
		return pager.Overlay{}
	}
	return rc.cs.ApplyToWindow(w, offset, keyColIndexes(w.Columns, rc.keyCols))
}

// ApplyResultChange stages one edit; returns staged count after apply.
func (s *Session) ApplyResultChange(rid string, c Change) (int, error) {
	if s.profile.ReadOnly {
		return 0, ErrWriteBlockedReadOnly()
	}
	rc, ok := s.resultChangesFor(rid)
	if !ok || rc == nil {
		return 0, errors.New("session: unknown result")
	}
	if !rc.editable {
		return 0, errChangeNotEditable
	}
	if err := rc.cs.Apply(c); err != nil {
		return 0, err
	}
	if p := s.ensurePager(); p != nil {
		p.InvalidateWindows(rid)
	}
	return rc.cs.Count(), nil
}

// ResultChangeCount returns staged change count for a result.
func (s *Session) ResultChangeCount(rid string) int {
	rc, ok := s.resultChangesFor(rid)
	if !ok || rc == nil || rc.cs == nil {
		return 0
	}
	return rc.cs.Count()
}

// DiscardResultChanges clears staged edits for a result.
func (s *Session) DiscardResultChanges(rid string) error {
	rc, ok := s.resultChangesFor(rid)
	if !ok || rc == nil || rc.cs == nil {
		return errors.New("session: unknown result")
	}
	rc.cs.Discard()
	if p := s.ensurePager(); p != nil {
		p.InvalidateWindows(rid)
	}
	return nil
}

// ResultInsertKey returns the temp row key after op insert (empty if none).
func (s *Session) ResultLastInsertKey(rid string) string {
	rc, ok := s.resultChangesFor(rid)
	if !ok || rc == nil || rc.cs == nil {
		return ""
	}
	return rc.cs.LastInsertKey()
}

func keyColIndexes(cols []driver.ColumnMeta, names []string) []int {
	out := make([]int, 0, len(names))
	for _, n := range names {
		for i, c := range cols {
			if c.Name == n {
				out = append(out, i)
				break
			}
		}
	}
	return out
}
