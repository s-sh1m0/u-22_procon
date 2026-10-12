import * as vscode from 'vscode'

// DiffGraph サーバー（compose.prod.yml の nginx → backend）との通信。
// GitHub トークンは拡張ホストだけが持ち、Authorization: Bearer で backend に渡す。

/** backend の解析で必要な GitHub スコープ（Web の OAuth と同じ） */
const GITHUB_SCOPES = ['repo', 'read:user']

/** 設定 diffgraph.serverUrl（末尾スラッシュなし） */
export function serverUrl(): string {
  const url = vscode.workspace
    .getConfiguration('diffgraph')
    .get<string>('serverUrl', 'http://localhost:20080')
  return url.replace(/\/+$/, '')
}

/**
 * VS Code 標準の GitHub 認証でトークンを取得する。
 * interactive=true なら未ログイン時にログインを求め、false なら取得済みのものだけを返す。
 */
export async function githubToken(interactive: boolean): Promise<string | undefined> {
  const session = await vscode.authentication.getSession(
    'github',
    GITHUB_SCOPES,
    interactive ? { createIfNone: true } : { silent: true },
  )
  return session?.accessToken
}

/** サーバーが起動しているかを確認する */
export async function isServerUp(): Promise<boolean> {
  try {
    const res = await fetch(`${serverUrl()}/health`, { signal: AbortSignal.timeout(3000) })
    return res.ok
  } catch {
    return false
  }
}

/** backend にリクエストを送る。path は '/api/...' 形式。 */
export function request(
  token: string | undefined,
  path: string,
  init: { method?: string; body?: string } = {},
): Promise<Response> {
  const headers: Record<string, string> = { Accept: 'application/json' }
  if (token) headers.Authorization = `Bearer ${token}`
  if (init.body !== undefined) headers['Content-Type'] = 'application/json'
  return fetch(`${serverUrl()}${path}`, { method: init.method ?? 'GET', headers, body: init.body })
}

/** エラー応答からユーザー向けメッセージを取り出す（echo の {"message": ...} 形式） */
export async function errorMessage(res: Response): Promise<string> {
  try {
    const json = (await res.json()) as { message?: string }
    if (json.message) return json.message
  } catch {
    // fall through
  }
  return `${res.status} ${res.statusText}`
}

/** 解析を開始して jobId を返す */
export async function startAnalysis(token: string, prUrl: string): Promise<string> {
  const res = await request(token, '/api/analyze', {
    method: 'POST',
    body: JSON.stringify({ pr_url: prUrl }),
  })
  if (!res.ok) throw new Error(await errorMessage(res))
  const job = (await res.json()) as { job_id: string }
  return job.job_id
}
