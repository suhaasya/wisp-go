export type ExplainNode = {
  node: string;
  relation?: string;
  cost?: string;
  rows?: string;
  time?: string;
  depth: number;
  children?: ExplainNode[];
};

type ExplainTreeProps = {
  root: ExplainNode;
};

function ExplainRow({ node }: { node: ExplainNode }) {
  return (
    <>
      <div class="font-mono text-12 text-ink-2" data-testid="explain-row">
        <span class="text-ink">{node.node}</span>
        {node.relation ? <span class="text-ink-3"> · {node.relation}</span> : null}
        {node.cost ? <span class="text-ink-3"> · cost {node.cost}</span> : null}
        {node.rows ? <span class="text-ink-3"> · rows {node.rows}</span> : null}
        {node.time ? <span class="text-ink-3"> · {node.time}</span> : null}
      </div>
      {node.children?.map((ch) => (
        <div
          key={`${node.depth}-${ch.node}-${ch.relation ?? ""}`}
          class="ml-3 border-l border-line pl-2"
        >
          <ExplainRow node={ch} />
        </div>
      ))}
    </>
  );
}

export function ExplainTree({ root }: ExplainTreeProps) {
  return (
    <div class="min-h-0 flex-1 overflow-auto p-3" data-testid="explain-tree">
      <ExplainRow node={root} />
    </div>
  );
}
