import { randomBytes } from 'node:crypto'
import * as vscode from 'vscode'
import { openDiff, type DiffDocumentProvider } from './diffEditor'
import type { HostToWebviewMessage, WebviewConfig, WebviewToHostMessage } from './protocol'
import { githubToken, request, serverUrl } from './server'

const VIEW_TYPE = 'diffgraph.graph'

// webview から中継を許す API。解析結果の閲覧に必要なものだけに絞る。
function isAllowedApiPath(path: string): boolean {
  return path.startsWith('/api/') || path === '/auth/me'
}

/**
 * 解析結果のグラフを表示する webview パネル（常に 1 枚を使い回す）。
 * 中身は Web と同じ React の解析画面（frontend の webview ビルド）で、
 * API は拡張ホストが Bearer を付けて中継し、diff は VS Code の diff エディタで開く。
 */
export class GraphPanel {
  private panel: vscode.WebviewPanel | undefined
  private jobId: string | undefined

  constructor(
    private readonly extensionUri: vscode.Uri,
    private readonly docs: DiffDocumentProvider,
  ) {}

  /** jobId の解析画面を表示する。既存パネルがあれば中身を差し替える。 */
  show(jobId: string, title: string): void {
    if (!this.panel) {
      this.panel = vscode.window.createWebviewPanel(VIEW_TYPE, title, vscode.ViewColumn.Active, {
        enableScripts: true,
        // タブを切り替えてもグラフの展開状態・選択を失わないようにする
        retainContextWhenHidden: true,
        localResourceRoots: [this.webviewRoot()],
      })
      this.panel.onDidDispose(() => {
        this.panel = undefined
        this.jobId = undefined
        void vscode.commands.executeCommand('setContext', 'diffgraph.hasAnalysis', false)
      })
      this.panel.webview.onDidReceiveMessage((msg: WebviewToHostMessage) => this.onMessage(msg))
    } else {
      this.panel.reveal()
    }
    this.jobId = jobId
    this.panel.title = title
    this.panel.webview.html = this.renderHtml(this.panel.webview, { jobId })
    void vscode.commands.executeCommand('setContext', 'diffgraph.hasAnalysis', true)
  }

  /** 表示中の解析結果を Web 版で開く */
  async openInBrowser(): Promise<void> {
    if (!this.jobId) {
      void vscode.window.showInformationMessage(
        'DiffGraph: 先に「DiffGraph: PR を解析」を実行してください。',
      )
      return
    }
    const url = `${serverUrl()}/analysis/${encodeURIComponent(this.jobId)}`
    await vscode.env.openExternal(vscode.Uri.parse(url))
  }

  private webviewRoot(): vscode.Uri {
    return vscode.Uri.joinPath(this.extensionUri, 'media', 'webview')
  }

  private async onMessage(msg: WebviewToHostMessage): Promise<void> {
    switch (msg.type) {
      case 'api':
        await this.relayApi(msg)
        return
      case 'openDiff':
        if (!this.jobId) return
        try {
          await openDiff(this.docs, this.jobId, msg.request)
        } catch (err) {
          void vscode.window.showErrorMessage(`DiffGraph: diff を開けませんでした: ${String(err)}`)
        }
        return
    }
  }

  private async relayApi(msg: Extract<WebviewToHostMessage, { type: 'api' }>): Promise<void> {
    const reply = (m: HostToWebviewMessage) => void this.panel?.webview.postMessage(m)
    if (!isAllowedApiPath(msg.path)) {
      reply({ type: 'apiError', id: msg.id, message: `許可されていない API です: ${msg.path}` })
      return
    }
    try {
      const token = await githubToken(false)
      const res = await request(token, msg.path, { method: msg.method, body: msg.body })
      reply({ type: 'apiResponse', id: msg.id, status: res.status, body: await res.text() })
    } catch {
      reply({
        type: 'apiError',
        id: msg.id,
        message: `DiffGraph サーバー（${serverUrl()}）に接続できません。サーバーが起動しているか確認してください。`,
      })
    }
  }

  private renderHtml(webview: vscode.Webview, config: WebviewConfig): string {
    const root = this.webviewRoot()
    const scriptUri = webview.asWebviewUri(vscode.Uri.joinPath(root, 'webview.js'))
    const styleUri = webview.asWebviewUri(vscode.Uri.joinPath(root, 'webview.css'))
    const nonce = randomBytes(16).toString('base64')
    const csp = [
      "default-src 'none'",
      `script-src 'nonce-${nonce}' ${webview.cspSource}`,
      // React Flow / Radix がインラインスタイルを使う
      `style-src ${webview.cspSource} 'unsafe-inline'`,
      `img-src ${webview.cspSource} data:`,
      `font-src ${webview.cspSource} data:`,
      // ELK の worker スクリプトを fetch して Blob URL から Worker を作る
      `connect-src ${webview.cspSource}`,
      'worker-src blob:',
    ].join('; ')
    // </script> 等で埋め込みが壊れないよう < をエスケープする
    const configJson = JSON.stringify(config).replace(/</g, '\\u003c')

    return `<!DOCTYPE html>
<html lang="ja">
<head>
  <meta charset="UTF-8" />
  <meta http-equiv="Content-Security-Policy" content="${csp}" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <link rel="stylesheet" href="${styleUri}" />
  <title>DiffGraph</title>
</head>
<body>
  <div id="root"></div>
  <script type="application/json" id="diffgraph-config">${configJson}</script>
  <script type="module" nonce="${nonce}" src="${scriptUri}"></script>
</body>
</html>`
  }
}
