import Wrap from './Wrap'
import GuestCta from './GuestCta'
import PrInputForm from './PrInputForm'
import HeroIllustration from './HeroIllustration'

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
            変更ファイルの羅列ではなく、モジュールの依存関係として PR
            を可視化。大規模レビューも、意味のあるクラスタに自動分割します。
          </p>

          {isAuthenticated ? <PrInputForm /> : <GuestCta />}
        </div>

        <HeroIllustration />
      </Wrap>
    </section>
  )
}
