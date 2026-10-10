import { useCallback, useEffect, useRef, useState } from "preact/hooks";
import { navigate } from "../../router";
import {
  applySettings,
  getSettings,
  loadSettings,
  saveSettings,
  type WispSettings,
} from "../../settings";
import { Button } from "../../ui/Button";
import { pushToast } from "../../ui/Toast";
import { AppearancePane } from "./appearance";
import { EditorPane } from "./editor";
import { GridPane } from "./grid";
import { PrivacyPane } from "./privacy";
import { SafetyPane } from "./safety";
import { AboutPane } from "./about";
import { ShortcutsPane } from "./shortcuts";
import { SETTINGS_SECTIONS, type SettingsSection } from "./types";

export function SettingsScreen() {
  const [section, setSection] = useState<SettingsSection>("appearance");
  const [doc, setDoc] = useState<WispSettings>(() => getSettings());
  const saveTimer = useRef<number | null>(null);

  useEffect(() => {
    void loadSettings()
      .then(setDoc)
      .catch(() => pushToast("Could not load settings", "error"));
    return () => {
      if (saveTimer.current !== null) {
        window.clearTimeout(saveTimer.current);
      }
    };
  }, []);

  const onPatch = useCallback((patch: Partial<WispSettings>) => {
    const next = { ...getSettings(), ...patch, version: getSettings().version ?? 1 };
    applySettings(next);
    setDoc(next);
    if (saveTimer.current !== null) {
      window.clearTimeout(saveTimer.current);
    }
    saveTimer.current = window.setTimeout(() => {
      void saveSettings(next).catch(() => pushToast("Could not save settings", "error"));
    }, 400);
  }, []);

  const onSectionKeyDown = (e: KeyboardEvent) => {
    const idx = SETTINGS_SECTIONS.findIndex((s) => s.id === section);
    if (e.key === "ArrowDown") {
      e.preventDefault();
      const next = SETTINGS_SECTIONS[Math.min(SETTINGS_SECTIONS.length - 1, idx + 1)];
      if (next) {
        setSection(next.id);
      }
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      const next = SETTINGS_SECTIONS[Math.max(0, idx - 1)];
      if (next) {
        setSection(next.id);
      }
    }
  };

  let pane = null;
  switch (section) {
    case "appearance":
      pane = <AppearancePane doc={doc} onPatch={onPatch} />;
      break;
    case "editor":
      pane = <EditorPane doc={doc} onPatch={onPatch} />;
      break;
    case "grid":
      pane = <GridPane doc={doc} onPatch={onPatch} />;
      break;
    case "safety":
      pane = <SafetyPane doc={doc} onPatch={onPatch} />;
      break;
    case "privacy":
      pane = <PrivacyPane doc={doc} onPatch={onPatch} />;
      break;
    case "shortcuts":
      pane = <ShortcutsPane doc={doc} onPatch={onPatch} />;
      break;
    case "about":
      pane = <AboutPane />;
      break;
  }

  return (
    <div class="grid flex-1 place-items-center overflow-auto bg-canvas p-5">
      <div class="w-full max-w-[720px] overflow-hidden rounded-dialog border border-line bg-panel shadow-lg">
        <header class="flex items-center gap-2 border-b border-line px-5 py-3">
          <h1 class="flex-1 text-15 font-semibold text-ink">Settings</h1>
          <Button label="Back" variant="ghost" size="sm" onClick={() => navigate("/connections")} />
        </header>
        <div class="flex min-h-[420px]">
          <nav
            class="flex w-[168px] shrink-0 flex-col gap-0.5 border-r border-line p-2"
            aria-label="Settings sections"
            onKeyDown={onSectionKeyDown as unknown as (e: Event) => void}
          >
            {SETTINGS_SECTIONS.map((s) => {
              const selected = s.id === section;
              return (
                <button
                  key={s.id}
                  type="button"
                  class={`rounded-control border-0 px-2.5 py-2 text-left text-13 ${
                    selected ? "bg-accent-soft font-medium text-ink" : "bg-transparent text-ink-2 hover:text-ink"
                  }`}
                  aria-current={selected ? "true" : undefined}
                  onClick={() => setSection(s.id)}
                >
                  {s.label}
                </button>
              );
            })}
          </nav>
          <div class="min-w-0 flex-1 px-5 py-4" role="region" aria-label={SETTINGS_SECTIONS.find((s) => s.id === section)?.label}>
            {pane}
          </div>
        </div>
      </div>
    </div>
  );
}
