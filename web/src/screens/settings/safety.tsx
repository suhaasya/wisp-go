import { Checkbox } from "../../ui/Checkbox";
import type { SettingsPaneProps } from "./types";

export function SafetyPane({ doc, onPatch }: SettingsPaneProps) {
  const on = doc.confirm_writes_on_production !== false;
  return (
    <div class="flex flex-col gap-3">
      <Checkbox
        label="Confirm writes on production connections"
        checked={on}
        onChange={(confirm_writes_on_production) => onPatch({ confirm_writes_on_production })}
      />
      <p class="text-12-5 text-ink-2">
        When enabled, destructive or write operations on production profiles require extra
        confirmation.
      </p>
    </div>
  );
}
