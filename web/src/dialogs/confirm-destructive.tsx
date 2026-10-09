import { useEffect, useState } from "preact/hooks";
import { Button } from "../ui/Button";
import { Input } from "../ui/Input";
import { Modal } from "../ui/Modal";
import { tokenClass, tokenizeSQL } from "./sql-highlight";
import type { DDLDependent } from "../workspace/ddl-client";

export type ConfirmDestructiveProps = {
  open: boolean;
  title: string;
  subtitle: string;
  sql: string;
  confirmName: string;
  needsTypedConfirm: boolean;
  dependents?: DDLDependent[];
  cascade?: boolean;
  cascadeEnabled?: boolean;
  warning?: string;
  onCascadeChange?: (v: boolean) => void;
  onCancel: () => void;
  onConfirm: (typedConfirm: string) => void;
  confirming?: boolean;
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

export function ConfirmDestructiveDialog(props: ConfirmDestructiveProps) {
  const [typed, setTyped] = useState("");
  useEffect(() => {
    if (!props.open) {
      setTyped("");
    }
  }, [props.open]);
  const typedOk = !props.needsTypedConfirm || typed === props.confirmName;
  const canRun = typedOk && !props.confirming;

  return (
    <Modal open={props.open} title={props.title} onClose={props.onCancel}>
      <p class="mb-3 text-12-5 text-ink-2">{props.subtitle}</p>
      {props.warning ? (
        <p class="mb-3 rounded-control border border-modified-line bg-modified px-3 py-2 text-12-5 text-ink">
          {props.warning}
        </p>
      ) : null}
      {props.dependents && props.dependents.length > 0 ? (
        <div class="mb-3 rounded-control border border-line bg-raise px-3 py-2 text-12-5 text-ink">
          <p class="mb-1 font-medium text-ink">Dependent objects</p>
          <ul class="m-0 list-disc pl-4 text-ink-2">
            {props.dependents.map((d) => (
              <li key={`${d.schema}.${d.name}`}>
                {d.kind} {d.schema}.{d.name}
              </li>
            ))}
          </ul>
        </div>
      ) : null}
      {props.cascadeEnabled ? (
        <label class="mb-3 flex items-center gap-2 text-12-5 text-ink">
          <input
            type="checkbox"
            checked={props.cascade === true}
            onChange={(e) => props.onCascadeChange?.((e.target as HTMLInputElement).checked)}
          />
          CASCADE (PostgreSQL only)
        </label>
      ) : null}
      <pre
        class="max-h-[280px] overflow-auto rounded-control border border-line-2 bg-raise p-3.5 font-mono text-12-5 leading-5 whitespace-pre-wrap"
        data-testid="confirm-destructive-sql"
      >
        <SqlHighlighted sql={props.sql} />
      </pre>
      {props.needsTypedConfirm ? (
        <p class="mt-3 text-13 text-ink-2">
          Type <span class="font-mono text-ink">{props.confirmName}</span> to run this statement.
        </p>
      ) : null}
      {props.needsTypedConfirm ? (
        <Input
          mono
          ariaLabel="Object name confirmation"
          value={typed}
          onInput={setTyped}
        />
      ) : null}
      <div class="mt-4 flex justify-end gap-2">
        <Button label="Cancel" onClick={props.onCancel} />
        <button
          type="button"
          data-testid="confirm-destructive-run"
          disabled={!canRun}
          class="rounded-control bg-env-prod px-3 py-1.5 text-13 font-medium text-white disabled:opacity-40"
          onClick={() => props.onConfirm(typed)}
        >
          {props.confirming ? "Running…" : "Run SQL"}
        </button>
      </div>
    </Modal>
  );
}
