import type { ClusterMode } from '@/types/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'

type Props = {
  clusterMode: ClusterMode
  onChangeClusterMode: (mode: ClusterMode) => void
  impactOnly: boolean
  onSetImpactOnly: (next: boolean) => void
  changedCount: number
  changedIndex: number | null
  onNextChanged: () => void
  onPrevChanged: () => void
  onExpandAll: () => void
  onCollapseAll: () => void
}

const CLUSTER_MODES: { value: ClusterMode; label: string }[] = [
  { value: 'louvain', label: 'Louvain' },
  { value: 'package', label: 'パッケージ' },
  { value: 'file', label: 'ファイル' },
]

const VIEW_MODES: { value: boolean; label: string }[] = [
  { value: true, label: '変更影響のみ' },
  { value: false, label: '全体' },
]

export default function GraphControls({
  clusterMode,
  onChangeClusterMode,
  impactOnly,
  onSetImpactOnly,
  changedCount,
  changedIndex,
  onNextChanged,
  onPrevChanged,
  onExpandAll,
  onCollapseAll,
}: Props) {
  return (
    <Card className="shadow-md border-stone-200 bg-white">
      <CardContent className="p-2 flex flex-col gap-2">
        <div className="flex overflow-hidden rounded border border-stone-200 text-xs">
          {VIEW_MODES.map((m, i) => (
            <button
              key={String(m.value)}
              className={[
                'flex-1 px-2 py-1.5 transition-colors',
                i > 0 ? 'border-l border-stone-200' : '',
                impactOnly === m.value
                  ? 'bg-stone-900 text-white'
                  : 'bg-white text-stone-600 hover:bg-stone-50',
              ].join(' ')}
              onClick={() => onSetImpactOnly(m.value)}
            >
              {m.label}
            </button>
          ))}
        </div>
        <div className="flex overflow-hidden rounded border border-stone-200 text-xs">
          {CLUSTER_MODES.map((m, i) => (
            <button
              key={m.value}
              className={[
                'flex-1 px-2 py-1.5 transition-colors',
                i > 0 ? 'border-l border-stone-200' : '',
                clusterMode === m.value
                  ? 'bg-stone-900 text-white'
                  : 'bg-white text-stone-600 hover:bg-stone-50',
              ].join(' ')}
              onClick={() => onChangeClusterMode(m.value)}
            >
              {m.label}
            </button>
          ))}
        </div>
        <div className="flex gap-1">
          <Button variant="outline" size="sm" className="h-7 flex-1 text-xs" onClick={onExpandAll}>
            すべて展開
          </Button>
          <Button
            variant="outline"
            size="sm"
            className="h-7 flex-1 text-xs"
            onClick={onCollapseAll}
          >
            折りたたむ
          </Button>
        </div>
        {changedCount > 0 && (
          <div className="flex items-center gap-1">
            <Button
              variant="outline"
              size="icon"
              className="size-7"
              onClick={onPrevChanged}
              aria-label="前の変更ノード"
            >
              <svg
                width="12"
                height="12"
                viewBox="0 0 12 12"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
                aria-hidden="true"
              >
                <path d="M7.5 2.5L4 6l3.5 3.5" />
              </svg>
            </Button>
            <span className="flex-1 text-center text-xs tabular-nums text-stone-600">
              {changedIndex !== null ? changedIndex + 1 : '-'} / {changedCount} 変更
            </span>
            <Button
              variant="outline"
              size="icon"
              className="size-7"
              onClick={onNextChanged}
              aria-label="次の変更ノード"
            >
              <svg
                width="12"
                height="12"
                viewBox="0 0 12 12"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
                aria-hidden="true"
              >
                <path d="M4.5 2.5L8 6l-3.5 3.5" />
              </svg>
            </Button>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
