import Wrap from './Wrap'
import SectionHead from './SectionHead'

type Mode = {
  label: string
  desc: string
  icon: string
  iconBox: string
  badge: string
}

const MODES: Mode[] = [
  {
    label: 'Louvain',
    desc: '依存関係の密度からコミュニティを自動検出',
    icon: '{}',
    iconBox: 'bg-teal-100 border-teal-300 text-teal-700',
    badge: '既定',
  },
  {
    label: 'パッケージ',
    desc: 'Go パッケージ単位でグルーピング',
    icon: '/',
    iconBox: 'bg-violet-100 border-violet-300 text-violet-600',
    badge: '',
  },
  {
    label: 'ファイル',
    desc: 'ソースファイル単位でグルーピング',
    icon: '[ ]',
    iconBox: 'bg-amber-100 border-amber-300 text-amber-600',
    badge: '',
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
          eyebrow="Smart clustering"
          title="依存関係から、自動でクラスタを生成。"
          lead="Louvain アルゴリズムが呼び出しグラフのコミュニティ構造を検出し、関連する関数群を意味のあるまとまりに自動分割。パッケージ・ファイル単位への切替もワンクリック。"
        />
        <div className="mt-12 grid grid-cols-1 gap-4 sm:grid-cols-3">
          {MODES.map((m) => (
            <div
              key={m.label}
              className="relative overflow-hidden rounded-[14px] border border-stone-200 bg-white p-[22px]"
            >
              <div className="flex items-center justify-between">
                <div
                  className={`flex size-[38px] items-center justify-center rounded-[10px] border font-mono text-sm font-semibold ${m.iconBox}`}
                >
                  {m.icon}
                </div>
                {m.badge && (
                  <span className="rounded-full bg-teal-100 px-2 py-0.5 text-[10px] font-medium text-teal-700">
                    {m.badge}
                  </span>
                )}
              </div>
              <div className="mt-4 text-base font-semibold tracking-[-0.01em] text-stone-900">
                {m.label}
              </div>
              <div className="mt-1 text-[13px] leading-normal text-stone-500">{m.desc}</div>
            </div>
          ))}
        </div>
      </Wrap>
    </section>
  )
}
