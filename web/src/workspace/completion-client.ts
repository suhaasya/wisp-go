import { request } from "../api";
import type { CompletionSchemaDTO } from "../editor/complete";

export async function fetchCompletionSchema(
  sessionId: string,
  schema: string,
  tables?: string[],
  signal?: AbortSignal,
): Promise<CompletionSchemaDTO> {
  const params = new URLSearchParams({ schema });
  if (tables && tables.length > 0) {
    params.set("tables", tables.join(","));
  }
  return request<CompletionSchemaDTO>(
    "GET",
    `/api/v1/sessions/${encodeURIComponent(sessionId)}/complete?${params}`,
    undefined,
    { signal },
  );
}
