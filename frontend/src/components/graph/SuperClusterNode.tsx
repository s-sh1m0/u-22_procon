import { Handle, Position, type NodeProps } from '@xyflow/react'
import type { SuperClusterFlowNode } from '@/types/graph'

export default function SuperClusterNode({ data }: NodeProps<SuperClusterFlowNode>) {
  return (
    <div
      className={[
        'relative rounded-xl border cursor-pointer transition-shadow hover:shadow-md',
        'w-[240px]',
        data.hasChanged ? 'ring-2 ring-amber-400' : '',
      ]
        .filter(Boolean)
        .join(' ')}
      style={{
        background: `color-mix(in srgb, ${data.clusterColorSoft} 50%, white)`,
        borderColor: `color-mix(in srgb, ${data.clusterColorHex} 40%, transparent)`,
        borderWidth: '1.5px',
      }}
    >
      <div className="px-4 py-3">
        <p className="truncate text-sm font-semibold" style={{ color: data.clusterColorHex }}>
          {data.label}
        </p>
        <div className="mt-1 flex items-center gap-2 text-xs text-stone-400">
          <span>{data.functionCount} 関数</span>
          {data.changedCount > 0 && (
            <span className="font-medium text-amber-500">{data.changedCount} 変更</span>
          )}
        </div>
        <p className="mt-1.5 text-[10px] text-stone-300">クリックで展開</p>
      </div>
      <Handle type="target" position={Position.Top} style={{ opacity: 0 }} />
      <Handle type="source" position={Position.Bottom} style={{ opacity: 0 }} />
    </div>
  )
}
