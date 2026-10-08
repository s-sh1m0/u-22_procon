export type PrRef = {
  owner: string
  repo: string
  number: number
  /** backend に渡す正規形（https://github.com/owner/repo/pull/123） */
  url: string
}

/**
 * GitHub の PR URL を解釈する。ブラウザからコピーした URL をそのまま使えるよう、
 * `/files` などの末尾やクエリ・フラグメントが付いていても PR 本体の URL に正規化する。
 * PR の URL でなければ undefined。
 */
export function normalizePrUrl(raw: string): PrRef | undefined {
  let u: URL
  try {
    u = new URL(raw.trim())
  } catch {
    return undefined
  }
  if (u.protocol !== 'https:' && u.protocol !== 'http:') return undefined
  if (u.hostname !== 'github.com') return undefined

  const m = /^\/([^/]+)\/([^/]+)\/pull\/(\d+)(?:\/.*)?$/.exec(u.pathname)
  if (!m) return undefined
  const [, owner, repo, numberStr] = m
  const number = Number(numberStr)
  if (!Number.isInteger(number) || number <= 0) return undefined

  return { owner, repo, number, url: `https://github.com/${owner}/${repo}/pull/${number}` }
}
