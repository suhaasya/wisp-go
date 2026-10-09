import { useEffect, useState } from "preact/hooks";
import { Button } from "../ui/Button";
import { FilePicker } from "../ui/file-picker";
import { Input } from "../ui/Input";
import { Modal } from "../ui/Modal";
import { onImportEvent } from "../editor/import-events";
import { pushToast } from "../ui/Toast";
import {
  cancelImport,
  importErrorsURL,
  previewImport,
  startImport,
  type ColumnMapping,
  type ImportPreviewResponse,
} from "../workspace/import-client";
import type { ImportKind } from "../workspace/import-types";

export type ImportDialogProps = {
  open: boolean;
  sessionId: string;
  schema: string;
  defaultTable: string;
  onClose: () => void;
  onComplete?: () => void;
};

export function ImportDialog(props: ImportDialogProps) {
  const [step, setStep] = useState(0);
  const [pickerOpen, setPickerOpen] = useState(false);
  const [path, setPath] = useState("");
  const [preview, setPreview] = useState<ImportPreviewResponse | null>(null);
  const [table, setTable] = useState(props.defaultTable);
  const [mapping, setMapping] = useState<ColumnMapping[]>([]);
  const [nullMarkers, setNullMarkers] = useState("\\N, NULL");
  const [batchSize, setBatchSize] = useState("500");
  const [stopOnError, setStopOnError] = useState(true);
  const [importId, setImportId] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    if (!props.open) {
      setStep(0);
      setPath("");
      setPreview(null);
      setImportId(null);
    }
  }, [props.open]);

  useEffect(() => {
    if (!importId) {
      return;
    }
    return onImportEvent((ev) => {
      if (ev.iid !== importId) {
        return;
      }
      if (ev.error) {
        pushToast(ev.error, "error");
        setImportId(null);
        setBusy(false);
        return;
      }
      if (ev.cancelled) {
        pushToast("Import cancelled", "default");
        setImportId(null);
        setBusy(false);
        return;
      }
      if (ev.done) {
        pushToast(`Import finished (${ev.rows ?? 0} rows)`, "success");
        setImportId(null);
        setBusy(false);
        props.onComplete?.();
        props.onClose();
        return;
      }
      if (typeof ev.rows === "number") {
        pushToast(`Importing… ${ev.rows.toLocaleString()} rows`, "default");
      }
    });
  }, [importId]);

  async function loadPreview(selected: string) {
    setPath(selected);
    setBusy(true);
    try {
      const doc = await previewImport(props.sessionId, selected);
      setPreview(doc);
      const fileCols = fileColumns(doc);
      setMapping(
        fileCols.map((c) => ({
          file_column: c,
          table_column: c,
        })),
      );
      setStep(1);
    } catch (e: unknown) {
      pushToast(e instanceof Error ? e.message : "Preview failed", "error");
    } finally {
      setBusy(false);
    }
  }

  async function runImport() {
    if (!preview || !path) {
      return;
    }
    setBusy(true);
    try {
      const kind = preview.kind as ImportKind;
      const fileCols = fileColumns(preview);
      const { import_id } = await startImport(props.sessionId, {
        path,
        kind,
        schema: props.schema,
        table: table.trim(),
        delimiter: preview.csv?.delimiter,
        has_header: preview.csv?.has_header ?? true,
        null_markers: nullMarkers.split(",").map((s) => s.trim()).filter(Boolean),
        batch_size: Number.parseInt(batchSize, 10) || 500,
        stop_on_error: stopOnError,
        mapping,
        file_columns: fileCols,
      });
      setImportId(import_id);
      pushToast("Import started", "default");
    } catch (e: unknown) {
      pushToast(e instanceof Error ? e.message : "Import failed", "error");
      setBusy(false);
    }
  }

  return (
    <>
      <Modal
        open={props.open}
        title="Import data"
        onClose={() => {
          if (importId) {
            void cancelImport(props.sessionId, importId);
          }
          props.onClose();
        }}
      >
        <div class="sheet mb-3 flex gap-2 text-12-5 text-ink-2">
          <span class={step === 0 ? "text-ink font-medium" : ""}>1. File</span>
          <span>→</span>
          <span class={step === 1 ? "text-ink font-medium" : ""}>2. Map</span>
          <span>→</span>
          <span class={step === 2 ? "text-ink font-medium" : ""}>3. Options</span>
        </div>
        {step === 0 ? (
          <div>
            <p class="mb-2 text-12-5 text-ink-2">Choose a CSV, JSON Lines, or SQL file under your home directory.</p>
            <Button label={path || "Choose file…"} onClick={() => setPickerOpen(true)} disabled={busy} />
            {preview ? (
              <p class="mt-2 font-mono text-12 text-ink-2">{path}</p>
            ) : null}
          </div>
        ) : null}
        {step === 1 && preview ? (
          <div>
            <Input ariaLabel="Target table" value={table} onInput={setTable} mono />
            <div class="mt-3 max-h-48 overflow-auto rounded-control border border-line">
              <table class="w-full border-collapse text-12-5">
                <thead>
                  <tr class="bg-raise text-12 text-ink-2">
                    <th class="px-2 py-1 text-left">File column</th>
                    <th class="px-2 py-1 text-left">Table column</th>
                  </tr>
                </thead>
                <tbody>
                  {mapping.map((m, i) => (
                    <tr key={m.file_column} class="border-t border-line-2">
                      <td class="px-2 py-1 font-mono">{m.file_column}</td>
                      <td class="p-1">
                        <input
                          class="input w-full font-mono text-12"
                          value={m.table_column}
                          aria-label={`Map ${m.file_column}`}
                          onInput={(e) => {
                            const next = [...mapping];
                            next[i] = { ...m, table_column: (e.currentTarget as HTMLInputElement).value };
                            setMapping(next);
                          }}
                        />
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            <PreviewTable preview={preview} />
          </div>
        ) : null}
        {step === 2 ? (
          <div class="space-y-3">
            <label class="block text-12-5 text-ink">
              NULL markers (comma-separated)
              <input
                class="input mt-1 w-full font-mono text-12"
                value={nullMarkers}
                onInput={(e) => setNullMarkers((e.currentTarget as HTMLInputElement).value)}
              />
            </label>
            <label class="block text-12-5 text-ink">
              Batch size
              <input
                class="input mt-1 w-full font-mono text-12"
                value={batchSize}
                onInput={(e) => setBatchSize((e.currentTarget as HTMLInputElement).value)}
              />
            </label>
            <label class="flex items-center gap-2 text-12-5 text-ink">
              <input
                type="checkbox"
                checked={stopOnError}
                onChange={(e) => setStopOnError((e.currentTarget as HTMLInputElement).checked)}
              />
              Stop on first error
            </label>
            {importId ? (
              <a
                class="text-12-5 text-accent hover:underline"
                href={importErrorsURL(props.sessionId, importId)}
              >
                Download error report
              </a>
            ) : null}
          </div>
        ) : null}
        <div class="mt-4 flex justify-between border-t border-line-2 pt-3">
          <Button
            label="Back"
            disabled={step === 0 || busy}
            onClick={() => setStep((s) => Math.max(0, s - 1))}
          />
          <div class="flex gap-2">
            <Button label="Cancel" onClick={props.onClose} />
            {step < 2 ? (
              <Button
                label="Next"
                variant="primary"
                disabled={!preview || busy}
                onClick={() => setStep((s) => s + 1)}
              />
            ) : (
              <Button
                label={busy ? "Importing…" : "Import"}
                variant="primary"
                disabled={busy || !table.trim()}
                onClick={() => void runImport()}
              />
            )}
          </div>
        </div>
      </Modal>
      <FilePicker
        open={pickerOpen}
        onClose={() => setPickerOpen(false)}
        onSelect={(p) => void loadPreview(p)}
        filter={(name) => /\.(csv|tsv|json|jsonl|ndjson|sql)$/i.test(name)}
      />
    </>
  );
}

function fileColumns(preview: ImportPreviewResponse): string[] {
  if (preview.csv) {
    return preview.csv.header;
  }
  if (preview.jsonl) {
    return preview.jsonl.columns;
  }
  return [];
}

function PreviewTable({ preview }: { preview: ImportPreviewResponse }) {
  const cols = fileColumns(preview);
  const rows = preview.csv?.rows ?? preview.jsonl?.rows ?? [];
  if (cols.length === 0) {
    return preview.sql ? (
      <pre class="mt-3 max-h-32 overflow-auto rounded-control border border-line bg-raise p-2 font-mono text-11 whitespace-pre-wrap">
        {preview.sql.sample.slice(0, 2000)}
      </pre>
    ) : null;
  }
  return (
    <div class="mt-3 max-h-40 overflow-auto rounded-control border border-line">
      <table class="w-full border-collapse text-11">
        <thead>
          <tr class="bg-raise">
            {cols.map((c) => (
              <th key={c} class="border-b border-line-2 px-2 py-1 text-left font-medium text-ink-2">
                {c}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.slice(0, 20).map((row, ri) => (
            <tr key={ri}>
              {cols.map((_, ci) => (
                <td key={ci} class="border-b border-line-2 px-2 py-1 font-mono text-ink">
                  {row[ci] ?? ""}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
