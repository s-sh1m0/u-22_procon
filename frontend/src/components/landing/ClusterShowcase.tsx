import Wrap from './Wrap'
import SectionHead from './SectionHead'

type Cluster = {
  label: string
  desc: string
  icon: string
  n: number
  bar: string
  iconBox: string
}

const CLUSTERS: Cluster[] = [
  {
    label: '追加',
    desc: '新規モジュール',
    icon: '+',
    n: 11,
    bar: 'bg-green-600',
    iconBox: 'bg-green-100 border-green-300 text-green-600',
  },
  {
    label: '削除',
    desc: '削除されたモジュール',
    icon: '−',
    n: 4,
    bar: 'bg-red-600',
    iconBox: 'bg-red-100 border-red-300 text-red-600',
  },
  {
    label: '修正',
    desc: 'ロジック変更を含む',
    icon: '~',
    n: 13,
    bar: 'bg-amber-600',
    iconBox: 'bg-amber-100 border-amber-300 text-amber-600',
  },
  {
    label: 'リファクタ',
    desc: '振る舞い保存・構造変更',
    icon: '↻',
    n: 6,
    bar: 'bg-violet-600',
    iconBox: 'bg-violet-100 border-violet-300 text-violet-600',
  },
]

export default function ClusterShowcase() {
  return (
    <section
      id="showcase"
      className="border-b border-stone-200 py-[clamp(72px,10vw,112px)] pb-[88px]"
    >
      <Wrap>
        <SectionHead
          eyebrow="4-way auto cluster"
          title="変更の種類で、自動的に4つに分ける。"
          lead="ファイル名やパスではなく、AST と呼び出しグラフから「変更の意味」を抽出。レビュアーは色とまとまりで読み進められます。"
        />
        <div className="mt-12 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {CLUSTERS.map((k) => (
            <div
              key={k.label}
              className="relative overflow-hidden rounded-[14px] border border-stone-200 bg-white p-[22px]"
            >
              <div className={`absolute inset-y-0 left-0 w-[3px] ${k.bar}`} />
              <div className="flex items-center justify-between">
                <div
                  className={`flex size-[38px] items-center justify-center rounded-[10px] border text-lg font-semibold ${k.iconBox}`}
                >
                  {k.icon}
                </div>
                <div className="font-mono text-[11px] text-stone-500">n={k.n}</div>
              </div>
              <div className="mt-4 text-base font-semibold tracking-[-0.01em] text-stone-900">
                {k.label}
              </div>
              <div className="mt-1 text-[13px] leading-normal text-stone-500">{k.desc}</div>
            </div>
          ))}
        </div>
      </Wrap>
    </section>
  )
}
