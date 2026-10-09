import { useEffect, useState } from "preact/hooks";
import { Grid } from "../grid/Grid";
import { fetchResultWindow } from "../grid/results-client";
import type { ColumnMeta, RowTotal } from "../grid/window";

type QueryResultGridProps = {
  sessionId: string;
  rid: string;
  primaryKey?: string[];
  editable?: boolean;
};

export function QueryResultGrid({
  sessionId,
  rid,
  primaryKey = [],
  editable = false,
}: QueryResultGridProps) {
  const [columns, setColumns] = useState<ColumnMeta[]>([]);
  const [total, setTotal] = useState<RowTotal>({ kind: "estimate", value: 0 });
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const ac = new AbortController();
    setColumns([]);
    setError(null);
    void fetchResultWindow(sessionId, rid, 0, 150, ac.signal)
      .then((win) => {
        setColumns(win.columns);
        setTotal(win.total);
      })
      .catch((err: unknown) => {
        if (ac.signal.aborted) {
          return;
        }
        setError(err instanceof Error ? err.message : "Could not load result");
      });
    return () => ac.abort();
  }, [sessionId, rid]);

  if (error) {
    return <p class="p-4 text-12 text-env-prod">{error}</p>;
  }
  if (columns.length === 0) {
    return <div class="flex flex-1 bg-canvas" aria-busy="true" />;
  }
  return (
    <div class="gridwrap flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden">
      <Grid
        sessionId={sessionId}
        rid={rid}
        columns={columns}
        total={total}
        schema=""
        table=""
        editable={editable && primaryKey.length > 0}
        primaryKey={primaryKey}
      />
    </div>
  );
}
