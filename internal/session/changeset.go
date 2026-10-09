package session

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/suhaasya/wisp-go/internal/driver"
	"github.com/suhaasya/wisp-go/internal/pager"
)

// Change is one staged edit operation from the page.
type Change struct {
	Op       string  `json:"op"`
	Row      []any   `json:"row,omitempty"`
	Col      string  `json:"col,omitempty"`
	Value    *string `json:"value,omitempty"`
	Null     bool    `json:"null,omitempty"`
	Original *string `json:"original,omitempty"`
	OrigNull bool    `json:"origNull,omitempty"`
}

// ChangeSet holds staged edits for one result.
type ChangeSet struct {
	keyCols []string

	updates  map[string]map[string]stagedCell
	inserts  map[string]map[string]stagedCell
	deletes  map[string]struct{}
	pkValues map[string][]any

	insertSeq     int
	lastInsertKey string
}

type stagedCell struct {
	value    *string
	null     bool
	original *string
	origNull bool
}

// NewChangeSet returns an empty change set for a table with the given key columns.
func NewChangeSet(keyCols []string) *ChangeSet {
	return &ChangeSet{
		keyCols:  append([]string(nil), keyCols...),
		updates:  make(map[string]map[string]stagedCell),
		inserts:  make(map[string]map[string]stagedCell),
		deletes:  make(map[string]struct{}),
		pkValues: make(map[string][]any),
	}
}

// LastInsertKey returns the temp row key from the most recent insert (tests).
func (cs *ChangeSet) LastInsertKey() string { return cs.lastInsertKey }

// InsertCell returns a staged value on an inserted row (tests).
func (cs *ChangeSet) InsertCell(insKey, col string) (string, bool) {
	cols, ok := cs.inserts[insKey]
	if !ok {
		return "", false
	}
	c, ok := cols[col]
	if !ok || c.null {
		return "", ok
	}
	if c.value == nil {
		return "", false
	}
	return *c.value, true
}

// Count returns the number of staged operations for display (insert=1, delete=1, each changed cell in updates).
func (cs *ChangeSet) Count() int {
	n := len(cs.deletes)
	for _, cols := range cs.updates {
		n += len(cols)
	}
	for range cs.inserts {
		n++
	}
	return n
}

// Discard clears all staged edits.
func (cs *ChangeSet) Discard() {
	cs.updates = make(map[string]map[string]stagedCell)
	cs.inserts = make(map[string]map[string]stagedCell)
	cs.deletes = make(map[string]struct{})
	cs.pkValues = make(map[string][]any)
	cs.insertSeq = 0
	cs.lastInsertKey = ""
}

func (cs *ChangeSet) rememberPK(key string, row []any) {
	if len(row) == 0 {
		return
	}
	cp := append([]any(nil), row...)
	cs.pkValues[key] = cp
}

// Apply merges one change into the set.
func (cs *ChangeSet) Apply(c Change) error {
	switch c.Op {
	case "insert":
		cs.insertSeq++
		key := fmt.Sprintf("__wisp:%d", cs.insertSeq)
		cs.inserts[key] = make(map[string]stagedCell)
		cs.lastInsertKey = key
		return nil
	case "set":
		return cs.applySet(c)
	case "delete":
		return cs.applyDelete(c)
	case "revert":
		return cs.applyRevert(c)
	default:
		return errors.New("changeset: unknown op")
	}
}

func (cs *ChangeSet) applySet(c Change) error {
	key, ins, err := cs.resolveRow(c.Row)
	if err != nil {
		return err
	}
	if _, deleted := cs.deletes[key]; deleted {
		return errors.New("changeset: row deleted")
	}
	if !ins {
		cs.rememberPK(key, c.Row)
	}
	cell := stagedCell{
		value:    c.Value,
		null:     c.Null,
		original: c.Original,
		origNull: c.OrigNull,
	}
	if ins {
		if cell.null {
			cs.inserts[key][c.Col] = cell
			return nil
		}
		cs.inserts[key][c.Col] = cell
		return nil
	}
	// existing row
	if cs.updates[key] == nil {
		cs.updates[key] = make(map[string]stagedCell)
	}
	prev, had := cs.updates[key][c.Col]
	if had {
		cell.original = prev.original
		cell.origNull = prev.origNull
	}
	if sameAsOriginal(cell) {
		delete(cs.updates[key], c.Col)
		if len(cs.updates[key]) == 0 {
			delete(cs.updates, key)
		}
		return nil
	}
	cs.updates[key][c.Col] = cell
	return nil
}

func sameAsOriginal(c stagedCell) bool {
	if c.origNull && c.null {
		return true
	}
	if c.origNull || c.null {
		return false
	}
	if c.original == nil || c.value == nil {
		return false
	}
	return *c.original == *c.value
}

func (cs *ChangeSet) applyDelete(c Change) error {
	key, ins, err := cs.resolveRow(c.Row)
	if err != nil {
		return err
	}
	if ins {
		delete(cs.inserts, key)
		return nil
	}
	cs.rememberPK(key, c.Row)
	delete(cs.updates, key)
	cs.deletes[key] = struct{}{}
	return nil
}

func (cs *ChangeSet) applyRevert(c Change) error {
	key, ins, err := cs.resolveRow(c.Row)
	if err != nil {
		return err
	}
	if ins {
		delete(cs.inserts[key], c.Col)
		return nil
	}
	if cols, ok := cs.updates[key]; ok {
		delete(cols, c.Col)
		if len(cols) == 0 {
			delete(cs.updates, key)
		}
	}
	return nil
}

func (cs *ChangeSet) resolveRow(row []any) (key string, insert bool, err error) {
	if len(row) == 0 {
		return "", false, errors.New("changeset: missing row key")
	}
	if s, ok := row[0].(string); ok && len(row) == 1 && strings.HasPrefix(s, "__wisp:") {
		if _, ok := cs.inserts[s]; ok {
			return s, true, nil
		}
		return "", false, errors.New("changeset: unknown insert row")
	}
	key = rowKeyFromPK(row)
	return key, false, nil
}

// ApplyToWindow merges staged values into w and returns overlay markers for this slice.
func (cs *ChangeSet) ApplyToWindow(w *driver.Window, windowOffset int, keyColIndexes []int) pager.Overlay {
	if w == nil {
		return pager.Overlay{}
	}
	colIndex := make(map[string]int, len(w.Columns))
	for i, c := range w.Columns {
		colIndex[c.Name] = i
	}
	var ov pager.Overlay
	for row := 0; row < w.Rows; row++ {
		key := pkKeyFromWindow(w, row, keyColIndexes)
		if _, del := cs.deletes[key]; del {
			ov.Del = append(ov.Del, row)
			continue
		}
		cols, ok := cs.updates[key]
		if !ok {
			continue
		}
		for colName, cell := range cols {
			ci, ok := colIndex[colName]
			if !ok {
				continue
			}
			applyCell(w, ci, row, cell)
			ov.Mod = append(ov.Mod, []int{row, ci})
		}
	}
	_ = windowOffset
	if len(cs.inserts) > 0 {
		appendInsertRows(w, cs, colIndex, &ov)
	}
	return ov
}

func pkKeyFromWindow(w *driver.Window, row int, keyColIndexes []int) string {
	parts := make([]any, len(keyColIndexes))
	for i, ki := range keyColIndexes {
		if s, ok := w.CellString(ki, row); ok {
			parts[i] = s
		} else {
			parts[i] = nil
		}
	}
	return rowKeyFromPK(parts)
}

func rowKeyFromPK(row []any) string {
	var b strings.Builder
	for i, v := range row {
		if i > 0 {
			b.WriteByte(0)
		}
		if v == nil {
			b.WriteString("\x00null")
			continue
		}
		switch x := v.(type) {
		case string:
			b.WriteString(x)
		case float64:
			b.WriteString(strconv.FormatFloat(x, 'f', -1, 64))
		case int64:
			b.WriteString(strconv.FormatInt(x, 10))
		case int:
			b.WriteString(strconv.Itoa(x))
		default:
			fmt.Fprint(&b, v)
		}
	}
	return b.String()
}

func applyCell(w *driver.Window, col, row int, cell stagedCell) {
	b := w.Column(col)
	if cell.null {
		b.SetNull(row)
		return
	}
	if cell.value == nil {
		return
	}
	v := *cell.value
	switch w.Columns[col].Type {
	case "int2", "int4", "int8", "integer", "bigint":
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			b.SetInt(row, n)
			return
		}
	case "bool", "boolean":
		switch strings.ToLower(v) {
		case "true", "t", "1":
			b.SetBool(row, true)
			return
		case "false", "f", "0":
			b.SetBool(row, false)
			return
		}
	}
	b.SetText(row, v)
}

func appendInsertRows(w *driver.Window, cs *ChangeSet, colIndex map[string]int, ov *pager.Overlay) {
	if len(cs.inserts) == 0 {
		return
	}
	start := w.Rows
	keys := make([]string, 0, len(cs.inserts))
	for k := range cs.inserts {
		keys = append(keys, k)
	}
	sortInsertKeys(keys)
	add := len(keys)
	if add == 0 {
		return
	}
	grown := driver.GrowWindow(w, add)
	*w = *grown
	for i, insKey := range keys {
		row := start + i
		ov.Ins = append(ov.Ins, row)
		cols := cs.inserts[insKey]
		for colName, cell := range cols {
			ci, ok := colIndex[colName]
			if !ok {
				continue
			}
			applyCell(w, ci, row, cell)
		}
	}
}

func sortInsertKeys(keys []string) {
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
}
