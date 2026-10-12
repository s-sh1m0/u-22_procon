import type { ReactNode } from 'react'
import { Card, CardContent } from '@/components/ui/card'

/**
 * ノード非選択時に表示する凡例。ノード右上のマークとエッジ色が表す diff の種類と、
 * interface メソッドから実装への点線エッジの意味を示す。
 */
export default function DiffLegend() {
  return (
    <Card className="border-stone-200 bg-white shadow-md">
      <CardContent className="flex flex-col gap-1.5 p-2.5 text-xs">
        <p className="font-semibold text-stone-700">凡例</p>
        <LegendRow label="追加された関数">
          <span className="flex size-3.5 items-center justify-center rounded-full bg-emerald-500 text-[9px] font-bold leading-none text-white">
            +
          </span>
        </LegendRow>
        <LegendRow label="変更された関数">
          <span className="size-2 rounded-full bg-amber-400" />
        </LegendRow>
        <LegendRow label="削除された関数">
          <span className="flex size-3.5 items-center justify-center rounded-full bg-stone-400 text-[9px] font-bold leading-none text-white">
            −
          </span>
        </LegendRow>
        <LegendRow label="追加された呼び出し">
          <span className="h-0.5 w-4 rounded-full bg-emerald-500" />
        </LegendRow>
        <LegendRow label="interface の実装（interface をクリックで展開）">
          <span className="w-4 border-t-2 border-dotted border-stone-400" />
        </LegendRow>
      </CardContent>
    </Card>
  )
}

function LegendRow({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="flex items-center gap-2">
      <span className="flex w-4 items-center justify-center" aria-hidden>
        {children}
      </span>
      <span className="text-stone-600">{label}</span>
    </div>
  )
}
