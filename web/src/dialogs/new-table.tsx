import { useEffect, useState } from "preact/hooks";
import { Button } from "../ui/Button";
import { Input } from "../ui/Input";
import { Modal } from "../ui/Modal";
import { previewDDL, type DDLPreviewBody } from "../workspace/ddl-client";

const PG_TYPES = ["bigint", "text", "boolean", "timestamptz", "jsonb", "uuid"];
const MY_TYPES = ["bigint", "varchar(255)", "text", "datetime", "json", "int"];

type Col = {
  name: string;
  type: string;
  nullable: boolean;
  default: string;
  primary: boolean;
};

export function NewTableDialog({
  open,
  sessionId,
  schema,
  engine,
  onClose,
  onCreated,
}: {
  open: boolean;
  sessionId: string;
  schema: string;
  engine: string;
  onClose: () => void;
  onCreated: (table: string) => void;
}) {
  const types = engine === "mysql" || engine === "mariadb" ? MY_TYPES : PG_TYPES;
  const [tableName, setTableName] = useState("");
  const [cols, setCols] = useState<Col[]>([
    { name: "id", type: types[0]!, nullable: false, default: "", primary: true },
  ]);
  const [sql, setSQL] = useState("");
  const [busy, setBusy] = useState(false);

  const body = (): DDLPreviewBody | null => {
    if (!tableName.trim()) {
      return null;
    }
    return {
      action: "create_table",
      schema,
      wizard: {
        table: tableName.trim(),
        columns: cols.map((c) => ({
          name: c.name,
          type: c.type,
          nullable: c.nullable,
          default: c.default || undefined,
          primary: c.primary,
        })),
      },
    };
  };

  useEffect(() => {
    if (!open) {
      return;
    }
    const b = body();
    if (!b) {
      setSQL("");
      return;
    }
    const ac = new AbortController();
    void previewDDL(sessionId, b, { signal: ac.signal })
      .then((r) => setSQL(r.sql))
      .catch(() => setSQL(""));
    return () => ac.abort();
  }, [open, tableName, cols, sessionId, schema]);

  function addColumn() {
    setCols([...cols, { name: "", type: types[0]!, nullable: true, default: "", primary: false }]);
  }

  return (
    <Modal open={open} title="New table" onClose={onClose}>
      <Input ariaLabel="Table name" value={tableName} onInput={setTableName} mono />
      <div class="mt-3 overflow-auto rounded-control border border-line">
        <table class="w-full border-collapse text-12-5">
          <thead>
            <tr class="bg-raise text-12 text-ink-2">
              <th class="px-2 py-1 text-left">Name</th>
              <th class="px-2 py-1 text-left">Type</th>
              <th class="px-2 py-1">PK</th>
            </tr>
          </thead>
          <tbody>
            {cols.map((c, i) => (
              <tr key={i} class="border-t border-line-2">
                <td class="p-1">
                  <input
                    class="input w-full font-mono text-12"
                    value={c.name}
                    onInput={(e) => {
                      const next = [...cols];
                      next[i] = { ...c, name: (e.target as HTMLInputElement).value };
                      setCols(next);
                    }}
                  />
                </td>
                <td class="p-1">
                  <select
                    class="input w-full font-mono text-12"
                    value={c.type}
                    onChange={(e) => {
                      const next = [...cols];
                      next[i] = { ...c, type: (e.target as HTMLSelectElement).value };
                      setCols(next);
                    }}
                  >
                    {types.map((t) => (
                      <option key={t} value={t}>
                        {t}
                      </option>
                    ))}
                  </select>
                </td>
                <td class="p-1 text-center">
                  <input
                    type="checkbox"
                    checked={c.primary}
                    onChange={(e) => {
                      const next = [...cols];
                      next[i] = { ...c, primary: (e.target as HTMLInputElement).checked };
                      setCols(next);
                    }}
                  />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <Button label="Add column" onClick={addColumn} />
      <pre
        class="mt-3 max-h-[180px] overflow-auto rounded-control border border-line bg-raise p-3 font-mono text-12 whitespace-pre-wrap"
        data-testid="new-table-preview"
      >
        {sql || "—"}
      </pre>
      <div class="mt-4 flex justify-end gap-2">
        <Button label="Cancel" onClick={onClose} />
        <Button
          label="Create"
          variant="primary"
          disabled={!tableName.trim() || busy}
          onClick={() => {
            const b = body();
            if (!b) {
              return;
            }
            setBusy(true);
            void import("../workspace/ddl-client").then(({ executeDDL }) =>
              executeDDL(sessionId, { ...b, confirm: true })
                .then(() => {
                  onCreated(tableName.trim());
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
