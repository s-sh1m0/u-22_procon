import Brand from '@/components/branding/Brand'
import { GitHubGlyph } from './glyphs'
import { btn, btnGhost } from './styles'

const REPO_URL = 'https://github.com/s-sh1m0/u-22_procon'

type Link = { label: string; href?: string }
type Col = { title: string; links: Link[] }

const COLS: Col[] = [
  {
    title: 'プロダクト',
    links: [
      { label: '機能', href: '#features' },
      { label: '仕組み', href: '#how' },
      { label: '料金' },
      { label: 'ロードマップ' },
      { label: '変更履歴' },
    ],
  },
  {
    title: 'ドキュメント',
    links: [
      { label: 'はじめに' },
      { label: 'API リファレンス' },
      { label: 'セルフホスト' },
      { label: 'CLI' },
    ],
  },
  {
    title: '会社',
    links: [
      { label: 'About' },
      { label: 'ブログ' },
      { label: 'お問い合わせ' },
      { label: 'プレスキット' },
    ],
  },
]

const linkStyle = { fontSize: 13, color: 'var(--ink-3)', textDecoration: 'none' } as const

function FooterLink({ link }: { link: Link }) {
  if (link.href) {
    return (
      <a href={link.href} style={{ ...linkStyle, cursor: 'pointer' }}>
        {link.label}
      </a>
    )
  }
  return <span style={{ ...linkStyle, cursor: 'default' }}>{link.label}</span>
}

export default function LandingFooter() {
  return (
    <footer
      style={{ padding: '72px 0 40px', background: 'white', borderTop: '1px solid var(--line)' }}
    >
      <div className="wrap footer-grid">
        <div>
          <Brand size={16} />
          <p
            style={{
              fontSize: 13,
              color: 'var(--ink-3)',
              margin: '14px 0 20px',
              lineHeight: 1.6,
              maxWidth: 280,
            }}
          >
            PR を変更行ではなく依存グラフとして読むためのレビュー支援ツール。
          </p>
          <div style={{ display: 'flex', gap: 10 }}>
            <a
              href={REPO_URL}
              target="_blank"
              rel="noreferrer"
              style={{ ...btnGhost, ...btn, height: 32, padding: '0 12px', fontSize: 12 }}
            >
              <GitHubGlyph size={12} />
              GitHub
            </a>
          </div>
        </div>
        {COLS.map((c) => (
          <div key={c.title}>
            <div
              style={{
                fontSize: 12,
                fontWeight: 600,
                color: 'var(--ink-2)',
                marginBottom: 12,
                letterSpacing: '-0.005em',
              }}
            >
              {c.title}
            </div>
            <ul
              style={{
                listStyle: 'none',
                padding: 0,
                margin: 0,
                display: 'flex',
                flexDirection: 'column',
                gap: 8,
              }}
            >
              {c.links.map((l) => (
                <li key={l.label}>
                  <FooterLink link={l} />
                </li>
              ))}
            </ul>
          </div>
        ))}
      </div>
      <div
        className="wrap"
        style={{
          marginTop: 56,
          paddingTop: 24,
          borderTop: '1px solid var(--line)',
          display: 'flex',
          justifyContent: 'space-between',
          gap: 12,
          flexWrap: 'wrap',
          fontSize: 12,
          color: 'var(--ink-3)',
        }}
      >
        <span>© 2026 diffmap. — u-22 プログラミングコンテスト</span>
        <span className="mono">v0.4.2-beta</span>
      </div>
    </footer>
  )
}
