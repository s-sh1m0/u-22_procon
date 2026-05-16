import { Handle, Position, type NodeProps } from '@xyflow/react'
import type { FunctionFlowNode } from '@/types/graph'

export default function FunctionNode({ data, selected }: NodeProps<FunctionFlowNode>) {
  const isAdded = data.diffStatus === 'added'
  const isRemoved = data.diffStatus === 'removed'
  return (
    <div
      className={[
        'relative overflow-hidden rounded-[6px] border bg-white text-left shadow-sm',
        'w-[200px]',
        isRemoved ? 'border-dashed border-stone-300 opacity-60 grayscale' : 'border-stone-200',
        selected ? 'ring-2 ring-stone-900' : '',
        !selected && isAdded ? 'ring-2 ring-emerald-500' : '',
        !selected && !isAdded && data.changed ? 'ring-2 ring-amber-400' : '',
      ]
        .filter(Boolean)
        .join(' ')}
    >
      <div
        className="absolute left-0 top-0 h-full w-[3px]"
        style={{ background: data.clusterColorHex }}
      />
      {isAdded && (
        <div
          className="absolute right-1.5 top-1.5 flex size-3.5 items-center justify-center rounded-full bg-emerald-500 text-[9px] font-bold leading-none text-white"
          title="新規追加"
        >
          +
        </div>
      )}
      {isRemoved && (
        <div
          className="absolute right-1.5 top-1.5 flex size-3.5 items-center justify-center rounded-full bg-stone-400 text-[9px] font-bold leading-none text-white"
          title="削除"
        >
          −
        </div>
      )}
      {!isAdded && !isRemoved && data.changed && (
        <div className="absolute right-1.5 top-1.5 size-2 rounded-full bg-amber-400" />
      )}
      <div className="pl-3 pr-4 py-2">
        <p
          className={[
            'truncate text-xs font-semibold',
            isRemoved ? 'text-stone-400 line-through' : 'text-stone-900',
          ].join(' ')}
        >
          {data.label}
        </p>
        <p className="truncate font-mono text-[10px] text-stone-400">{data.packagePath}</p>
        <p className="truncate font-mono text-[10px] text-stone-300">
          {data.file.split('/').pop()}:{data.line}
        </p>
      </div>
      <Handle type="target" position={Position.Top} style={{ opacity: 0 }} />
      <Handle type="source" position={Position.Bottom} style={{ opacity: 0 }} />
    </div>
  )
}
