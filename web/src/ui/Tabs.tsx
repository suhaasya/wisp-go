import { cx } from "./cx";

export interface TabItem {
  id: string;
  label: string;
}

export interface TabsProps {
  tabs: TabItem[];
  selectedId: string;
  "aria-label"?: string;
  onSelect: (id: string) => void;
}

export function Tabs({ tabs, selectedId, "aria-label": ariaLabel, onSelect }: TabsProps) {
  return (
    <div class="flex gap-0.5 border-b border-line px-4" role="tablist" aria-label={ariaLabel}>
      {tabs.map((tab) => {
        const selected = tab.id === selectedId;
        return (
          <button
            key={tab.id}
            type="button"
            role="tab"
            id={`tab-${tab.id}`}
            aria-selected={selected}
            aria-controls={`panel-${tab.id}`}
            class={cx(
              "mb-[-1px] border-0 border-b-2 bg-transparent px-2.5 py-2.5 pb-2 text-13",
              selected
                ? "border-accent font-medium text-ink"
                : "border-transparent text-ink-2 hover:text-ink",
            )}
            onClick={() => onSelect(tab.id)}
          >
            {tab.label}
          </button>
        );
      })}
    </div>
  );
}
