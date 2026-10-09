import { useEffect, useState } from "preact/hooks";
import { fetchTableDDL, fetchTableDetails, type TableDetailsResponse } from "../schema-client";
import { StructureEditor } from "./edit";

export function StructureTab({
  sessionId,
  schema,
  table,
}: {
  sessionId: string;
  schema: string;
  table: string;
}) {
  const [details, setDetails] = useState<TableDetailsResponse | null>(null);
  const [baselineDDL, setBaselineDDL] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [ddlLoading, setDDLLoading] = useState(true);
  const [reloadKey, setReloadKey] = useState(0);

  useEffect(() => {
    const ac = new AbortController();
    setError(null);
    void fetchTableDetails(sessionId, schema, table, { signal: ac.signal, refresh: reloadKey > 0 })
      .then(setDetails)
      .catch((e: unknown) => {
        if (!ac.signal.aborted) {
          setError(e instanceof Error ? e.message : "Could not load structure");
        }
      });
    setDDLLoading(true);
    void fetchTableDDL(sessionId, schema, table, { signal: ac.signal, refresh: reloadKey > 0 })
      .then((r) => setBaselineDDL(r.ddl))
      .catch(() => setBaselineDDL(""))
      .finally(() => {
        if (!ac.signal.aborted) {
          setDDLLoading(false);
        }
      });
    return () => ac.abort();
  }, [sessionId, schema, table, reloadKey]);

  if (error) {
    return <p class="p-4 text-13 text-env-prod">{error}</p>;
  }
  if (!details) {
    return <p class="p-4 text-13 text-ink-2">Loading structure…</p>;
  }

  return (
    <StructureEditor
      sessionId={sessionId}
      schema={schema}
      table={table}
      details={details}
      baselineDDL={baselineDDL}
      ddlLoading={ddlLoading}
      onRefresh={() => setReloadKey((k) => k + 1)}
    />
  );
}
