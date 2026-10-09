import { Menu, type MenuItem } from "../../ui/Menu";
import { useState } from "preact/hooks";

export type ObjectMenuAction =
  | "open-data"
  | "open-structure"
  | "copy-name"
  | "copy-select"
  | "refresh";

const MENU_ITEMS: MenuItem[] = [
  { id: "open-data", label: "Open data" },
  { id: "open-structure", label: "Open structure" },
  { id: "copy-name", label: "Copy name" },
  { id: "copy-select", label: "Copy SELECT" },
  { id: "refresh", label: "Refresh" },
];

export type ObjectContextMenuProps = {
  open: boolean;
  x: number;
  y: number;
  onClose: () => void;
  onAction: (action: ObjectMenuAction) => void;
};

export function ObjectContextMenu({ open, x, y, onClose, onAction }: ObjectContextMenuProps) {
  const [activeIndex, setActiveIndex] = useState(0);
  return (
    <Menu
      open={open}
      x={x}
      y={y}
      items={MENU_ITEMS}
      activeIndex={activeIndex}
      onActiveIndexChange={setActiveIndex}
      onClose={onClose}
      onSelect={(id) => onAction(id as ObjectMenuAction)}
    />
  );
}
