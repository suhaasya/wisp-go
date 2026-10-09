import { Button } from "../ui/Button";
import { Modal } from "../ui/Modal";
import { tokenClass, tokenizeSQL } from "./sql-highlight";

export type ReviewSQLProps = {
  open: boolean;
  title: string;
  subtitle: string;
  sql: string;
  warning?: string;
  onClose: () => void;
  onCommit: () => void;
  committing?: boolean;
};

function SqlHighlighted({ sql }: { sql: string }) {
  const tokens = tokenizeSQL(sql);
  return (
    <>
      {tokens.map((tok, idx) => (
        <span key={idx} class={tokenClass(tok.kind)}>
          {tok.text}
        </span>
      ))}
    </>
  );
}

export function ReviewSQLDialog(props: ReviewSQLProps) {
  return (
    <Modal open={props.open} title={props.title} onClose={props.onClose}>
      <p class="mb-3 text-12-5 text-ink-2">{props.subtitle}</p>
      {props.warning ? (
        <p class="mb-3 rounded-control border border-modified-line bg-modified px-3 py-2 text-12-5 text-ink">
          {props.warning}
        </p>
      ) : null}
      <pre class="max-h-[360px] overflow-auto rounded-control border border-line-2 bg-raise p-3.5 font-mono text-12-5 leading-5 whitespace-pre-wrap">
        <SqlHighlighted sql={props.sql} />
      </pre>
      <div class="mt-3 flex items-center gap-2 border-t border-line-2 pt-3">
        <span class="flex-1 text-12 text-ink-2">Nothing is written until you commit.</span>
        <Button label="Keep editing" onClick={props.onClose} />
        <Button
          label={props.committing ? "Committing…" : "Commit changes"}
          variant="primary"
          disabled={props.committing}
          onClick={props.onCommit}
        />
      </div>
    </Modal>
  );
}
