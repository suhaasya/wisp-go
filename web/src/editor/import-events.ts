export type ImportSSEPayload = {
  sid?: string;
  iid?: string;
  rows?: number;
  bytes?: number;
  done?: boolean;
  cancelled?: boolean;
  error?: string;
};

type Listener = (ev: ImportSSEPayload) => void;

const listeners = new Set<Listener>();

export function onImportEvent(fn: Listener): () => void {
  listeners.add(fn);
  return () => listeners.delete(fn);
}

export function emitImportEvent(ev: ImportSSEPayload): void {
  for (const fn of listeners) {
    fn(ev);
  }
}
