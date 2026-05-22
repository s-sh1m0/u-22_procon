import GuestCta from './GuestCta'
import PrInputForm from './PrInputForm'
import HeroIllustration from './HeroIllustration'
import { pillStyle } from './styles'

function Stat({ n, l }: { n: string; l: string }) {
  return (
    <div>
      <div
        style={{
          fontSize: 22,
          fontWeight: 600,
          color: 'var(--ink)',
          fontVariantNumeric: 'tabular-nums',
          letterSpacing: '-0.02em',
        }}
      >
        {n}
      </div>
      <div style={{ fontSize: 12, color: 'var(--ink-3)', marginTop: 2 }}>{l}</div>
    </div>
  )
}

const Divider = () => <div style={{ width: 1, height: 32, background: 'var(--line)' }} />

export default function LandingHero({ isAuthenticated }: { isAuthenticated: boolean }) {
  return (
    <section
      style={{ position: 'relative', overflow: 'hidden', borderBottom: '1px solid var(--line)' }}
    >
      {/* faint dotted grid */}
      <div
        style={{
          position: 'absolute',
          inset: 0,
          backgroundImage: 'radial-gradient(var(--line-strong) 1px, transparent 1px)',
          backgroundSize: '24px 24px',
          opacity: 0.35,
          maskImage: 'radial-gradient(80% 60% at 50% 30%, black 0%, transparent 75%)',
          WebkitMaskImage: 'radial-gradient(80% 60% at 50% 30%, black 0%, transparent 75%)',
        }}
      />
      <div
        className="wrap hero-grid"
        style={{
          position: 'relative',
          paddingTop: 'clamp(56px, 8vw, 88px)',
          paddingBottom: 'clamp(64px, 9vw, 96px)',
        }}
      >
        <div>
          <div style={pillStyle}>
            <span style={{ width: 6, height: 6, borderRadius: 99, background: 'var(--brand)' }} />
            <span>Go プロジェクト対応 · ベータ公開中</span>
          </div>
          <h1
            style={{
              fontSize: 'clamp(38px, 7vw, 60px)',
              lineHeight: 1.02,
              fontWeight: 600,
              letterSpacing: '-0.04em',
              margin: '20px 0 18px',
            }}
          >
            プルリクを、
            <br />
            <span style={{ color: 'var(--brand)' }}>構造</span>として読む。
          </h1>
          <p
            style={{
              fontSize: 17,
              lineHeight: 1.6,
              color: 'var(--ink-2)',
              margin: '0 0 32px',
              maxWidth: 500,
            }}
          >
            変更ファイルの羅列ではなく、モジュールの依存関係として PR を可視化。100
            ファイル超のレビューも、4 つの意味のあるクラスタに自動分割します。
          </p>

          {isAuthenticated ? <PrInputForm /> : <GuestCta />}

          <div
            style={{
              marginTop: 40,
              display: 'flex',
              gap: 28,
              alignItems: 'center',
              flexWrap: 'wrap',
            }}
          >
            <Stat n="2,400+" l="解析された PR" />
            <Divider />
            <Stat n="38%" l="レビュー時間短縮" />
            <Divider />
            <Stat n="< 5s" l="平均解析時間" />
          </div>
        </div>

        <HeroIllustration />
      </div>
    </section>
  )
}
