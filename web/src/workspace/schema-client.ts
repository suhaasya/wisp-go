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
  kind?: string;
  columns: {
    name: string;
    type: string;
    nullable?: boolean;
    default?: string;
    comment?: string;
    ordinal?: number;
  }[];
  primary_key?: string[];
  foreign_keys?: {
    name: string;
    columns: string[];
    ref_schema: string;
    ref_table: string;
    ref_columns: string[];
  }[];
  indexes?: {
    name: string;
    columns: string[];
    unique?: boolean;
    primary?: boolean;
  }[];
  checks?: { name: string; expression: string }[];
  triggers?: { name: string; timing: string; events: string; enabled?: boolean }[];
};

export async function fetchTableDetails(
  sessionId: string,
  schema: string,
  table: string,
  opts?: { refresh?: boolean; signal?: AbortSignal },
): Promise<TableDetailsResponse> {
  const q = opts?.refresh ? "?refresh=true" : "";
  const path = `/api/v1/sessions/${encodeURIComponent(sessionId)}/tables/${encodeURIComponent(schema)}/${encodeURIComponent(table)}${q}`;
  return request<TableDetailsResponse>("GET", path, undefined, { signal: opts?.signal });
}

export type TableDDLResponse = { ddl: string };

export async function fetchTableDDL(
  sessionId: string,
  schema: string,
  table: string,
  opts?: { refresh?: boolean; signal?: AbortSignal },
): Promise<TableDDLResponse> {
  const q = opts?.refresh ? "?refresh=true" : "";
  const path = `/api/v1/sessions/${encodeURIComponent(sessionId)}/tables/${encodeURIComponent(schema)}/${encodeURIComponent(table)}/ddl${q}`;
  return request<TableDDLResponse>("GET", path, undefined, { signal: opts?.signal });
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
