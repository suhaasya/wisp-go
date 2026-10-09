export type Matchable = {
  id: string;
  text: string;
};

export type PalettePrefixMode = "all" | "actions" | "connections" | "snippets";

const RECENCY_WEIGHT = 8;

/** Subsequence fuzzy score; 0 means no match. Higher is better. */
export function fuzzyScore(query: string, haystack: string): number {
  const q = query.trim().toLowerCase();
  const h = haystack.toLowerCase();
  if (q === "") {
    return 1;
  }
  let score = 0;
  let hi = 0;
  let run = 0;
  for (let qi = 0; qi < q.length; qi++) {
    const ch = q[qi]!;
    let found = -1;
    for (let i = hi; i < h.length; i++) {
      if (h[i] === ch) {
        found = i;
        break;
      }
    }
    if (found < 0) {
      return 0;
    }
    if (found === hi) {
      run++;
      score += 5 + run;
    } else {
      run = 1;
      score += 2;
    }
    if (found === 0) {
      score += 10;
    }
    const prev = found > 0 ? h[found - 1] : " ";
    if (prev === " " || prev === "." || prev === "_") {
      score += 6;
    }
    hi = found + 1;
  }
  if (h.startsWith(q)) {
    score += 20;
  }
  return score;
}

export function parsePaletteQuery(raw: string): { mode: PalettePrefixMode; query: string } {
  const t = raw.trimStart();
  if (t.startsWith(">")) {
    return { mode: "actions", query: t.slice(1).trimStart() };
  }
  if (t.startsWith("@")) {
    return { mode: "connections", query: t.slice(1).trimStart() };
  }
  if (t.startsWith("#")) {
    return { mode: "snippets", query: t.slice(1).trimStart() };
  }
  return { mode: "all", query: t };
}

export function rankByQuery(
  query: string,
  items: Matchable[],
  recency: Record<string, number>,
): Matchable[] {
  const q = query.trim().toLowerCase();
  const scored = items
    .map((item) => {
      const base = fuzzyScore(q, item.text);
      if (q !== "" && base === 0) {
        return null;
      }
      const boost = (recency[item.id] ?? 0) / 1_000_000_000;
      return { item, score: base + boost * RECENCY_WEIGHT };
    })
    .filter((x): x is { item: Matchable; score: number } => x !== null);
  scored.sort((a, b) => b.score - a.score || a.item.text.localeCompare(b.item.text));
  return scored.map((s) => s.item);
}

const RECENCY_KEY = "wisp.palette.recency";

export function loadRecency(): Record<string, number> {
  if (typeof localStorage === "undefined") {
    return {};
  }
  try {
    const raw = localStorage.getItem(RECENCY_KEY);
    if (!raw) {
      return {};
    }
    return JSON.parse(raw) as Record<string, number>;
  } catch {
    return {};
  }
}

export function touchRecency(id: string): void {
  if (typeof localStorage === "undefined") {
    return;
  }
  const map = loadRecency();
  map[id] = Date.now();
  const keys = Object.keys(map).sort((a, b) => (map[b] ?? 0) - (map[a] ?? 0));
  const trimmed: Record<string, number> = {};
  for (const k of keys.slice(0, 200)) {
    trimmed[k] = map[k]!;
  }
  try {
    localStorage.setItem(RECENCY_KEY, JSON.stringify(trimmed));
  } catch {
    // quota
  }
}
