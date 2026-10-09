import { request } from "../api";

export type StructureChangeOp = {
  op: string;
  column?: string;
  new_name?: string;
  column_def?: {
    name: string;
    type: string;
    nullable?: boolean;
    default?: string;
    comment?: string;
  };
  previous?: StructureChangeOp["column_def"];
  index_name?: string;
  index_columns?: string[];
  index_unique?: boolean;
  index_concurrent?: boolean;
};

export type StructureReviewResponse = {
  sql: string;
  count: number;
  statements?: number;
  schema?: string;
  table?: string;
  needsDeleteConfirm?: boolean;
  warning?: string;
};

export async function postStructureChange(
  sessionId: string,
  schema: string,
  table: string,
  body: StructureChangeOp,
  opts?: { signal?: AbortSignal },
): Promise<{ count: number }> {
  const path = `/api/v1/sessions/${encodeURIComponent(sessionId)}/tables/${encodeURIComponent(schema)}/${encodeURIComponent(table)}/structure/changes`;
  return request("POST", path, body, { signal: opts?.signal });
}

export async function discardStructureChanges(
  sessionId: string,
  schema: string,
  table: string,
  opts?: { signal?: AbortSignal },
): Promise<void> {
  const path = `/api/v1/sessions/${encodeURIComponent(sessionId)}/tables/${encodeURIComponent(schema)}/${encodeURIComponent(table)}/structure/changes`;
  await request("DELETE", path, undefined, { signal: opts?.signal });
}

export async function fetchStructureReviewSQL(
  sessionId: string,
  schema: string,
  table: string,
  opts?: { signal?: AbortSignal },
): Promise<StructureReviewResponse> {
  const path = `/api/v1/sessions/${encodeURIComponent(sessionId)}/tables/${encodeURIComponent(schema)}/${encodeURIComponent(table)}/structure/changes/sql`;
  return request("GET", path, undefined, { signal: opts?.signal });
}

export async function commitStructureChanges(
  sessionId: string,
  schema: string,
  table: string,
  body: { confirm?: boolean; tableConfirm?: string },
  opts?: { signal?: AbortSignal },
): Promise<{ committed: number }> {
  const path = `/api/v1/sessions/${encodeURIComponent(sessionId)}/tables/${encodeURIComponent(schema)}/${encodeURIComponent(table)}/structure/changes/commit`;
  return request("POST", path, body, { signal: opts?.signal });
}
