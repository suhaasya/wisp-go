import { useCallback, useEffect, useRef, useState } from "preact/hooks";
import { ApiError } from "../../api";
import { shellConnection } from "../../state";
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
  const resizeRef = useRef<{ startX: number; startW: number } | null>(null);

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

  return (
    <aside
      class="relative flex shrink-0 flex-col border-r border-line bg-sidebar"
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
      {schemasError ? (
        <div class="flex flex-col gap-2 px-2.5 py-2 text-12-5 text-env-prod">
          <p>{schemasError}</p>
          {!sessionId.startsWith("mock-") ? (
            <button type="button" class="text-left text-accent underline" onClick={() => void loadSchemas()}>
              Retry
            </button>
          ) : null}
        </div>
      ) : (
        <SchemaTree
          sessionId={sessionId}
          schema={activeSchema}
          filterQuery={filter}
          expanded={ui.expanded}
          onExpandedChange={setExpanded}
          selectedObject={selectedObject}
          onSelectObject={(schema, name) => setSelectedObject({ schema, name })}
          onOpenTable={onOpenTable}
        />
      )}
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
