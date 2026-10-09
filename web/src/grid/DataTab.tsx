import { useEffect, useState } from "preact/hooks";
import type { TableSortKey } from "../state";
import { appliedBrowseQuery, browseQueryKey, updateTableTabBrowseQuery } from "../workspace/browse-query";
import { FilterBar } from "../workspace/filter-bar";
import { shouldLoadTableResult } from "../workspace/table-load";
import { updateTableDataTabState } from "../workspace/tabs";
import type { WorkspaceTab } from "../state";
import { Grid } from "./Grid";
import { closeResult, openTableResult, type OpenResultResponse } from "./results-client";

type TableDataTab = Extract<WorkspaceTab, { kind: "table_data" }>;

export function DataTab({ sessionId, tab }: { sessionId: string; tab: TableDataTab }) {
  const [open, setOpen] = useState<OpenResultResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const applied = appliedBrowseQuery(tab);
  const appliedKey = browseQueryKey(applied);
  const loadResults = shouldLoadTableResult(tab);
  const [sortDraft, setSortDraft] = useState<TableSortKey[]>(() => applied.sort);

  useEffect(() => {
    setSortDraft(applied.sort);
  }, [tab.id, appliedKey]);

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
      <FilterBar
        columns={open.columns}
        applied={applied}
        sortDraft={sortDraft}
        onApply={(query) => handleApply({ ...query, sort: sortDraft })}
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
        initialScrollTop={tab.scrollTop}
        onScrollTopChange={(top) => updateTableDataTabState(tab.id, { scrollTop: top })}
      />
    </div>
  );
}
