import { route } from "./router";
import {
  activeWorkspaceTabId,
  currentSession,
  readOnly,
  workspaceTabs,
  type WorkspaceTab,
} from "./state";
import { closeActiveTab, newQueryTab, activateWorkspaceTab } from "./workspace/tabs";

export type ActionContext = {
  routeName: string;
  sessionId: string | null;
  activeTab: WorkspaceTab | null;
  readOnly: boolean;
};

export type ActionDef = {
  id: string;
  title: string;
  shortcut?: string;
  paletteSection: "actions";
  when: (ctx: ActionContext) => boolean;
  run: (ctx: ActionContext) => void;
};

const runners = new Map<string, (ctx: ActionContext) => void>();

const defs: ActionDef[] = [
  {
    id: "new-query",
    title: "New query tab",
    shortcut: "⌥T",
    paletteSection: "actions",
    when: (ctx) => ctx.routeName === "workspace" && ctx.sessionId !== null,
    run: () => {
      const tab = newQueryTab();
      workspaceTabs.value = [...workspaceTabs.value, tab];
      activateWorkspaceTab(tab.id);
    },
  },
  {
    id: "close-tab",
    title: "Close active tab",
    shortcut: "⌥W",
    paletteSection: "actions",
    when: (ctx) => ctx.routeName === "workspace" && (workspaceTabs.value.length > 0),
    run: () => {
      closeActiveTab();
    },
  },
  {
    id: "export-data",
    title: "Export data",
    paletteSection: "actions",
    when: (ctx) => ctx.activeTab?.kind === "table_data",
    run: (ctx) => {
      runners.get("export-data")?.(ctx);
    },
  },
  {
    id: "import-data",
    title: "Import data",
    paletteSection: "actions",
    when: (ctx) => ctx.activeTab?.kind === "table_data" && !ctx.readOnly,
    run: (ctx) => {
      runners.get("import-data")?.(ctx);
    },
  },
  {
    id: "review-sql",
    title: "Review SQL",
    shortcut: "⌘S",
    paletteSection: "actions",
    when: (ctx) => ctx.activeTab?.kind === "table_data" && !ctx.readOnly,
    run: (ctx) => {
      runners.get("review-sql")?.(ctx);
    },
  },
  {
    id: "discard-changes",
    title: "Discard staged changes",
    paletteSection: "actions",
    when: (ctx) => ctx.activeTab?.kind === "table_data" && !ctx.readOnly,
    run: (ctx) => {
      runners.get("discard-changes")?.(ctx);
    },
  },
  {
    id: "add-row",
    title: "Add row",
    paletteSection: "actions",
    when: (ctx) => ctx.activeTab?.kind === "table_data" && !ctx.readOnly,
    run: (ctx) => {
      runners.get("add-row")?.(ctx);
    },
  },
  {
    id: "duplicate-row",
    title: "Duplicate row",
    paletteSection: "actions",
    when: (ctx) => ctx.activeTab?.kind === "table_data" && !ctx.readOnly,
    run: (ctx) => {
      runners.get("duplicate-row")?.(ctx);
    },
  },
  {
    id: "delete-row",
    title: "Delete row",
    paletteSection: "actions",
    when: (ctx) => ctx.activeTab?.kind === "table_data" && !ctx.readOnly,
    run: (ctx) => {
      runners.get("delete-row")?.(ctx);
    },
  },
];

/** Toolbar and palette share these action ids. */
export const TOOLBAR_ACTION_IDS = [
  "add-row",
  "duplicate-row",
  "delete-row",
  "discard-changes",
  "review-sql",
  "import-data",
  "export-data",
] as const;

export function actionContext(): ActionContext {
  const r = route.value.name;
  const sessionId = currentSession.value;
  const tabs = workspaceTabs.value;
  const activeId = activeWorkspaceTabId.value;
  const activeTab = tabs.find((t) => t.id === activeId) ?? null;
  return {
    routeName: r,
    sessionId,
    activeTab,
    readOnly: readOnly.value,
  };
}

export function listActions(ctx: ActionContext = actionContext()): ActionDef[] {
  return defs.filter((d) => d.when(ctx));
}

export function runAction(id: string, ctx: ActionContext = actionContext()): void {
  const def = defs.find((d) => d.id === id);
  if (!def || !def.when(ctx)) {
    return;
  }
  def.run(ctx);
}

export function registerActionRunner(id: string, fn: (ctx: ActionContext) => void): () => void {
  runners.set(id, fn);
  return () => runners.delete(id);
}

export function allActionDefs(): ActionDef[] {
  return defs.slice();
}
