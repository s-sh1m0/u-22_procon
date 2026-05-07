import type { NodeKind } from '@/types/graph'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'

type Props = {
  nodeKind: NodeKind
  onChangeNodeKind: (kind: NodeKind) => void
  onFitChanged: () => void
}

export default function GraphControls({ nodeKind, onChangeNodeKind, onFitChanged }: Props) {
  return (
    <Card className="shadow-md border-stone-200 bg-white">
      <CardContent className="p-2 flex flex-col gap-2">
        <div className="flex overflow-hidden rounded border border-stone-200 text-xs">
          <button
            className={[
              'px-3 py-1.5 transition-colors',
              nodeKind === 'function'
                ? 'bg-stone-900 text-white'
                : 'bg-white text-stone-600 hover:bg-stone-50',
            ].join(' ')}
            onClick={() => onChangeNodeKind('function')}
          >
            関数
          </button>
          <button
            className={[
              'border-l border-stone-200 px-3 py-1.5 transition-colors',
              nodeKind === 'file'
                ? 'bg-stone-900 text-white'
                : 'bg-white text-stone-600 hover:bg-stone-50',
            ].join(' ')}
            onClick={() => onChangeNodeKind('file')}
          >
            ファイル
          </button>
        </div>
        <Button variant="outline" size="sm" className="h-7 w-full text-xs" onClick={onFitChanged}>
          変更ノードに寄せる
        </Button>
      </CardContent>
    </Card>
  )
}
