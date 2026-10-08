import { createContext, useContext } from 'react'
import type { DiffFile, DiffStatus } from '@/types/api'

/** エディタ（VS Code）で開く diff と、スクロール先の関数の情報。 */
export type OpenDiffRequest = {
  file: DiffFile
  /** 関数名（タブタイトル用） */
  functionName: string
  /** 関数の定義行（1 始まり） */
  line: number
  /** removed の関数は変更前のファイル側に行がある */
  diffStatus: DiffStatus
}

/**
 * アプリの外側にあるエディタとの連携口。VS Code 拡張の webview でだけ提供され、
 * Web では null（従来どおりパネル内の Monaco で diff を表示する）。
 */
export type EditorBridge = {
  openDiff: (req: OpenDiffRequest) => void
}

export const EditorBridgeContext = createContext<EditorBridge | null>(null)

export function useEditorBridge(): EditorBridge | null {
  return useContext(EditorBridgeContext)
}
