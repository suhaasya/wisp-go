import { Segmented } from "../../ui/Segmented";
import type { SettingsPaneProps } from "./types";

export function AppearancePane({ doc, onPatch }: SettingsPaneProps) {
  const theme = doc.theme === "dark" || doc.theme === "light" ? doc.theme : "system";
  const density = doc.density === "comfortable" ? "comfortable" : "compact";

  return (
    <div class="grid grid-cols-6 gap-x-3.5 gap-y-4">
      <div class="col-span-6">
        <span class="mb-1 block text-12-5 font-medium text-ink-2">Theme</span>
        <Segmented
          aria-label="Theme"
          options={[
            { value: "light", label: "Light" },
            { value: "dark", label: "Dark" },
            { value: "system", label: "System" },
          ]}
          value={theme}
          onChange={(theme) => onPatch({ theme: theme as "light" | "dark" | "system" })}
        />
      </div>
      <div class="col-span-6">
        <span class="mb-1 block text-12-5 font-medium text-ink-2">Density</span>
        <Segmented
          aria-label="Density"
          options={[
            { value: "compact", label: "Compact" },
            { value: "comfortable", label: "Comfortable" },
          ]}
          value={density}
          onChange={(density) => onPatch({ density: density as "compact" | "comfortable" })}
        />
      </div>
    </div>
  );
}
