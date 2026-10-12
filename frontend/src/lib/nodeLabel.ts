/**
 * ノードの表示名を返す。メソッドはレシーバ型を前に付ける（例: `AbsAction.TargetContains`）。
 * 同名のメソッドを持つ型が並んだときに、ファイル名と行番号を見なくても区別できるようにするため。
 */
export function nodeDisplayName(node: { name: string; receiver?: string }): string {
  return node.receiver ? `${node.receiver}.${node.name}` : node.name
}
