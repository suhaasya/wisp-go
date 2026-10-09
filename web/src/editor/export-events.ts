export type ExportSSEPayload = {
  sid?: string;
  eid?: string;
  rows?: number;
  done?: boolean;
  cancelled?: boolean;
  error?: string;
};

type Listener = (ev: ExportSSEPayload) => void;

const listeners = new Set<Listener>();

export function onExportEvent(fn: Listener): () => void {
  listeners.add(fn);
  return () => listeners.delete(fn);
}

export function emitExportEvent(ev: ExportSSEPayload): void {
  for (const fn of listeners) {
    fn(ev);
  }
}
