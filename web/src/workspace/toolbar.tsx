import { runAction } from "../actions";
import { Button } from "../ui/Button";

export type DataToolbarProps = {
  stagedCount: number;
  disabled?: boolean;
};

export function DataToolbar(props: DataToolbarProps) {
  const disabled = props.disabled === true;
  const n = props.stagedCount;
  return (
    <div class="toolbar flex shrink-0 items-center gap-2 border-b border-line px-2.5 py-2">
      <Button
        label="Add row"
        disabled={disabled}
        onClick={() => runAction("add-row")}
      />
      <Button
        label="Duplicate row"
        disabled={disabled}
        onClick={() => runAction("duplicate-row")}
      />
      <Button
        label="Delete row"
        disabled={disabled}
        onClick={() => runAction("delete-row")}
      />
      <Button
        label="Discard"
        disabled={disabled || n === 0}
        onClick={() => runAction("discard-changes")}
      />
      <Button label="Import" onClick={() => runAction("import-data")} />
      <Button label="Export" onClick={() => runAction("export-data")} />
      <span class="flex-1" />
      <Button
        label={n > 0 ? `Review SQL (${n})` : "Review SQL"}
        variant="warning"
        disabled={disabled || n === 0}
        onClick={() => runAction("review-sql")}
      />
    </div>
  );
}
