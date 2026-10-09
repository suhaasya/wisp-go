import { useEffect, useRef } from "preact/hooks";
import type { CompletionSchemaDTO } from "../editor/complete";

const benchSchema: CompletionSchemaDTO = {
  schema: "public",
  tables: [
    {
      name: "users",
      columns: [
        { name: "id", type: "int4" },
        { name: "email", type: "text" },
      ],
    },
  ],
};

/** Playwright: schema completion after FROM and alias dot. */
export function QueryCompleteBench() {
  const hostRef = useRef<HTMLDivElement>(null);
  const viewRef = useRef<import("@codemirror/view").EditorView | null>(null);

  useEffect(() => {
    let destroyed = false;
    void import("../editor/index").then(({ mountSqlEditor }) => {
      if (destroyed || !hostRef.current) {
        return;
      }
      const handle = mountSqlEditor({
        parent: hostRef.current,
        value: "SELECT * FROM ",
        engine: "postgres",
        completion: benchSchema,
        onChange: () => {},
      });
      viewRef.current = handle.view;
    });
    return () => {
      destroyed = true;
      viewRef.current?.destroy();
    };
  }, []);

  return (
    <div class="flex min-h-0 flex-1 flex-col p-4" data-testid="query-complete-bench">
      <div
        ref={hostRef}
        class="h-[250px] border border-line bg-panel"
        data-testid="sql-editor-host"
      />
    </div>
  );
}
