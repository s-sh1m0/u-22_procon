import type { DiffStatus, PRInfo } from '@/types/api'

// buildBlobUrl はノードの定義位置を指す GitHub blob URL を組み立てる。
//
// file はリポジトリルート相対パス（バックエンドが filepath.Rel(repoRoot, ...) で算出）。
// removed ノードの定義行は base 側にしか存在しないため base SHA を、
// それ以外（added / existing）は head SHA を ref に使う。
// SHA が空の場合はブランチ ref にフォールバックし、ref が一切無ければ null を返す。
export function buildBlobUrl(
  pr: PRInfo,
  file: string,
  line: number,
  diffStatus: DiffStatus,
): string | null {
  if (!file) return null

  const ref = diffStatus === 'removed' ? pr.base_sha || pr.base_ref : pr.head_sha || pr.head_ref
  if (!ref) return null

  const path = file.split('/').map(encodeURIComponent).join('/')
  const base = `https://github.com/${pr.owner}/${pr.repo}/blob/${ref}/${path}`
  return line > 0 ? `${base}#L${line}` : base
}
