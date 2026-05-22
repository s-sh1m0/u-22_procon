import type { ReactNode } from 'react'
import Wrap from './Wrap'
import SectionHead from './SectionHead'

function FeatureArt1() {
  return (
    <svg viewBox="0 0 200 120" className="h-full w-full" aria-hidden="true">
      <circle
        cx="100"
        cy="60"
        r="50"
        fill="none"
        stroke="#5eead4"
        strokeWidth="1"
        strokeDasharray="2 3"
        opacity="0.6"
      />
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
        y="22"
        textAnchor="middle"
        fontSize="9"
        fill="#0f766e"
        fontFamily="JetBrains Mono"
      >
        1 hop
      </text>
      <text
        x="100"
        y="6"
        textAnchor="middle"
        fontSize="9"
        fill="#0f766e"
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
      <rect x="20" y="20" width="100" height="80" rx="8" fill="white" stroke="#d6d3d1" />
      <text x="30" y="38" fontSize="9" fill="#78716c" fontFamily="JetBrains Mono">
        func Charge()
      </text>
      <rect x="30" y="46" width="80" height="2" fill="#78716c" opacity="0.2" />
      <rect x="30" y="52" width="60" height="2" fill="#78716c" opacity="0.2" />
      <rect x="30" y="58" width="70" height="2" fill="#78716c" opacity="0.2" />
      <path d="M120 60 L150 40" stroke="#0f766e" strokeWidth="1.5" />
      <g>
        <rect x="140" y="20" width="50" height="36" rx="8" fill="#0f766e" />
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
