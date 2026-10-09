export type CellRef = { row: number; col: number };

export type SelectionState =
  | { kind: "none" }
  | { kind: "cell"; anchor: CellRef; focus: CellRef }
  | { kind: "row"; rows: Set<number> };

export function emptySelection(): SelectionState {
  return { kind: "none" };
}

export function selectCell(row: number, col: number): SelectionState {
  const ref = { row, col };
  return { kind: "cell", anchor: ref, focus: ref };
}

export function extendCellSelection(state: SelectionState, row: number, col: number): SelectionState {
  if (state.kind === "cell") {
    return { kind: "cell", anchor: state.anchor, focus: { row, col } };
  }
  return selectCell(row, col);
}

export function selectRow(row: number): SelectionState {
  return { kind: "row", rows: new Set([row]) };
}

export function toggleRowSelection(state: SelectionState, row: number, extend: boolean): SelectionState {
  if (state.kind === "row" && extend) {
    const rows = new Set(state.rows);
    if (rows.has(row)) {
      rows.delete(row);
    } else {
      rows.add(row);
    }
    return { kind: "row", rows };
  }
  return selectRow(row);
}

export function selectionRange(state: SelectionState): {
  rowMin: number;
  rowMax: number;
  colMin: number;
  colMax: number;
} | null {
  if (state.kind === "cell") {
    return {
      rowMin: Math.min(state.anchor.row, state.focus.row),
      rowMax: Math.max(state.anchor.row, state.focus.row),
      colMin: Math.min(state.anchor.col, state.focus.col),
      colMax: Math.max(state.anchor.col, state.focus.col),
    };
  }
  if (state.kind === "row" && state.rows.size > 0) {
    const rows = [...state.rows];
    const rowMin = Math.min(...rows);
    const rowMax = Math.max(...rows);
    return { rowMin, rowMax, colMin: 0, colMax: Number.MAX_SAFE_INTEGER };
  }
  return null;
}

export function rowIsSelected(state: SelectionState, row: number): boolean {
  if (state.kind === "row") {
    return state.rows.has(row);
  }
  if (state.kind === "cell") {
    const r = selectionRange(state);
    return r !== null && row >= r.rowMin && row <= r.rowMax;
  }
  return false;
}

export function cellInSelection(state: SelectionState, row: number, col: number): boolean {
  const r = selectionRange(state);
  if (!r) {
    return false;
  }
  if (col > r.colMax && r.colMax === Number.MAX_SAFE_INTEGER) {
    return row >= r.rowMin && row <= r.rowMax;
  }
  return row >= r.rowMin && row <= r.rowMax && col >= r.colMin && col <= r.colMax;
}
