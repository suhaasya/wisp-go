import { request } from "../api";
import type { OpenResultResponse } from "../grid/results-client";

export async function openServerToolsProcesses(
  sessionId: string,
  signal?: AbortSignal,
): Promise<OpenResultResponse> {
  return request<OpenResultResponse>(
    "POST",
    `/api/v1/sessions/${encodeURIComponent(sessionId)}/server-tools/processes`,
    undefined,
    { signal },
  );
}

export async function cancelServerProcess(
  sessionId: string,
  pid: string,
  terminate: boolean,
  signal?: AbortSignal,
): Promise<void> {
  await request<void>(
    "POST",
    `/api/v1/sessions/${encodeURIComponent(sessionId)}/server-tools/processes/${encodeURIComponent(pid)}/cancel`,
    { terminate },
    { signal },
  );
}

export type ServerToolsVariable = { name: string; value: string };

export async function fetchServerToolsVariables(
  sessionId: string,
  q: string,
  signal?: AbortSignal,
): Promise<ServerToolsVariable[]> {
  const params = q.trim() ? `?q=${encodeURIComponent(q.trim())}` : "";
  const out = await request<{ variables: ServerToolsVariable[] }>(
    "GET",
    `/api/v1/sessions/${encodeURIComponent(sessionId)}/server-tools/variables${params}`,
    undefined,
    { signal },
  );
  return out.variables;
}

export type ServerToolsUsersResponse = {
  columns: string[];
  rows: string[][];
};

export async function fetchServerToolsUsers(
  sessionId: string,
  signal?: AbortSignal,
): Promise<ServerToolsUsersResponse> {
  return request<ServerToolsUsersResponse>(
    "GET",
    `/api/v1/sessions/${encodeURIComponent(sessionId)}/server-tools/users`,
    undefined,
    { signal },
  );
}
