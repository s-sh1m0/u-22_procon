import type { DiffStatus } from '@/types/api'

/**
 * ノードが diff に関与しているか（追加 / 削除 / 変更のいずれか）を判定する。
 *
 * 「変更ノード」の単一の真実のソース（SSOT）。影響フィルタ・クラスタ集計・
 * レビュー優先度など複数箇所で同じ基準を使うため、ここに一本化する。
 *
 * 引数は `GraphNode` 全体ではなく必要なフィールドのみを構造的に受け取り、
 * 部分的なノード表現からも呼べるようにする。
 */
export function isChanged(node: { changed: boolean; diff_status: DiffStatus }): boolean {
  return node.changed || node.diff_status !== 'existing'
}
