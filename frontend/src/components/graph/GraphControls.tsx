import type { ClusterMode } from '@/types/api'

type Props = {
  clusterMode: ClusterMode
  onChangeClusterMode: (mode: ClusterMode) => void
  impactOnly: boolean
  onSetImpactOnly: (next: boolean) => void
  changedCount: number
  changedIndex: number | null
  onNextChanged: () => void
  onPrevChanged: () => void
  onExpandAll: () => void
  onCollapseAll: () => void
}

const CLUSTER_MODES: { value: ClusterMode; label: string }[] = [
  { value: 'louvain', label: 'Louvain' },
  { value: 'package', label: 'Package' },
  { value: 'file', label: 'File' },
]

const VIEW_MODES: { value: boolean; label: string }[] = [
  { value: true, label: '変更影響' },
  { value: false, label: '全体' },
]

function SegmentedControl<T extends string | boolean>({
  items,
  value,
  onChange,
}: {
  items: { value: T; label: string }[]
  value: T
  onChange: (v: T) => void
}) {
  return (
    <div className="flex rounded-lg bg-stone-100 p-0.5">
      {items.map((m) => (
        <button
          key={String(m.value)}
          type="button"
          aria-pressed={value === m.value}
          className={[
            'flex-1 rounded-md px-2.5 py-1 text-[11px] font-medium transition-all',
            value === m.value
              ? 'bg-white text-stone-900 shadow-sm'
              : 'text-stone-500 hover:text-stone-700',
          ].join(' ')}
          onClick={() => onChange(m.value)}
        >
          {m.label}
        </button>
      ))}
    </div>
  )
}

function IconButton({
  onClick,
  label,
  disabled,
  children,
}: {
  onClick: () => void
  label: string
  disabled?: boolean
  children: React.ReactNode
}) {
  return (
    <button
      type="button"
      className="flex size-7 items-center justify-center rounded-md text-stone-500 transition-colors hover:bg-stone-100 hover:text-stone-700 disabled:opacity-30 disabled:hover:bg-transparent disabled:hover:text-stone-500"
      onClick={onClick}
      aria-label={label}
      disabled={disabled}
    >
      {children}
    </button>
  )
}

export default function GraphControls({
  clusterMode,
  onChangeClusterMode,
  impactOnly,
  onSetImpactOnly,
  changedCount,
  changedIndex,
  onNextChanged,
  onPrevChanged,
  onExpandAll,
  onCollapseAll,
}: Props) {
  return (
    <div className="flex flex-col gap-2.5 rounded-xl border border-stone-200/60 bg-white/90 p-2.5 shadow-lg shadow-stone-950/[.06] backdrop-blur-md">
      <div>
        <div className="mb-1 px-0.5 text-[10px] font-medium tracking-wider text-stone-400">
          表示
        </div>
        <SegmentedControl items={VIEW_MODES} value={impactOnly} onChange={onSetImpactOnly} />
      </div>

      <div className="h-px bg-stone-100" />

      <div>
        <div className="mb-1 px-0.5 text-[10px] font-medium tracking-wider text-stone-400">
          クラスタ
        </div>
        <SegmentedControl
          items={CLUSTER_MODES}
          value={clusterMode}
          onChange={onChangeClusterMode}
        />
      </div>

      <div className="h-px bg-stone-100" />

      <div className="flex gap-1">
        <button
          type="button"
          className="flex h-7 flex-1 items-center justify-center gap-1 rounded-md text-[11px] font-medium text-stone-600 transition-colors hover:bg-stone-100 hover:text-stone-900"
          onClick={onExpandAll}
        >
          <svg
            width="12"
            height="12"
            viewBox="0 0 12 12"
            fill="none"
            stroke="currentColor"
            strokeWidth="1.5"
            strokeLinecap="round"
            aria-hidden="true"
          >
            <path d="M2 4.5L6 1.5L10 4.5" />
            <path d="M2 7.5L6 10.5L10 7.5" />
          </svg>
          展開
        </button>
        <div className="w-px bg-stone-100" />
        <button
          type="button"
          className="flex h-7 flex-1 items-center justify-center gap-1 rounded-md text-[11px] font-medium text-stone-600 transition-colors hover:bg-stone-100 hover:text-stone-900"
          onClick={onCollapseAll}
        >
          <svg
            width="12"
            height="12"
            viewBox="0 0 12 12"
            fill="none"
            stroke="currentColor"
            strokeWidth="1.5"
            strokeLinecap="round"
            aria-hidden="true"
          >
            <path d="M2 2L6 5L10 2" />
            <path d="M2 10L6 7L10 10" />
          </svg>
          折畳
        </button>
      </div>

      {changedCount > 0 && (
        <>
          <div className="h-px bg-stone-100" />
          <div className="flex items-center gap-1">
            <IconButton onClick={onPrevChanged} label="前の変更ノード">
              <svg
                width="10"
                height="10"
                viewBox="0 0 10 10"
                fill="none"
                stroke="currentColor"
                strokeWidth="1.5"
                strokeLinecap="round"
                strokeLinejoin="round"
                aria-hidden="true"
              >
                <path d="M6.5 1.5L3 5l3.5 3.5" />
              </svg>
            </IconButton>
            <div className="flex flex-1 items-center justify-center gap-1.5">
              <span className="text-[11px] tabular-nums text-stone-900">
                {changedIndex !== null ? changedIndex + 1 : '-'}
              </span>
              <span className="text-[10px] text-stone-400">/</span>
              <span className="text-[11px] tabular-nums text-stone-500">{changedCount}</span>
              <span className="text-[10px] text-stone-400">変更</span>
            </div>
            <IconButton onClick={onNextChanged} label="次の変更ノード">
              <svg
                width="10"
                height="10"
                viewBox="0 0 10 10"
                fill="none"
                stroke="currentColor"
                strokeWidth="1.5"
                strokeLinecap="round"
                strokeLinejoin="round"
                aria-hidden="true"
              >
                <path d="M3.5 1.5L7 5l-3.5 3.5" />
              </svg>
            </IconButton>
          </div>
        </>
      )}
    </div>
  )
}
