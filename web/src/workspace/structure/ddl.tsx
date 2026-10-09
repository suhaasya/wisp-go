import { Button } from "../../ui/Button";
import { pushToast } from "../../ui/Toast";

export function StructureDDLPanel({
  ddl,
  loading,
  title = "DDL",
  baselineDDL = "",
}: {
  ddl: string;
  loading: boolean;
  title?: string;
  baselineDDL?: string;
}) {
  const text = ddl || baselineDDL;
  async function onCopy() {
    if (!text) {
      return;
    }
    try {
      await navigator.clipboard.writeText(text);
      pushToast("DDL copied", "success");
    } catch {
      pushToast("Could not copy DDL", "error");
    }
  }

  return (
    <div
      class="flex w-[340px] shrink-0 flex-col self-start overflow-hidden rounded-panel border border-line bg-panel"
      data-testid="structure-ddl-panel"
    >
      <h5
        class="flex items-center justify-between gap-2 border-b border-line-2 px-3 py-2.5 text-12-5 font-semibold text-ink"
        data-testid="structure-ddl-title"
      >
        {title}
        <Button label="Copy" disabled={!text || loading} onClick={() => void onCopy()} />
      </h5>
      <pre class="max-h-[min(70vh,640px)] overflow-auto p-3 font-mono text-12 leading-[19px] whitespace-pre-wrap text-ink">
        {loading ? "Loading…" : text || "—"}
      </pre>
    </div>
  );
}
