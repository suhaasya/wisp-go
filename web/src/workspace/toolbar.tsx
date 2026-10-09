import { Button } from "../ui/Button";

export type DataToolbarProps = {
  stagedCount: number;
  disabled?: boolean;
  onAddRow: () => void;
  onDuplicateRow: () => void;
  onDeleteRow: () => void;
  onDiscard: () => void;
  onReviewSQL: () => void;
  onExport?: () => void;
};

export function DataToolbar(props: DataToolbarProps) {
  const disabled = props.disabled === true;
  const n = props.stagedCount;
  return (
    <div class="toolbar flex shrink-0 items-center gap-2 border-b border-line px-2.5 py-2">
      <Button label="Add row" disabled={disabled} onClick={props.onAddRow} />
      <Button label="Duplicate row" disabled={disabled} onClick={props.onDuplicateRow} />
      <Button label="Delete row" disabled={disabled} onClick={props.onDeleteRow} />
      <Button label="Discard" disabled={disabled || n === 0} onClick={props.onDiscard} />
      <Button label="Export" onClick={() => props.onExport?.()} />
      <span class="flex-1" />
      <Button
        label={n > 0 ? `Review SQL (${n})` : "Review SQL"}
        variant="warning"
        disabled={disabled || n === 0}
        onClick={props.onReviewSQL}
      />
    </div>
  );
}
