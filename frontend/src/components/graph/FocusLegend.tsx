import { Card, CardContent } from '@/components/ui/card'
import { FOCUS_CALLER_COLOR, FOCUS_CALLEE_COLOR } from '@/lib/graphFocus'

type Props = {
  callerCount: number
  calleeCount: number
}

/** フォーカスモード中に表示する凡例。選択ノードの直接 caller / callee と方向色を示す。 */
export default function FocusLegend({ callerCount, calleeCount }: Props) {
  return (
    <Card className="border-stone-200 bg-white shadow-md">
      <CardContent className="flex flex-col gap-1.5 p-2.5 text-xs">
        <p className="font-semibold text-stone-700">フォーカス中</p>
        <LegendRow color={FOCUS_CALLER_COLOR} label="呼び出し元" count={callerCount} />
        <LegendRow color={FOCUS_CALLEE_COLOR} label="呼び出し先" count={calleeCount} />
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
