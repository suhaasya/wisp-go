export type QuerySSEPayload = {
  sid?: string;
  qid?: string;
  done?: boolean;
  cancelled?: boolean;
  stmt_index?: number;
  stmt_total?: number;
  stmt_start?: number;
  rid?: string;
  rows_affected?: number;
  duration_ms?: number;
  error?: string;
  error_detail?: string;
  position?: number;
  needs_confirm?: boolean;
};

type Listener = (payload: QuerySSEPayload) => void;

const listeners = new Set<Listener>();

export function onQueryEvent(fn: Listener): () => void {
  listeners.add(fn);
  return () => listeners.delete(fn);
}

export function emitQueryEvent(payload: QuerySSEPayload): void {
  for (const fn of listeners) {
    fn(payload);
  }
}
