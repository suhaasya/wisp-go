import type { WispSettings } from "../../settings";

export type SettingsSection =
  | "appearance"
  | "editor"
  | "grid"
  | "safety"
  | "privacy"
  | "shortcuts"
  | "about";

export type SettingsPaneProps = {
  doc: WispSettings;
  onPatch: (patch: Partial<WispSettings>) => void;
};

export const SETTINGS_SECTIONS: { id: SettingsSection; label: string }[] = [
  { id: "appearance", label: "Appearance" },
  { id: "editor", label: "Editor" },
  { id: "grid", label: "Grid" },
  { id: "safety", label: "Data safety" },
  { id: "privacy", label: "Privacy" },
  { id: "shortcuts", label: "Shortcuts" },
  { id: "about", label: "About & feedback" },
];
