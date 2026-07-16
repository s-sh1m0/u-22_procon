import type { Cycle } from '@/types/api'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'

type Props = {
  cycles: Cycle[]
  onSelectCycleNode: (nodeId: string) => void
}

// 新規循環参照のリストをアラートで表示する。is_new=true のものだけ列挙。
// クリックで該当ノードに fitView する。
export default function CycleAlert({ cycles, onSelectCycleNode }: Props) {
  const newCycles = cycles.filter((c) => c.is_new)
  if (newCycles.length === 0) return null

  return (
    <Alert variant="destructive" className="mt-2">
      <svg
        xmlns="http://www.w3.org/2000/svg"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
        className="size-4"
        aria-hidden="true"
      >
        <circle cx="12" cy="12" r="10" />
        <line x1="12" y1="8" x2="12" y2="12" />
        <line x1="12" y1="16" x2="12.01" y2="16" />
      </svg>
      <AlertTitle>新規循環参照を {newCycles.length} 件検出しました</AlertTitle>
      <AlertDescription>
        <ul className="mt-1 space-y-1 text-xs">
          {newCycles.map((c) => (
            <li key={c.id}>
              <button
                type="button"
                className="text-left underline-offset-2 hover:underline"
                onClick={() => onSelectCycleNode(c.nodes[0])}
                title={c.nodes.join(' → ')}
              >
                cycle #{c.id} ({c.nodes.length} ノード): {c.nodes.slice(0, 3).join(' → ')}
                {c.nodes.length > 3 ? ' …' : ''}
              </button>
            </li>
          ))}
        </ul>
      </AlertDescription>
    </Alert>
  )
}
