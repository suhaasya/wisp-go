import { useEffect, useState } from "preact/hooks";
import { registerActionRunner } from "../actions";
import { readOnly, shellConnection, stagedChangeCount, type TableSortKey } from "../state";
import { pkRowValues } from "../grid/edit";
import {
  commitResultChanges,
  discardResultChanges,
  fetchReviewSQL,
  postResultChange,
  type ReviewSQLResponse,
} from "../grid/results-client";
import { ReviewSQLDialog } from "../dialogs/review-sql";
import { TypedConfirmDialog } from "../dialogs/typed-confirm";
import { pushToast } from "../ui/Toast";
import { buildColumns } from "../grid/columns";
import { DataToolbar } from "../workspace/toolbar";
import { TableModeToolbar } from "../workspace/structure/mode-toolbar";
import { appliedBrowseQuery, browseQueryKey, updateTableTabBrowseQuery } from "../workspace/browse-query";
import { FilterBar } from "../workspace/filter-bar";
import {
  DetailPanel,
  loadDetailPanelWidth,
  saveDetailPanelWidth,
} from "../workspace/detail-panel";
import { fetchTableDetails } from "../workspace/schema-client";
import { shouldLoadTableResult } from "../workspace/table-load";
import { updateTableDataTabState } from "../workspace/tabs";
import type { WorkspaceTab } from "../state";
import { Grid, type GridCellAccess } from "./Grid";
import { closeResult, openTableResult, type OpenResultResponse } from "./results-client";
import { emptySelection, primaryRow, type SelectionState } from "./selection";
import { ExportDialog } from "../dialogs/export";
import { ImportDialog } from "../dialogs/import";
import { onExportEvent } from "../editor/export-events";
import {
  cancelFileExport,
  exportDownloadURL,
  fetchExportText,
  startFileExport,
  type ExportQuery,
} from "../workspace/export-client";
import { DEFAULT_WINDOW_ROWS } from "./window";

const EXPORT_PAGE_ROW_HEIGHT = 26;

type TableDataTab = Extract<WorkspaceTab, { kind: "table_data" }>;

export function DataTab({ sessionId, tab }: { sessionId: string; tab: TableDataTab }) {
  const [open, setOpen] = useState<OpenResultResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const applied = appliedBrowseQuery(tab);
  const appliedKey = browseQueryKey(applied);
  const loadResults = shouldLoadTableResult(tab);
  const [sortDraft, setSortDraft] = useState<TableSortKey[]>(() => applied.sort);
  const [detailOpen, setDetailOpen] = useState(true);
  const [detailWidth, setDetailWidth] = useState(loadDetailPanelWidth);
  const [selection, setSelection] = useState<SelectionState>(emptySelection());
  const [gridAccess, setGridAccess] = useState<GridCellAccess | null>(null);
  const [primaryKey, setPrimaryKey] = useState<string[]>([]);
  const [foreignKeys, setForeignKeys] = useState<
    import("../workspace/fk-jump").ForeignKeyMeta[]
  >([]);
  const [stagedCount, setStagedCount] = useState(0);
  const [reviewOpen, setReviewOpen] = useState(false);
  const [review, setReview] = useState<ReviewSQLResponse | null>(null);
  const [typedOpen, setTypedOpen] = useState(false);
  const [committing, setCommitting] = useState(false);
  const [exportOpen, setExportOpen] = useState(false);
  const [importOpen, setImportOpen] = useState(false);
  const [fileExportId, setFileExportId] = useState<string | null>(null);

  useEffect(() => {
    stagedChangeCount.value = stagedCount;
    return () => {
      stagedChangeCount.value = 0;
    };
  }, [stagedCount]);

  useEffect(() => {
    const unsubs = [
      registerActionRunner("export-data", () => setExportOpen(true)),
      registerActionRunner("import-data", () => setImportOpen(true)),
      registerActionRunner("review-sql", () => void openReview()),
      registerActionRunner("discard-changes", () => void onDiscard()),
      registerActionRunner("add-row", () => void onAddRow()),
      registerActionRunner("duplicate-row", () => void onDuplicateRow()),
      registerActionRunner("delete-row", () => void onDeleteRow()),
    ];
    return () => unsubs.forEach((u) => u());
  }, [sessionId, open?.rid, stagedCount, selection, gridAccess]);

  useEffect(() => {
    setStagedCount(0);
  }, [open?.rid]);

  useEffect(() => {
    if (!fileExportId) {
      return;
    }
    return onExportEvent((ev) => {
      if (ev.eid !== fileExportId) {
        return;
      }
      if (ev.error) {
        pushToast(ev.error, "error");
        setFileExportId(null);
        return;
      }
      if (ev.cancelled) {
        pushToast("Export cancelled", "default");
        setFileExportId(null);
        return;
      }
      if (ev.done) {
        pushToast(`Export finished (${ev.rows ?? 0} rows)`, "success");
        setFileExportId(null);
        setExportOpen(false);
        return;
      }
      if (typeof ev.rows === "number") {
        pushToast(`Exporting… ${ev.rows.toLocaleString()} rows`, "default");
      }
    });
  }, [fileExportId]);

  useEffect(() => {
    setSortDraft(applied.sort);
  }, [tab.id, appliedKey]);

  useEffect(() => {
    const ac = new AbortController();
    void fetchTableDetails(sessionId, tab.schema, tab.table, { signal: ac.signal })
      .then((det) => {
        setPrimaryKey(det.primary_key ?? []);
        setForeignKeys(det.foreign_keys ?? []);
      })
      .catch(() => {});
    return () => ac.abort();
  }, [sessionId, tab.schema, tab.table]);

  useEffect(() => {
    let rid: string | null = null;
    const ac = new AbortController();
    setOpen(null);
    setError(null);
    if (!loadResults) {
      return () => {
        ac.abort();
      };
    }
    void openTableResult(sessionId, tab.schema, tab.table, applied, ac.signal)
      .then((res) => {
        rid = res.rid;
        setOpen(res);
      })
      .catch((e: unknown) => {
        if (ac.signal.aborted) {
          return;
        }
        setError(e instanceof Error ? e.message : "Could not open table");
      });
    return () => {
      ac.abort();
      if (rid) {
        void closeResult(sessionId, rid).catch(() => {});
      }
    };
  }, [sessionId, tab.schema, tab.table, appliedKey, loadResults]);

  function handleApply(query: Parameters<typeof updateTableTabBrowseQuery>[1]) {
    updateTableTabBrowseQuery(tab.id, query);
  }

  function onDetailWidthChange(w: number) {
    const clamped = Math.min(520, Math.max(220, w));
    setDetailWidth(clamped);
    saveDetailPanelWidth(clamped);
  }

  const detailRow = primaryRow(selection);
  const editable = primaryKey.length > 0 && !readOnly.value;
  const gridColumns = buildColumns(open?.columns ?? []);

  async function runChange(body: Parameters<typeof postResultChange>[2]) {
    if (!open) {
      return;
    }
    const resp = await postResultChange(sessionId, open.rid, body);
    setStagedCount(resp.count);
    gridAccess?.invalidateWindows();
  }

  async function onAddRow() {
    await runChange({ op: "insert" });
  }

  async function onDuplicateRow() {
    const picked =
      detailRow ?? (selection.kind === "row" ? [...selection.rows][0] : undefined);
    if (picked === undefined || !open || !gridAccess) {
      return;
    }
    const row = picked;
    const ins = await postResultChange(sessionId, open.rid, { op: "insert" });
    const key = ins.insertKey;
    if (!key) {
      setStagedCount(ins.count);
      return;
    }
    let count = ins.count;
    for (let ci = 0; ci < gridColumns.length; ci++) {
      const col = gridColumns[ci]!;
      const raw = gridAccess.getCell(row, ci);
      if (raw === null || raw === undefined) {
        continue;
      }
      const r = await postResultChange(sessionId, open.rid, {
        op: "set",
        row: [key],
        col: col.id,
        value: String(raw),
      });
      count = r.count;
    }
    setStagedCount(count);
    gridAccess.invalidateWindows();
  }

  async function onDeleteRow() {
    const picked =
      detailRow ?? (selection.kind === "row" ? [...selection.rows][0] : undefined);
    if (picked === undefined || !gridAccess) {
      return;
    }
    const row = picked;
    const ctx = { primaryKey, columns: gridColumns, getCell: gridAccess.getCell.bind(gridAccess) };
    await runChange({ op: "delete", row: pkRowValues(ctx, row) });
  }

  async function openReview() {
    if (!open || stagedCount === 0) {
      return;
    }
    try {
      const data = await fetchReviewSQL(sessionId, open.rid);
      setReview(data);
      setReviewOpen(true);
    } catch (e: unknown) {
      pushToast(e instanceof Error ? e.message : "Could not load review SQL", "error");
    }
  }

  async function runCommit(tableConfirm?: string) {
    if (!open) {
      return;
    }
    setCommitting(true);
    try {
      const res = await commitResultChanges(sessionId, open.rid, { tableConfirm });
      setReviewOpen(false);
      setTypedOpen(false);
      setReview(null);
      setStagedCount(0);
      gridAccess?.invalidateWindows();
      pushToast(`Committed ${res.committed} change${res.committed === 1 ? "" : "s"}`, "success");
    } catch (e: unknown) {
      pushToast(e instanceof Error ? e.message : "Commit failed", "error");
    } finally {
      setCommitting(false);
    }
  }

  function onCommitClick() {
    if (review?.needsDeleteConfirm) {
      setTypedOpen(true);
      return;
    }
    void runCommit();
  }

  function onExportDownload(query: ExportQuery) {
    window.location.assign(exportDownloadURL(sessionId, query));
    setExportOpen(false);
  }

  async function onExportClipboard(query: ExportQuery) {
    try {
      const text = await fetchExportText(sessionId, query);
      await navigator.clipboard.writeText(text);
      pushToast("Copied export to clipboard", "success");
      setExportOpen(false);
    } catch (e: unknown) {
      pushToast(e instanceof Error ? e.message : "Clipboard export failed", "error");
    }
  }

  async function onExportFile(query: ExportQuery & { path: string }) {
    try {
      const { export_id } = await startFileExport(sessionId, query);
      setFileExportId(export_id);
      pushToast("Writing export to disk…", "default");
    } catch (e: unknown) {
      pushToast(e instanceof Error ? e.message : "Could not start export", "error");
    }
  }

  async function onDiscard() {
    if (!open) {
      return;
    }
    await discardResultChanges(sessionId, open.rid);
    setStagedCount(0);
    gridAccess?.invalidateWindows();
  }

  if (!loadResults) {
    return (
      <div class="flex flex-1 flex-col items-center justify-center p-6 text-13 text-ink-2">
        Tab restored — click the tab again to load data.
      </div>
    );
  }

  if (error) {
    return <p class="p-4 text-13 text-env-prod">{error}</p>;
  }
  if (!open) {
    return <p class="p-4 text-13 text-ink-2">Loading table…</p>;
  }
  return (
    <div class="flex min-h-0 flex-1 flex-col overflow-hidden">
      <TableModeToolbar schema={tab.schema} table={tab.table} mode="data" />
      {primaryKey.length === 0 ? (
        <p class="border-b border-line bg-raise px-3 py-2 text-12-5 text-ink-2">
          Inline editing requires a primary or unique key on this table.
        </p>
      ) : null}
      <DataToolbar stagedCount={stagedCount} disabled={!editable} />
      <FilterBar
        columns={open.columns}
        applied={applied}
        sortDraft={sortDraft}
        onApply={(query) => handleApply({ ...query, sort: sortDraft })}
      />
      <div class="flex min-h-0 min-w-0 flex-1 overflow-hidden">
        <Grid
          sessionId={sessionId}
          rid={open.rid}
          columns={open.columns}
          total={open.total}
          schema={tab.schema}
          table={tab.table}
          sortDraft={sortDraft}
          onSortDraftChange={setSortDraft}
          initialScrollTop={tab.scrollTop}
          onScrollTopChange={(top) => updateTableDataTabState(tab.id, { scrollTop: top })}
          onSelectionChange={setSelection}
          onGridReady={setGridAccess}
          onToggleDetail={() => setDetailOpen((v) => !v)}
          editable={editable}
          primaryKey={primaryKey}
          onStagedCount={setStagedCount}
        />
        <DetailPanel
          open={detailOpen}
          width={detailWidth}
          onWidthChange={onDetailWidthChange}
          sessionId={sessionId}
          rid={open.rid}
          schema={tab.schema}
          table={tab.table}
          row={detailRow}
          columns={open.columns}
          primaryKey={primaryKey}
          foreignKeys={foreignKeys}
          getCell={(row, col) => gridAccess?.getCell(row, col) ?? null}
          isTruncated={(row, col) => gridAccess?.isTruncated(row, col) ?? false}
          editable={editable}
          onCellCommit={async (colIndex, colName, next) => {
            if (!open || !gridAccess || detailRow === null) {
              return;
            }
            const raw = gridAccess.getCell(detailRow, colIndex);
            const orig =
              raw === null || raw === undefined
                ? { origNull: true }
                : { original: String(raw) };
            const resp = await postResultChange(sessionId, open.rid, {
              op: "set",
              row: pkRowValues(
                { primaryKey, columns: gridColumns, getCell: gridAccess.getCell.bind(gridAccess) },
                detailRow,
              ),
              col: colName,
              ...orig,
              ...next,
            });
            setStagedCount(resp.count);
            gridAccess.invalidateWindows();
          }}
        />
      </div>
      <ReviewSQLDialog
        open={reviewOpen}
        title={`Review ${review?.count ?? stagedCount} changes to ${tab.table}`}
        subtitle={
          shellConnection.value
            ? `${shellConnection.value.name} · ${shellConnection.value.subtitle}`
            : `${tab.schema}.${tab.table}`
        }
        sql={review?.sql ?? ""}
        warning={review?.warning}
        onClose={() => setReviewOpen(false)}
        onCommit={onCommitClick}
        committing={committing}
      />
      <TypedConfirmDialog
        open={typedOpen}
        tableName={review?.table ?? tab.table}
        onCancel={() => setTypedOpen(false)}
        onConfirm={() => void runCommit(review?.table ?? tab.table)}
      />
      <ImportDialog
        open={importOpen}
        sessionId={sessionId}
        schema={tab.schema}
        defaultTable={tab.table}
        onClose={() => setImportOpen(false)}
        onComplete={() => gridAccess?.invalidateWindows()}
      />
      <ExportDialog
        open={exportOpen}
        sessionId={sessionId}
        rid={open.rid}
        tableLabel={`${tab.schema}.${tab.table}`}
        selection={selection}
        pageOffset={Math.floor((tab.scrollTop ?? 0) / EXPORT_PAGE_ROW_HEIGHT)}
        pageLimit={DEFAULT_WINDOW_ROWS}
        onClose={() => {
          setExportOpen(false);
          if (fileExportId) {
            void cancelFileExport(sessionId, fileExportId);
          }
        }}
        onDownload={onExportDownload}
        onClipboard={(q) => void onExportClipboard(q)}
        onFileExport={(q) => void onExportFile(q)}
      />
    </div>
  );
}
