export const API_BASE_URL: string = import.meta.env.VITE_API_BASE_URL ?? ''

export class ApiError extends Error {
  readonly status: number

  constructor(status: number, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

/**
 * API リクエストの送信手段。Web はブラウザの fetch（Cookie 認証）を使い、
 * VS Code 拡張の webview は拡張ホストへの中継（Bearer 認証）に差し替える。
 */
export type ApiTransport = (path: string, init?: RequestInit) => Promise<Response>

const fetchTransport: ApiTransport = (path, init) =>
  fetch(API_BASE_URL + path, {
    ...init,
    credentials: 'include',
    headers: { Accept: 'application/json', ...init?.headers },
  })

let transport: ApiTransport = fetchTransport

/** API の送信手段を差し替える。アプリの描画前に一度だけ呼ぶ想定。 */
export function setApiTransport(t: ApiTransport): void {
  transport = t
}

async function extractErrorMessage(res: Response): Promise<string> {
  try {
    const json = (await res.json()) as { message?: string }
    if (json.message) return json.message
  } catch {
    // fall through
  }
  return `${res.status} ${res.statusText}`
}

export async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await transport(path, init)
  if (!res.ok) {
    const msg = await extractErrorMessage(res)
    throw new ApiError(res.status, msg)
  }
  return res.json() as Promise<T>
}

export async function apiFetchVoid(path: string, init?: RequestInit): Promise<void> {
  const res = await transport(path, init)
  if (!res.ok) {
    const msg = await extractErrorMessage(res)
    throw new ApiError(res.status, msg)
  }
}
