import Brand from '@/components/branding/Brand'
import { GitHubGlyph } from './glyphs'
import { btn, btnInk, navLinkStyle } from './styles'

const REPO_URL = 'https://github.com/s-sh1m0/u-22_procon'

type Props = { isAuthenticated: boolean; login?: string | null }

export default function LandingNav({ isAuthenticated, login }: Props) {
  return (
    <nav
      style={{
        borderBottom: '1px solid var(--line)',
        background: 'rgba(250,250,249,0.85)',
        backdropFilter: 'blur(8px)',
        WebkitBackdropFilter: 'blur(8px)',
        position: 'sticky',
        top: 0,
        zIndex: 50,
      }}
    >
      <div
        className="wrap"
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          height: 64,
        }}
      >
        <Brand size={17} />
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 28,
            fontSize: 13,
            color: 'var(--ink-2)',
          }}
        >
          <a className="nav-mid" style={navLinkStyle} href="#features">
            機能
          </a>
          <a className="nav-mid" style={navLinkStyle} href="#how">
            仕組み
          </a>
          <a
            className="nav-mid"
            style={navLinkStyle}
            href={REPO_URL}
            target="_blank"
            rel="noreferrer"
          >
            <GitHubGlyph size={13} /> GitHub
          </a>
          <div className="nav-mid" style={{ width: 1, height: 18, background: 'var(--line)' }} />
          {isAuthenticated ? (
            <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
              <span className="mono" style={{ fontSize: 12, color: 'var(--ink-3)' }}>
                {login}
              </span>
              <div
                style={{
                  width: 30,
                  height: 30,
                  borderRadius: 99,
                  background: 'linear-gradient(135deg, var(--brand) 0%, var(--brand-deep) 100%)',
                  color: 'white',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontSize: 11,
                  fontWeight: 600,
                  textTransform: 'uppercase',
                }}
              >
                {login?.[0] ?? '?'}
              </div>
            </div>
          ) : (
            <a
              href="/auth/github"
              style={{ ...btn, ...btnInk, height: 34, padding: '0 14px', fontSize: 13 }}
            >
              ログイン
            </a>
          )}
        </div>
      </div>
    </nav>
  )
}
