import type { ReactNode } from 'react'
import Wrap from './Wrap'
import SectionHead from './SectionHead'
import { GitHubGlyph } from './glyphs'

function StepMock1() {
  return (
    <div className="w-full">
      <div className="mb-1.5 font-mono text-[10px] text-stone-500">PR URL</div>
      <div className="flex items-center gap-2 rounded-lg border-[1.5px] border-teal-700 bg-white px-2.5 py-2">
        <GitHubGlyph size={12} className="text-stone-500" />
        <span className="flex-1 overflow-hidden font-mono text-[11px] text-ellipsis whitespace-nowrap text-stone-700">
          github.com/kobold/checkout-service/pull/1284
        </span>
        <span className="text-[10px] font-medium text-green-600">✓</span>
      </div>
      <div className="mt-2 flex justify-end">
        <div className="rounded-md bg-stone-900 px-3 py-1.5 text-[11px] text-white">
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
    <svg viewBox="0 0 240 160" className="w-full" aria-hidden="true">
      <g stroke="#d6d3d1" strokeWidth="1" fill="none">
        <line x1="60" y1="40" x2="120" y2="80" />
        <line x1="180" y1="40" x2="120" y2="80" />
        <line x1="60" y1="120" x2="120" y2="80" />
        <line x1="180" y1="120" x2="120" y2="80" />
      </g>
      <g>
        {files.map(([x, y, label]) => (
          <g key={label}>
            <rect x={x} y={y} width="48" height="24" rx="4" fill="white" stroke="#d6d3d1" />
            <text
              x={x + 24}
              y={y + 16}
              textAnchor="middle"
              fontSize="9"
              fill="#44403c"
              fontFamily="Inter Tight"
            >
              {label}
            </text>
          </g>
        ))}
        <circle cx="120" cy="80" r="14" fill="#0f766e" />
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
  const bars: [string, string, string, number][] = [
    ['追加', 'bg-green-100 border-green-300', 'bg-green-600', 70],
    ['修正', 'bg-amber-100 border-amber-300', 'bg-amber-600', 90],
    ['リファクタ', 'bg-violet-100 border-violet-300', 'bg-violet-600', 45],
    ['削除', 'bg-red-100 border-red-300', 'bg-red-600', 30],
  ]
  return (
    <div className="flex w-full flex-col gap-1.5">
      {bars.map(([lbl, track, fill, w]) => (
        <div key={lbl} className="flex items-center gap-2">
          <span className="w-16 text-[11px] text-stone-700">{lbl}</span>
          <div className={`h-2.5 flex-1 overflow-hidden rounded-full border ${track}`}>
            <div className={`h-full opacity-85 ${fill}`} style={{ width: `${w}%` }} />
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
      className="border-b border-stone-200 bg-linear-to-b from-stone-50 to-white py-[clamp(72px,10vw,112px)]"
    >
      <Wrap>
        <SectionHead
          eyebrow="How it works"
          title="3 ステップで、PR を構造化。"
          lead="バックエンドが Go AST を解析し、フロントが xyflow でレンダリング。"
        />
        <div className="mt-14 grid grid-cols-1 gap-5 lg:grid-cols-3">
          {STEPS.map((s) => (
            <div
              key={s.n}
              className="flex min-h-[380px] flex-col gap-4 rounded-2xl border border-stone-200 bg-white p-6"
            >
              <div className="flex items-baseline gap-3">
                <span className="shrink-0 font-mono text-[11px] tracking-[0.06em] text-teal-700">
                  {s.n}
                </span>
                <h3 className="m-0 text-[19px] font-semibold tracking-[-0.015em]">{s.title}</h3>
              </div>
              <p className="m-0 text-[13.5px] leading-relaxed text-stone-700">{s.desc}</p>
              <div className="mt-auto flex flex-1 items-center justify-center rounded-xl border border-stone-200 bg-stone-100 p-4">
                {s.mock}
              </div>
            </div>
          ))}
        </div>
      </Wrap>
    </section>
  )
}
