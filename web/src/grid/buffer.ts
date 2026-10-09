import { DEFAULT_WINDOW_ROWS, windowOffsetForRow, type WindowPayload } from "./window";

export type FetchWindowFn = (offset: number, signal: AbortSignal) => Promise<WindowPayload>;

export type WindowBufferOptions = {
  windowRows?: number;
  maxWindows?: number;
  fetch: FetchWindowFn;
  /** Called after a window is loaded or cache changes (schedule grid relayout). */
  onUpdate?: () => void;
};

export type ViewportOffsets = {
  required: number[];
  prefetch: number[];
};

/** Pure helper: which window offsets to load and prefetch for a viewport. */
export function offsetsNeededForViewport(
  firstRow: number,
  lastRow: number,
  windowRows: number,
  cachedOffsets: number[],
): ViewportOffsets {
  if (firstRow < 0) {
    firstRow = 0;
  }
  if (lastRow < firstRow) {
    lastRow = firstRow;
  }
  const startOff = windowOffsetForRow(firstRow, windowRows);
  const endOff = windowOffsetForRow(lastRow, windowRows);
  const required: number[] = [];
  for (let o = startOff; o <= endOff; o += windowRows) {
    required.push(o);
  }
  const cached = new Set(cachedOffsets);
  const half = windowRows / 2;
  const prefetch: number[] = [];
  if (firstRow - startOff < half) {
    const prev = startOff - windowRows;
    if (prev >= 0 && !cached.has(prev) && !required.includes(prev)) {
      prefetch.push(prev);
    }
  }
  const lastWindowEnd = endOff + windowRows - 1;
  if (lastWindowEnd - lastRow < half) {
    const next = endOff + windowRows;
    if (!cached.has(next) && !required.includes(next)) {
      prefetch.push(next);
    }
  }
  return { required, prefetch };
}

type CacheEntry = {
  payload: WindowPayload;
  lastTouch: number;
};

export type WindowBuffer = {
  ensureViewport: (firstRow: number, lastRow: number) => Promise<void>;
  getWindow: (offset: number) => WindowPayload | undefined;
  cellValue: (row: number, col: number) => unknown | undefined;
  isRowLoaded: (row: number) => boolean;
  cachedOffsets: () => number[];
  lastFetchTime: () => Date | null;
  destroy: () => void;
};

export function createWindowBuffer(opts: WindowBufferOptions): WindowBuffer {
  const windowRows = opts.windowRows ?? DEFAULT_WINDOW_ROWS;
  const maxWindows = opts.maxWindows ?? 3;
  const cache = new Map<number, CacheEntry>();
  let touchSeq = 0;
  let lastFetch: Date | null = null;
  let generation = 0;
  let inflight: AbortController | null = null;

  function touch(offset: number): void {
    const e = cache.get(offset);
    if (e) {
      e.lastTouch = ++touchSeq;
    }
  }

  function evictIfNeeded(): void {
    while (cache.size > maxWindows) {
      let oldestOff = -1;
      let oldestTouch = Infinity;
      for (const [off, ent] of cache) {
        if (ent.lastTouch < oldestTouch) {
          oldestTouch = ent.lastTouch;
          oldestOff = off;
        }
      }
      if (oldestOff >= 0) {
        cache.delete(oldestOff);
      } else {
        break;
      }
    }
  }

  async function loadOffset(offset: number, signal: AbortSignal): Promise<void> {
    if (cache.has(offset)) {
      touch(offset);
      return;
    }
    const payload = await opts.fetch(offset, signal);
    if (signal.aborted) {
      return;
    }
    cache.set(offset, { payload, lastTouch: ++touchSeq });
    lastFetch = new Date();
    evictIfNeeded();
    opts.onUpdate?.();
  }

  async function ensureViewport(firstRow: number, lastRow: number): Promise<void> {
    const gen = ++generation;
    inflight?.abort();
    const ac = new AbortController();
    inflight = ac;
    const cached = [...cache.keys()];
    const { required, prefetch } = offsetsNeededForViewport(firstRow, lastRow, windowRows, cached);
    const toLoad = [...required, ...prefetch];
    for (const off of toLoad) {
      if (gen !== generation) {
        ac.abort();
        throw new DOMException("superseded", "AbortError");
      }
      try {
        await loadOffset(off, ac.signal);
      } catch (e) {
        if (ac.signal.aborted || gen !== generation) {
          throw e;
        }
        throw e;
      }
    }
    if (inflight === ac) {
      inflight = null;
    }
  }

  function getWindow(offset: number): WindowPayload | undefined {
    touch(offset);
    return cache.get(offset)?.payload;
  }

  function windowForRow(row: number): WindowPayload | undefined {
    const off = windowOffsetForRow(row, windowRows);
    return getWindow(off);
  }

  function isRowLoaded(row: number): boolean {
    const w = windowForRow(row);
    if (!w) {
      return false;
    }
    const idx = row - w.offset;
    const col0 = w.data[0];
    return idx >= 0 && col0 !== undefined && idx < col0.length;
  }

  function cellValue(row: number, col: number): unknown | undefined {
    const w = windowForRow(row);
    if (!w) {
      return undefined;
    }
    const idx = row - w.offset;
    const colData = w.data[col];
    if (!colData || idx < 0 || idx >= colData.length) {
      return undefined;
    }
    if (w.nulls?.[col]?.includes(idx)) {
      return null;
    }
    return colData[idx];
  }

  return {
    ensureViewport,
    getWindow,
    cellValue,
    isRowLoaded,
    cachedOffsets: () => [...cache.keys()],
    lastFetchTime: () => lastFetch,
    destroy: () => {
      generation++;
      inflight?.abort();
      inflight = null;
      cache.clear();
    },
  };
}
