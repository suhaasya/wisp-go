import type { TableSortKey } from "../state";

/** Header click toggles primary sort; Shift+click adds or toggles secondary keys. */
export function toggleSortDraft(
  current: TableSortKey[],
  column: string,
  shiftKey: boolean,
): TableSortKey[] {
  const idx = current.findIndex((s) => s.column === column);
  if (!shiftKey) {
    if (idx === 0) {
      const only = current[0]!;
      return [{ column, desc: !only.desc }];
    }
    return [{ column, desc: false }];
  }
  if (idx >= 0) {
    const next = [...current];
    const item = next[idx]!;
    next[idx] = { column, desc: !item.desc };
    return next;
  }
  return [...current, { column, desc: false }];
}

export function sortRank(sort: TableSortKey[], column: string): number {
  const idx = sort.findIndex((s) => s.column === column);
  return idx < 0 ? -1 : idx;
}
