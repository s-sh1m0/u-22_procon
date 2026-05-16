import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'

type Props = {
  onFitChanged: () => void
  onExpandAll: () => void
  onCollapseAll: () => void
}

export default function GraphControls({ onFitChanged, onExpandAll, onCollapseAll }: Props) {
  return (
    <Card className="shadow-md border-stone-200 bg-white">
      <CardContent className="p-2 flex flex-col gap-2">
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
