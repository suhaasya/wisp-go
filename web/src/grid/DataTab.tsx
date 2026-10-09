import { useEffect, useState } from "preact/hooks";
import type { TableBrowseQuery, TableSortKey } from "../state";
import {
  appliedBrowseQuery,
  browseQueryKey,
  updateTableTabBrowseQuery,
} from "../workspace/browse-query";
import { FilterBar } from "../workspace/filter-bar";
import { Grid } from "./Grid";
import { closeResult, openTableResult, type OpenResultResponse } from "./results-client";
import type { WorkspaceTableTab } from "../state";

export function DataTab({ sessionId, tab }: { sessionId: string; tab: WorkspaceTableTab }) {
  const [open, setOpen] = useState<OpenResultResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const applied = appliedBrowseQuery(tab);
  const appliedKey = browseQueryKey(applied);

  const [sortDraft, setSortDraft] = useState<TableSortKey[]>(() => applied.sort);

  useEffect(() => {
    setSortDraft(applied.sort);
  }, [tab.id, appliedKey]);

  useEffect(() => {
    let rid: string | null = null;
    const ac = new AbortController();
    setOpen(null);
    setError(null);
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
  }, [sessionId, tab.schema, tab.table, appliedKey]);

  function handleApply(query: TableBrowseQuery) {
    updateTableTabBrowseQuery(tab.id, query);
  }

  if (error) {
    return <p class="p-4 text-13 text-env-prod">{error}</p>;
  }
  if (!open) {
    return <p class="p-4 text-13 text-ink-2">Loading table…</p>;
  }
  return (
    <div class="flex min-h-0 flex-1 flex-col overflow-hidden">
      <FilterBar
        columns={open.columns}
        applied={applied}
        sortDraft={sortDraft}
        onApply={handleApply}
      />
      <Grid
        sessionId={sessionId}
        rid={open.rid}
        columns={open.columns}
        total={open.total}
        schema={tab.schema}
        table={tab.table}
        sortDraft={sortDraft}
        onSortDraftChange={setSortDraft}
      />
    </div>
  );
}
