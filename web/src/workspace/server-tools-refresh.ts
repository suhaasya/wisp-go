export const PROCESS_REFRESH_MS = 2000;

export type ProcessRefreshOptions = {
  intervalMs?: number;
  isHidden?: () => boolean;
  tick: () => void;
};

/** Polls process list while the document is visible; pauses when hidden. */
export function startProcessRefresh(options: ProcessRefreshOptions): () => void {
  const intervalMs = options.intervalMs ?? PROCESS_REFRESH_MS;
  const isHidden = options.isHidden ?? (() => document.hidden);
  const id = window.setInterval(() => {
    if (!isHidden()) {
      options.tick();
    }
  }, intervalMs);
  const onVis = () => {
    if (!isHidden()) {
      options.tick();
    }
  };
  document.addEventListener("visibilitychange", onVis);
  return () => {
    window.clearInterval(id);
    document.removeEventListener("visibilitychange", onVis);
  };
}
