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

// 変更ノード巡回のイメージ。現在地の変更関数だけ amber リングで強調する。
const NAV_STEPS = ['Parse', 'Charge', 'Notify']

function FeatureArt3() {
  return (
    <div className="flex w-full flex-col items-center gap-3 font-mono text-[10px]">
      <div className="flex items-center gap-2">
        {NAV_STEPS.map((name, i) => (
          <div
            key={name}
            className={[
              'relative rounded-md border border-stone-200 bg-white py-1.5 pl-2 pr-3.5',
              i === 1 ? 'text-stone-900 ring-2 ring-amber-400' : 'text-stone-500',
            ].join(' ')}
          >
            <span className="absolute right-1 top-1 size-1.5 rounded-full bg-amber-400" />
            {name}
          </div>
        ))}
      </div>
      <div className="flex items-center gap-2 rounded-lg border border-stone-200 bg-white px-2.5 py-1 shadow-sm">
        <span className="text-stone-500">‹</span>
        <span className="tabular-nums text-stone-900">2</span>
        <span className="text-stone-400">/ 3 変更</span>
        <span className="text-stone-500">›</span>
      </div>
    </div>
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
    title: '矢印で変更を順にたどる',
    desc: 'ナビゲーターの ‹ › で、変更された関数へ 1 つずつ移動。折りたたまれたクラスタも自動で開き、画面の中央に寄せます。',
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
