export type ColumnMeta = {
  name: string;
  type: string;
};

export type RowTotal = {
  kind: "exact" | "estimate" | string;
  value: number;
};

export type WindowOverlay = {
  mod: number[][];
  ins: number[];
  del: number[];
};

/** Columnar window payload from GET .../results/{rid}/window */
export type WindowPayload = {
  rid: string;
  offset: number;
  total: RowTotal;
  columns: ColumnMeta[];
  data: unknown[][];
  nulls?: number[][];
  truncated?: number[][];
  overlay?: WindowOverlay;
};

export const DEFAULT_WINDOW_ROWS = 150;

export function windowOffsetForRow(row: number, windowRows: number): number {
  if (row < 0) {
    return 0;
  }
  return Math.floor(row / windowRows) * windowRows;
}
