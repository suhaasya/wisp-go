import { Button } from "../../ui/Button";

export type ConnectionsEmptyProps = {
  filtered: boolean;
  onNewConnection: () => void;
};

export function ConnectionsEmpty({ filtered, onNewConnection }: ConnectionsEmptyProps) {
  return (
    <div class="flex flex-col items-center justify-center gap-3 py-16 text-center">
      <h2 class="text-15 font-semibold text-ink">
        {filtered ? "No connections match your filters" : "No saved connections yet"}
      </h2>
      <p class="max-w-md text-13 text-ink-2">
        {filtered
          ? "Try clearing search or changing folder and environment filters."
          : "Create a connection or paste a URL anywhere to import one (coming soon)."}
      </p>
      {!filtered ? (
        <Button label="New connection" variant="primary" kbd="Alt+N" onClick={onNewConnection} />
      ) : null}
      {!filtered ? (
        <p class="text-12-5 text-ink-3">
          Example:{" "}
          <code class="font-mono text-11-5 text-ink-2">
            postgres://app@db.example.com:5432/shop?sslmode=require
          </code>
        </p>
      ) : null}
    </div>
  );
}
