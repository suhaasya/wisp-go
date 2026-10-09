const listeners = new Set<(sessionId: string) => void>();

export function onSchemaInvalidated(listener: (sessionId: string) => void): () => void {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

export function emitSchemaInvalidated(sessionId: string): void {
  for (const fn of listeners) {
    fn(sessionId);
  }
}
