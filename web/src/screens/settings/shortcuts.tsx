import {
  BINDING_IDS,
  BINDING_LABELS,
  DEFAULT_KEYBINDINGS,
  findKeybindingConflicts,
  formatBindingDisplay,
  isBrowserReservedBinding,
  parseBinding,
  resolveKeybindings,
  type BindingId,
} from "../../keys";
import { Input } from "../../ui/Input";
import type { SettingsPaneProps } from "./types";

export function ShortcutsPane({ doc, onPatch }: SettingsPaneProps) {
  const overrides = { ...(doc.keybindings ?? {}) };
  const resolved = resolveKeybindings(overrides);
  const conflicts = findKeybindingConflicts(resolved);

  function setOverride(id: BindingId, raw: string) {
    const trimmed = raw.trim();
    const next = { ...overrides };
    if (!trimmed || trimmed === DEFAULT_KEYBINDINGS[id]) {
      delete next[id];
    } else {
      const parsed = parseBinding(trimmed);
      if (!parsed || isBrowserReservedBinding(trimmed)) {
        return;
      }
      next[id] = trimmed;
    }
    onPatch({ keybindings: next });
  }

  return (
    <div class="flex flex-col gap-3">
      {conflicts.length > 0 ? (
        <p class="rounded-control border border-env-prod bg-deleted/30 px-3 py-2 text-12-5 text-ink" role="alert">
          Conflicting shortcuts:{" "}
          {conflicts.map((c) => `${c.chord} (${c.ids.map((id) => BINDING_LABELS[id]).join(", ")})`).join("; ")}
        </p>
      ) : null}
      <ul class="list-none divide-y divide-line rounded-dialog border border-line">
        {BINDING_IDS.filter((id) => !id.startsWith("switch-tab-") || id === "switch-tab-1").map((id) => (
          <li key={id} class="flex flex-wrap items-center gap-2 px-3 py-2">
            <span class="min-w-[140px] flex-1 text-13 text-ink">
              {id === "switch-tab-1" ? "Switch to tab 1–9" : BINDING_LABELS[id]}
            </span>
            <span class="text-12 text-ink-3">{formatBindingDisplay(resolved[id])}</span>
            <div class="w-[160px]">
              <Input
                mono
                ariaLabel={`Shortcut for ${BINDING_LABELS[id]}`}
                placeholder={DEFAULT_KEYBINDINGS[id]}
                value={overrides[id] ?? ""}
                onInput={(v) => setOverride(id, v)}
              />
            </div>
          </li>
        ))}
      </ul>
      <p class="text-12-5 text-ink-2">
        Use Mod for ⌘/Ctrl, Alt, Shift, and key names like Enter or Escape (e.g. Mod-K, Alt-T).
        Browser shortcuts Mod+W, Mod+T, Mod+N, Mod+L, and Mod+R cannot be assigned.
      </p>
    </div>
  );
}
