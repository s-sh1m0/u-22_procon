import { Card, CardContent } from '@/components/ui/card'
import { FOCUS_CALLER_COLOR, FOCUS_CALLEE_COLOR } from '@/lib/graphFocus'

const CHANGED_COLOR = '#fbbf24' // amber-400: 変更ノードのリング色と揃える

type Props = {
  originCount: number
  upstreamCount: number
  downstreamCount: number
}

/**
 * 影響範囲モード中の凡例。変更ノードを起点に上流（呼び出し元）/ 下流（呼び出し先）へ
 * 多ホップで波及した件数と方向色を示す。
 */
export default function BlastLegend({ originCount, upstreamCount, downstreamCount }: Props) {
  return (
    <Card className="border-stone-200 bg-white shadow-md">
      <CardContent className="flex flex-col gap-1.5 p-2.5 text-xs">
        <p className="font-semibold text-stone-700">影響範囲</p>
        <LegendRow color={CHANGED_COLOR} label="変更箇所" count={originCount} />
        <LegendRow color={FOCUS_CALLER_COLOR} label="上流 (要確認)" count={upstreamCount} />
        <LegendRow color={FOCUS_CALLEE_COLOR} label="下流 (依存先)" count={downstreamCount} />
      </CardContent>
    </Card>
  )
}

function LegendRow({ color, label, count }: { color: string; label: string; count: number }) {
  return (
    <div className="flex items-center gap-2">
      <span className="h-0.5 w-4 rounded-full" style={{ background: color }} aria-hidden />
      <span className="text-stone-600">{label}</span>
      <span className="ml-auto font-mono text-stone-400">{count}</span>
    </div>
  )
}
