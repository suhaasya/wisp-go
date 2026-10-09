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

export type CellValueResponse = {
  value: string;
  truncated?: boolean;
  full?: boolean;
  json?: boolean;
};

export async function fetchCellValue(
  sessionId: string,
  rid: string,
  row: number,
  col: string,
  signal?: AbortSignal,
): Promise<CellValueResponse> {
  const q = new URLSearchParams({ row: String(row), col });
  return request<CellValueResponse>(
    "GET",
    `/api/v1/sessions/${encodeURIComponent(sessionId)}/results/${encodeURIComponent(rid)}/value?${q}`,
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

export type ChangeOp = "set" | "insert" | "delete" | "revert";

export type ChangeRequest = {
  op: ChangeOp;
  row?: unknown[];
  col?: string;
  value?: string;
  null?: boolean;
  original?: string;
  origNull?: boolean;
  default?: boolean;
};

export type ChangeResponse = {
  count: number;
  insertKey?: string;
};

export async function postResultChange(
  sessionId: string,
  rid: string,
  body: ChangeRequest,
  signal?: AbortSignal,
): Promise<ChangeResponse> {
  return request<ChangeResponse>(
    "POST",
    `/api/v1/sessions/${encodeURIComponent(sessionId)}/results/${encodeURIComponent(rid)}/changes`,
    body,
    { signal },
  );
}

export type ReviewSQLResponse = {
  sql: string;
  count: number;
  statements: number;
  table: string;
  schema: string;
  needsDeleteConfirm?: boolean;
  warning?: string;
};

export async function fetchReviewSQL(
  sessionId: string,
  rid: string,
  signal?: AbortSignal,
): Promise<ReviewSQLResponse> {
  return request<ReviewSQLResponse>(
    "GET",
    `/api/v1/sessions/${encodeURIComponent(sessionId)}/results/${encodeURIComponent(rid)}/changes/sql`,
    undefined,
    { signal },
  );
}

export type CommitChangesResponse = { committed: number };

export async function commitResultChanges(
  sessionId: string,
  rid: string,
  body: { tableConfirm?: string },
  signal?: AbortSignal,
): Promise<CommitChangesResponse> {
  return request<CommitChangesResponse>(
    "POST",
    `/api/v1/sessions/${encodeURIComponent(sessionId)}/results/${encodeURIComponent(rid)}/changes/commit`,
    body,
    { signal },
  );
}

export async function discardResultChanges(
  sessionId: string,
  rid: string,
  signal?: AbortSignal,
): Promise<ChangeResponse> {
  return request<ChangeResponse>(
    "DELETE",
    `/api/v1/sessions/${encodeURIComponent(sessionId)}/results/${encodeURIComponent(rid)}/changes`,
    undefined,
    { signal },
  );
}
