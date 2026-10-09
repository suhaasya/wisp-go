import { useEffect, useState } from "preact/hooks";
import { Grid } from "./Grid";
import { closeResult, openTableResult, type OpenResultResponse } from "./results-client";
import type { WorkspaceTableTab } from "../state";

export function DataTab({ sessionId, tab }: { sessionId: string; tab: WorkspaceTableTab }) {
  const [open, setOpen] = useState<OpenResultResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let rid: string | null = null;
    const ac = new AbortController();
    setOpen(null);
    setError(null);
    void openTableResult(sessionId, tab.schema, tab.table, ac.signal)
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
  }, [sessionId, tab.schema, tab.table]);

  if (error) {
    return <p class="p-4 text-13 text-env-prod">{error}</p>;
  }
  if (!open) {
    return <p class="p-4 text-13 text-ink-2">Loading table…</p>;
  }
  return (
    <Grid
      sessionId={sessionId}
      rid={open.rid}
      columns={open.columns}
      total={open.total}
      schema={tab.schema}
      table={tab.table}
    />
  );
}
