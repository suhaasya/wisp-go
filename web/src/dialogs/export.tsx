import { useState } from "preact/hooks";
import { Button } from "../ui/Button";
import { Modal } from "../ui/Modal";
import type { ExportFormat, ExportQuery, ExportScope } from "../workspace/export-client";
import { selectionRange, type SelectionState } from "../grid/selection";

export type ExportDialogProps = {
  open: boolean;
  sessionId: string;
  rid: string;
  tableLabel: string;
  selection: SelectionState;
  pageOffset: number;
  pageLimit: number;
  onClose: () => void;
  onDownload: (query: ExportQuery) => void;
  onClipboard: (query: ExportQuery) => void;
  onFileExport: (query: ExportQuery & { path: string }) => void;
};

const CLIPBOARD_MAX_ROWS = 10_000;

function selectionRowCount(selection: SelectionState): number | null {
  const range = selectionRange(selection);
  if (!range) {
    return null;
  }
  return range.rowMax - range.rowMin + 1;
}

export function ExportDialog(props: ExportDialogProps) {
  const selRows = selectionRowCount(props.selection);
  const hasSelection = selRows !== null;
  const [format, setFormat] = useState<ExportFormat>("csv");
  const [scope, setScope] = useState<ExportScope>(hasSelection ? "selection" : "result");
  const [dest, setDest] = useState<"download" | "clipboard" | "file">("download");
  const [path, setPath] = useState("~/Downloads/wisp-export.csv");

  function buildQuery(): ExportQuery {
    const q: ExportQuery = {
      format,
      scope,
      rid: props.rid,
    };
    if (scope === "page") {
      q.offset = props.pageOffset;
      q.limit = props.pageLimit;
    }
    if (scope === "selection") {
      const range = selectionRange(props.selection);
      if (range) {
        q.rowMin = range.rowMin;
        q.rowMax = range.rowMax;
        if (range.colMax !== Number.MAX_SAFE_INTEGER) {
          q.colMin = range.colMin;
          q.colMax = range.colMax;
        }
      }
    }
    return q;
  }

  function onRun() {
    const query = buildQuery();
    if (dest === "clipboard") {
      props.onClipboard(query);
      return;
    }
    if (dest === "file") {
      props.onFileExport({ ...query, path });
      return;
    }
    props.onDownload(query);
  }

  const clipboardDisabled =
    scope === "selection" && selRows !== null && selRows > CLIPBOARD_MAX_ROWS;

  return (
    <Modal open={props.open} title="Export data" onClose={props.onClose}>
      <p class="mb-3 text-12-5 text-ink-2">{props.tableLabel}</p>
      <fieldset class="mb-3 space-y-2">
        <legend class="text-12 font-medium text-ink">Source</legend>
        <label class="flex items-center gap-2 text-12-5 text-ink">
          <input
            type="radio"
            name="export-scope"
            checked={scope === "result"}
            onChange={() => setScope("result")}
          />
          Whole result
        </label>
        <label class="flex items-center gap-2 text-12-5 text-ink">
          <input
            type="radio"
            name="export-scope"
            checked={scope === "page"}
            onChange={() => setScope("page")}
          />
          Current page
        </label>
        <label class="flex items-center gap-2 text-12-5 text-ink-2">
          <input
            type="radio"
            name="export-scope"
            disabled={!hasSelection}
            checked={scope === "selection"}
            onChange={() => setScope("selection")}
          />
          Selection{hasSelection ? ` (${selRows} rows)` : ""}
        </label>
      </fieldset>
      <label class="mb-3 block text-12-5 text-ink">
        Format
        <select
          aria-label="Format"
          class="mt-1 w-full rounded-control border border-line-2 bg-panel px-2 py-1.5 text-12-5"
          value={format}
          onChange={(e) => setFormat((e.currentTarget as HTMLSelectElement).value as ExportFormat)}
        >
          <option value="csv">CSV</option>
          <option value="json">JSON</option>
          <option value="sql">SQL INSERT</option>
          <option value="markdown">Markdown</option>
        </select>
      </label>
      <fieldset class="mb-3 space-y-2">
        <legend class="text-12 font-medium text-ink">Destination</legend>
        <label class="flex items-center gap-2 text-12-5 text-ink">
          <input
            type="radio"
            name="export-dest"
            checked={dest === "download"}
            onChange={() => setDest("download")}
          />
          Browser download
        </label>
        <label class="flex items-center gap-2 text-12-5 text-ink-2">
          <input
            type="radio"
            name="export-dest"
            checked={dest === "clipboard"}
            disabled={clipboardDisabled}
            onChange={() => setDest("clipboard")}
          />
          Clipboard (under {CLIPBOARD_MAX_ROWS.toLocaleString()} rows)
        </label>
        <label class="flex items-center gap-2 text-12-5 text-ink">
          <input
            type="radio"
            name="export-dest"
            checked={dest === "file"}
            onChange={() => setDest("file")}
          />
          Local file (large exports)
        </label>
      </fieldset>
      {dest === "file" ? (
        <label class="mb-3 block text-12-5 text-ink">
          Path
          <input
            class="mt-1 w-full rounded-control border border-line-2 bg-panel px-2 py-1.5 font-mono text-12-5"
            value={path}
            onInput={(e) => setPath((e.currentTarget as HTMLInputElement).value)}
          />
        </label>
      ) : null}
      <div class="flex justify-end gap-2 border-t border-line-2 pt-3">
        <Button label="Cancel" onClick={props.onClose} />
        <Button label="Export" variant="primary" onClick={onRun} />
      </div>
    </Modal>
  );
}
