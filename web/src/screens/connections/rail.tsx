import { envLabel, type ProfileEnv } from "../../connections";
import { cx } from "../../ui/cx";

const envOrder: ProfileEnv[] = ["production", "staging", "development", "local"];

function envDotClass(env: ProfileEnv): string {
  switch (env) {
    case "local":
      return "bg-env-local";
    case "development":
      return "bg-env-dev";
    case "staging":
      return "bg-env-staging";
    case "production":
      return "bg-env-prod";
  }
}

export type ConnectionsRailProps = {
  folders: string[];
  folderCounts: Map<string, number>;
  envCounts: Map<ProfileEnv, number>;
  activeFolder: string | "all";
  activeEnv: ProfileEnv | "all";
  onFolder: (folder: string | "all") => void;
  onEnv: (env: ProfileEnv | "all") => void;
};

export function ConnectionsRail({
  folders,
  folderCounts,
  envCounts,
  activeFolder,
  activeEnv,
  onFolder,
  onEnv,
}: ConnectionsRailProps) {
  return (
    <nav
      class="flex w-[210px] flex-none flex-col gap-0.5 border-r border-line bg-sidebar px-2.5 py-3.5"
      aria-label="Connection folders"
    >
      <h3 class="mx-2 mb-1.5 mt-2.5 text-12-5 font-medium text-ink-3">Folders</h3>
      <RailLink
        label="All connections"
        count={folderCounts.get("all") ?? 0}
        active={activeFolder === "all"}
        onClick={() => onFolder("all")}
      />
      {folders.map((folder) => (
        <RailLink
          key={folder}
          label={folder}
          count={folderCounts.get(folder) ?? 0}
          active={activeFolder === folder}
          onClick={() => onFolder(folder)}
        />
      ))}
      <h3 class="mx-2 mb-1.5 mt-2.5 text-12-5 font-medium text-ink-3">Environment</h3>
      {envOrder.map((env) => (
        <button
          key={env}
          type="button"
          onClick={() => onEnv(activeEnv === env ? "all" : env)}
          class={cx(
            "flex w-full items-center justify-between rounded-[6px] px-2 py-1.5 text-left text-13 text-ink",
            activeEnv === env && "bg-accent-soft font-medium",
          )}
        >
          <span class="flex items-center gap-2">
            <span class={cx("h-2 w-2 rounded-full", envDotClass(env))} aria-hidden="true" />
            {envLabel(env)}
          </span>
          <span class="text-ink-3">{envCounts.get(env) ?? 0}</span>
        </button>
      ))}
    </nav>
  );
}

function RailLink({
  label,
  count,
  active,
  onClick,
}: {
  label: string;
  count: number;
  active: boolean;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      class={cx(
        "flex w-full items-center justify-between rounded-[6px] px-2 py-1.5 text-left text-13 text-ink",
        active && "bg-accent-soft font-medium",
      )}
    >
      <span>{label}</span>
      <span class="text-ink-3">{count}</span>
    </button>
  );
}
