import Wrap from './Wrap'
import GuestCta from './GuestCta'
import PrInputForm from './PrInputForm'
import HeroIllustration from './HeroIllustration'

function Stat({ n, l }: { n: string; l: string }) {
  return (
    <div>
      <div className="text-[22px] font-semibold tracking-[-0.02em] text-stone-900 tabular-nums">
        {n}
      </div>
      <div className="mt-0.5 text-xs text-stone-500">{l}</div>
    </div>
  )
}

const Divider = () => <div className="h-8 w-px bg-stone-200" />

export default function LandingHero({ isAuthenticated }: { isAuthenticated: boolean }) {
  return (
    <section className="relative overflow-hidden border-b border-stone-200">
      {/* faint dotted grid */}
      <div
        className="pointer-events-none absolute inset-0 opacity-35"
        style={{
          backgroundImage: 'radial-gradient(#d6d3d1 1px, transparent 1px)',
          backgroundSize: '24px 24px',
          maskImage: 'radial-gradient(80% 60% at 50% 30%, black 0%, transparent 75%)',
          WebkitMaskImage: 'radial-gradient(80% 60% at 50% 30%, black 0%, transparent 75%)',
        }}
      />
      <Wrap className="relative grid items-center gap-11 pt-[clamp(56px,8vw,88px)] pb-[clamp(64px,9vw,96px)] lg:grid-cols-[1.05fr_1fr] lg:gap-16">
        <div>
          <div className="inline-flex items-center gap-2 rounded-full border border-teal-200 bg-teal-100 px-3 py-1.5 text-xs font-medium text-teal-900">
            <span className="size-1.5 rounded-full bg-teal-700" />
            <span>Go プロジェクト対応 · ベータ公開中</span>
          </div>
          <h1 className="m-0 mt-5 mb-[18px] text-[clamp(38px,7vw,60px)] font-semibold leading-[1.02] tracking-[-0.04em] text-balance">
            プルリクを、
            <br />
            <span className="text-teal-700">構造</span>として読む。
          </h1>
          <p className="m-0 mb-8 max-w-[500px] text-[17px] leading-relaxed text-stone-700">
            変更ファイルの羅列ではなく、モジュールの依存関係として PR を可視化。100
            ファイル超のレビューも、4 つの意味のあるクラスタに自動分割します。
          </p>

          {isAuthenticated ? <PrInputForm /> : <GuestCta />}

          <div className="mt-10 flex flex-wrap items-center gap-7">
            <Stat n="2,400+" l="解析された PR" />
            <Divider />
            <Stat n="38%" l="レビュー時間短縮" />
            <Divider />
            <Stat n="< 5s" l="平均解析時間" />
          </div>
        </div>

        <HeroIllustration />
      </Wrap>
    </section>
  )
}
