import AppMark from '@/components/branding/AppMark'
import { Button } from '@/components/ui/button'
import { API_BASE_URL } from '@/lib/api'
import Wrap from './Wrap'
import { GitHubGlyph } from './glyphs'

export default function CtaStrip({ isAuthenticated }: { isAuthenticated: boolean }) {
  return (
    <section className="relative overflow-hidden bg-stone-900 py-[clamp(72px,9vw,104px)] text-white">
      <div
        className="pointer-events-none absolute inset-0"
        style={{
          backgroundImage:
            'radial-gradient(120% 80% at 80% 50%, rgba(15,118,110,0.32) 0%, transparent 60%), radial-gradient(80% 60% at 0% 100%, rgba(94,234,212,0.08) 0%, transparent 60%)',
        }}
      />
      <Wrap className="relative grid items-center gap-10 lg:grid-cols-[1.4fr_1fr] lg:gap-14">
        <div>
          <h2 className="m-0 mb-4 text-[clamp(34px,5vw,48px)] font-semibold leading-[1.05] tracking-[-0.035em]">
            次の大規模 PR から、
            <br />
            <span className="text-teal-300">構造を読もう。</span>
          </h2>
          <p className="m-0 mb-8 max-w-[480px] text-base leading-relaxed text-white/65">
            GitHub と連携するだけですぐに使えます。パブリックリポジトリは無料で解析可能。
          </p>
          {isAuthenticated ? (
            <Button
              asChild
              className="h-13 bg-teal-700 px-[22px] text-[15px] text-white shadow-[0_4px_14px_rgba(15,118,110,0.45)] hover:bg-teal-800"
            >
              <a href="#top">新しい PR を解析する →</a>
            </Button>
          ) : (
            <Button
              asChild
              className="h-13 gap-2 bg-white px-[22px] text-[15px] text-stone-900 shadow-[0_4px_14px_rgba(0,0,0,0.25)] hover:bg-white/90"
            >
              <a href={`${API_BASE_URL}/auth/github`}>
                <GitHubGlyph size={16} />
                GitHub で始める
              </a>
            </Button>
          )}
        </div>
        <div className="hidden justify-center lg:flex">
          <AppMark size={200} shadow />
        </div>
      </Wrap>
    </section>
  )
}
