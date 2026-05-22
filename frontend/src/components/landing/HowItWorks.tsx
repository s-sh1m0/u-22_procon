import type { ReactNode } from 'react'
import SectionHead from './SectionHead'
import { GitHubGlyph } from './glyphs'

function StepMock1() {
  return (
    <div style={{ width: '100%' }}>
      <div className="mono" style={{ fontSize: 10, color: 'var(--ink-3)', marginBottom: 6 }}>
        PR URL
      </div>
      <div
        style={{
          background: 'white',
          border: '1.5px solid var(--brand)',
          borderRadius: 8,
          padding: '8px 10px',
          display: 'flex',
          alignItems: 'center',
          gap: 8,
        }}
      >
        <GitHubGlyph size={12} color="var(--ink-3)" />
        <span
          className="mono"
          style={{
            fontSize: 11,
            color: 'var(--ink-2)',
            flex: 1,
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            whiteSpace: 'nowrap',
          }}
        >
          github.com/kobold/checkout-service/pull/1284
        </span>
        <span style={{ fontSize: 10, color: 'var(--added)', fontWeight: 500 }}>✓</span>
      </div>
      <div style={{ marginTop: 8, display: 'flex', justifyContent: 'flex-end' }}>
        <div
          style={{
            background: 'var(--ink)',
            color: 'white',
            padding: '6px 12px',
            borderRadius: 6,
            fontSize: 11,
          }}
        >
          解析を開始 →
        </div>
      </div>
    </div>
  )
}

function StepMock2() {
  const files: [number, number, string][] = [
    [36, 28, 'handler.go'],
    [156, 28, 'model.go'],
    [36, 108, 'repo.go'],
    [156, 108, 'util.go'],
  ]
  return (
    <svg viewBox="0 0 240 160" style={{ width: '100%' }} aria-hidden="true">
      <g stroke="var(--line-strong)" strokeWidth="1" fill="none">
        <line x1="60" y1="40" x2="120" y2="80" />
        <line x1="180" y1="40" x2="120" y2="80" />
        <line x1="60" y1="120" x2="120" y2="80" />
        <line x1="180" y1="120" x2="120" y2="80" />
      </g>
      <g>
        {files.map(([x, y, label]) => (
          <g key={label}>
            <rect
              x={x}
              y={y}
              width="48"
              height="24"
              rx="4"
              fill="white"
              stroke="var(--line-strong)"
            />
            <text
              x={x + 24}
              y={y + 16}
              textAnchor="middle"
              fontSize="9"
              fill="var(--ink-2)"
              fontFamily="Inter Tight"
            >
              {label}
            </text>
          </g>
        ))}
        <circle cx="120" cy="80" r="14" fill="var(--brand)" />
        <text
          x="120"
          y="84"
          textAnchor="middle"
          fontSize="9"
          fill="white"
          fontFamily="JetBrains Mono"
          fontWeight="600"
        >
          AST
        </text>
      </g>
    </svg>
  )
}

function StepMock3() {
  const bars: [string, string, string, string, number][] = [
    ['追加', 'var(--added)', 'var(--added-bg)', 'var(--added-bd)', 70],
    ['修正', 'var(--modified)', 'var(--modified-bg)', 'var(--modified-bd)', 90],
    ['リファクタ', 'var(--refactor)', 'var(--refactor-bg)', 'var(--refactor-bd)', 45],
    ['削除', 'var(--removed)', 'var(--removed-bg)', 'var(--removed-bd)', 30],
  ]
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 6, width: '100%' }}>
      {bars.map(([lbl, c, bg, bd, w]) => (
        <div key={lbl} style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <span style={{ width: 64, fontSize: 11, color: 'var(--ink-2)' }}>{lbl}</span>
          <div
            style={{
              flex: 1,
              height: 10,
              background: bg,
              border: `1px solid ${bd}`,
              borderRadius: 99,
              overflow: 'hidden',
            }}
          >
            <div style={{ width: `${w}%`, height: '100%', background: c, opacity: 0.85 }} />
          </div>
        </div>
      ))}
    </div>
  )
}

type Step = { n: string; title: string; desc: string; mock: ReactNode }

const STEPS: Step[] = [
  {
    n: '01',
    title: 'PR を貼り付ける',
    desc: 'GitHub の PR URL を貼るだけ。OAuth 連携でプライベートリポジトリも対応。',
    mock: <StepMock1 />,
  },
  {
    n: '02',
    title: '依存グラフを構築',
    desc: 'go/ast で関数・型を抽出し、呼び出しグラフを双方向に辿ります。変更行を AST ノードにマッピング。',
    mock: <StepMock2 />,
  },
  {
    n: '03',
    title: 'クラスタごとに読む',
    desc: '追加 → 修正 → リファクタ → 削除 の順に、レビュアーの認知に沿った読み方を提案します。',
    mock: <StepMock3 />,
  },
]

export default function HowItWorks() {
  return (
    <section
      id="how"
      style={{
        padding: 'clamp(72px, 10vw, 112px) 0',
        borderBottom: '1px solid var(--line)',
        background: 'linear-gradient(180deg, var(--bg) 0%, white 100%)',
      }}
    >
      <div className="wrap">
        <SectionHead
          eyebrow="How it works"
          title="3 ステップで、PR を構造化。"
          lead="バックエンドが Go AST を解析し、フロントが xyflow でレンダリング。"
        />
        <div className="grid-3" style={{ marginTop: 56 }}>
          {STEPS.map((s) => (
            <div
              key={s.n}
              style={{
                background: 'white',
                border: '1px solid var(--line)',
                borderRadius: 16,
                padding: 24,
                display: 'flex',
                flexDirection: 'column',
                gap: 16,
                minHeight: 380,
              }}
            >
              <div style={{ display: 'flex', alignItems: 'baseline', gap: 12 }}>
                <span
                  className="mono"
                  style={{
                    fontSize: 11,
                    color: 'var(--brand)',
                    letterSpacing: '0.06em',
                    flexShrink: 0,
                  }}
                >
                  {s.n}
                </span>
                <h3 style={{ fontSize: 19, fontWeight: 600, letterSpacing: '-0.015em', margin: 0 }}>
                  {s.title}
                </h3>
              </div>
              <p style={{ fontSize: 13.5, color: 'var(--ink-2)', lineHeight: 1.6, margin: 0 }}>
                {s.desc}
              </p>
              <div
                style={{
                  marginTop: 'auto',
                  flex: 1,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  background: 'var(--bg-alt)',
                  border: '1px solid var(--line)',
                  borderRadius: 12,
                  padding: 16,
                }}
              >
                {s.mock}
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
