import type { OpenDiffRequest } from '@/lib/editorBridge'

// VS Code 拡張の webview と拡張ホストの間でやり取りするメッセージ。
// 拡張側（extension/src/protocol.ts）と同じ形を保つこと。

/** webview → 拡張ホスト */
export type WebviewToHostMessage =
  /** API リクエストの中継依頼。拡張ホストが Bearer を付けて backend に送る */
  | { type: 'api'; id: number; method: string; path: string; body?: string }
  /** 選択した関数の diff を VS Code の diff エディタで開く */
  | { type: 'openDiff'; request: OpenDiffRequest }

/** 拡張ホスト → webview */
export type HostToWebviewMessage =
  | { type: 'apiResponse'; id: number; status: number; body: string }
  /** backend に到達できなかった等、HTTP 応答が得られなかった場合 */
  | { type: 'apiError'; id: number; message: string }

/** 拡張が HTML に埋め込む初期設定（`<script type="application/json" id="diffgraph-config">`） */
export type WebviewConfig = {
  jobId: string
}
