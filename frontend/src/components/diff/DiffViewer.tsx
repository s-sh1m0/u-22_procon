import { lazy, Suspense, useCallback, useEffect, useState } from 'react'
import type { MonacoDiffEditor } from '@monaco-editor/react'
import type { DiffFile, DiffStatus } from '@/types/api'

const DiffEditor = lazy(() =>
  import('@monaco-editor/react').then((m) => ({ default: m.DiffEditor })),
)

const STATUS_LABELS: Record<string, { label: string; className: string }> = {
  added: { label: 'Added', className: 'bg-green-100 text-green-700' },
  modified: { label: 'Modified', className: 'bg-blue-100 text-blue-700' },
  removed: { label: 'Removed', className: 'bg-red-100 text-red-700' },
  renamed: { label: 'Renamed', className: 'bg-amber-100 text-amber-700' },
}

type Props = {
  file: DiffFile
  /** スクロール先の関数の定義行（1 始まり） */
  line: number
  /** removed の関数は変更前のファイル側に行がある */
  diffStatus: DiffStatus
}

export default function DiffViewer({ file, line, diffStatus }: Props) {
  const [editor, setEditor] = useState<MonacoDiffEditor | null>(null)
  const handleMount = useCallback((e: MonacoDiffEditor) => setEditor(e), [])

  // 関数の定義行を上端付近に出す。削除された関数は変更前（左）側にしか無いので左側をスクロールする
  // （左右のスクロールは連動する）。差分の計算後に左右の行揃えの余白が入って位置がずれるので、
  // 計算が終わったらもう一度スクロールする。
  const reveal = useCallback(() => {
    if (!editor || line < 1) return
    const side = diffStatus === 'removed' ? editor.getOriginalEditor() : editor.getModifiedEditor()
    side.setPosition({ lineNumber: line, column: 1 })
    side.revealLineNearTop(line)
  }, [editor, line, diffStatus])

  useEffect(() => {
    if (!editor) return
    const sub = editor.onDidUpdateDiff(reveal)
    return () => sub.dispose()
  }, [editor, reveal])

  // 別の関数を選んだとき。同じファイル内なら差分は計算し直されないので、ここでスクロールする。
  useEffect(() => {
    reveal()
  }, [reveal, file.filename, line, diffStatus])

  const statusInfo = STATUS_LABELS[file.status] ?? {
    label: file.status,
    className: 'bg-stone-100 text-stone-600',
  }

  const pathLabel =
    file.status === 'renamed' && file.previous_name
      ? `${file.previous_name} → ${file.filename}`
      : file.filename

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      <div className="flex items-center gap-2 border-b border-stone-200 bg-stone-50 px-3 py-2">
        <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${statusInfo.className}`}>
          {statusInfo.label}
        </span>
        <span className="truncate font-mono text-xs text-stone-500" title={pathLabel}>
          {pathLabel}
        </span>
      </div>
      <div className="min-h-0 flex-1">
        <Suspense
          fallback={
            <div className="flex h-full items-center justify-center text-xs text-stone-400">
              エディタを読み込み中…
            </div>
          }
        >
          <DiffEditor
            height="100%"
            language="go"
            original={file.before_content}
            modified={file.after_content}
            onMount={handleMount}
            options={{
              readOnly: true,
              renderSideBySide: true,
              minimap: { enabled: false },
              scrollBeyondLastLine: false,
              automaticLayout: true,
              fontSize: 12,
              lineNumbers: 'on',
            }}
          />
        </Suspense>
      </div>
    </div>
  )
}
