// Static "analysis result" preview shown beside the hero copy.

type Node = { x: number; y: number; w: number; label: string; c: string; bg: string; bd: string }

const NODES: Node[] = [
  { x: 80, y: 36, w: 90, label: 'CheckoutForm', c: '#d97706', bg: '#fef3c7', bd: '#fcd34d' },
  { x: 200, y: 36, w: 100, label: 'PaymentPicker', c: '#16a34a', bg: '#dcfce7', bd: '#86efac' },
  { x: 320, y: 36, w: 84, label: 'ReceiptModal', c: '#16a34a', bg: '#dcfce7', bd: '#86efac' },
  { x: 160, y: 152, w: 96, label: 'PaymentService', c: '#7c3aed', bg: '#ede9fe', bd: '#c4b5fd' },
  { x: 280, y: 152, w: 90, label: 'StripeAdapter', c: '#16a34a', bg: '#dcfce7', bd: '#86efac' },
  { x: 80, y: 152, w: 88, label: 'useCheckoutFlow', c: '#7c3aed', bg: '#ede9fe', bd: '#c4b5fd' },
  { x: 340, y: 152, w: 88, label: 'FeeCalculator', c: '#d97706', bg: '#fef3c7', bd: '#fcd34d' },
  {
    x: 240,
    y: 272,
    w: 104,
    label: 'LegacyStripeClient',
    c: '#dc2626',
    bg: '#fee2e2',
    bd: '#fca5a5',
  },
  { x: 80, y: 272, w: 88, label: 'PaymentRepo', c: '#d97706', bg: '#fef3c7', bd: '#fcd34d' },
]

const LEGEND: [string, string][] = [
  ['#16a34a', '追加 3'],
  ['#dc2626', '削除 1'],
  ['#d97706', '修正 3'],
  ['#7c3aed', 'リファクタ 2'],
]

export default function HeroIllustration() {
  return (
    <div style={{ position: 'relative' }}>
      <div
        style={{
          position: 'relative',
          background: 'white',
          borderRadius: 16,
          border: '1px solid var(--line)',
          boxShadow: '0 24px 60px rgba(15,23,42,0.08), 0 2px 8px rgba(15,23,42,0.04)',
          overflow: 'hidden',
        }}
      >
        {/* window chrome */}
        <div
          style={{
            padding: '12px 16px',
            borderBottom: '1px solid var(--line)',
            display: 'flex',
            alignItems: 'center',
            gap: 10,
          }}
        >
          <div style={{ display: 'flex', gap: 6 }}>
            {['#ef4444', '#f59e0b', '#10b981'].map((c) => (
              <div
                key={c}
                style={{ width: 10, height: 10, borderRadius: 99, background: c, opacity: 0.55 }}
              />
            ))}
          </div>
          <div className="mono" style={{ fontSize: 11, color: 'var(--ink-3)' }}>
            kobold/checkout-service · #1284 · 34 files
          </div>
          <div style={{ marginLeft: 'auto', fontSize: 11, color: 'var(--ink-3)' }}>解析完了</div>
        </div>

        {/* graph */}
        <svg viewBox="0 0 520 360" style={{ width: '100%', display: 'block' }} aria-hidden="true">
          <g fontFamily='"JetBrains Mono", monospace' fontSize="9" fill="#a8a29e">
            <text x="16" y="44">
              UI
            </text>
            <text x="16" y="160">
              DOMAIN
            </text>
            <text x="16" y="276">
              DATA
            </text>
          </g>
          <g stroke="#e7e5e4" strokeWidth="1" strokeDasharray="2 4">
            <line x1="0" y1="100" x2="520" y2="100" />
            <line x1="0" y1="220" x2="520" y2="220" />
          </g>
          <g fill="none" strokeWidth="1.6" opacity="0.65">
            <path d="M120 60 C120 110, 180 130, 200 160" stroke="#d97706" />
            <path d="M240 60 C240 110, 220 130, 220 160" stroke="#16a34a" />
            <path d="M360 60 C360 110, 340 130, 320 160" stroke="#16a34a" />
            <path d="M200 200 C200 240, 260 250, 280 280" stroke="#dc2626" />
            <path d="M320 200 C320 240, 280 250, 280 280" stroke="#7c3aed" />
            <path d="M120 200 C120 230, 160 250, 200 220" stroke="#7c3aed" />
            <path d="M380 200 C380 240, 340 260, 320 280" stroke="#d97706" />
          </g>
          {NODES.map((n) => (
            <g key={n.label}>
              <rect x={n.x} y={n.y} width={n.w} height={30} rx="6" fill={n.bg} stroke={n.bd} />
              <rect x={n.x} y={n.y} width={3} height={30} rx="1" fill={n.c} />
              <text
                x={n.x + 10}
                y={n.y + 19}
                fontSize="9.5"
                fill={n.c}
                fontFamily='"Inter Tight", sans-serif'
                fontWeight="600"
              >
                {n.label}
              </text>
            </g>
          ))}
        </svg>

        {/* cluster legend */}
        <div
          style={{
            borderTop: '1px solid var(--line)',
            padding: '10px 14px',
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            background: 'var(--bg-alt)',
          }}
        >
          <div style={{ display: 'flex', gap: 14, fontSize: 11, flexWrap: 'wrap' }}>
            {LEGEND.map(([c, l]) => (
              <span
                key={l}
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  gap: 5,
                  color: 'var(--ink-2)',
                }}
              >
                <span style={{ width: 8, height: 8, borderRadius: 2, background: c }} />
                {l}
              </span>
            ))}
          </div>
          <span className="mono" style={{ fontSize: 10, color: 'var(--ink-3)' }}>
            2.4s
          </span>
        </div>
      </div>
    </div>
  )
}
