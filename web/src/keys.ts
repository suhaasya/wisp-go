/** Stable ids for user-configurable shortcuts (stored in settings.keybindings). */
export type BindingId =
  | "command-palette"
  | "review-sql"
  | "run-statement"
  | "run-all"
  | "cancel-query"
  | "new-query-tab"
  | "close-tab"
  | "new-connection"
  | "switch-tab-1"
  | "switch-tab-2"
  | "switch-tab-3"
  | "switch-tab-4"
  | "switch-tab-5"
  | "switch-tab-6"
  | "switch-tab-7"
  | "switch-tab-8"
  | "switch-tab-9"
  | "shortcut-help";

export const BINDING_IDS: BindingId[] = [
  "command-palette",
  "review-sql",
  "run-statement",
  "run-all",
  "cancel-query",
  "new-query-tab",
  "close-tab",
  "new-connection",
  "switch-tab-1",
  "switch-tab-2",
  "switch-tab-3",
  "switch-tab-4",
  "switch-tab-5",
  "switch-tab-6",
  "switch-tab-7",
  "switch-tab-8",
  "switch-tab-9",
  "shortcut-help",
];

export const BINDING_LABELS: Record<BindingId, string> = {
  "command-palette": "Command palette",
  "review-sql": "Review SQL",
  "run-statement": "Run statement",
  "run-all": "Run all",
  "cancel-query": "Cancel query",
  "new-query-tab": "New query tab",
  "close-tab": "Close tab",
  "new-connection": "New connection",
  "switch-tab-1": "Switch to tab 1",
  "switch-tab-2": "Switch to tab 2",
  "switch-tab-3": "Switch to tab 3",
  "switch-tab-4": "Switch to tab 4",
  "switch-tab-5": "Switch to tab 5",
  "switch-tab-6": "Switch to tab 6",
  "switch-tab-7": "Switch to tab 7",
  "switch-tab-8": "Switch to tab 8",
  "switch-tab-9": "Switch to tab 9",
  "shortcut-help": "Keyboard shortcuts",
};

export const DEFAULT_KEYBINDINGS: Record<BindingId, string> = {
  "command-palette": "Mod-K",
  "review-sql": "Mod-S",
  "run-statement": "Mod-Enter",
  "run-all": "Shift-Mod-Enter",
  "cancel-query": "Escape",
  "new-query-tab": "Alt-T",
  "close-tab": "Alt-W",
  "new-connection": "Alt-N",
  "switch-tab-1": "Alt-1",
  "switch-tab-2": "Alt-2",
  "switch-tab-3": "Alt-3",
  "switch-tab-4": "Alt-4",
  "switch-tab-5": "Alt-5",
  "switch-tab-6": "Alt-6",
  "switch-tab-7": "Alt-7",
  "switch-tab-8": "Alt-8",
  "switch-tab-9": "Alt-9",
  "shortcut-help": "Mod-/",
};

export type ParsedBinding = {
  mod: boolean;
  alt: boolean;
  shift: boolean;
  key: string;
};

const RESERVED_MOD_KEYS = new Set(["w", "t", "n", "l", "r"]);

export function isMacPlatform(): boolean {
  if (typeof navigator === "undefined") {
    return false;
  }
  return /Mac|iPhone|iPod|iPad/i.test(navigator.platform);
}

export function parseBinding(raw: string): ParsedBinding | null {
  const t = raw.trim();
  if (!t) {
    return null;
  }
  const parts = t.split("-");
  let mod = false;
  let alt = false;
  let shift = false;
  let keyPart = "";
  for (const p of parts) {
    const u = p.trim();
    if (!u) {
      return null;
    }
    const lower = u.toLowerCase();
    if (lower === "mod") {
      mod = true;
      continue;
    }
    if (lower === "alt") {
      alt = true;
      continue;
    }
    if (lower === "shift") {
      shift = true;
      continue;
    }
    if (keyPart) {
      return null;
    }
    keyPart = u;
  }
  if (!keyPart) {
    return null;
  }
  const key = normalizeKeyName(keyPart);
  if (!key) {
    return null;
  }
  return { mod, alt, shift, key };
}

function normalizeKeyName(part: string): string | null {
  const k = part.length === 1 ? part.toLowerCase() : part.toLowerCase();
  if (k === "/") {
    return "/";
  }
  if (k === "enter") {
    return "enter";
  }
  if (k === "escape" || k === "esc") {
    return "escape";
  }
  if (/^[1-9a-z]$/.test(k)) {
    return k;
  }
  return null;
}

export function formatBinding(b: ParsedBinding): string {
  const parts: string[] = [];
  if (b.shift) {
    parts.push("Shift");
  }
  if (b.mod) {
    parts.push("Mod");
  }
  if (b.alt) {
    parts.push("Alt");
  }
  const keyLabel =
    b.key === "enter" ? "Enter" : b.key === "escape" ? "Escape" : b.key === "/" ? "/" : b.key.toUpperCase();
  parts.push(keyLabel);
  return parts.join("-");
}

export function isBrowserReservedBinding(raw: string): boolean {
  const p = parseBinding(raw);
  if (!p || !p.mod || p.alt || p.shift) {
    return false;
  }
  return RESERVED_MOD_KEYS.has(p.key);
}

export function resolveKeybindings(overrides: Record<string, string> = {}): Record<BindingId, string> {
  const out = { ...DEFAULT_KEYBINDINGS };
  for (const id of BINDING_IDS) {
    const custom = overrides[id];
    if (!custom) {
      continue;
    }
    const parsed = parseBinding(custom);
    if (!parsed || isBrowserReservedBinding(custom)) {
      continue;
    }
    out[id] = formatBinding(parsed);
  }
  return out;
}

export type KeybindingConflict = { chord: string; ids: BindingId[] };

export function findKeybindingConflicts(resolved: Record<BindingId, string>): KeybindingConflict[] {
  const byChord = new Map<string, BindingId[]>();
  for (const id of BINDING_IDS) {
    const chord = resolved[id];
    const list = byChord.get(chord) ?? [];
    list.push(id);
    byChord.set(chord, list);
  }
  const conflicts: KeybindingConflict[] = [];
  for (const [chord, ids] of byChord) {
    if (ids.length > 1) {
      conflicts.push({ chord, ids });
    }
  }
  return conflicts;
}

export function formatBindingDisplay(raw: string, opts?: { mac?: boolean }): string {
  const p = parseBinding(raw);
  if (!p) {
    return raw;
  }
  const mac = opts?.mac ?? isMacPlatform();
  if (mac) {
    let s = "";
    if (p.shift) {
      s += "⇧";
    }
    if (p.mod) {
      s += "⌘";
    }
    if (p.alt) {
      s += "⌥";
    }
    if (p.key === "enter") {
      return `${s}↩`;
    }
    if (p.key === "escape") {
      return `${s}Esc`;
    }
    if (p.key === "/") {
      return `${s}/`;
    }
    return s + p.key.toUpperCase();
  }
  const parts: string[] = [];
  if (p.shift) {
    parts.push("Shift");
  }
  if (p.mod) {
    parts.push("Ctrl");
  }
  if (p.alt) {
    parts.push("Alt");
  }
  if (p.key === "enter") {
    parts.push("Enter");
  } else if (p.key === "escape") {
    parts.push("Esc");
  } else if (p.key === "/") {
    parts.push("/");
  } else {
    parts.push(p.key.toUpperCase());
  }
  return parts.join("+");
}

/** Physical key for Alt chords (macOS Option often changes e.key, not e.code). */
export function bindingKeyToKeyboardCode(key: string): string | null {
  if (key === "enter") {
    return "Enter";
  }
  if (key === "escape") {
    return "Escape";
  }
  if (key === "/") {
    return "Slash";
  }
  if (key.length === 1 && /[a-z]/i.test(key)) {
    return `Key${key.toUpperCase()}`;
  }
  if (key.length === 1 && /[0-9]/.test(key)) {
    return `Digit${key}`;
  }
  return null;
}

export function eventMatchesBinding(e: KeyboardEvent, binding: string): boolean {
  const p = parseBinding(binding);
  if (!p) {
    return false;
  }
  const mod = e.metaKey || e.ctrlKey;
  if (p.mod !== mod) {
    return false;
  }
  if (p.alt !== e.altKey) {
    return false;
  }
  if (p.shift !== e.shiftKey) {
    return false;
  }
  const ek = e.key.length === 1 ? e.key.toLowerCase() : e.key.toLowerCase();
  if (p.key === "enter") {
    return ek === "enter";
  }
  if (p.key === "escape") {
    return ek === "escape";
  }
  if (p.key === "/") {
    return e.key === "/" || e.key === "?";
  }
  if (ek === p.key) {
    return true;
  }
  if (p.alt && !p.mod) {
    const code = bindingKeyToKeyboardCode(p.key);
    if (code && e.code === code) {
      return true;
    }
  }
  return false;
}

/** CodeMirror 6 key names for run/cancel extensions. */
export function bindingToCodeMirrorKey(binding: string): string | null {
  const p = parseBinding(binding);
  if (!p) {
    return null;
  }
  const parts: string[] = [];
  if (p.shift) {
    parts.push("Shift");
  }
  if (p.mod) {
    parts.push("Mod");
  }
  if (p.key === "enter") {
    parts.push("Enter");
    return parts.join("-");
  }
  if (p.key === "escape") {
    return "Escape";
  }
  return null;
}

export function editorRunKeyBindings(resolved: Record<BindingId, string>): {
  runStatement: string;
  runAll: string;
  cancel: string;
} {
  return {
    runStatement: bindingToCodeMirrorKey(resolved["run-statement"]) ?? "Mod-Enter",
    runAll: bindingToCodeMirrorKey(resolved["run-all"]) ?? "Shift-Mod-Enter",
    cancel: bindingToCodeMirrorKey(resolved["cancel-query"]) ?? "Escape",
  };
}

export function isTypingTarget(target: EventTarget | null): boolean {
  const el = target as HTMLElement | null;
  if (!el) {
    return false;
  }
  const tag = el.tagName;
  if (tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT") {
    return true;
  }
  return el.isContentEditable;
}
