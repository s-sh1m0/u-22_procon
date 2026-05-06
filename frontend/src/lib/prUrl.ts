export type ParsedPrUrl = {
  owner: string
  repo: string
  number: number
}

export function parsePrUrl(raw: string): ParsedPrUrl | null {
  let u: URL
  try {
    u = new URL(raw.trim())
  } catch {
    return null
  }
  if (u.protocol !== 'https:' && u.protocol !== 'http:') return null
  if (u.hostname !== 'github.com') return null

  const parts = u.pathname.replace(/^\//, '').split('/')
  if (parts.length !== 4 || parts[2] !== 'pull') return null

  const [owner, repo, , numberStr] = parts
  if (!owner || !repo) return null

  const number = parseInt(numberStr, 10)
  if (!Number.isInteger(number) || number <= 0) return null

  return { owner, repo, number }
}
