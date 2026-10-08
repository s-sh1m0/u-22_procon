// webview と拡張ホストの間でやり取りするメッセージ。
// webview 側（frontend/src/webview/protocol.ts, frontend/src/lib/editorBridge.ts,
// frontend/src/types/api.ts の DiffFile）と同じ形を保つこと。
// webview 側は React に依存するため import せず、必要な形だけをここに写している。

/** backend の /api/diff/:jobId が返す 1 ファイル分の差分 */
export type DiffFile = {
  filename: string
  previous_name?: string
  status: 'added' | 'modified' | 'removed' | 'renamed'
  before_content: string
  after_content: string
}

/** 選択された関数の diff をエディタで開く依頼 */
export type OpenDiffRequest = {
  file: DiffFile
  functionName: string
  /** 関数の定義行（1 始まり） */
  line: number
  /** removed の関数は変更前のファイル側に行がある */
  diffStatus: 'added' | 'removed' | 'existing'
}

/** webview → 拡張ホスト */
export type WebviewToHostMessage =
  | { type: 'api'; id: number; method: string; path: string; body?: string }
  | { type: 'openDiff'; request: OpenDiffRequest }

/** 拡張ホスト → webview */
export type HostToWebviewMessage =
  | { type: 'apiResponse'; id: number; status: number; body: string }
  | { type: 'apiError'; id: number; message: string }

/** 拡張が HTML に埋め込む初期設定 */
export type WebviewConfig = {
  jobId: string
}
