export type ExportScope = "result" | "page" | "selection";
export type ExportFormat = "csv" | "json" | "sql" | "markdown";

export type ExportQuery = {
  format: ExportFormat;
  scope: ExportScope;
  rid: string;
  offset?: number;
  limit?: number;
  rowMin?: number;
  rowMax?: number;
  colMin?: number;
  colMax?: number;
  path?: string;
};

function appendInt(params: URLSearchParams, key: string, value: number | undefined): void {
  if (value === undefined || Number.isNaN(value)) {
    return;
  }
  params.set(key, String(value));
}

export function buildExportQuery(params: ExportQuery): URLSearchParams {
  const q = new URLSearchParams();
  q.set("format", params.format);
  q.set("scope", params.scope);
  q.set("rid", params.rid);
  appendInt(q, "offset", params.offset);
  appendInt(q, "limit", params.limit);
  appendInt(q, "row_min", params.rowMin);
  appendInt(q, "row_max", params.rowMax);
  appendInt(q, "col_min", params.colMin);
  appendInt(q, "col_max", params.colMax);
  if (params.path) {
    q.set("path", params.path);
  }
  return q;
}

export function exportDownloadURL(sessionId: string, params: ExportQuery): string {
  const q = buildExportQuery(params);
  return `/api/v1/sessions/${encodeURIComponent(sessionId)}/export?${q.toString()}`;
}

export async function fetchExportText(sessionId: string, params: ExportQuery, signal?: AbortSignal): Promise<string> {
  const res = await fetch(exportDownloadURL(sessionId, params), {
    credentials: "include",
    signal,
  });
  if (!res.ok) {
    const body = await res.text();
    throw new Error(body || `Export failed (${res.status})`);
  }
  return res.text();
}

export async function startFileExport(
  sessionId: string,
  params: ExportQuery & { path: string },
): Promise<{ export_id: string }> {
  const res = await fetch(exportDownloadURL(sessionId, params), { credentials: "include" });
  if (!res.ok) {
    const body = await res.text();
    throw new Error(body || `Export failed (${res.status})`);
  }
  return (await res.json()) as { export_id: string };
}

export async function cancelFileExport(sessionId: string, exportId: string): Promise<void> {
  const res = await fetch(
    `/api/v1/sessions/${encodeURIComponent(sessionId)}/export/${encodeURIComponent(exportId)}/cancel`,
    { method: "POST", credentials: "include" },
  );
  if (!res.ok && res.status !== 204) {
    throw new Error(`Cancel failed (${res.status})`);
  }
}
