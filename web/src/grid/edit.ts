import type { GridColumn } from "./columns";
import { isEnumType, isNumericType } from "./columns";
import { mountBoolEditor } from "./editors/bool";
import { mountDateTimeEditor } from "./editors/datetime";
import { mountEnumEditor } from "./editors/enum";
import { mountJSONEditor } from "./editors/json";
import { mountTextEditor } from "./editors/text";
import type { ChangeRequest } from "./results-client";

export type EditContext = {
  sessionId: string;
  rid: string;
  primaryKey: string[];
  columns: GridColumn[];
  getCell: (row: number, col: number) => unknown;
  postChange: (body: ChangeRequest) => Promise<{ count: number }>;
  onStagedCount: (n: number) => void;
  invalidateWindows: () => void;
};

function isBoolType(type: string): boolean {
  const t = type.toLowerCase();
  return t === "bool" || t === "boolean";
}

function isDateTimeType(type: string): boolean {
  const t = type.toLowerCase();
  return t.includes("timestamp") || t.includes("date") || t === "time";
}

function isJSONType(type: string): boolean {
  const t = type.toLowerCase();
  return t === "json" || t === "jsonb";
}

export function pkRowValues(
  ctx: Pick<EditContext, "primaryKey" | "columns" | "getCell">,
  row: number,
): unknown[] {
  const out: unknown[] = [];
  for (const pk of ctx.primaryKey) {
    const colIndex = ctx.columns.findIndex((c) => c.id === pk);
    if (colIndex < 0) {
      continue;
    }
    const v = ctx.getCell(row, colIndex);
    if (typeof v === "string" && ctx.columns[colIndex]?.meta.type.includes("int")) {
      const n = Number(v);
      out.push(Number.isFinite(n) ? n : v);
    } else {
      out.push(v);
    }
  }
  return out;
}

function originalFields(raw: unknown): Pick<ChangeRequest, "original" | "origNull"> {
  if (raw === null || raw === undefined) {
    return { origNull: true };
  }
  return { original: String(raw) };
}

export function openCellEditor(
  ctx: EditContext,
  row: number,
  colIndex: number,
  cellEl: HTMLElement,
  enumValues?: string[],
  onTab?: () => void,
): () => void {
  const col = ctx.columns[colIndex];
  if (!col) {
    return () => {};
  }
  const raw = ctx.getCell(row, colIndex);
  const rowKey = pkRowValues(ctx, row);
  const orig = originalFields(raw);
  let disposed = false;

  const finish = (cleanup: () => void) => {
    if (disposed) {
      return;
    }
    disposed = true;
    cleanup();
  };

  const commitSet = async (body: Partial<ChangeRequest>) => {
    const resp = await ctx.postChange({
      op: "set",
      row: rowKey,
      col: col.id,
      ...orig,
      ...body,
    });
    ctx.onStagedCount(resp.count);
    ctx.invalidateWindows();
  };

  const onCancel = () => finish(() => {});

  const typ = col.meta.type;
  if (isBoolType(typ)) {
    const initial = raw === null ? null : String(raw) === "true";
    return mountBoolEditor(cellEl, initial, (v) => {
      void commitSet({ value: v ? "true" : "false", null: false }).finally(() =>
        finish(() => {}),
      );
    }, onCancel);
  }
  if (isEnumType(typ) && enumValues && enumValues.length > 0) {
    return mountEnumEditor(cellEl, raw === null ? enumValues[0]! : String(raw), enumValues, (v) => {
      void commitSet({ value: v, null: false }).finally(() => finish(() => {}));
    }, onCancel);
  }
  if (isDateTimeType(typ)) {
    return mountDateTimeEditor(cellEl, raw === null ? "" : String(raw), (v) => {
      void commitSet(v === "" ? { null: true } : { value: v, null: false }).finally(() =>
        finish(() => {}),
      );
    }, onCancel);
  }
  if (isJSONType(typ)) {
    return mountJSONEditor(cellEl, raw === null ? "" : String(raw), (v) => {
      void commitSet(v === "" ? { null: true } : { value: v, null: false }).finally(() =>
        finish(() => {}),
      );
    }, onCancel);
  }
  return mountTextEditor(
    cellEl,
    raw === null ? "" : String(raw),
    (v) => {
      if (v === "" && isNumericType(typ)) {
        void commitSet({ null: true }).finally(() => finish(() => {}));
        return;
      }
      void commitSet(v === "" ? { null: true } : { value: v, null: false }).finally(() =>
        finish(() => {}),
      );
    },
    onCancel,
    onTab,
  );
}

export async function setCellDefault(ctx: EditContext, row: number, colIndex: number): Promise<void> {
  const col = ctx.columns[colIndex];
  if (!col) {
    return;
  }
  const raw = ctx.getCell(row, colIndex);
  const resp = await ctx.postChange({
    op: "set",
    row: pkRowValues(ctx, row),
    col: col.id,
    default: true,
    ...originalFields(raw),
  });
  ctx.onStagedCount(resp.count);
  ctx.invalidateWindows();
}

export async function setCellNull(ctx: EditContext, row: number, colIndex: number): Promise<void> {
  const col = ctx.columns[colIndex];
  if (!col) {
    return;
  }
  const raw = ctx.getCell(row, colIndex);
  const resp = await ctx.postChange({
    op: "set",
    row: pkRowValues(ctx, row),
    col: col.id,
    null: true,
    ...originalFields(raw),
  });
  ctx.onStagedCount(resp.count);
  ctx.invalidateWindows();
}

export async function revertCell(ctx: EditContext, row: number, colIndex: number): Promise<void> {
  const col = ctx.columns[colIndex];
  if (!col) {
    return;
  }
  const resp = await ctx.postChange({
    op: "revert",
    row: pkRowValues(ctx, row),
    col: col.id,
  });
  ctx.onStagedCount(resp.count);
  ctx.invalidateWindows();
}
