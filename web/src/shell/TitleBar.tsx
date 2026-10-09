import { formatBindingDisplay, resolveKeybindings } from "../keys";
import { getSettings } from "../settings";
import { navigate } from "../router";
import { openCommandPalette, shellConnection } from "../state";
import { Button } from "../ui/Button";
import { cx } from "../ui/cx";

const envClass = {
  local: "bg-env-local",
  dev: "bg-env-dev",
  staging: "bg-env-staging",
  prod: "bg-env-prod",
} as const;

export function TitleBar() {
  const conn = shellConnection.value;
  const env = conn?.env ?? null;
  const stripClass = env ? envClass[env] : "bg-transparent";

  return (
    <header class="relative flex h-10 shrink-0 items-center gap-3 border-b border-line bg-sidebar px-3">
      <div
        class={cx("pointer-events-none absolute right-0 bottom-0 left-0 h-0.5", stripClass)}
        aria-hidden
      />
      <div class="flex min-w-0 flex-1 items-center justify-center gap-2 text-12-5">
        {conn ? (
          <>
            {env ? (
              <span
                class={cx(
                  "inline-flex items-center rounded-full px-2 py-0.5 text-11-5 font-semibold text-on-accent",
                  envClass[env],
                )}
              >
                {env}
              </span>
            ) : null}
            <span class="truncate font-semibold text-ink">{conn.name}</span>
            <span class="truncate text-ink-2">{conn.subtitle}</span>
          </>
        ) : (
          <span class="font-semibold text-ink">Wisp</span>
        )}
      </div>
      <div class="ml-auto flex shrink-0 items-center gap-1">
        <Button label="Settings" variant="ghost" size="sm" onClick={() => navigate("/settings")} />
        <Button
          label="Command palette"
          variant="ghost"
          size="sm"
          kbd={formatBindingDisplay(resolveKeybindings(getSettings().keybindings ?? {})["command-palette"])}
          onClick={() => openCommandPalette()}
        />
      </div>
    </header>
  );
}
