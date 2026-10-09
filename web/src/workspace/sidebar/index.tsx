import { useCallback, useEffect, useRef, useState } from "preact/hooks";
import { ApiError } from "../../api";
import { ConfirmDestructiveDialog } from "../../dialogs/confirm-destructive";
import { NewSchemaDialog } from "../../dialogs/new-schema";
import { Input } from "../../ui/Input";
import { Button } from "../../ui/Button";
import { readOnly, shellConnection } from "../../state";
import { executeDDL, previewDDL, type DDLPreviewResponse } from "../ddl-client";
import { fetchSchemas } from "../schema-client";
import { SchemaSelector, useFilterFocusShortcut } from "./filter";
import {
  clampSidebarWidth,
  loadSidebarUiState,
  saveSidebarUiState,
  type SidebarUiState,
} from "./sidebar-state";
import { SchemaTree, type OpenTableRequest, type SchemaGroupId } from "./tree";

export type SchemaSidebarProps = {
  sessionId: string;
  profileId: string | null;
  engine: string;
  onOpenTable: (req: OpenTableRequest) => void;
};

const FILTER_INPUT_ID = "schema-sidebar-filter";

export function SchemaSidebar({
  sessionId,
  profileId,
  engine,
  onOpenTable,
}: SchemaSidebarProps) {
  const storageKey = profileId ?? sessionId;
  const [ui, setUi] = useState<SidebarUiState>(() => loadSidebarUiState(storageKey));
  const [filter, setFilter] = useState("");
  const [databases, setDatabases] = useState<string[]>([]);
  const [schemas, setSchemas] = useState<string[]>([]);
  const [selectedObject, setSelectedObject] = useState<{ schema: string; name: string } | null>(
    null,
  );
  const [schemasError, setSchemasError] = useState<string | null>(null);
  const [newSchemaOpen, setNewSchemaOpen] = useState(false);
  const [newTableOpen, setNewTableOpen] = useState(false);
  const [NewTableDialogComp, setNewTableDialogComp] = useState<
    typeof import("../../dialogs/new-table").NewTableDialog | null
  >(null);
  const [ddlFlow, setDdlFlow] = useState<{
    action: "truncate" | "drop_table" | "rename_table";
    schema: string;
    table: string;
  } | null>(null);
  const [renameTo, setRenameTo] = useState("");
  const [renameReady, setRenameReady] = useState(false);
  const [cascade, setCascade] = useState(false);
  const [ddlPreview, setDdlPreview] = useState<DDLPreviewResponse | null>(null);
  const [ddlBusy, setDdlBusy] = useState(false);
  const [treeReload, setTreeReload] = useState(0);
  const resizeRef = useRef<{ startX: number; startW: number } | null>(null);

  useEffect(() => {
    if (!newTableOpen) {
      return;
    }
    void import("../../dialogs/new-table").then((m) => setNewTableDialogComp(() => m.NewTableDialog));
  }, [newTableOpen]);

  useFilterFocusShortcut(FILTER_INPUT_ID, true);

  const persist = useCallback(
    (patch: Partial<SidebarUiState> | ((prev: SidebarUiState) => SidebarUiState)) => {
      setUi((prev) => {
        const next = typeof patch === "function" ? patch(prev) : { ...prev, ...patch };
        saveSidebarUiState(storageKey, next);
        return next;
      });
    },
    [storageKey],
  );

  useEffect(() => {
    setUi(loadSidebarUiState(storageKey));
  }, [storageKey]);

  const loadSchemas = useCallback(async () => {
    setSchemasError(null);
    try {
      if (sessionId.startsWith("mock-")) {
        setSchemasError("Connect to a real database to browse schemas.");
        return;
      }
      const snap = await fetchSchemas(sessionId);
      const dbs = snap.databases ?? [];
      const sc = snap.schemas ?? dbs;
      setDatabases(dbs);
      setSchemas(sc);
      const infoDb = shellConnection.value?.name ?? "";
      persist((prev) => ({
        ...prev,
        selectedDatabase: prev.selectedDatabase || infoDb || dbs[0] || "",
        selectedSchema: prev.selectedSchema || sc[0] || "",
      }));
    } catch (err) {
      let msg = "Could not load schemas.";
      if (err instanceof ApiError) {
        if (err.code === "session_not_found") {
          msg = "Session expired. Go back to Connections and connect again.";
        } else if (err.message) {
          msg = err.message;
        }
      }
      setSchemasError(msg);
    }
  }, [sessionId, persist]);

  useEffect(() => {
    void loadSchemas();
  }, [loadSchemas]);

  const activeSchema =
    engine === "postgres"
      ? ui.selectedSchema || schemas[0] || "public"
      : ui.selectedDatabase || databases[0] || shellConnection.value?.name || "";

  const onResizePointerDown = (e: PointerEvent) => {
    e.preventDefault();
    resizeRef.current = { startX: e.clientX, startW: ui.width };
    const onMove = (ev: PointerEvent) => {
      const start = resizeRef.current;
      if (!start) {
        return;
      }
      const next = clampSidebarWidth(start.startW + (ev.clientX - start.startX));
      persist({ width: next });
    };
    const onUp = () => {
      resizeRef.current = null;
      window.removeEventListener("pointermove", onMove);
      window.removeEventListener("pointerup", onUp);
    };
    window.addEventListener("pointermove", onMove);
    window.addEventListener("pointerup", onUp);
  };

  const setExpanded = (expanded: Record<SchemaGroupId, boolean>) => {
    persist({ expanded });
  };

  useEffect(() => {
    if (!ddlFlow) {
      setDdlPreview(null);
      return;
    }
    if (ddlFlow.action === "rename_table" && !renameTo.trim()) {
      setDdlPreview(null);
      return;
    }
    const ac = new AbortController();
    void previewDDL(
      sessionId,
      {
        action: ddlFlow.action,
        schema: ddlFlow.schema,
        table: ddlFlow.table,
        new_name: ddlFlow.action === "rename_table" ? renameTo.trim() : undefined,
        cascade: ddlFlow.action === "drop_table" ? cascade : undefined,
      },
      { signal: ac.signal },
    )
      .then(setDdlPreview)
      .catch(() => setDdlPreview(null));
    return () => ac.abort();
  }, [ddlFlow, sessionId, renameTo, cascade]);

  const refreshTree = () => setTreeReload((n) => n + 1);

  return (
    <aside
      class="relative flex h-full min-h-0 shrink-0 flex-col overflow-hidden border-r border-line bg-sidebar"
      style={{ width: `${ui.width}px` }}
    >
      <SchemaSelector
        engine={engine}
        databases={databases.length ? databases : [ui.selectedDatabase].filter(Boolean)}
        schemas={schemas}
        database={ui.selectedDatabase || databases[0] || ""}
        schema={activeSchema}
        filter={filter}
        filterInputId={FILTER_INPUT_ID}
        onDatabaseChange={(selectedDatabase) => persist({ selectedDatabase })}
        onSchemaChange={(selectedSchema) => persist({ selectedSchema })}
        onFilterChange={setFilter}
      />
      {!schemasError && !readOnly.value ? (
        <div class="flex shrink-0 gap-2 border-b border-line px-2.5 py-2">
          <Button label="New table" onClick={() => setNewTableOpen(true)} />
          <Button
            label={engine === "mysql" || engine === "mariadb" ? "New database" : "New schema"}
            onClick={() => setNewSchemaOpen(true)}
          />
        </div>
      ) : null}
      {schemasError ? (
        <div class="flex shrink-0 flex-col gap-2 px-2.5 py-2 text-12-5 text-env-prod">
          <p>{schemasError}</p>
          {!sessionId.startsWith("mock-") ? (
            <button type="button" class="text-left text-accent underline" onClick={() => void loadSchemas()}>
              Retry
            </button>
          ) : null}
        </div>
      ) : (
        <SchemaTree
          key={treeReload}
          sessionId={sessionId}
          schema={activeSchema}
          filterQuery={filter}
          expanded={ui.expanded}
          onExpandedChange={setExpanded}
          selectedObject={selectedObject}
          onSelectObject={(schema, name) => setSelectedObject({ schema, name })}
          onOpenTable={onOpenTable}
          readOnly={readOnly.value}
          onCatalogDDL={(req) => {
            setRenameTo("");
            setCascade(false);
            const action =
              req.action === "drop"
                ? "drop_table"
                : req.action === "rename"
                  ? "rename_table"
                  : "truncate";
            setRenameReady(req.action !== "rename");
            setDdlFlow({ action, schema: req.schema, table: req.table });
          }}
        />
      )}
      <NewSchemaDialog
        open={newSchemaOpen}
        sessionId={sessionId}
        engine={engine}
        onClose={() => setNewSchemaOpen(false)}
        onCreated={() => void loadSchemas()}
      />
      {NewTableDialogComp ? (
        <NewTableDialogComp
          open={newTableOpen}
          sessionId={sessionId}
          schema={activeSchema}
          engine={engine}
          onClose={() => setNewTableOpen(false)}
          onCreated={() => refreshTree()}
        />
      ) : null}
      <ConfirmDestructiveDialog
        open={ddlFlow !== null && renameReady && (ddlFlow.action !== "rename_table" || renameTo.trim() !== "")}
        title={
          ddlFlow?.action === "drop_table"
            ? `Drop ${ddlFlow.table}`
            : ddlFlow?.action === "truncate"
              ? `Truncate ${ddlFlow?.table ?? ""}`
              : `Rename ${ddlFlow?.table ?? ""}`
        }
        subtitle={shellConnection.value?.subtitle ?? activeSchema}
        sql={ddlPreview?.sql ?? ""}
        confirmName={ddlPreview?.confirm_name ?? ddlFlow?.table ?? ""}
        needsTypedConfirm={ddlPreview?.needs_typed_confirm === true}
        dependents={ddlPreview?.dependents}
        warning={ddlPreview?.warning}
        cascade={cascade}
        cascadeEnabled={ddlFlow?.action === "drop_table" && engine === "postgres"}
        onCascadeChange={setCascade}
        confirming={ddlBusy}
        onCancel={() => setDdlFlow(null)}
        onConfirm={(typedConfirm) => {
          if (!ddlFlow || !ddlPreview) {
            return;
          }
          setDdlBusy(true);
          void executeDDL(sessionId, {
            action: ddlFlow.action,
            schema: ddlFlow.schema,
            table: ddlFlow.table,
            new_name: ddlFlow.action === "rename_table" ? renameTo.trim() : undefined,
            cascade: ddlFlow.action === "drop_table" ? cascade : undefined,
            confirm: true,
            object_confirm: typedConfirm || ddlPreview.confirm_name,
          })
            .then(() => {
              setDdlFlow(null);
              refreshTree();
              void loadSchemas();
            })
            .finally(() => setDdlBusy(false));
        }}
      />
      {ddlFlow?.action === "rename_table" ? (
        <div class="fixed inset-0 z-30 flex items-center justify-center bg-canvas/80 p-4">
          <div class="w-full max-w-sm rounded-panel border border-line bg-panel p-4">
            <p class="mb-2 text-13 font-medium text-ink">Rename table</p>
            <Input mono ariaLabel="New table name" value={renameTo} onInput={setRenameTo} />
            <div class="mt-3 flex justify-end gap-2">
              <Button label="Cancel" onClick={() => setDdlFlow(null)} />
              <Button
                label="Preview"
                variant="primary"
                disabled={!renameTo.trim()}
                onClick={() => setRenameReady(true)}
              />
            </div>
          </div>
        </div>
      ) : null}
      <div
        role="separator"
        aria-orientation="vertical"
        aria-label="Resize sidebar"
        class="absolute right-0 top-0 z-10 h-full w-1 cursor-col-resize touch-none hover:bg-accent-soft"
        onPointerDown={onResizePointerDown}
      />
    </aside>
  );
}

export { filterNames } from "./filter";
export { buildSidebarRows } from "./tree";
