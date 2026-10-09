import { useEffect, useState } from "preact/hooks";
import { Button } from "../ui/Button";
import { Input } from "../ui/Input";
import { Modal } from "../ui/Modal";
import { previewDDL } from "../workspace/ddl-client";

export function NewSchemaDialog({
  open,
  sessionId,
  engine,
  onClose,
  onCreated,
}: {
  open: boolean;
  sessionId: string;
  engine: string;
  onClose: () => void;
  onCreated: () => void;
}) {
  const isMy = engine === "mysql" || engine === "mariadb";
  const [name, setName] = useState("");
  const [sql, setSQL] = useState("");
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    if (!open) {
      setName("");
      setSQL("");
      return;
    }
    const ac = new AbortController();
    if (name.trim()) {
      void previewDDL(
        sessionId,
        { action: "create_schema", schema: name.trim() },
        { signal: ac.signal },
      )
        .then((r) => setSQL(r.sql))
        .catch(() => setSQL(""));
    } else {
      setSQL("");
    }
    return () => ac.abort();
  }, [open, name, sessionId]);

  return (
    <Modal open={open} title={isMy ? "New database" : "New schema"} onClose={onClose}>
      <Input ariaLabel="Name" value={name} onInput={setName} mono />
      <pre class="mt-3 max-h-[200px] overflow-auto rounded-control border border-line bg-raise p-3 font-mono text-12 whitespace-pre-wrap">
        {sql || "—"}
      </pre>
      <div class="mt-4 flex justify-end gap-2">
        <Button label="Cancel" onClick={onClose} />
        <Button
          label="Create"
          variant="primary"
          disabled={!name.trim() || busy}
          onClick={() => {
            setBusy(true);
            void import("../workspace/ddl-client").then(({ executeDDL }) =>
              executeDDL(sessionId, {
                action: "create_schema",
                schema: name.trim(),
                confirm: true,
              })
                .then(() => {
                  onCreated();
                  onClose();
                })
                .finally(() => setBusy(false)),
            );
          }}
        />
      </div>
    </Modal>
  );
}
