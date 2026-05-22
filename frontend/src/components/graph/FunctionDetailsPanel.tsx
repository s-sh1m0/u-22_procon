import type { AnyFlowNode, LayerKind } from '@/types/graph'
import type { Cluster, DiffFile, PRInfo } from '@/types/api'
import { getClusterColor } from '@/lib/clusterColors'
import { buildBlobUrl } from '@/lib/githubLinks'
import DiffViewer from '@/components/diff/DiffViewer'

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
  diff: DiffFile | undefined
  pr: PRInfo
  onClose: () => void
}

export default function FunctionDetailsPanel({ node, cluster, layer, diff, pr, onClose }: Props) {
  if (!node || node.type !== 'function') return null

  const color = cluster ? getClusterColor(cluster.id) : null
  const changed = node.data.changed as boolean
  const blobUrl = buildBlobUrl(pr, node.data.file, node.data.line, node.data.diffStatus)

  return (
    <div
      className={`flex flex-shrink-0 flex-col border-l border-stone-200 bg-white overflow-hidden ${diff ? 'w-[700px]' : 'w-80'}`}
    >
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

      <div className="p-4 space-y-4 border-b border-stone-100">
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

        <div>
          <p className="break-all text-base font-bold text-stone-900">{node.data.label}</p>
          <p className="mt-1 break-all font-mono text-xs text-stone-500">{node.data.packagePath}</p>
        </div>
        <div className="flex items-center justify-between gap-2">
          <p className="break-all font-mono text-xs text-stone-400">
            {node.data.file.split('/').pop()}:{node.data.line}
          </p>
          {blobUrl && (
            <a
              href={blobUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="flex flex-shrink-0 items-center gap-1 text-xs font-medium text-teal-600 hover:text-teal-700 hover:underline"
              title="GitHub で該当行を開く"
            >
              <svg
                width="13"
                height="13"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
                aria-hidden="true"
              >
                <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6" />
                <polyline points="15 3 21 3 21 9" />
                <line x1="10" y1="14" x2="21" y2="3" />
              </svg>
              GitHub
            </a>
          )}
        </div>
      </div>

      {diff ? (
        <div className="flex min-h-0 flex-1">
          <DiffViewer file={diff} />
        </div>
      ) : (
        <div className="px-4 py-3 text-xs text-stone-400">このファイルに変更はありません</div>
      )}
    </div>
  )
}
