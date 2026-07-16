import type { NodeProps } from '@xyflow/react'
import type { ClusterFlowNode } from '@/types/graph'

export default function ClusterGroup({ data }: NodeProps<ClusterFlowNode>) {
  return (
    <div
      className="size-full rounded-xl border"
      style={{
        background: `color-mix(in srgb, ${data.clusterColorSoft} 60%, white)`,
        border: `1.5px solid color-mix(in srgb, ${data.clusterColorHex} 35%, transparent)`,
      }}
    >
      <div className="flex items-center justify-between px-3 pt-2">
        <span className="text-xs font-semibold" style={{ color: data.clusterColorHex }}>
          {data.label}
        </span>
        <span
          className="cursor-pointer rounded px-1 text-[10px] text-stone-400 hover:text-stone-600"
          title="折りたたむ"
          aria-label="クラスタを折りたたむ"
        >
          −
        </span>
      </div>
    </div>
  )
}
