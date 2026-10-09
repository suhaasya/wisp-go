import { Input } from "../../ui/Input";
import type { SettingsPaneProps } from "./types";

function intInput(
  label: string,
  value: number | undefined,
  onChange: (n: number) => void,
) {
  return (
    <div class="col-span-3">
      <Input
        label={label}
        mono
        value={value !== undefined ? String(value) : ""}
        onInput={(v) => {
          const n = Number.parseInt(v, 10);
          if (!Number.isNaN(n)) {
            onChange(n);
          }
        }}
      />
    </div>
  );
}

export function GridPane({ doc, onPatch }: SettingsPaneProps) {
  return (
    <div class="grid grid-cols-6 gap-x-3.5 gap-y-4">
      {intInput("Rows per window", doc.window_rows, (window_rows) => onPatch({ window_rows }))}
      {intInput("Windows in memory", doc.windows_in_memory, (windows_in_memory) =>
        onPatch({ windows_in_memory }),
      )}
    </div>
  );
}
