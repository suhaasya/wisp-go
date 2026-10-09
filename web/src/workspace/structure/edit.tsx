import { useCallback, useEffect, useState } from "preact/hooks";
import { readOnly, shellConnection, stagedChangeCount } from "../../state";
import { Button } from "../../ui/Button";
import { ReviewSQLDialog } from "../../dialogs/review-sql";
import { TypedConfirmDialog } from "../../dialogs/typed-confirm";
import { pushToast } from "../../ui/Toast";
import type { TableDetailsResponse } from "../schema-client";
import {
  commitStructureChanges,
  discardStructureChanges,
  fetchStructureReviewSQL,
  postStructureChange,
  type StructureReviewResponse,
} from "../structure-changes-client";
import { StructureColumnsSection } from "./columns";
import { StructureDDLPanel } from "./ddl";
import { StructureIndexesSection } from "./indexes";
import {
  StructureChecksSection,
  StructureForeignKeysSection,
  StructureTriggersSection,
} from "./meta-sections";
import { TableModeToolbar } from "./mode-toolbar";

type RowMark = "mod" | "ins" | "";

export function StructureEditor({
  sessionId,
  schema,
  table,
  details,
  baselineDDL,
  ddlLoading,
  onRefresh,
}: {
  sessionId: string;
  schema: string;
  table: string;
  details: TableDetailsResponse;
  baselineDDL: string;
  ddlLoading: boolean;
  onRefresh: () => void;
}) {
  const [pendingCount, setPendingCount] = useState(0);
  const [pendingSQL, setPendingSQL] = useState("");
  const [rowMarks, setRowMarks] = useState<Record<string, RowMark>>({});
  const [reviewOpen, setReviewOpen] = useState(false);
  const [review, setReview] = useState<StructureReviewResponse | null>(null);
  const [typedOpen, setTypedOpen] = useState(false);
  const [committing, setCommitting] = useState(false);
  const disabled = readOnly.value;

  const refreshPending = useCallback(async () => {
    try {
      const r = await fetchStructureReviewSQL(sessionId, schema, table);
      setPendingCount(r.count ?? 0);
      setPendingSQL(r.count > 0 ? r.sql : "");
    } catch {
      setPendingCount(0);
      setPendingSQL("");
    }
  }, [sessionId, schema, table]);

  useEffect(() => {
    stagedChangeCount.value = pendingCount;
    return () => {
      stagedChangeCount.value = 0;
    };
  }, [pendingCount]);

  useEffect(() => {
    void refreshPending();
  }, [refreshPending]);

  async function stage(op: Parameters<typeof postStructureChange>[3], mark?: { key: string; kind: RowMark }) {
    if (disabled) {
      return;
    }
    const res = await postStructureChange(sessionId, schema, table, op);
    setPendingCount(res.count);
    if (mark) {
      setRowMarks((m) => ({ ...m, [mark.key]: mark.kind }));
    }
    await refreshPending();
  }

  async function onDiscard() {
    await discardStructureChanges(sessionId, schema, table);
    setPendingCount(0);
    setPendingSQL("");
    setRowMarks({});
  }

  async function openReview() {
    if (pendingCount === 0) {
      return;
    }
    const r = await fetchStructureReviewSQL(sessionId, schema, table);
    setReview(r);
    setReviewOpen(true);
  }

  async function runCommit(tableConfirm?: string) {
    setCommitting(true);
    try {
      const res = await commitStructureChanges(sessionId, schema, table, {
        confirm: true,
        tableConfirm,
      });
      setReviewOpen(false);
      setTypedOpen(false);
      setReview(null);
      setPendingCount(0);
      setPendingSQL("");
      setRowMarks({});
      pushToast(`Applied ${res.committed} statement${res.committed === 1 ? "" : "s"}`, "success");
      onRefresh();
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

  async function onAddColumn() {
    const name = `col_${Date.now() % 100000}`;
    await stage(
      {
        op: "add_column",
        column_def: { name, type: "text", nullable: true },
      },
      { key: name, kind: "ins" },
    );
  }

  return (
    <div class="flex min-h-0 flex-1 flex-col overflow-hidden">
      <TableModeToolbar schema={schema} table={table} mode="structure" />
      <div class="toolbar flex shrink-0 items-center gap-2 border-b border-line px-2.5 py-2">
        <Button label="Add column" disabled={disabled} onClick={() => void onAddColumn()} />
        <Button label="Discard" disabled={disabled || pendingCount === 0} onClick={() => void onDiscard()} />
        <span class="flex-1" />
        <Button
          label={pendingCount > 0 ? `Review SQL (${pendingCount})` : "Review SQL"}
          variant="warning"
          disabled={disabled || pendingCount === 0}
          onClick={() => void openReview()}
        />
      </div>
      <div class="grid min-h-0 flex-1 grid-cols-[1fr_340px] gap-4 overflow-auto p-3.5 content-start">
        <div class="flex min-w-0 flex-col gap-4">
          <EditableColumns
            details={details}
            rowMarks={rowMarks}
            disabled={disabled}
            onAlter={(column, field, value) =>
              void stageAlterColumn(sessionId, schema, table, column, field, value, details, stage, setRowMarks)
            }
          />
          <StructureIndexesSection details={details} />
          <StructureForeignKeysSection details={details} />
          <StructureChecksSection details={details} />
          <StructureTriggersSection details={details} />
        </div>
        <StructureDDLPanel
          ddl={pendingCount > 0 ? pendingSQL : ""}
          loading={ddlLoading && pendingCount === 0}
          title={pendingCount > 0 ? "Pending changes" : "DDL"}
          baselineDDL={baselineDDL}
        />
      </div>
      <ReviewSQLDialog
        open={reviewOpen}
        title={`Review ${review?.count ?? pendingCount} structure changes to ${table}`}
        subtitle={
          shellConnection.value
            ? `${shellConnection.value.name} · ${shellConnection.value.subtitle}`
            : `${schema}.${table}`
        }
        sql={review?.sql ?? pendingSQL}
        warning={review?.warning}
        onClose={() => setReviewOpen(false)}
        onCommit={onCommitClick}
        committing={committing}
      />
      <TypedConfirmDialog
        open={typedOpen}
        tableName={review?.table ?? table}
        onCancel={() => setTypedOpen(false)}
        onConfirm={() => void runCommit(review?.table ?? table)}
      />
    </div>
  );
}

function EditableColumns({
  details,
  rowMarks,
  disabled,
  onAlter,
}: {
  details: TableDetailsResponse;
  rowMarks: Record<string, RowMark>;
  disabled: boolean;
  onAlter: (column: string, field: "type" | "nullable" | "default", value: string) => void;
}) {
  return (
    <StructureColumnsSection
      details={details}
      editable={!disabled}
      rowMarks={rowMarks}
      onFieldChange={onAlter}
    />
  );
}

async function stageAlterColumn(
  _sessionId: string,
  _schema: string,
  _table: string,
  column: string,
  field: "type" | "nullable" | "default",
  value: string,
  details: TableDetailsResponse,
  stage: (op: Parameters<typeof postStructureChange>[3], mark?: { key: string; kind: RowMark }) => Promise<void>,
  setRowMarks: (fn: (m: Record<string, RowMark>) => Record<string, RowMark>) => void,
) {
  const col = details.columns.find((c) => c.name === column);
  if (!col) {
    return;
  }
  if (field === "type" && value !== col.type) {
    await stage(
      {
        op: "alter_column_type",
        column,
        column_def: { name: column, type: value },
        previous: { name: column, type: col.type, nullable: col.nullable, default: col.default },
      },
      { key: column, kind: "mod" },
    );
    setRowMarks((m) => ({ ...m, [column]: "mod" }));
    return;
  }
  if (field === "nullable") {
    const nullable = value === "Yes";
    if (nullable !== (col.nullable !== false)) {
      await stage(
        {
          op: "alter_column_nullable",
          column,
          column_def: { name: column, type: col.type, nullable },
        },
        { key: column, kind: "mod" },
      );
    }
    return;
  }
  if (field === "default" && value !== (col.default ?? "")) {
    await stage(
      {
        op: "alter_column_default",
        column,
        column_def: { name: column, type: col.type, default: value },
      },
      { key: column, kind: "mod" },
    );
  }
}
