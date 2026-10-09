import { request } from "../api";
import type { TableBrowseQuery } from "../state";
import type { ColumnMeta, RowTotal, WindowPayload } from "./window";

export type OpenResultResponse = {
  rid: string;
  columns: ColumnMeta[];
  total: RowTotal;
};

function sourceFromBrowse(schema: string, table: string, browse?: TableBrowseQuery) {
  const base = { kind: "table" as const, schema, table };
  if (!browse) {
    return base;
  }
  const payload: Record<string, unknown> = { ...base };
  if (browse.rawMode && browse.rawWhere.trim()) {
    payload.rawWhere = browse.rawWhere.trim();
  } else if (browse.filter.conditions.some((c) => c.column && (c.op === "is null" || c.value))) {
    payload.filter = browse.filter;
  }
  if (browse.sort.length > 0) {
    payload.sort = browse.sort;
  }
  return payload;
}

export async function openTableResult(
  sessionId: string,
  schema: string,
  table: string,
  browse?: TableBrowseQuery,
  signal?: AbortSignal,
): Promise<OpenResultResponse> {
  return request<OpenResultResponse>(
    "POST",
    `/api/v1/sessions/${encodeURIComponent(sessionId)}/results`,
    {
      source: sourceFromBrowse(schema, table, browse),
    },
    { signal },
  );
}

export async function fetchResultWindow(
  sessionId: string,
  rid: string,
  offset: number,
  limit: number,
  signal?: AbortSignal,
): Promise<WindowPayload> {
  const q = new URLSearchParams({
    offset: String(offset),
    limit: String(limit),
  });
  return request<WindowPayload>(
    "GET",
    `/api/v1/sessions/${encodeURIComponent(sessionId)}/results/${encodeURIComponent(rid)}/window?${q}`,
    undefined,
    { signal },
  );
}

export async function closeResult(sessionId: string, rid: string): Promise<void> {
  await request(
    "DELETE",
    `/api/v1/sessions/${encodeURIComponent(sessionId)}/results/${encodeURIComponent(rid)}`,
  );
}
