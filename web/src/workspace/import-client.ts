import { request } from "../api";
import type { ImportKind } from "./import-types";

export type ImportPreviewResponse = {
  kind: ImportKind;
  path: string;
  csv?: {
    encoding: string;
    delimiter: string;
    header: string[];
    rows: string[][];
    has_header: boolean;
    size_bytes: number;
  };
  jsonl?: {
    columns: string[];
    rows: string[][];
    size_bytes: number;
  };
  sql?: {
    size_bytes: number;
    sample: string;
    stmt_count: number;
  };
  inferred?: { name: string; type: string }[];
};

export type ColumnMapping = { file_column: string; table_column: string };

export type ImportStartBody = {
  path: string;
  kind: ImportKind;
  schema: string;
  table: string;
  delimiter?: string;
  has_header?: boolean;
  null_markers?: string[];
  batch_size?: number;
  tx_mode?: string;
  stop_on_error?: boolean;
  confirm?: boolean;
  mapping: ColumnMapping[];
  file_columns: string[];
};

export function previewImport(
  sessionId: string,
  path: string,
  signal?: AbortSignal,
): Promise<ImportPreviewResponse> {
  return request("POST", `/api/v1/sessions/${encodeURIComponent(sessionId)}/import/preview`, { path }, {
    signal,
  });
}

export function startImport(sessionId: string, body: ImportStartBody): Promise<{ import_id: string }> {
  return request("POST", `/api/v1/sessions/${encodeURIComponent(sessionId)}/import/start`, body);
}

export function cancelImport(sessionId: string, importId: string): Promise<void> {
  return request(
    "POST",
    `/api/v1/sessions/${encodeURIComponent(sessionId)}/import/${encodeURIComponent(importId)}/cancel`,
    undefined,
  );
}

export function importErrorsURL(sessionId: string, importId: string): string {
  return `/api/v1/sessions/${encodeURIComponent(sessionId)}/import/${encodeURIComponent(importId)}/errors`;
}
