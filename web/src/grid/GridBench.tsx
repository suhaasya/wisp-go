import { Grid } from "./Grid";
import type { WindowPayload } from "./window";
import { DEFAULT_WINDOW_ROWS } from "./window";

const TOTAL = 1_000_000;

function syntheticWindow(offset: number): WindowPayload {
  const n = Math.min(DEFAULT_WINDOW_ROWS, TOTAL - offset);
  const ids: number[] = [];
  const names: string[] = [];
  for (let i = 0; i < n; i++) {
    const g = offset + i;
    ids.push(g);
    names.push(`row-${g}`);
  }
  return {
    rid: "bench",
    offset,
    total: { kind: "exact", value: TOTAL },
    columns: [
      { name: "id", type: "int8" },
      { name: "name", type: "text" },
    ],
    data: [ids, names],
  };
}

/** Bench-only grid with 1M synthetic rows (tab memory gate). */
export function GridBench() {
  return (
    <div class="flex h-[600px] min-h-0 flex-col">
      <Grid
        sessionId="bench"
        rid="bench"
        schema="public"
        table="bench"
        columns={[
          { name: "id", type: "int8" },
          { name: "name", type: "text" },
        ]}
        total={{ kind: "exact", value: TOTAL }}
        fetchWindow={async (offset) => syntheticWindow(offset)}
      />
    </div>
  );
}
