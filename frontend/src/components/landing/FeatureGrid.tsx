import type { ReactNode } from 'react'
import SectionHead from './SectionHead'

function FeatureArt1() {
  return (
    <svg viewBox="0 0 200 120" style={{ width: '100%', height: '100%' }} aria-hidden="true">
      <circle
        cx="100"
        cy="60"
        r="50"
        fill="none"
        stroke="var(--brand-light)"
        strokeWidth="1"
        strokeDasharray="2 3"
        opacity="0.6"
      />
      <circle
        cx="100"
        cy="60"
        r="32"
        fill="none"
        stroke="var(--brand-light)"
        strokeWidth="1"
        strokeDasharray="2 3"
        opacity="0.8"
      />
      <circle cx="100" cy="60" r="14" fill="var(--brand)" />
      <text
        x="100"
        y="64"
        textAnchor="middle"
        fontSize="9"
        fill="white"
        fontFamily="JetBrains Mono"
        fontWeight="600"
      >
        0
      </text>
      <text
        x="100"
        y="22"
        textAnchor="middle"
        fontSize="9"
        fill="var(--brand)"
        fontFamily="JetBrains Mono"
      >
        1 hop
      </text>
      <text
        x="100"
        y="6"
        textAnchor="middle"
        fontSize="9"
        fill="var(--brand)"
        fontFamily="JetBrains Mono"
        opacity="0.7"
      >
        2 hop
      </text>
    </svg>
  )
}

function FeatureArt2() {
  return (
    <div className="mono" style={{ fontSize: 10, color: 'var(--ink-2)', width: '100%' }}>
      <div
        style={{
          background: 'white',
          borderRadius: 6,
          border: '1px solid var(--line)',
          overflow: 'hidden',
        }}
      >
        <div
          style={{
            padding: '4px 8px',
            borderBottom: '1px solid var(--line)',
            color: 'var(--ink-3)',
          }}
        >
          PaymentService.go
        </div>
        <div style={{ padding: '4px 8px', background: 'var(--added-bg)', color: 'var(--added)' }}>
          + func (s *PaymentService) Charge(...
        </div>
        <div style={{ padding: '4px 8px', color: 'var(--ink-2)' }}>
          &nbsp;&nbsp;return s.adapter.Process(...)
        </div>
        <div
          style={{ padding: '4px 8px', background: 'var(--removed-bg)', color: 'var(--removed)' }}
        >
          − func (s *PaymentService) Pay(...
        </div>
        <div style={{ padding: '4px 8px', color: 'var(--ink-2)' }}>{'}'}</div>
      </div>
    </div>
  )
}

function FeatureArt3() {
  return (
    <svg viewBox="0 0 200 120" style={{ width: '100%', height: '100%' }} aria-hidden="true">
      <rect x="20" y="20" width="100" height="80" rx="8" fill="white" stroke="var(--line-strong)" />
      <text x="30" y="38" fontSize="9" fill="var(--ink-3)" fontFamily="JetBrains Mono">
        func Charge()
      </text>
      <rect x="30" y="46" width="80" height="2" fill="var(--ink-3)" opacity="0.2" />
      <rect x="30" y="52" width="60" height="2" fill="var(--ink-3)" opacity="0.2" />
      <rect x="30" y="58" width="70" height="2" fill="var(--ink-3)" opacity="0.2" />
      <path d="M120 60 L150 40" stroke="var(--brand)" strokeWidth="1.5" />
      <g>
        <rect x="140" y="20" width="50" height="36" rx="8" fill="var(--brand)" />
        <circle cx="151" cy="32" r="3" fill="white" />
        <rect x="158" y="30" width="22" height="2" fill="white" opacity="0.7" />
        <rect x="158" y="34" width="18" height="2" fill="white" opacity="0.7" />
        <rect x="158" y="38" width="20" height="2" fill="white" opacity="0.7" />
        <text x="146" y="50" fontSize="7" fill="white" opacity="0.8" fontFamily="JetBrains Mono">
          @ shimo-dev
        </text>
      </g>
    </svg>
  )
}

type Feature = { title: string; desc: string; art: ReactNode }

const FEATURES: Feature[] = [
  {
    title: '影響範囲を hop で測る',
    desc: '変更ノードから caller/callee を双方向に辿り、何が壊れる可能性があるかを段階的に可視化。',
    art: <FeatureArt1 />,
  },
  {
    title: 'AST 単位の diff',
    desc: '行ベースではなく関数・型単位で差分を表示。Monaco エディタによる VSCode 同等の体験。',
    art: <FeatureArt2 />,
  },
  {
    title: 'コメントは AST に紐づく',
    desc: 'rebase してもコメントが迷子にならない。ノード単位で議論を継続できます。',
    art: <FeatureArt3 />,
  },
]

export default function FeatureGrid() {
  return (
    <section
      id="features"
      style={{ padding: 'clamp(72px, 10vw, 112px) 0', borderBottom: '1px solid var(--line)' }}
    >
      <div className="wrap">
        <SectionHead
          eyebrow="Features"
          title="グラフだけじゃない。"
          lead="解析の仕方そのものを設計し直したから、レビューの一歩一歩が短くなります。"
        />
        <div className="grid-3" style={{ marginTop: 48 }}>
          {FEATURES.map((f) => (
            <div
              key={f.title}
              style={{
                background: 'white',
                border: '1px solid var(--line)',
                borderRadius: 16,
                padding: 24,
                display: 'flex',
                flexDirection: 'column',
                gap: 16,
              }}
            >
              <div
                style={{
                  height: 140,
                  background: 'var(--brand-tint)',
                  border: '1px solid #ccfbf1',
                  borderRadius: 10,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  padding: 16,
                }}
              >
                {f.art}
              </div>
              <h3 style={{ fontSize: 18, fontWeight: 600, letterSpacing: '-0.015em', margin: 0 }}>
                {f.title}
              </h3>
              <p style={{ fontSize: 13.5, color: 'var(--ink-2)', lineHeight: 1.6, margin: 0 }}>
                {f.desc}
              </p>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
