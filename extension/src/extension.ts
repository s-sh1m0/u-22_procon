import * as vscode from 'vscode'
import { DIFF_SCHEME, DiffDocumentProvider } from './diffEditor'
import { GraphPanel } from './graphPanel'
import { normalizePrUrl } from './prUrl'
import { githubToken, isServerUp, serverUrl, startAnalysis } from './server'

export function activate(context: vscode.ExtensionContext): void {
  const docs = new DiffDocumentProvider()
  const panel = new GraphPanel(context.extensionUri, docs)

  context.subscriptions.push(
    vscode.workspace.registerTextDocumentContentProvider(DIFF_SCHEME, docs),
    vscode.commands.registerCommand('diffgraph.analyzePR', () => analyzePR(panel)),
    vscode.commands.registerCommand('diffgraph.openInBrowser', () => panel.openInBrowser()),
  )
}

export function deactivate(): void {}

/** コマンド「DiffGraph: PR を解析」 */
async function analyzePR(panel: GraphPanel): Promise<void> {
  if (!(await isServerUp())) {
    const openSettings = '設定を開く'
    const choice = await vscode.window.showErrorMessage(
      `DiffGraph サーバー（${serverUrl()}）に接続できません。リポジトリ直下で ` +
        '`docker compose -f compose.prod.yml up -d` を実行してから、もう一度お試しください。',
      openSettings,
    )
    if (choice === openSettings) {
      await vscode.commands.executeCommand('workbench.action.openSettings', 'diffgraph.serverUrl')
    }
    return
  }

  // クリップボードに PR の URL があれば初期値にする（ブラウザからコピーしてすぐ実行できるように）
  const clipboard = await vscode.env.clipboard.readText()
  const input = await vscode.window.showInputBox({
    title: 'DiffGraph: PR を解析',
    prompt: '解析する Go の GitHub PR の URL',
    placeHolder: 'https://github.com/owner/repo/pull/123',
    value: normalizePrUrl(clipboard)?.url ?? '',
    ignoreFocusOut: true,
    validateInput: (v) =>
      normalizePrUrl(v)
        ? undefined
        : 'https://github.com/owner/repo/pull/123 の形式で入力してください',
  })
  const pr = input === undefined ? undefined : normalizePrUrl(input)
  if (!pr) return

  let token: string | undefined
  try {
    token = await githubToken(true)
  } catch {
    // ログインがキャンセルされた
    return
  }
  if (!token) return

  try {
    const jobId = await vscode.window.withProgress(
      { location: vscode.ProgressLocation.Notification, title: 'DiffGraph: 解析を開始しています…' },
      () => startAnalysis(token, pr.url),
    )
    panel.show(jobId, `DiffGraph: ${pr.owner}/${pr.repo}#${pr.number}`)
  } catch (err) {
    void vscode.window.showErrorMessage(
      `DiffGraph: 解析を開始できませんでした: ${err instanceof Error ? err.message : String(err)}`,
    )
  }
}
