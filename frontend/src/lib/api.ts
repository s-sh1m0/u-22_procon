export class ApiError extends Error {
  readonly status: number

  constructor(status: number, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
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
  const res = await fetch(path, {
    ...init,
    headers: { Accept: 'application/json', ...init?.headers },
  })
  if (!res.ok) {
    const msg = await extractErrorMessage(res)
    throw new ApiError(res.status, msg)
  }
  return res.json() as Promise<T>
}

export async function apiFetchVoid(path: string, init?: RequestInit): Promise<void> {
  const res = await fetch(path, {
    ...init,
    headers: { Accept: 'application/json', ...init?.headers },
  })
  if (!res.ok) {
    const msg = await extractErrorMessage(res)
    throw new ApiError(res.status, msg)
  }
}
