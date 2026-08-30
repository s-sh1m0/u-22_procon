import type { ReactNode } from 'react'
import Wrap from './Wrap'
import SectionHead from './SectionHead'

function FeatureArt1() {
  return (
    <svg viewBox="0 0 200 120" className="h-full w-full" aria-hidden="true">
      <circle
        cx="100"
        cy="60"
        r="32"
        fill="none"
        stroke="#5eead4"
        strokeWidth="1"
        strokeDasharray="2 3"
        opacity="0.8"
      />
      <circle cx="100" cy="60" r="14" fill="#0f766e" />
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
        y="18"
        textAnchor="middle"
        fontSize="9"
        fill="#0f766e"
        fontFamily="JetBrains Mono"
      >
        1 hop
      </text>
    </svg>
  )
}

function FeatureArt2() {
  return (
    <div className="w-full font-mono text-[10px] text-stone-700">
      <div className="overflow-hidden rounded-md border border-stone-200 bg-white">
        <div className="border-b border-stone-200 px-2 py-1 text-stone-500">PaymentService.go</div>
        <div className="bg-green-100 px-2 py-1 text-green-600">
          + func (s *PaymentService) Charge(...
        </div>
        <div className="px-2 py-1 text-stone-700">&nbsp;&nbsp;return s.adapter.Process(...)</div>
        <div className="bg-red-100 px-2 py-1 text-red-600">− func (s *PaymentService) Pay(...</div>
        <div className="px-2 py-1 text-stone-700">{'}'}</div>
      </div>
    </div>
  )
}

function FeatureArt3() {
  return (
    <svg viewBox="0 0 200 120" className="h-full w-full" aria-hidden="true">
      <g stroke="#d6d3d1" strokeWidth="1" fill="none">
        <path d="M60 30 L100 60" />
        <path d="M140 30 L100 60" />
        <path d="M100 60 L60 90" />
        <path d="M100 60 L140 90" />
      </g>
      <rect x="40" y="18" width="40" height="24" rx="4" fill="white" stroke="#d6d3d1" />
      <text
        x="60"
        y="34"
        textAnchor="middle"
        fontSize="8"
        fill="#44403c"
        fontFamily="JetBrains Mono"
      >
        funcA
      </text>
      <rect x="120" y="18" width="40" height="24" rx="4" fill="white" stroke="#d6d3d1" />
      <text
        x="140"
        y="34"
        textAnchor="middle"
        fontSize="8"
        fill="#44403c"
        fontFamily="JetBrains Mono"
      >
        funcB
      </text>
      <circle cx="100" cy="60" r="14" fill="#ef4444" />
      <text
        x="100"
        y="64"
        textAnchor="middle"
        fontSize="8"
        fill="white"
        fontFamily="JetBrains Mono"
        fontWeight="600"
      >
        ↻
      </text>
      <rect x="40" y="78" width="40" height="24" rx="4" fill="white" stroke="#d6d3d1" />
      <text
        x="60"
        y="94"
        textAnchor="middle"
        fontSize="8"
        fill="#44403c"
        fontFamily="JetBrains Mono"
      >
        funcC
      </text>
      <rect x="120" y="78" width="40" height="24" rx="4" fill="#fee2e2" stroke="#fca5a5" />
      <text
        x="140"
        y="94"
        textAnchor="middle"
        fontSize="8"
        fill="#dc2626"
        fontFamily="JetBrains Mono"
      >
        funcD
      </text>
      <path d="M140 78 L100 74" stroke="#ef4444" strokeWidth="2" fill="none" />
    </svg>
  )
}

type Feature = { title: string; desc: string; art: ReactNode }

const FEATURES: Feature[] = [
  {
    title: '影響範囲を hop で測る',
    desc: '変更ノードをクリックすると、直接の caller / callee を自動ハイライト。影響範囲をひと目で把握。',
    art: <FeatureArt1 />,
  },
  {
    title: 'インライン diff ビューア',
    desc: '関数ノードから対応ファイルの差分を即座に表示。Monaco エディタによる VSCode 同等の体験。',
    art: <FeatureArt2 />,
  },
  {
    title: '新規循環参照を自動検出',
    desc: 'PR で新たに生じた循環依存を赤でハイライト。構造リスクをレビュー前に把握できます。',
    art: <FeatureArt3 />,
  },
]

export default function FeatureGrid() {
  return (
    <section id="features" className="border-b border-stone-200 py-[clamp(72px,10vw,112px)]">
      <Wrap>
        <SectionHead
          eyebrow="Features"
          title="グラフだけじゃない。"
          lead="解析の仕方そのものを設計し直したから、レビューの一歩一歩が短くなります。"
        />
        <div className="mt-12 grid grid-cols-1 gap-5 lg:grid-cols-3">
          {FEATURES.map((f) => (
            <div
              key={f.title}
              className="flex flex-col gap-4 rounded-2xl border border-stone-200 bg-white p-6"
            >
              <div className="flex h-[140px] items-center justify-center rounded-[10px] border border-teal-100 bg-teal-50 p-4">
                {f.art}
              </div>
              <h3 className="m-0 text-[18px] font-semibold tracking-[-0.015em]">{f.title}</h3>
              <p className="m-0 text-[13.5px] leading-relaxed text-stone-700">{f.desc}</p>
            </div>
          ))}
        </div>
      </Wrap>
    </section>
  )
}
