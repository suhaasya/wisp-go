import {
  deriveTags,
  engineAbbrev,
  engineBadgeClass,
  envBorderClass,
  envLabel,
  formatHost,
  formatLastUsed,
  type ConnectionProfile,
} from "../../connections";
import { cx } from "../../ui/cx";

export type ConnectionCardProps = {
  profile: ConnectionProfile;
  tabIndex: number;
  focused: boolean;
  onFocus: () => void;
  onConnect: () => void;
  onContextMenu: (e: MouseEvent) => void;
};

export function ConnectionCard({
  profile,
  tabIndex,
  focused,
  onFocus,
  onConnect,
  onContextMenu,
}: ConnectionCardProps) {
  const tags = deriveTags(profile);
  const lastUsed = profile.env === "production" ? null : formatLastUsed(profile.last_used);
  const showProdPill = profile.env === "production";

  return (
    <button
      type="button"
      tabIndex={tabIndex}
      data-connection-id={profile.id}
      onFocus={onFocus}
      onDblClick={onConnect}
      onKeyDown={(e) => {
        if (e.key === "Enter") {
          e.preventDefault();
          onConnect();
        }
      }}
      onContextMenu={(e) => {
        e.preventDefault();
        onContextMenu(e as unknown as MouseEvent);
      }}
      class={cx(
        "relative w-full cursor-pointer rounded-card border border-line bg-panel py-3 pr-3.5 pl-[18px] text-left",
        "before:absolute before:top-2.5 before:bottom-2.5 before:left-0 before:w-1 before:rounded-r before:content-['']",
        envBorderClass(profile.env),
        "hover:border-ink-3",
        focused && "border-ink-3 ring-2 ring-accent ring-offset-1 ring-offset-panel",
      )}
    >
      <div class="flex items-center gap-2">
        <span
          class={cx(
            "grid h-[26px] w-[26px] flex-none place-items-center rounded-[6px] font-mono text-11-5 font-semibold text-on-accent",
            engineBadgeClass(profile.engine),
          )}
        >
          {engineAbbrev(profile.engine)}
        </span>
        <span class="font-semibold text-ink">{profile.name}</span>
        {showProdPill ? (
          <span class="ml-auto text-11-5 text-ink-2">{envLabel(profile.env)}</span>
        ) : lastUsed ? (
          <span class="ml-auto text-11-5 text-ink-3">{lastUsed}</span>
        ) : null}
      </div>
      <div class="mt-2 truncate font-mono text-11-5 text-ink-2">{formatHost(profile)}</div>
      {tags.length > 0 ? (
        <div class="mt-1.5 flex flex-wrap gap-1.5 text-11-5 text-ink-3">
          {tags.map((tag) => (
            <span key={tag} class="rounded-[5px] border border-line bg-raise px-1.5 text-ink-2">
              {tag}
            </span>
          ))}
        </div>
      ) : null}
    </button>
  );
}
