import type { ClusterMode } from '@/types/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'

type Props = {
  clusterMode: ClusterMode
  onChangeClusterMode: (mode: ClusterMode) => void
  onFitChanged: () => void
  onExpandAll: () => void
  onCollapseAll: () => void
}

const CLUSTER_MODES: { value: ClusterMode; label: string }[] = [
  { value: 'louvain', label: 'Louvain' },
  { value: 'package', label: 'パッケージ' },
  { value: 'file', label: 'ファイル' },
]

export default function GraphControls({
  clusterMode,
  onChangeClusterMode,
  onFitChanged,
  onExpandAll,
  onCollapseAll,
}: Props) {
  return (
    <Card className="shadow-md border-stone-200 bg-white">
      <CardContent className="p-2 flex flex-col gap-2">
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
        <Button variant="outline" size="sm" className="h-7 w-full text-xs" onClick={onFitChanged}>
          変更ノードに寄せる
        </Button>
      </CardContent>
    </Card>
  )
}
