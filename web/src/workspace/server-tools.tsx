import { useCallback, useEffect, useMemo, useRef, useState } from "preact/hooks";
import { readOnly } from "../state";
import { Button } from "../ui/Button";
import { Input } from "../ui/Input";
import { Modal } from "../ui/Modal";
import { Grid, type GridCellAccess } from "../grid/Grid";
import { fetchResultWindow } from "../grid/results-client";
import type { ColumnMeta, RowTotal } from "../grid/window";
import { emptySelection, primaryRow, type SelectionState } from "../grid/selection";
import { pushToast } from "../ui/Toast";
import {
  cancelServerProcess,
  fetchServerToolsUsers,
  fetchServerToolsVariables,
  openServerToolsProcesses,
  type ServerToolsUsersResponse,
  type ServerToolsVariable,
} from "./server-tools-client";
import { startProcessRefresh } from "./server-tools-refresh";

type Section = "processes" | "variables" | "users";

function pidColumnIndex(columns: ColumnMeta[]): number {
  for (let i = 0; i < columns.length; i++) {
    const n = columns[i]!.name.toLowerCase();
    if (n === "pid" || n === "id") {
      return i;
    }
  }
  return 0;
}

function highlightColumnIndex(columns: ColumnMeta[]): number {
  for (let i = 0; i < columns.length; i++) {
    if (columns[i]!.name === "wisp_highlight") {
      return i;
    }
  }
  return -1;
}

export function ServerTools({ sessionId }: { sessionId: string }) {
  const [section, setSection] = useState<Section>("processes");
  const [rid, setRid] = useState<string | null>(null);
  const [columns, setColumns] = useState<ColumnMeta[]>([]);
  const [total, setTotal] = useState<RowTotal>({ kind: "estimate", value: 0 });
  const [processError, setProcessError] = useState<string | null>(null);
  const [refreshing, setRefreshing] = useState(false);
  const [lastRefresh, setLastRefresh] = useState<string>("—");
  const [selection, setSelection] = useState<SelectionState>(emptySelection());
  const gridAccessRef = useRef<GridCellAccess | null>(null);
  const highlightColRef = useRef(-1);

  const [varQuery, setVarQuery] = useState("");
  const [variables, setVariables] = useState<ServerToolsVariable[]>([]);
  const [varError, setVarError] = useState<string | null>(null);

  const [users, setUsers] = useState<ServerToolsUsersResponse | null>(null);
  const [usersError, setUsersError] = useState<string | null>(null);

  const [confirmOpen, setConfirmOpen] = useState(false);
  const [terminateAction, setTerminateAction] = useState(false);
  const [actionPid, setActionPid] = useState<string | null>(null);
  const [acting, setActing] = useState(false);

  const loadProcesses = useCallback(
    (signal?: AbortSignal) => {
      setRefreshing(true);
      return openServerToolsProcesses(sessionId, signal)
        .then((open) => {
          setRid(open.rid);
          setColumns(open.columns);
          setTotal(open.total);
          highlightColRef.current = highlightColumnIndex(open.columns);
          setProcessError(null);
          setLastRefresh(new Date().toLocaleTimeString());
        })
        .catch((err: unknown) => {
          if (signal?.aborted) {
            return;
          }
          setProcessError(err instanceof Error ? err.message : "Could not load processes");
        })
        .finally(() => {
          if (!signal?.aborted) {
            setRefreshing(false);
          }
        });
    },
    [sessionId],
  );

  useEffect(() => {
    if (section !== "processes") {
      return;
    }
    const ac = new AbortController();
    void loadProcesses(ac.signal);
    const stop = startProcessRefresh({
      tick: () => {
        void loadProcesses();
      },
    });
    return () => {
      ac.abort();
      stop();
    };
  }, [section, loadProcesses]);

  useEffect(() => {
    if (section !== "variables") {
      return;
    }
    const ac = new AbortController();
    setVarError(null);
    void fetchServerToolsVariables(sessionId, varQuery, ac.signal)
      .then(setVariables)
      .catch((err: unknown) => {
        if (!ac.signal.aborted) {
          setVarError(err instanceof Error ? err.message : "Could not load variables");
        }
      });
    return () => ac.abort();
  }, [section, sessionId, varQuery]);

  useEffect(() => {
    if (section !== "users") {
      return;
    }
    const ac = new AbortController();
    setUsersError(null);
    void fetchServerToolsUsers(sessionId, ac.signal)
      .then(setUsers)
      .catch((err: unknown) => {
        if (!ac.signal.aborted) {
          setUsersError(err instanceof Error ? err.message : "Could not load users");
        }
      });
    return () => ac.abort();
  }, [section, sessionId]);

  const pidCol = useMemo(() => pidColumnIndex(columns), [columns]);

  const rowTone = useCallback(
    (row: number): "blocked" | "long_running" | null => {
      const hi = highlightColRef.current;
      const access = gridAccessRef.current;
      if (hi < 0 || !access) {
        return null;
      }
      const raw = access.getCell(row, hi);
      if (raw === "blocked") {
        return "blocked";
      }
      if (raw === "long_running") {
        return "long_running";
      }
      return null;
    },
    [rid, columns],
  );

  const selectedPid = (): string | null => {
    const row = primaryRow(selection);
    const access = gridAccessRef.current;
    if (row === null || !access) {
      return null;
    }
    const raw = access.getCell(row, pidCol);
    if (raw === null || raw === undefined) {
      return null;
    }
    return String(raw);
  };

  const openConfirm = (terminate: boolean) => {
    const pid = selectedPid();
    if (!pid) {
      pushToast("Select a process row first.", "error");
      return;
    }
    setActionPid(pid);
    setTerminateAction(terminate);
    setConfirmOpen(true);
  };

  const runConfirm = () => {
    if (!actionPid) {
      return;
    }
    setActing(true);
    void cancelServerProcess(sessionId, actionPid, terminateAction)
      .then(() => {
        pushToast(terminateAction ? "Process terminated." : "Cancel sent.", "success");
        setConfirmOpen(false);
        void loadProcesses();
      })
      .catch((err: unknown) => {
        pushToast(err instanceof Error ? err.message : "Action failed", "error");
      })
      .finally(() => setActing(false));
  };

  const ro = readOnly.value;

  return (
    <div class="flex min-h-0 flex-1 flex-col overflow-hidden">
      <div class="flex shrink-0 items-center gap-2 border-b border-line px-2.5 py-2">
        {(["processes", "variables", "users"] as const).map((s) => (
          <Button
            key={s}
            label={s === "processes" ? "Processes" : s === "variables" ? "Variables" : "Users"}
            variant={section === s ? "primary" : "ghost"}
            onClick={() => setSection(s)}
          />
        ))}
      </div>
      {section === "processes" ? (
        <>
          <div class="toolbar flex shrink-0 items-center gap-2 border-b border-line px-2.5 py-2">
            <span class="text-12-5 text-ink-2">
              {refreshing ? "Refreshing…" : `Updated ${lastRefresh}`}
            </span>
            <span class="flex-1" />
            <Button label="Cancel query" disabled={ro} onClick={() => openConfirm(false)} />
            <Button
              label="Terminate"
              variant="danger"
              disabled={ro}
              onClick={() => openConfirm(true)}
            />
          </div>
          {processError ? (
            <p class="p-4 text-13 text-env-prod">{processError}</p>
          ) : rid && columns.length > 0 ? (
            <div class="gridwrap flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden">
              <Grid
                sessionId={sessionId}
                rid={rid}
                columns={columns}
                total={total}
                schema=""
                table=""
                rowTone={rowTone}
                onSelectionChange={setSelection}
                onGridReady={(access) => {
                  gridAccessRef.current = access;
                }}
                fetchWindow={(offset, signal) =>
                  fetchResultWindow(sessionId, rid, offset, 150, signal)
                }
              />
            </div>
          ) : (
            <div class="flex flex-1 bg-canvas" aria-busy="true" />
          )}
        </>
      ) : null}
      {section === "variables" ? (
        <div class="flex min-h-0 flex-1 flex-col overflow-hidden p-3">
          <Input
            id="server-tools-var-search"
            label="Search variables"
            value={varQuery}
            onInput={(v) => setVarQuery(v)}
          />
          {varError ? (
            <p class="mt-3 text-13 text-env-prod">{varError}</p>
          ) : (
            <div class="mt-3 min-h-0 flex-1 overflow-auto rounded-control border border-line">
              <table class="w-full border-collapse text-left text-13">
                <thead class="sticky top-0 bg-raise text-12-5 text-ink-2">
                  <tr>
                    <th class="border-b border-line px-3 py-2 font-medium">Name</th>
                    <th class="border-b border-line px-3 py-2 font-medium">Value</th>
                  </tr>
                </thead>
                <tbody>
                  {variables.map((v) => (
                    <tr key={v.name} class="border-b border-line-2 hover:bg-raise">
                      <td class="px-3 py-1.5 font-mono text-12-5">{v.name}</td>
                      <td class="px-3 py-1.5 font-mono text-12-5 text-ink-2">{v.value}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      ) : null}
      {section === "users" ? (
        <div class="flex min-h-0 flex-1 flex-col overflow-hidden p-3">
          {usersError ? (
            <p class="text-13 text-env-prod">{usersError}</p>
          ) : users ? (
            <div class="min-h-0 flex-1 overflow-auto rounded-control border border-line">
              <table class="w-full border-collapse text-left text-13">
                <thead class="sticky top-0 bg-raise text-12-5 text-ink-2">
                  <tr>
                    {users.columns.map((c) => (
                      <th key={c} class="border-b border-line px-3 py-2 font-medium">
                        {c}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {users.rows.map((row, ri) => (
                    <tr key={ri} class="border-b border-line-2 hover:bg-raise">
                      {row.map((cell, ci) => (
                        <td key={ci} class="px-3 py-1.5 font-mono text-12-5">
                          {cell}
                        </td>
                      ))}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <div class="flex flex-1 bg-canvas" aria-busy="true" />
          )}
        </div>
      ) : null}
      <Modal
        open={confirmOpen}
        title={terminateAction ? "Terminate process" : "Cancel query"}
        onClose={() => setConfirmOpen(false)}
        footer={
          <div class="flex justify-end gap-2">
            <Button label="Back" variant="ghost" onClick={() => setConfirmOpen(false)} />
            <Button
              label={terminateAction ? "Terminate" : "Cancel query"}
              variant="danger"
              disabled={acting}
              onClick={runConfirm}
            />
          </div>
        }
      >
        <p class="text-13 text-ink-2">
          {terminateAction
            ? `Terminate backend ${actionPid ?? ""}? This ends the connection.`
            : `Send cancel to process ${actionPid ?? ""}?`}
        </p>
      </Modal>
    </div>
  );
}
