import SectionHead from './SectionHead'

type Cluster = {
  label: string
  desc: string
  c: string
  bg: string
  bd: string
  icon: string
  n: number
}

const CLUSTERS: Cluster[] = [
  {
    label: '追加',
    desc: '新規モジュール',
    c: 'var(--added)',
    bg: 'var(--added-bg)',
    bd: 'var(--added-bd)',
    icon: '+',
    n: 11,
  },
  {
    label: '削除',
    desc: '削除されたモジュール',
    c: 'var(--removed)',
    bg: 'var(--removed-bg)',
    bd: 'var(--removed-bd)',
    icon: '−',
    n: 4,
  },
  {
    label: '修正',
    desc: 'ロジック変更を含む',
    c: 'var(--modified)',
    bg: 'var(--modified-bg)',
    bd: 'var(--modified-bd)',
    icon: '~',
    n: 13,
  },
  {
    label: 'リファクタ',
    desc: '振る舞い保存・構造変更',
    c: 'var(--refactor)',
    bg: 'var(--refactor-bg)',
    bd: 'var(--refactor-bd)',
    icon: '↻',
    n: 6,
  },
]

export default function ClusterShowcase() {
  return (
    <section
      id="showcase"
      style={{ padding: 'clamp(72px, 10vw, 112px) 0 88px', borderBottom: '1px solid var(--line)' }}
    >
      <div className="wrap">
        <SectionHead
          eyebrow="4-way auto cluster"
          title="変更の種類で、自動的に4つに分ける。"
          lead="ファイル名やパスではなく、AST と呼び出しグラフから「変更の意味」を抽出。レビュアーは色とまとまりで読み進められます。"
        />
        <div className="grid-4" style={{ marginTop: 48 }}>
          {CLUSTERS.map((k) => (
            <div
              key={k.label}
              style={{
                background: 'white',
                border: '1px solid var(--line)',
                borderRadius: 14,
                padding: 22,
                position: 'relative',
                overflow: 'hidden',
              }}
            >
              <div
                style={{ position: 'absolute', inset: '0 auto 0 0', width: 3, background: k.c }}
              />
              <div
                style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}
              >
                <div
                  style={{
                    width: 38,
                    height: 38,
                    borderRadius: 10,
                    background: k.bg,
                    border: `1px solid ${k.bd}`,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    color: k.c,
                    fontWeight: 600,
                    fontSize: 18,
                  }}
                >
                  {k.icon}
                </div>
                <div className="mono" style={{ fontSize: 11, color: 'var(--ink-3)' }}>
                  n={k.n}
                </div>
              </div>
              <div
                style={{
                  fontSize: 16,
                  fontWeight: 600,
                  color: 'var(--ink)',
                  marginTop: 16,
                  letterSpacing: '-0.01em',
                }}
              >
                {k.label}
              </div>
              <div style={{ fontSize: 13, color: 'var(--ink-3)', marginTop: 4, lineHeight: 1.5 }}>
                {k.desc}
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
