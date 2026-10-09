import { Checkbox } from "../../ui/Checkbox";
import { Input } from "../../ui/Input";
import type { SettingsPaneProps } from "./types";

export function PrivacyPane({ doc, onPatch }: SettingsPaneProps) {
  return (
    <div class="grid grid-cols-6 gap-x-3.5 gap-y-4">
      <div class="col-span-6">
        <Checkbox
          label="Telemetry"
          checked={doc.telemetry === true}
          onChange={(telemetry) => onPatch({ telemetry })}
        />
        <p class="mt-1 text-12-5 text-ink-2">No telemetry is sent today; this reserves the preference.</p>
      </div>
      <div class="col-span-3">
        <Input
          label="Idle exit (minutes)"
          mono
          value={doc.idle_exit_min !== undefined ? String(doc.idle_exit_min) : ""}
          onInput={(v) => {
            const n = Number.parseInt(v, 10);
            if (!Number.isNaN(n)) {
              onPatch({ idle_exit_min: n });
            }
          }}
        />
      </div>
      <div class="col-span-3">
        <Input
          label="Go memory limit (MB)"
          mono
          value={doc.go_mem_limit_mb !== undefined ? String(doc.go_mem_limit_mb) : ""}
          onInput={(v) => {
            const n = Number.parseInt(v, 10);
            if (!Number.isNaN(n)) {
              onPatch({ go_mem_limit_mb: n });
            }
          }}
        />
      </div>
    </div>
  );
}
