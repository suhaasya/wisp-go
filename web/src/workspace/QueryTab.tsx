import { useEffect, useRef } from "preact/hooks";
import { onSchemaInvalidated } from "../editor/schema-events";
import {
  mergeCompletionSchema,
  missingTableNames,
  referencedTableNames,
  type CompletionSchemaDTO,
} from "../editor/complete";
import { sessionEngine, workspaceProfileId } from "../state";
import type { WorkspaceTab } from "../state";
import { fetchCompletionSchema } from "./completion-client";
import { loadSidebarUiState } from "./sidebar/sidebar-state";
import { updateQueryTabText } from "./tabs";

type QueryTab = Extract<WorkspaceTab, { kind: "query" }>;

function activeSchemaName(profileId: string): string {
  const ui = loadSidebarUiState(profileId);
  return ui.selectedSchema || "public";
}

export function QueryTab({ sessionId, tab }: { sessionId: string; tab: QueryTab }) {
  const hostRef = useRef<HTMLDivElement>(null);
  const handleRef = useRef<{
    destroy: () => void;
    setCompletionSchema: (data: CompletionSchemaDTO) => void;
  } | null>(null);
  const schemaDataRef = useRef<CompletionSchemaDTO>({ schema: "public", tables: [] });
  const hydrateTimerRef = useRef<number | null>(null);

  useEffect(() => {
    const host = hostRef.current;
    if (!host) {
      return;
    }
    let cancelled = false;
    const profileId = workspaceProfileId.value ?? "";
    const schema = activeSchemaName(profileId);

    const loadSchema = (extra?: string[]) => {
      void fetchCompletionSchema(sessionId, schema, extra).then((data) => {
        if (cancelled) {
          return;
        }
        schemaDataRef.current = mergeCompletionSchema(schemaDataRef.current, data);
        handleRef.current?.setCompletionSchema(schemaDataRef.current);
      });
    };

    void import("../editor/index").then(({ mountSqlEditor }) => {
      if (cancelled || !hostRef.current) {
        return;
      }
      const engine = sessionEngine.value;
      const sqlEngine =
        engine === "mysql" || engine === "mariadb" ? engine : "postgres";
      handleRef.current = mountSqlEditor({
        parent: hostRef.current,
        value: tab.queryText ?? "",
        engine: sqlEngine,
        schema,
        onChange: (value) => {
          updateQueryTabText(tab.id, value);
          if (hydrateTimerRef.current !== null) {
            window.clearTimeout(hydrateTimerRef.current);
          }
          hydrateTimerRef.current = window.setTimeout(() => {
            hydrateTimerRef.current = null;
            const loaded = new Set(schemaDataRef.current.tables.map((t) => t.name));
            const need = missingTableNames(loaded, referencedTableNames(value));
            if (need.length > 0) {
              loadSchema(need);
            }
          }, 400);
        },
      });
      hostRef.current.dataset.editorMounted = "true";
      loadSchema();
    });

    const offSchema = onSchemaInvalidated((sid) => {
      if (sid !== sessionId) {
        return;
      }
      schemaDataRef.current = { schema, tables: [] };
      loadSchema();
    });

    return () => {
      cancelled = true;
      offSchema();
      if (hydrateTimerRef.current !== null) {
        window.clearTimeout(hydrateTimerRef.current);
      }
      handleRef.current?.destroy();
      handleRef.current = null;
      if (hostRef.current) {
        delete hostRef.current.dataset.editorMounted;
      }
    };
  }, [sessionId, tab.id]);

  return (
    <div class="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden">
      <div
        class="editor flex h-[250px] shrink-0 border-b border-line bg-panel font-mono text-13 leading-[21px]"
        data-testid="sql-editor"
      >
        <div ref={hostRef} class="min-h-0 min-w-0 flex-1" data-testid="sql-editor-host" />
      </div>
      <div class="flex min-h-0 flex-1 flex-col overflow-hidden bg-canvas p-4">
        <p class="text-12 text-ink-3">Run statement or script (LW-032).</p>
      </div>
    </div>
  );
}
