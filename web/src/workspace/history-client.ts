import type { HistoryEntry } from "./history-drawer";

export async function fetchHistory(
  profileId: string,
  q: string,
  signal?: AbortSignal,
): Promise<HistoryEntry[]> {
  const params = new URLSearchParams({ profile: profileId });
  if (q.trim()) {
    params.set("q", q.trim());
  }
  const res = await fetch(`/api/v1/history?${params}`, {
    credentials: "same-origin",
    signal,
  });
  if (!res.ok) {
    throw new Error(`history failed: ${res.status}`);
  }
  const text = await res.text();
  if (!text.trim()) {
    return [];
  }
  const out: HistoryEntry[] = [];
  for (const line of text.split("\n")) {
    if (!line.trim()) {
      continue;
    }
    out.push(JSON.parse(line) as HistoryEntry);
  }
  return out;
}
