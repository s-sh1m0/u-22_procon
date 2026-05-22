import AppMark from '@/components/branding/AppMark'
import { GitHubGlyph } from './glyphs'
import { btn } from './styles'

export default function CtaStrip({ isAuthenticated }: { isAuthenticated: boolean }) {
  return (
    <section
      style={{
        padding: 'clamp(72px, 9vw, 104px) 0',
        background: 'var(--ink)',
        color: 'white',
        position: 'relative',
        overflow: 'hidden',
      }}
    >
      <div
        style={{
          position: 'absolute',
          inset: 0,
          backgroundImage:
            'radial-gradient(120% 80% at 80% 50%, rgba(15,118,110,0.32) 0%, transparent 60%), radial-gradient(80% 60% at 0% 100%, rgba(94,234,212,0.08) 0%, transparent 60%)',
        }}
      />
      <div className="wrap cta-grid" style={{ position: 'relative' }}>
        <div>
          <h2
            style={{
              fontSize: 'clamp(34px, 5vw, 48px)',
              fontWeight: 600,
              letterSpacing: '-0.035em',
              lineHeight: 1.05,
              margin: '0 0 16px',
            }}
          >
            次の大規模 PR から、
            <br />
            <span style={{ color: 'var(--brand-light)' }}>構造を読もう。</span>
          </h2>
          <p
            style={{
              fontSize: 16,
              color: 'rgba(255,255,255,0.65)',
              margin: '0 0 32px',
              maxWidth: 480,
              lineHeight: 1.6,
            }}
          >
            GitHub と連携して 30 秒で開始。公開リポジトリは無料、Pro プラン招待ベータ受付中。
          </p>
          {isAuthenticated ? (
            <a
              href="#top"
              style={{
                ...btn,
                background: 'var(--brand)',
                color: 'white',
                height: 52,
                padding: '0 22px',
                fontSize: 15,
                boxShadow: '0 4px 14px rgba(15,118,110,0.45)',
              }}
            >
              新しい PR を解析する →
            </a>
          ) : (
            <a
              href="/auth/github"
              style={{
                ...btn,
                background: 'white',
                color: 'var(--ink)',
                height: 52,
                padding: '0 22px',
                fontSize: 15,
                boxShadow: '0 4px 14px rgba(0,0,0,0.25)',
              }}
            >
              <GitHubGlyph size={16} />
              GitHub で始める
            </a>
          )}
        </div>
        <div className="cta-art" style={{ display: 'flex', justifyContent: 'center' }}>
          <AppMark size={200} bg="var(--brand)" fg="white" accent="white" shadow />
        </div>
      </div>
    </section>
  )
}
