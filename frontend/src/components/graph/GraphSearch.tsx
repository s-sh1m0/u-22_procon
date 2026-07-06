import { useMemo, useRef, useState, useCallback, useEffect } from 'react'
import { Input } from '@/components/ui/input'

export type SearchCandidate = {
  id: string
  name: string
  package: string
}

type Props = {
  candidates: SearchCandidate[]
  onJump: (id: string) => void
}

// 候補ドロップダウンに一度に表示する最大件数。大規模グラフで数千件を
// DOM に出すとキー入力ごとに重くなるため上限を設ける（超過分は件数のみ通知）。
const MAX_RESULTS = 40

/**
 * グラフ上の関数を名前 / パッケージで絞り込み、選択でそのノードへジャンプする検索ボックス。
 * ジャンプ後の展開・選択・中央寄せは親（onJump）側で行う。
 */
export default function GraphSearch({ candidates, onJump }: Props) {
  const [query, setQuery] = useState('')
  const [open, setOpen] = useState(false)
  const [activeIndex, setActiveIndex] = useState(0)
  const containerRef = useRef<HTMLDivElement>(null)

  const q = query.trim().toLowerCase()

  const { results, total } = useMemo(() => {
    if (!q) return { results: [] as SearchCandidate[], total: 0 }
    const matched: SearchCandidate[] = []
    let count = 0
    for (const c of candidates) {
      if (c.name.toLowerCase().includes(q) || c.package.toLowerCase().includes(q)) {
        count++
        if (matched.length < MAX_RESULTS) matched.push(c)
      }
    }
    return { results: matched, total: count }
  }, [candidates, q])

  // 外側クリックでドロップダウンを閉じる。
  useEffect(() => {
    if (!open) return
    const onPointerDown = (e: PointerEvent) => {
      if (!containerRef.current?.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('pointerdown', onPointerDown)
    return () => document.removeEventListener('pointerdown', onPointerDown)
  }, [open])

  const jump = useCallback(
    (c: SearchCandidate) => {
      onJump(c.id)
      setQuery('')
      setOpen(false)
    },
    [onJump],
  )

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === 'Escape') {
        setQuery('')
        setOpen(false)
        return
      }
      if (results.length === 0) return
      if (e.key === 'ArrowDown') {
        e.preventDefault()
        setActiveIndex((i) => (i + 1) % results.length)
      } else if (e.key === 'ArrowUp') {
        e.preventDefault()
        setActiveIndex((i) => (i - 1 + results.length) % results.length)
      } else if (e.key === 'Enter') {
        e.preventDefault()
        const c = results[activeIndex]
        if (c) jump(c)
      }
    },
    [results, activeIndex, jump],
  )

  const showDropdown = open && q.length > 0

  return (
    <div ref={containerRef} className="w-64">
      <Input
        type="text"
        value={query}
        onChange={(e) => {
          setQuery(e.target.value)
          setActiveIndex(0)
          setOpen(true)
        }}
        onFocus={() => setOpen(true)}
        onKeyDown={handleKeyDown}
        placeholder="関数・パッケージを検索"
        aria-label="関数・パッケージを検索"
        className="h-8 bg-white text-xs shadow-md"
      />
      {showDropdown && (
        <div className="mt-1 max-h-72 overflow-y-auto rounded-lg border border-stone-200 bg-white shadow-md">
          {results.length === 0 ? (
            <p className="px-2.5 py-2 text-xs text-stone-400">一致する関数がありません</p>
          ) : (
            <ul>
              {results.map((c, i) => (
                <li key={c.id}>
                  <button
                    type="button"
                    // pointerdown による外側クリック判定より前に選択を確定させたいので
                    // クリックで確定する（onMouseDown だと input の blur と競合しにくい）。
                    onClick={() => jump(c)}
                    onMouseEnter={() => setActiveIndex(i)}
                    className={[
                      'flex w-full flex-col items-start gap-0.5 px-2.5 py-1.5 text-left transition-colors',
                      i === activeIndex ? 'bg-stone-100' : 'hover:bg-stone-50',
                    ].join(' ')}
                  >
                    <span className="font-mono text-xs text-stone-800">{c.name}</span>
                    <span className="font-mono text-[10px] text-stone-400">{c.package}</span>
                  </button>
                </li>
              ))}
              {total > results.length && (
                <li className="px-2.5 py-1.5 text-[10px] text-stone-400">
                  他 {total - results.length} 件（絞り込んでください）
                </li>
              )}
            </ul>
          )}
        </div>
      )}
    </div>
  )
}
