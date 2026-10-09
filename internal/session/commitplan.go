package session

import (
	"sort"

	"github.com/suhaasya/wisp-go/internal/sqlgen"
)

func (cs *ChangeSet) CommitPlan() sqlgen.CommitPlan {
	var plan sqlgen.CommitPlan
	for key := range cs.deletes {
		pk := cs.pkValues[key]
		plan.Deletes = append(plan.Deletes, sqlgen.RowDelete{
			KeyColumns: append([]string(nil), cs.keyCols...),
			KeyValues:  append([]any(nil), pk...),
		})
	}
	sort.Slice(plan.Deletes, func(i, j int) bool {
		return rowKeyFromPK(plan.Deletes[i].KeyValues) < rowKeyFromPK(plan.Deletes[j].KeyValues)
	})

	for key, cols := range cs.updates {
		pk := cs.pkValues[key]
		cells := make(map[string]sqlgen.CellChange, len(cols))
		for col, sc := range cols {
			cells[col] = stagedToCell(sc)
		}
		plan.Updates = append(plan.Updates, sqlgen.RowUpdate{
			KeyColumns: append([]string(nil), cs.keyCols...),
			KeyValues:  append([]any(nil), pk...),
			Cells:      cells,
		})
	}
	sort.Slice(plan.Updates, func(i, j int) bool {
		return rowKeyFromPK(plan.Updates[i].KeyValues) < rowKeyFromPK(plan.Updates[j].KeyValues)
	})

	for insKey, cols := range cs.inserts {
		cells := make(map[string]sqlgen.CellChange, len(cols))
		for col, sc := range cols {
			cells[col] = stagedToCell(sc)
		}
		_ = insKey
		plan.Inserts = append(plan.Inserts, sqlgen.RowInsert{Cells: cells})
	}
	sort.Slice(plan.Inserts, func(i, j int) bool {
		return insertSortKey(plan.Inserts[i]) < insertSortKey(plan.Inserts[j])
	})
	return plan
}

func stagedToCell(sc stagedCell) sqlgen.CellChange {
	return sqlgen.CellChange{
		Null:     sc.null,
		Value:    sc.value,
		OrigNull: sc.origNull,
		Original: sc.original,
	}
}

func insertSortKey(ins sqlgen.RowInsert) string {
	cols := sortedInsertCols(ins.Cells)
	if len(cols) == 0 {
		return ""
	}
	return cols[0]
}

func sortedInsertCols(m map[string]sqlgen.CellChange) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
