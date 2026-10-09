export type QueryMessage = {
  id: string;
  stmtIndex: number;
  stmtTotal: number;
  durationMs?: number;
  rowsAffected?: number;
  error?: string;
  errorDetail?: string;
  position?: number;
};

type MessagesPanelProps = {
  lines: QueryMessage[];
  onErrorClick?: (position: number) => void;
};

export function MessagesPanel({ lines, onErrorClick }: MessagesPanelProps) {
  if (lines.length === 0) {
    return (
      <p class="p-4 text-12 text-ink-3" data-testid="query-messages-empty">
        No messages yet.
      </p>
    );
  }
  return (
    <ul class="min-h-0 flex-1 list-none overflow-auto p-3 text-12" data-testid="query-messages">
      {lines.map((line) => (
        <li key={line.id} class="border-b border-line py-2 font-mono">
          <span class="text-ink-3">
            Statement {line.stmtIndex}/{line.stmtTotal}
          </span>
          {line.durationMs !== undefined ? (
            <span class="text-ink-2"> · {line.durationMs} ms</span>
          ) : null}
          {line.rowsAffected !== undefined && line.rowsAffected > 0 ? (
            <span class="text-ink-2"> · {line.rowsAffected} rows affected</span>
          ) : null}
          {line.error ? (
            <button
              type="button"
              class="mt-1 block text-left text-env-prod hover:underline"
              onClick={() => {
                if (line.position !== undefined) {
                  onErrorClick?.(line.position);
                }
              }}
            >
              {line.error}: {line.errorDetail ?? ""}
            </button>
          ) : null}
        </li>
      ))}
    </ul>
  );
}
