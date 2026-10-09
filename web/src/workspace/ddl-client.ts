import { request } from "../api";

export type DDLDependent = { schema: string; name: string; kind: string };

export type DDLPreviewResponse = {
  sql: string;
  dependents?: DDLDependent[];
  needs_typed_confirm?: boolean;
  confirm_name?: string;
  warning?: string;
};

export type DDLPreviewBody = {
  action: string;
  schema: string;
  table?: string;
  new_name?: string;
  cascade?: boolean;
  wizard?: {
    table: string;
    columns: {
      name: string;
      type: string;
      nullable?: boolean;
      default?: string;
      primary?: boolean;
    }[];
    indexes?: { name: string; columns: string[]; unique?: boolean }[];
  };
};

export async function previewDDL(
  sessionId: string,
  body: DDLPreviewBody,
  opts?: { signal?: AbortSignal },
): Promise<DDLPreviewResponse> {
  return request(
    "POST",
    `/api/v1/sessions/${encodeURIComponent(sessionId)}/ddl/preview`,
    body,
    { signal: opts?.signal },
  );
}

export async function executeDDL(
  sessionId: string,
  body: DDLPreviewBody & { confirm?: boolean; object_confirm?: string },
  opts?: { signal?: AbortSignal },
): Promise<{ ok: boolean }> {
  return request(
    "POST",
    `/api/v1/sessions/${encodeURIComponent(sessionId)}/ddl/execute`,
    body,
    { signal: opts?.signal },
  );
}
