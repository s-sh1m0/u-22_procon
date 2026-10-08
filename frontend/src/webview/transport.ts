import type { ApiTransport } from '@/lib/api'
import type { HostToWebviewMessage, WebviewToHostMessage } from './protocol'

/** acquireVsCodeApi() が返すオブジェクトのうち、webview で使う部分。 */
export type VsCodeApi = {
  postMessage(message: WebviewToHostMessage): void
}

// HTTP 仕様上ボディを持てないステータス。new Response() に本文を渡すと例外になる。
const NULL_BODY_STATUSES = new Set([101, 204, 205, 304])

/**
 * API リクエストを拡張ホストに postMessage で中継する ApiTransport を作る。
 * GitHub トークンは拡張ホストだけが持ち、webview には渡さない。
 * webview のオリジン（vscode-webview://）から backend を直接叩かないので CORS も不要。
 */
export function createHostTransport(vscode: VsCodeApi, target: EventTarget = window): ApiTransport {
  let nextId = 1
  const pending = new Map<
    number,
    { resolve: (res: Response) => void; reject: (err: Error) => void }
  >()

  target.addEventListener('message', (event) => {
    const msg = (event as MessageEvent<HostToWebviewMessage>).data
    if (msg?.type !== 'apiResponse' && msg?.type !== 'apiError') return
    const p = pending.get(msg.id)
    if (!p) return
    pending.delete(msg.id)

    if (msg.type === 'apiError') {
      p.reject(new Error(msg.message))
      return
    }
    p.resolve(
      new Response(NULL_BODY_STATUSES.has(msg.status) ? null : msg.body, {
        status: msg.status,
        headers: { 'Content-Type': 'application/json' },
      }),
    )
  })

  return (path, init) =>
    new Promise<Response>((resolve, reject) => {
      const id = nextId++
      pending.set(id, { resolve, reject })
      vscode.postMessage({
        type: 'api',
        id,
        method: init?.method ?? 'GET',
        path,
        body: typeof init?.body === 'string' ? init.body : undefined,
      })
    })
}
