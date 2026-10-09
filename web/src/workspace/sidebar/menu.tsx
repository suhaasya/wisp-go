import { Menu, type MenuItem } from "../../ui/Menu";
import { useMemo, useState } from "preact/hooks";

export type ObjectMenuAction =
  | "open-data"
  | "open-structure"
  | "copy-name"
  | "copy-select"
  | "refresh"
  | "truncate"
  | "drop"
  | "rename";

const BASE_ITEMS: MenuItem[] = [
  { id: "open-data", label: "Open data" },
  { id: "open-structure", label: "Open structure" },
  { id: "copy-name", label: "Copy name" },
  { id: "copy-select", label: "Copy SELECT" },
  { id: "refresh", label: "Refresh" },
];

const DESTRUCTIVE_ITEMS: MenuItem[] = [
  { id: "truncate", label: "Truncate…" },
  { id: "rename", label: "Rename…" },
  { id: "drop", label: "Drop…" },
];

export type ObjectContextMenuProps = {
  open: boolean;
  x: number;
  y: number;
  readOnly?: boolean;
  tablesOnly?: boolean;
  onClose: () => void;
  onAction: (action: ObjectMenuAction) => void;
};

export function ObjectContextMenu({
  open,
  x,
  y,
  readOnly,
  tablesOnly,
  onClose,
  onAction,
}: ObjectContextMenuProps) {
  const [activeIndex, setActiveIndex] = useState(0);
  const items = useMemo(() => {
    if (readOnly || !tablesOnly) {
      return BASE_ITEMS;
    }
    return [...BASE_ITEMS, ...DESTRUCTIVE_ITEMS];
  }, [readOnly, tablesOnly]);
  return (
    <Menu
      open={open}
      x={x}
      y={y}
      items={items}
      activeIndex={activeIndex}
      onActiveIndexChange={setActiveIndex}
      onClose={onClose}
      onSelect={(id) => onAction(id as ObjectMenuAction)}
    />
  );
}
