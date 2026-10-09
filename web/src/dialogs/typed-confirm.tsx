import { useState } from "preact/hooks";
import { Button } from "../ui/Button";
import { Input } from "../ui/Input";
import { Modal } from "../ui/Modal";

export type TypedConfirmProps = {
  open: boolean;
  tableName: string;
  onCancel: () => void;
  onConfirm: () => void;
};

export function TypedConfirmDialog(props: TypedConfirmProps) {
  const [typed, setTyped] = useState("");
  const ok = typed === props.tableName;
  return (
    <Modal open={props.open} title="Confirm delete" onClose={props.onCancel}>
      <p class="mb-3 text-13 text-ink-2">
        This change set includes row deletes. Type the table name{" "}
        <span class="font-mono text-ink">{props.tableName}</span> to commit.
      </p>
      <Input
        mono
        ariaLabel="Table name confirmation"
        value={typed}
        onInput={setTyped}
      />
      <div class="mt-4 flex justify-end gap-2">
        <Button label="Cancel" onClick={props.onCancel} />
        <Button label="Continue" variant="primary" disabled={!ok} onClick={props.onConfirm} />
      </div>
    </Modal>
  );
}
