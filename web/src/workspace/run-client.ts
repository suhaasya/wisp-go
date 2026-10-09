import { request } from "../api";
import type { ExplainNode } from "./results/explain";

export type RunMode = "cursor" | "selection" | "all";

export type RunBody = {
  sql: string;
  mode: RunMode;
  cursor?: number;
  sel_start?: number;
  sel_end?: number;
  confirm?: boolean;
};

export async function postRun(sessionId: string, body: RunBody): Promise<{ qid: string }> {
  return request("POST", `/api/v1/sessions/${encodeURIComponent(sessionId)}/run`, body);
}

export async function postCancelRun(sessionId: string, qid: string): Promise<void> {
  await request(
    "POST",
    `/api/v1/sessions/${encodeURIComponent(sessionId)}/cancel/${encodeURIComponent(qid)}`,
  );
}

export type ExplainBody = {
  sql: string;
  analyze?: boolean;
  confirm?: boolean;
};

export async function postExplain(
  sessionId: string,
  body: ExplainBody,
): Promise<{ root: ExplainNode }> {
  return request("POST", `/api/v1/sessions/${encodeURIComponent(sessionId)}/explain`, body);
}
