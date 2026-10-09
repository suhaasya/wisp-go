import type { WindowBuffer } from "./buffer";
import type { GridColumn } from "./columns";
import { selectionRange, type SelectionState } from "./selection";
import { DEFAULT_WINDOW_ROWS } from "./window";

export type CopyFormat = "tsv" | "csv" | "insert";

export type GetCellText = (row: number, colIndex: number) => string | null;

const LARGE_COPY_ROWS = 500;
const CHUNK_ROWS = 200;

function escapeCsv(value: string): string {
  if (/[",\n\r]/.test(value)) {
    return `"${value.replace(/"/g, '""')}"`;
  }
  return value;
}

function formatCell(value: string | null, format: CopyFormat): string {
  if (value === null) {
    return format === "insert" ? "NULL" : "";
  }
  return value;
}

export function formatSelectionSync(
  state: SelectionState,
  columns: GridColumn[],
  getCell: GetCellText,
  format: CopyFormat,
  tableName?: string,
): string {
  const range = selectionRange(state);
  if (!range) {
    return "";
  }
  const visibleCols = columns.filter((c) => !c.hidden);
  const colMin = range.colMax === Number.MAX_SAFE_INTEGER ? 0 : range.colMin;
  const colMax =
    range.colMax === Number.MAX_SAFE_INTEGER ? visibleCols.length - 1 : range.colMax;
  const lines: string[] = [];
  for (let row = range.rowMin; row <= range.rowMax; row++) {
    const cells: string[] = [];
    for (let ci = colMin; ci <= colMax; ci++) {
      const text = getCell(row, ci);
      cells.push(formatCell(text, format));
    }
    if (format === "insert" && tableName) {
      const names = visibleCols.slice(colMin, colMax + 1).map((c) => c.meta.name);
      lines.push(
        `INSERT INTO ${tableName} (${names.join(", ")}) VALUES (${cells.join(", ")});`,
      );
    } else if (format === "csv") {
      lines.push(cells.map(escapeCsv).join(","));
    } else {
      lines.push(cells.join("\t"));
    }
  }
  return lines.join("\n");
}

export async function copySelection(
  state: SelectionState,
  columns: GridColumn[],
  buffer: WindowBuffer,
  format: CopyFormat,
  tableName: string,
  getCellText: GetCellText,
): Promise<void> {
  const range = selectionRange(state);
  if (!range) {
    return;
  }
  const rowCount = range.rowMax - range.rowMin + 1;
  if (rowCount <= LARGE_COPY_ROWS) {
    const text = formatSelectionSync(state, columns, getCellText, format, tableName);
    await navigator.clipboard.writeText(text);
    return;
  }
  const parts: string[] = [];
  for (let start = range.rowMin; start <= range.rowMax; start += CHUNK_ROWS) {
    const end = Math.min(range.rowMax, start + CHUNK_ROWS - 1);
    await buffer.ensureViewport(
      windowOffsetForChunk(start),
      end,
    );
    await yieldToMain();
    const chunkState: SelectionState = {
      kind: "cell",
      anchor: { row: start, col: range.colMin },
      focus: { row: end, col: range.colMax === Number.MAX_SAFE_INTEGER ? columns.length - 1 : range.colMax },
    };
    parts.push(formatSelectionSync(chunkState, columns, getCellText, format, tableName));
  }
  await navigator.clipboard.writeText(parts.join("\n"));
}

function windowOffsetForChunk(row: number): number {
  return Math.floor(row / DEFAULT_WINDOW_ROWS) * DEFAULT_WINDOW_ROWS;
}

function yieldToMain(): Promise<void> {
  return new Promise((resolve) => {
    setTimeout(resolve, 0);
  });
}
