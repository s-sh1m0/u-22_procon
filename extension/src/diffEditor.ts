import * as path from 'node:path'
import * as vscode from 'vscode'
import type { OpenDiffRequest } from './protocol'

/** 解析結果のファイル内容を表す読み取り専用ドキュメントの URI スキーム */
export const DIFF_SCHEME = 'diffgraph'

/**
 * backend の /api/diff が返した変更前後の内容を、仮想ドキュメントとして VS Code に見せる。
 * 審査員の手元に PR のリポジトリが無くても diff エディタで読めるようにするため、
 * ワークスペースのファイルではなく解析結果の内容を使う。
 */
export class DiffDocumentProvider implements vscode.TextDocumentContentProvider {
  private readonly contents = new Map<string, string>()

  provideTextDocumentContent(uri: vscode.Uri): string {
    return this.contents.get(uri.toString()) ?? ''
  }

  set(uri: vscode.Uri, content: string): void {
    this.contents.set(uri.toString(), content)
  }
}

function docUri(jobId: string, side: 'before' | 'after', filename: string): vscode.Uri {
  return vscode.Uri.from({ scheme: DIFF_SCHEME, path: `/${jobId}/${side}/${filename}` })
}

/**
 * 選択された関数の diff を、グラフの隣のプレビュータブで開き、関数の定義行を表示する。
 * preserveFocus でフォーカスはグラフに残し、続けて別の関数を選べるようにする
 * （プレビュータブなので、次の選択で同じタブの中身が差し替わる）。
 */
export async function openDiff(
  docs: DiffDocumentProvider,
  jobId: string,
  req: OpenDiffRequest,
): Promise<void> {
  const { file } = req
  const before = docUri(jobId, 'before', file.previous_name ?? file.filename)
  const after = docUri(jobId, 'after', file.filename)
  docs.set(before, file.before_content)
  docs.set(after, file.after_content)

  const line = Math.max(0, req.line - 1)
  const range = new vscode.Range(line, 0, line, 0)
  const title = `${path.posix.basename(file.filename)}: ${req.functionName} (DiffGraph)`

  // diff エディタの selection は変更後（右）側に効く。削除された関数は変更前（左）側にしか無いので、
  // 開いた後に左側のエディタを直接スクロールする。
  const removed = req.diffStatus === 'removed'
  await vscode.commands.executeCommand('vscode.diff', before, after, title, {
    viewColumn: vscode.ViewColumn.Beside,
    preview: true,
    preserveFocus: true,
    selection: removed ? undefined : range,
  } satisfies vscode.TextDocumentShowOptions)

  if (removed) {
    const left = vscode.window.visibleTextEditors.find(
      (e) => e.document.uri.toString() === before.toString(),
    )
    left?.revealRange(range, vscode.TextEditorRevealType.InCenter)
  }
}
