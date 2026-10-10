import { actionContext, listActions, runAction } from "./actions";
import {
  type BindingId,
  BINDING_LABELS,
  eventMatchesBinding,
  formatBindingDisplay,
  isTypingTarget,
  resolveKeybindings,
} from "./keys";
import { navigate, route } from "./router";
import { openCommandPalette, toggleShortcutHelp, workspaceTabs } from "./state";
import { getSettings } from "./settings";
import { activateTabAtIndex, closeActiveTab, newQueryTab, activateWorkspaceTab } from "./workspace/tabs";
export type QueryKeyRunners = {
  runStatement?: () => void;
  runAll?: () => void;
  cancel?: () => void;
};

let queryRunners: QueryKeyRunners = {};

export function registerQueryKeyRunners(next: QueryKeyRunners): () => void {
  queryRunners = { ...queryRunners, ...next };
  return () => {
    for (const k of Object.keys(next) as (keyof QueryKeyRunners)[]) {
      delete queryRunners[k];
    }
  };
}

function resolved(): Record<BindingId, string> {
  return resolveKeybindings(getSettings().keybindings ?? {});
}

export function handleGlobalKeydown(e: KeyboardEvent): void {
  const map = resolved();
  const ctx = actionContext();
  const typing = isTypingTarget(e.target);

  const tryBinding = (
    id: BindingId,
    run: () => void,
    opts?: { allowInField?: boolean; when?: () => boolean },
  ) => {
    if (!eventMatchesBinding(e, map[id])) {
      return false;
    }
    if (opts?.when && !opts.when()) {
      return false;
    }
    if (typing && !opts?.allowInField) {
      return false;
    }
    e.preventDefault();
    run();
    return true;
  };

  if (
    tryBinding(
      "shortcut-help",
      () => {
        toggleShortcutHelp();
      },
      { allowInField: true },
    )
  ) {
    return;
  }

  if (
    tryBinding(
      "command-palette",
      () => {
        openCommandPalette();
      },
      { allowInField: true },
    )
  ) {
    return;
  }

  if (
    tryBinding(
      "review-sql",
      () => {
        runAction("review-sql", ctx);
      },
      { when: () => listActions(ctx).some((a) => a.id === "review-sql") },
    )
  ) {
    return;
  }

  if (ctx.routeName === "workspace" && ctx.sessionId) {
    for (let i = 1; i <= 9; i++) {
      const id = `switch-tab-${i}` as BindingId;
      if (
        tryBinding(id, () => {
          activateTabAtIndex(i - 1);
        })
      ) {
        return;
      }
    }
    if (
      tryBinding("new-query-tab", () => {
        const tab = newQueryTab();
        workspaceTabs.value = [...workspaceTabs.value, tab];
        activateWorkspaceTab(tab.id);
      })
    ) {
      return;
    }
    if (tryBinding("close-tab", () => closeActiveTab())) {
      return;
    }
  }

  if (route.value.name === "connections") {
    if (
      tryBinding("new-connection", () => {
        navigate("/connections/new");
      })
    ) {
      return;
    }
  }

  if (ctx.activeTab?.kind === "query") {
    if (tryBinding("run-statement", () => queryRunners.runStatement?.())) {
      return;
    }
    if (tryBinding("run-all", () => queryRunners.runAll?.())) {
      return;
    }
    if (tryBinding("cancel-query", () => queryRunners.cancel?.(), { allowInField: true })) {
      return;
    }
  }

  // PWA tab cycle (unchanged from workspace)
  if (
    typeof window !== "undefined" &&
    window.matchMedia("(display-mode: standalone)").matches &&
    e.ctrlKey &&
    e.key === "Tab"
  ) {
    e.preventDefault();
    const tabs = workspaceTabs.value;
    const activeId = ctx.activeTab?.id;
    const idx = tabs.findIndex((t) => t.id === activeId);
    const next = idx < 0 ? 0 : (idx + (e.shiftKey ? -1 : 1) + tabs.length) % tabs.length;
    activateTabAtIndex(next);
  }
}

/** Ordered ids for cheat sheet (hide tab 2-9 duplicates in display — show pattern). */
export function cheatSheetBindings(): { id: BindingId; label: string; chord: string }[] {
  const map = resolved();
  const show: BindingId[] = [
    "command-palette",
    "shortcut-help",
    "new-connection",
    "new-query-tab",
    "close-tab",
    "switch-tab-1",
    "review-sql",
    "run-statement",
    "run-all",
    "cancel-query",
  ];
  return show.map((id) => ({
    id,
    label: BINDING_LABELS[id],
    chord: formatBindingDisplay(map[id]),
  }));
}
