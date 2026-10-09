import { request } from "../api";
import type { ColumnMeta, RowTotal, WindowPayload } from "./window";

export type OpenResultResponse = {
  rid: string;
  columns: ColumnMeta[];
  total: RowTotal;
};

export async function openTableResult(
  sessionId: string,
  schema: string,
  table: string,
  signal?: AbortSignal,
): Promise<OpenResultResponse> {
  return request<OpenResultResponse>(
    "POST",
    `/api/v1/sessions/${encodeURIComponent(sessionId)}/results`,
    {
      source: { kind: "table", schema, table },
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
