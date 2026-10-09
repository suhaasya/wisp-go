import { useEffect, useRef, useState } from "preact/hooks";
import type { ColumnMeta } from "../grid/window";
import { fetchCellValue } from "../grid/results-client";
import { Input } from "../ui/Input";
import { buildForeignKeyJump, type ForeignKeyMeta } from "./fk-jump";
import { openFilteredTableTab } from "./tabs";

const WIDTH_KEY = "wisp.detailPanelWidth";
const MIN_W = 220;
const MAX_W = 520;
const DEFAULT_W = 280;

export function loadDetailPanelWidth(): number {
  if (typeof localStorage === "undefined") {
    return DEFAULT_W;
  }
  const n = Number(localStorage.getItem(WIDTH_KEY));
  if (!Number.isFinite(n)) {
    return DEFAULT_W;
  }
  return Math.min(MAX_W, Math.max(MIN_W, n));
}

export function saveDetailPanelWidth(w: number): void {
  if (typeof localStorage === "undefined") {
    return;
  }
  localStorage.setItem(WIDTH_KEY, String(Math.min(MAX_W, Math.max(MIN_W, w))));
}

export function formatPreviewValue(v: unknown): string {
  if (v === null || v === undefined) {
    return "NULL";
  }
  return String(v);
}

export type DetailPanelProps = {
  open: boolean;
  width: number;
  onWidthChange: (w: number) => void;
  sessionId: string;
  rid: string;
  schema: string;
  table: string;
  row: number | null;
  columns: ColumnMeta[];
  primaryKey: string[];
  foreignKeys: ForeignKeyMeta[];
  getCell: (row: number, col: number) => unknown;
  isTruncated: (row: number, col: number) => boolean;
  editable?: boolean;
  onCellCommit?: (
    colIndex: number,
    colName: string,
    next: { value?: string; null?: boolean; default?: boolean },
  ) => void | Promise<void>;
};

export function DetailPanel(props: DetailPanelProps) {
  const cacheRef = useRef(new Map<string, string>());
  const inflightRef = useRef<AbortController | null>(null);
  const [focusedCol, setFocusedCol] = useState<string | null>(null);
  const [fullText, setFullText] = useState<string | null>(null);
  const [fullJSON, setFullJSON] = useState(false);

  useEffect(() => {
    cacheRef.current.clear();
    inflightRef.current?.abort();
    inflightRef.current = null;
    setFocusedCol(null);
    setFullText(null);
    setFullJSON(false);
  }, [props.row]);

  if (!props.open) {
    return null;
  }

  const row = props.row;
  const pkLabel =
    row !== null && props.primaryKey.length > 0
      ? props.primaryKey
          .map((name) => {
            const ci = props.columns.findIndex((c) => c.name === name);
            if (ci < 0) {
              return "?";
            }
            return formatPreviewValue(props.getCell(row, ci));
          })
          .join(", ")
      : row !== null
        ? String(row + 1)
        : "—";

  async function onFieldFocus(colName: string, colIndex: number) {
    if (row === null) {
      return;
    }
    setFocusedCol(colName);
    const cacheKey = `${row}:${colName}`;
    const cached = cacheRef.current.get(cacheKey);
    if (cached !== undefined) {
      setFullText(cached);
      setFullJSON(looksLikeJSON(cached));
      return;
    }
    if (!props.isTruncated(row, colIndex)) {
      const preview = formatPreviewValue(props.getCell(row, colIndex));
      setFullText(preview);
      setFullJSON(looksLikeJSON(preview));
      return;
    }
    inflightRef.current?.abort();
    const ac = new AbortController();
    inflightRef.current = ac;
    try {
      const out = await fetchCellValue(props.sessionId, props.rid, row, colName, ac.signal);
      if (ac.signal.aborted) {
        return;
      }
      cacheRef.current.set(cacheKey, out.value);
      setFullText(out.value);
      setFullJSON(Boolean(out.json));
    } catch {
      if (!ac.signal.aborted) {
        setFullText(formatPreviewValue(props.getCell(row, colIndex)));
      }
    }
  }

  function fkForColumn(colName: string): ForeignKeyMeta | undefined {
    return props.foreignKeys.find((fk) => fk.columns[0] === colName);
  }

  return (
    <aside
      class="relative flex shrink-0 flex-col overflow-auto border-l border-line bg-panel px-3.5 py-3"
      style={{ width: `${props.width}px` }}
      data-testid="row-detail-panel"
    >
      <div
        class="absolute left-0 top-0 z-10 h-full w-1 cursor-col-resize"
        role="separator"
        aria-orientation="vertical"
        onPointerDown={(e) => {
          const startX = e.clientX;
          const startW = props.width;
          const onMove = (ev: PointerEvent) => {
            const next = startW - (ev.clientX - startX);
            props.onWidthChange(next);
          };
          const onUp = () => {
            window.removeEventListener("pointermove", onMove);
            window.removeEventListener("pointerup", onUp);
            saveDetailPanelWidth(props.width);
          };
          window.addEventListener("pointermove", onMove);
          window.addEventListener("pointerup", onUp);
        }}
      />
      <h4 class="m-0 text-13 font-semibold text-ink">Row {pkLabel}</h4>
      <p class="sub mb-3 text-11-5 text-ink-3">
        {props.schema}.{props.table}, selected row
      </p>
      {row === null ? (
        <p class="text-12 text-ink-2">Select a row in the grid.</p>
      ) : (
        props.columns.map((col, colIndex) => {
          const preview = formatPreviewValue(props.getCell(row, colIndex));
          const fk = fkForColumn(col.name);
          const localVals: Record<string, string | null> = {};
          if (fk) {
            for (const c of fk.columns) {
              const i = props.columns.findIndex((x) => x.name === c);
              localVals[c] = i >= 0 ? formatPreviewValue(props.getCell(row, i)) : null;
            }
          }
          const jump = fk ? buildForeignKeyJump(fk, localVals) : null;
          const showFull = focusedCol === col.name && fullText !== null;
          const display = showFull && fullJSON ? prettyJSON(fullText) : showFull ? fullText! : preview;
          return (
            <div key={col.name} class="mb-2.5">
              <label class="mb-0.5 flex justify-between text-11-5 text-ink-2">
                <span>{col.name}</span>
                <em class="font-mono text-[10.5px] not-italic text-ink-3">{col.type}</em>
              </label>
              {showFull && fullJSON ? (
                <div>
                  <textarea
                    class="input w-full rounded-control border border-line bg-panel px-2 py-1 font-mono text-12 text-ink read-only:opacity-100"
                    readOnly
                    rows={6}
                    value={display}
                    onFocus={() => void onFieldFocus(col.name, colIndex)}
                  />
                  <button
                    type="button"
                    class="mt-1 text-11-5 text-accent hover:underline"
                    onClick={() => void navigator.clipboard.writeText(display)}
                  >
                    Copy
                  </button>
                </div>
              ) : (
                <Input
                  mono
                  value={display === "NULL" && props.editable ? "" : display}
                  readOnly={!props.editable || props.primaryKey.includes(col.name)}
                  ariaLabel={col.name}
                  onFocus={() => void onFieldFocus(col.name, colIndex)}
                  onBlur={(e) => {
                    if (!props.editable || !props.onCellCommit || props.primaryKey.includes(col.name)) {
                      return;
                    }
                    const v = (e.target as HTMLInputElement).value;
                    const raw = props.getCell(row, colIndex);
                    const wasNull = raw === null || raw === undefined;
                    if (v === "" && wasNull) {
                      return;
                    }
                    if (v === "" && !wasNull) {
                      void props.onCellCommit(colIndex, col.name, { null: true });
                      return;
                    }
                    if (String(raw) === v) {
                      return;
                    }
                    void props.onCellCommit(colIndex, col.name, { value: v });
                  }}
                />
              )}
              {jump ? (
                <button
                  type="button"
                  class="fk mt-1 text-11-5 text-accent hover:underline"
                  onClick={() => openFilteredTableTab(jump.refSchema, jump.refTable, jump.browseQuery)}
                >
                  {jump.linkLabel}
                </button>
              ) : null}
            </div>
          );
        })
      )}
    </aside>
  );
}

function looksLikeJSON(s: string): boolean {
  const t = s.trim();
  return (t.startsWith("{") && t.endsWith("}")) || (t.startsWith("[") && t.endsWith("]"));
}

function prettyJSON(s: string): string {
  try {
    return JSON.stringify(JSON.parse(s), null, 2);
  } catch {
    return s;
  }
}
