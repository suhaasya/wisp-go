import { useCallback, useEffect, useRef, useState } from "preact/hooks";
import { onQueryEvent } from "../editor/query-events";
import { onSchemaInvalidated } from "../editor/schema-events";
import {
  mergeCompletionSchema,
  missingTableNames,
  referencedTableNames,
  type CompletionSchemaDTO,
} from "../editor/complete";
import { registerQueryKeyRunners } from "../keydispatch";
import { editorRunKeyBindings, formatBindingDisplay, resolveKeybindings } from "../keys";
import { getSettings } from "../settings";
import { readOnly, sessionEngine, workspaceProfileId } from "../state";
import { HistoryDrawer } from "./history-drawer";
import { fetchSnippets, saveSnippet, type SnippetsDocument } from "./snippets-client";
import type { WorkspaceTab } from "../state";
import { Button } from "../ui/Button";
import { Modal } from "../ui/Modal";
import { fetchCompletionSchema } from "./completion-client";
import { QueryResultGrid } from "./QueryResultGrid";
import { postCancelRun, postExplain, postRun, type RunBody } from "./run-client";
import { ExplainTree, type ExplainNode } from "./results/explain";
import { MessagesPanel, type QueryMessage } from "./results/messages";
import { ResultTabs, type QueryPanelTab } from "./results/tabs";
import { loadSidebarUiState } from "./sidebar/sidebar-state";
import { updateQueryTabText } from "./tabs";

type QueryTab = Extract<WorkspaceTab, { kind: "query" }>;

function activeSchemaName(profileId: string): string {
  const ui = loadSidebarUiState(profileId);
  return ui.selectedSchema || "public";
}

export function QueryTab({ sessionId, tab }: { sessionId: string; tab: QueryTab }) {
  const hostRef = useRef<HTMLDivElement>(null);
  const sqlTextRef = useRef(tab.queryText ?? "");
  const handleRef = useRef<{
    destroy: () => void;
    setCompletionSchema: (data: CompletionSchemaDTO) => void;
    setRunErrorAt: (n: number) => void;
    clearRunError: () => void;
    getSelectionRange: () => { start: number; end: number };
    getCursor: () => number;
    insertText: (text: string) => void;
  } | null>(null);
  const schemaDataRef = useRef<CompletionSchemaDTO>({ schema: "public", tables: [] });
  const hydrateTimerRef = useRef<number | null>(null);

  const [running, setRunning] = useState(false);
  const [activeQid, setActiveQid] = useState<string | null>(null);
  const [panelTabs, setPanelTabs] = useState<QueryPanelTab[]>([
    { id: "messages", kind: "messages" },
    { id: "explain", kind: "explain" },
  ]);
  const [activePanelId, setActivePanelId] = useState("messages");
  const [messages, setMessages] = useState<QueryMessage[]>([]);
  const [explainRoot, setExplainRoot] = useState<ExplainNode | null>(null);
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [historyOpen, setHistoryOpen] = useState(false);
  const [snippetOpen, setSnippetOpen] = useState(false);
  const [snippetName, setSnippetName] = useState("");
  const pendingRunRef = useRef<RunBody | null>(null);
  const lastRunBodyRef = useRef<RunBody | null>(null);
  const resultCounterRef = useRef(0);

  const startRun = useCallback(
    (body: RunBody) => {
      handleRef.current?.clearRunError();
      lastRunBodyRef.current = body;
      setRunning(true);
      void postRun(sessionId, body)
        .then(({ qid }) => {
          setActiveQid(qid);
        })
        .catch(() => {
          setRunning(false);
        });
    },
    [sessionId],
  );

  const buildRunBody = useCallback(
    (mode: RunBody["mode"], confirm = false): RunBody | null => {
      const sql = sqlTextRef.current;
      const h = handleRef.current;
      if (!h) {
        return null;
      }
      const base: RunBody = { sql, mode, confirm };
      if (mode === "cursor") {
        base.cursor = h.getCursor();
      }
      if (mode === "selection") {
        const sel = h.getSelectionRange();
        if (sel.end <= sel.start) {
          return null;
        }
        base.sel_start = sel.start;
        base.sel_end = sel.end;
      }
      return base;
    },
    [],
  );

  const onRunStatement = useCallback(() => {
    const body = buildRunBody("cursor");
    if (body) {
      startRun(body);
    }
  }, [buildRunBody, startRun]);

  const onRunAll = useCallback(() => {
    const body = buildRunBody("all");
    if (body) {
      startRun(body);
    }
  }, [buildRunBody, startRun]);

  const onCancel = useCallback(() => {
    if (activeQid) {
      void postCancelRun(sessionId, activeQid);
    }
  }, [activeQid, sessionId]);

  useEffect(() => {
    return registerQueryKeyRunners({
      runStatement: onRunStatement,
      runAll: onRunAll,
      cancel: onCancel,
    });
  }, [onRunStatement, onRunAll, onCancel]);

  const keybindingsKey = JSON.stringify(getSettings().keybindings ?? {});

  const onExplain = useCallback(
    (analyze = false, confirm = false) => {
      const h = handleRef.current;
      if (!h) {
        return;
      }
      const sel = h.getSelectionRange();
      const stmt =
        sel.end > sel.start
          ? sqlTextRef.current.slice(sel.start, sel.end).trim()
          : sqlTextRef.current.trim();
      void postExplain(sessionId, { sql: stmt, analyze, confirm })
        .then((out) => {
          setExplainRoot(out.root);
          setActivePanelId("explain");
        })
        .catch((err: unknown) => {
          const code =
            err && typeof err === "object" && "code" in err
              ? String((err as { code: string }).code)
              : "";
          if (code === "needs_confirmation") {
            pendingRunRef.current = { sql: stmt, mode: "all", confirm: false };
            setConfirmOpen(true);
          }
        });
    },
    [sessionId],
  );

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
      void fetchSnippets()
        .catch((): SnippetsDocument => ({ version: 1, snippets: [] }))
        .then((doc) => {
          if (cancelled || !hostRef.current) {
            return;
          }
          const resolved = resolveKeybindings(getSettings().keybindings ?? {});
          handleRef.current = mountSqlEditor({
            parent: hostRef.current,
            value: tab.queryText ?? "",
            engine: sqlEngine,
            schema,
            snippets: (doc.snippets ?? []).map((s) => ({ name: s.name, tags: s.tags })),
            runKeys: {
              onRunStatement,
              onRunAll,
              onCancel,
            },
            runKeyBindings: editorRunKeyBindings(resolved),
            onChange: (value) => {
              sqlTextRef.current = value;
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
          sqlTextRef.current = tab.queryText ?? "";
          hostRef.current!.dataset.editorMounted = "true";
          loadSchema();
        });
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
  }, [sessionId, tab.id, onRunStatement, onRunAll, onCancel, keybindingsKey]);

  useEffect(() => {
    const off = onQueryEvent((ev) => {
      if (ev.sid !== sessionId) {
        return;
      }
      if (ev.qid && ev.qid === activeQid) {
        if (ev.rid) {
          resultCounterRef.current += 1;
          const n = resultCounterRef.current;
          const id = ev.rid;
          setPanelTabs((prev) => {
            const next = prev.filter((t) => t.id !== id);
            next.splice(next.length - 2, 0, {
              id,
              kind: "result",
              rid: id,
              label: `Result ${n}`,
              ms: ev.duration_ms,
            });
            return next;
          });
          setActivePanelId(id);
        }
        if (ev.stmt_index && (ev.duration_ms !== undefined || ev.error || ev.rows_affected)) {
          const msg: QueryMessage = {
            id: `${ev.qid}-${ev.stmt_index}-${Date.now()}`,
            stmtIndex: ev.stmt_index,
            stmtTotal: ev.stmt_total ?? ev.stmt_index,
            durationMs: ev.duration_ms,
            rowsAffected: ev.rows_affected,
            error: ev.error,
            errorDetail: ev.error_detail,
            position: ev.position,
          };
          setMessages((prev) => [...prev, msg]);
          if (ev.error && ev.position !== undefined) {
            handleRef.current?.setRunErrorAt(ev.position);
          }
        }
        if (ev.needs_confirm) {
          setRunning(false);
          if (lastRunBodyRef.current) {
            pendingRunRef.current = lastRunBodyRef.current;
          }
          setConfirmOpen(true);
        }
        if (ev.done) {
          setRunning(false);
          setActiveQid(null);
        }
      }
    });
    return off;
  }, [sessionId, activeQid]);

  const activeResult = panelTabs.find(
    (t): t is Extract<QueryPanelTab, { kind: "result" }> =>
      t.id === activePanelId && t.kind === "result",
  );
  const meta =
    activeResult?.ms !== undefined
      ? `${activeResult.rowsLabel ?? "—"} · ${activeResult.ms} ms`
      : undefined;

  const runKbd = formatBindingDisplay(
    resolveKeybindings(getSettings().keybindings ?? {})["run-statement"],
  );

  return (
    <div class="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden">
      <div class="toolbar flex shrink-0 items-center gap-2 border-b border-line px-2.5 py-2">
        <Button
          label="Run statement"
          variant="primary"
          kbd={runKbd}
          disabled={running}
          onClick={onRunStatement}
        />
        <Button label="Run all" disabled={running} onClick={onRunAll} />
        <Button label="Explain" disabled={running} onClick={onExplain} />
        {running ? (
          <Button label="Cancel" variant="ghost" onClick={onCancel} />
        ) : null}
        <Button label="History" onClick={() => setHistoryOpen(true)} />
        <Button label="Save snippet" onClick={() => setSnippetOpen(true)} />
        <span class="flex-1" />
      </div>
      <div
        class="editor flex h-[250px] shrink-0 border-b border-line bg-panel font-mono text-13 leading-[21px]"
        data-testid="sql-editor"
      >
        <div ref={hostRef} class="h-full min-h-0 min-w-0 flex-1" data-testid="sql-editor-host" />
      </div>
      <ResultTabs
        tabs={panelTabs}
        activeId={activePanelId}
        meta={meta}
        onSelect={setActivePanelId}
      />
      <div class="relative flex min-h-0 flex-1 flex-col overflow-hidden bg-canvas">
        <HistoryDrawer
          open={historyOpen}
          profileId={workspaceProfileId.value ?? ""}
          onClose={() => setHistoryOpen(false)}
          onInsert={(sql) => handleRef.current?.insertText(sql)}
          onRun={(sql) => {
            handleRef.current?.insertText(sql);
            startRun({ sql, mode: "all", confirm: false });
          }}
        />
        {activePanelId === "messages" ? (
          <MessagesPanel
            lines={messages}
            onErrorClick={(pos) => handleRef.current?.setRunErrorAt(pos)}
          />
        ) : null}
        {activePanelId === "explain" && explainRoot ? (
          <ExplainTree root={explainRoot} />
        ) : null}
        {activeResult ? (
          <QueryResultGrid
            sessionId={sessionId}
            rid={activeResult.rid}
            editable={!readOnly.value}
          />
        ) : null}
      </div>
      <Modal
        open={snippetOpen}
        title="Save snippet"
        onClose={() => setSnippetOpen(false)}
        footer={
          <div class="flex justify-end gap-2">
            <Button label="Cancel" onClick={() => setSnippetOpen(false)} />
            <Button
              label="Save"
              variant="primary"
              onClick={() => {
                const name = snippetName.trim();
                const sql = sqlTextRef.current.trim();
                if (!name || !sql) {
                  return;
                }
                void saveSnippet({ name, sql }).then(() => setSnippetOpen(false));
              }}
            />
          </div>
        }
      >
        <label class="block text-12 text-ink-2">
          Name
          <input
            class="input mt-1 w-full font-mono text-13"
            value={snippetName}
            onInput={(e) => setSnippetName((e.target as HTMLInputElement).value)}
          />
        </label>
      </Modal>
      <Modal
        open={confirmOpen}
        title="Confirm write"
        description="This statement changes data. Run it anyway?"
        onClose={() => setConfirmOpen(false)}
        footer={
          <div class="flex justify-end gap-2">
            <Button label="Cancel" onClick={() => setConfirmOpen(false)} />
            <Button
              label="Run"
              variant="primary"
              onClick={() => {
                setConfirmOpen(false);
                const pending = pendingRunRef.current;
                if (pending) {
                  startRun({ ...pending, confirm: true });
                  pendingRunRef.current = null;
                  return;
                }
                const body = buildRunBody("cursor", true);
                if (body) {
                  startRun(body);
                }
              }}
            />
          </div>
        }
      />
    </div>
  );
}
