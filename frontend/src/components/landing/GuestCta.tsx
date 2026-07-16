import { Button } from '@/components/ui/button'
import { API_BASE_URL } from '@/lib/api'
import { CheckGlyph, GitHubGlyph } from './glyphs'

export default function GuestCta() {
  return (
    <div className="flex flex-col items-start gap-3.5">
      <div className="flex flex-wrap items-center gap-2.5">
        <Button
          asChild
          className="h-13 gap-2 bg-stone-900 px-[22px] text-[15px] text-white shadow-[0_4px_14px_rgba(28,25,23,0.18),inset_0_1px_0_rgba(255,255,255,0.08)] hover:bg-stone-800"
        >
          <a href={`${API_BASE_URL}/auth/github`}>
            <GitHubGlyph size={16} />
            GitHub で始める
            <span className="ml-1 opacity-50">→</span>
          </a>
        </Button>
        <Button
          asChild
          variant="outline"
          className="h-13 border-stone-300 px-[22px] text-[15px] text-stone-900"
        >
          <a href="#how">デモを見る</a>
        </Button>
      </div>
      <div className="flex flex-wrap items-center gap-3.5 text-xs text-stone-500">
        <span className="inline-flex items-center gap-1.5">
          <CheckGlyph className="text-green-600" />
          無料で使い始められる
        </span>
        <span className="inline-flex items-center gap-1.5">
          <CheckGlyph className="text-green-600" />
          パブリックリポジトリのみ要求
        </span>
      </div>
    </div>
  )
}
