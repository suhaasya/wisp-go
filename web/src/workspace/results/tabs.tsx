export type QueryPanelTab =
  | { id: string; kind: "result"; rid: string; label: string; rowsLabel?: string; ms?: number }
  | { id: "messages"; kind: "messages" }
  | { id: "explain"; kind: "explain" };

type ResultTabsProps = {
  tabs: QueryPanelTab[];
  activeId: string;
  meta?: string;
  onSelect: (id: string) => void;
};

export function ResultTabs({ tabs, activeId, meta, onSelect }: ResultTabsProps) {
  return (
    <div
      class="rtabs flex h-[34px] shrink-0 items-center gap-0.5 border-b border-line px-2.5"
      role="tablist"
    >
      {tabs.map((tab) => {
        const label =
          tab.kind === "result"
            ? tab.label
            : tab.kind === "messages"
              ? "Messages"
              : "Explain";
        return (
          <button
            key={tab.id}
            type="button"
            role="tab"
            aria-selected={activeId === tab.id}
            class={`border-0 bg-transparent px-2.5 py-2 text-13 ${
              activeId === tab.id
                ? "border-b-2 border-accent font-medium text-ink"
                : "border-b-2 border-transparent text-ink-2"
            }`}
            onClick={() => onSelect(tab.id)}
          >
            {label}
          </button>
        );
      })}
      {meta ? <span class="meta ml-auto text-11-5 text-ink-3">{meta}</span> : null}
    </div>
  );
}
