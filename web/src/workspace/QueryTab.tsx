import { useEffect, useRef } from "preact/hooks";
import { sessionEngine } from "../state";
import type { WorkspaceTab } from "../state";
import { updateQueryTabText } from "./tabs";

type QueryTab = Extract<WorkspaceTab, { kind: "query" }>;

export function QueryTab({ sessionId: _sessionId, tab }: { sessionId: string; tab: QueryTab }) {
  const hostRef = useRef<HTMLDivElement>(null);
  const handleRef = useRef<{ destroy: () => void } | null>(null);

  useEffect(() => {
    const host = hostRef.current;
    if (!host) {
      return;
    }
    let cancelled = false;
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
        onChange: (value) => {
          updateQueryTabText(tab.id, value);
        },
      });
      hostRef.current.dataset.editorMounted = "true";
    });
    return () => {
      cancelled = true;
      handleRef.current?.destroy();
      handleRef.current = null;
      if (hostRef.current) {
        delete hostRef.current.dataset.editorMounted;
      }
    };
  }, [tab.id]);

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
