import { Handle, Position, type NodeProps } from '@xyflow/react'
import type { FileFlowNode } from '@/types/graph'

export default function FileNode({ data, selected }: NodeProps<FileFlowNode>) {
  return (
    <div
      className={[
        'relative overflow-hidden rounded-[6px] border border-stone-200 bg-white text-left shadow-sm',
        'w-[220px]',
        selected ? 'ring-2 ring-stone-900' : '',
        !selected && data.changed ? 'ring-2 ring-amber-400' : '',
      ]
        .filter(Boolean)
        .join(' ')}
    >
      <div
        className="absolute left-0 top-0 h-full w-[3px]"
        style={{ background: data.clusterColorHex }}
      />
      {data.changed && (
        <div className="absolute right-1.5 top-1.5 size-2 rounded-full bg-amber-400" />
      )}
      <div className="pl-3 pr-4 py-2 space-y-0.5">
        <p className="truncate text-xs font-semibold text-stone-900">{data.fileName}</p>
        <p className="truncate font-mono text-[10px] text-stone-400">{data.packagePath}</p>
        <div className="flex gap-2 font-mono text-[10px] text-stone-400">
          <span>{data.functionCount} fns</span>
          {data.changedCount > 0 && (
            <span className="text-amber-500">{data.changedCount} changed</span>
          )}
        </div>
      </div>
      <Handle type="target" position={Position.Top} style={{ opacity: 0 }} />
      <Handle type="source" position={Position.Bottom} style={{ opacity: 0 }} />
    </div>
  )
}
