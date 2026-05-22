import { CheckGlyph, GitHubGlyph } from './glyphs'
import { btn, btnGhost, btnInk } from './styles'

export default function GuestCta() {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 14, alignItems: 'flex-start' }}>
      <div style={{ display: 'flex', gap: 10, alignItems: 'center', flexWrap: 'wrap' }}>
        <a
          href="/auth/github"
          style={{
            ...btn,
            ...btnInk,
            height: 52,
            padding: '0 22px',
            fontSize: 15,
            boxShadow: '0 4px 14px rgba(28,25,23,0.18), inset 0 1px 0 rgba(255,255,255,0.08)',
          }}
        >
          <GitHubGlyph size={16} />
          GitHub で始める
          <span style={{ opacity: 0.5, marginLeft: 4 }}>→</span>
        </a>
        <a href="#how" style={{ ...btn, ...btnGhost, height: 52, padding: '0 22px', fontSize: 15 }}>
          デモを見る
        </a>
      </div>
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 14,
          fontSize: 12,
          color: 'var(--ink-3)',
          flexWrap: 'wrap',
        }}
      >
        <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}>
          <CheckGlyph color="var(--added)" />
          無料で使い始められる
        </span>
        <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}>
          <CheckGlyph color="var(--added)" />
          パブリックリポジトリのみ要求
        </span>
      </div>
    </div>
  )
}
