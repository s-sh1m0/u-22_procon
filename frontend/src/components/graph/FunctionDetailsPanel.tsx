import type { AnyFlowNode, LayerKind } from '@/types/graph'
import type { Cluster } from '@/types/api'
import { getClusterColor } from '@/lib/clusterColors'

const LAYER_LABELS: Record<LayerKind, string> = {
  ui: 'UI',
  domain: 'Domain',
  data: 'Data',
  infra: 'Infra',
  other: 'Other',
}

type Props = {
  node: AnyFlowNode | null
  cluster: Cluster | null
  layer: LayerKind
  onClose: () => void
}

export default function FunctionDetailsPanel({ node, cluster, layer, onClose }: Props) {
  if (!node || (node.type !== 'function' && node.type !== 'file')) return null

  const color = cluster ? getClusterColor(cluster.id) : null
  const changed = node.data.changed as boolean

  return (
    <div className="flex w-80 flex-shrink-0 flex-col border-l border-stone-200 bg-white overflow-auto">
      <div className="flex items-center justify-between border-b border-stone-100 px-4 py-3">
        <span className="text-sm font-semibold text-stone-800">詳細</span>
        <button
          onClick={onClose}
          className="rounded p-1 text-stone-400 hover:bg-stone-100 hover:text-stone-600"
          aria-label="閉じる"
        >
          <svg
            width="14"
            height="14"
            viewBox="0 0 14 14"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
          >
            <path d="M2 2l10 10M12 2L2 12" />
          </svg>
        </button>
      </div>

      <div className="p-4 space-y-4">
        <div className="flex flex-wrap gap-1.5">
          {cluster && color && (
            <span
              className="rounded-full px-2 py-0.5 text-xs font-medium"
              style={{ background: color.soft, color: color.ink }}
            >
              {cluster.label}
            </span>
          )}
          <span className="rounded-full bg-stone-100 px-2 py-0.5 text-xs font-medium text-stone-600">
            {LAYER_LABELS[layer]}
          </span>
          {changed && (
            <span className="rounded-full bg-amber-100 px-2 py-0.5 text-xs font-medium text-amber-700">
              changed
            </span>
          )}
        </div>

        {node.type === 'function' && (
          <>
            <div>
              <p className="break-all text-base font-bold text-stone-900">{node.data.label}</p>
              <p className="mt-1 break-all font-mono text-xs text-stone-500">
                {node.data.packagePath}
              </p>
            </div>
            <p className="font-mono text-xs text-stone-400">
              {node.data.file.split('/').pop()}:{node.data.line}
            </p>
          </>
        )}

        {node.type === 'file' && (
          <>
            <div>
              <p className="break-all text-base font-bold text-stone-900">{node.data.fileName}</p>
              <p className="mt-1 break-all font-mono text-xs text-stone-500">
                {node.data.packagePath}
              </p>
            </div>
            <div className="space-y-1 text-sm text-stone-600">
              <p>{node.data.functionCount} 関数</p>
              {node.data.changedCount > 0 && (
                <p className="text-amber-600">{node.data.changedCount} 件変更あり</p>
              )}
            </div>
          </>
        )}
      </div>
    </div>
  )
}
