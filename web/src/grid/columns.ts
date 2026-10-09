import type { ColumnMeta } from "./window";

export type GridColumn = {
  id: string;
  meta: ColumnMeta;
  width: number;
  hidden: boolean;
};

const DEFAULT_WIDTH = 140;
const MIN_WIDTH = 60;
const RN_WIDTH = 44;

export function buildColumns(meta: ColumnMeta[]): GridColumn[] {
  return meta.map((m) => ({
    id: m.name,
    meta: m,
    width: DEFAULT_WIDTH,
    hidden: false,
  }));
}

export function columnWindow(
  columns: GridColumn[],
  scrollLeft: number,
  viewportWidth: number,
): { active: GridColumn[]; start: number; visible: GridColumn[]; leadWidth: number } {
  const active = columns.filter((c) => !c.hidden);
  const { start, visible } = visibleColumnIndices(columns, scrollLeft, viewportWidth);
  let leadWidth = 0;
  for (let i = 0; i < start; i++) {
    leadWidth += active[i]!.width;
  }
  return { active, start, visible, leadWidth };
}

export function visibleColumnIndices(
  columns: GridColumn[],
  scrollLeft: number,
  viewportWidth: number,
  bufferCols = 2,
): { start: number; end: number; visible: GridColumn[] } {
  const active = columns.filter((c) => !c.hidden);
  if (active.length === 0) {
    return { start: 0, end: 0, visible: [] };
  }
  let x = RN_WIDTH;
  let startIdx = 0;
  let endIdx = active.length;
  for (let i = 0; i < active.length; i++) {
    const col = active[i]!;
    const left = x;
    const right = x + col.width;
    if (right > scrollLeft && left < scrollLeft) {
      startIdx = i;
    }
    if (left >= scrollLeft + viewportWidth) {
      endIdx = i;
      break;
    }
    x = right;
  }
  startIdx = Math.max(0, startIdx - bufferCols);
  endIdx = Math.min(active.length, endIdx + bufferCols);
  return { start: startIdx, end: endIdx, visible: active.slice(startIdx, endIdx) };
}

export function totalBodyWidth(columns: GridColumn[]): number {
  let w = RN_WIDTH;
  for (const c of columns) {
    if (!c.hidden) {
      w += c.width;
    }
  }
  return w;
}

export function clampColumnWidth(width: number): number {
  return Math.max(MIN_WIDTH, Math.min(480, width));
}

export function rowNumberWidth(): number {
  return RN_WIDTH;
}

export function isNumericType(type: string): boolean {
  const t = type.toLowerCase();
  return (
    t.includes("int") ||
    t.includes("numeric") ||
    t.includes("decimal") ||
    t.includes("float") ||
    t.includes("double") ||
    t.includes("real") ||
    t === "money" ||
    t === "number"
  );
}

export function isEnumType(type: string): boolean {
  return type.includes("_") && !type.includes("[]") && !type.includes("timestamp");
}
