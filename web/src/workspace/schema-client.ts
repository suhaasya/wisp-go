import { request } from "../api";

export type SchemasResponse = {
  databases?: string[];
  schemas?: string[];
};

export type ObjectsResponse = {
  names: string[];
};

export type ObjectKind =
  | "table"
  | "view"
  | "matview"
  | "function"
  | "procedure"
  | "sequence"
  | "enum";

export async function fetchSchemas(
  sessionId: string,
  opts?: { refresh?: boolean; signal?: AbortSignal },
): Promise<SchemasResponse> {
  const q = opts?.refresh ? "?refresh=true" : "";
  return request<SchemasResponse>(
    "GET",
    `/api/v1/sessions/${encodeURIComponent(sessionId)}/schemas${q}`,
    undefined,
    { signal: opts?.signal },
  );
}

export type TableDetailsResponse = {
  schema: string;
  table: string;
  columns: { name: string; type: string }[];
  primary_key?: string[];
  foreign_keys?: {
    name: string;
    columns: string[];
    ref_schema: string;
    ref_table: string;
    ref_columns: string[];
  }[];
};

export async function fetchTableDetails(
  sessionId: string,
  schema: string,
  table: string,
  opts?: { signal?: AbortSignal },
): Promise<TableDetailsResponse> {
  const path = `/api/v1/sessions/${encodeURIComponent(sessionId)}/tables/${encodeURIComponent(schema)}/${encodeURIComponent(table)}`;
  return request<TableDetailsResponse>("GET", path, undefined, { signal: opts?.signal });
}

export async function fetchObjects(
  sessionId: string,
  schema: string,
  kind: ObjectKind,
  opts?: { refresh?: boolean; signal?: AbortSignal },
): Promise<string[]> {
  const params = new URLSearchParams({ schema, kind });
  if (opts?.refresh) {
    params.set("refresh", "true");
  }
  const out = await request<ObjectsResponse>(
    "GET",
    `/api/v1/sessions/${encodeURIComponent(sessionId)}/objects?${params}`,
    undefined,
    { signal: opts?.signal },
  );
  return out.names ?? [];
}
