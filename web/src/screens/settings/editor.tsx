import { Input } from "../../ui/Input";
import type { SettingsPaneProps } from "./types";

function numField(
  doc: SettingsPaneProps["doc"],
  key: "result_cap_mb" | "keep_alive_sec",
  label: string,
  onPatch: SettingsPaneProps["onPatch"],
) {
  const raw = doc[key];
  const value = typeof raw === "number" ? String(raw) : "";
  return (
    <div class="col-span-3">
      <Input
        label={label}
        mono
        value={value}
        onInput={(v) => {
          const n = Number.parseInt(v, 10);
          if (!Number.isNaN(n)) {
            onPatch({ [key]: n });
          }
        }}
      />
    </div>
  );
}

export function EditorPane({ doc, onPatch }: SettingsPaneProps) {
  return (
    <div class="grid grid-cols-6 gap-x-3.5 gap-y-4">
      {numField(doc, "result_cap_mb", "Result cap (MB)", onPatch)}
      {numField(doc, "keep_alive_sec", "Keep-alive (seconds)", onPatch)}
    </div>
  );
}
